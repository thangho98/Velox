// Package translate provides subtitle translation using DeepL (primary) and
// Google Translate (fallback). On-demand translation for SRT/VTT subtitles.
package translate

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

// Translator translates text between languages.
type Translator interface {
	// Translate translates a batch of text strings to the target language.
	// Returns translated strings in the same order.
	Translate(ctx context.Context, texts []string, targetLang string) ([]string, error)
	Name() string
}

// ContextualTranslator is an optional capability: translators that implement
// this receive prior/following cues as read-only context to improve dialog
// continuity at forced-cut boundaries. LLM-based translators implement this;
// DeepL/Google do not, because they translate each item independently.
type ContextualTranslator interface {
	Translator
	TranslateWithContext(ctx context.Context, texts, prior, following []string, targetLang string) ([]string, error)
}

type batchSizer interface {
	MaxBatchSize() int
}

// SRTCue represents a single subtitle cue from an SRT file.
// StartSec/EndSec hold the parsed timing in seconds; both are -1 when the
// timing line is malformed, which signals chunkCuesByGap to fall back to
// fixed-size batching.
type SRTCue struct {
	Index    string // "1", "2", etc.
	Timing   string // "00:01:23,456 --> 00:01:25,789"
	Text     string // Can be multi-line
	StartSec float64
	EndSec   float64
}

// parseSRTTiming converts a single SRT timing line into start/end seconds.
// Returns ok=false when the format deviates from "HH:MM:SS,mmm --> HH:MM:SS,mmm".
func parseSRTTiming(line string) (start, end float64, ok bool) {
	sep := " --> "
	i := strings.Index(line, sep)
	if i < 0 {
		return 0, 0, false
	}
	s, sOk := parseSRTTimestamp(strings.TrimSpace(line[:i]))
	e, eOk := parseSRTTimestamp(strings.TrimSpace(line[i+len(sep):]))
	if !sOk || !eOk {
		return 0, 0, false
	}
	return s, e, true
}

func parseSRTTimestamp(s string) (float64, bool) {
	// Format: HH:MM:SS,mmm
	if len(s) < 12 {
		return 0, false
	}
	h, err1 := strconv.Atoi(s[0:2])
	m, err2 := strconv.Atoi(s[3:5])
	sec, err3 := strconv.Atoi(s[6:8])
	ms, err4 := strconv.Atoi(s[9:12])
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, false
	}
	if s[2] != ':' || s[5] != ':' || s[8] != ',' {
		return 0, false
	}
	return float64(h)*3600 + float64(m)*60 + float64(sec) + float64(ms)/1000, true
}

// ParseSRT parses an SRT file into cues.
func ParseSRT(content string) []SRTCue {
	var cues []SRTCue
	scanner := bufio.NewScanner(strings.NewReader(content))

	var current SRTCue
	state := 0 // 0=index, 1=timing, 2=text

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		switch state {
		case 0: // Expecting index
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			current.Index = trimmed
			state = 1

		case 1: // Expecting timing
			if strings.Contains(line, "-->") {
				current.Timing = line
				if start, end, ok := parseSRTTiming(line); ok {
					current.StartSec = start
					current.EndSec = end
				} else {
					current.StartSec = -1
					current.EndSec = -1
				}
				state = 2
			}

		case 2: // Collecting text lines
			if strings.TrimSpace(line) == "" {
				if current.Text != "" {
					cues = append(cues, current)
				}
				current = SRTCue{}
				state = 0
			} else {
				if current.Text != "" {
					current.Text += "\n"
				}
				current.Text += line
			}
		}
	}

	// Don't forget the last cue
	if current.Text != "" {
		cues = append(cues, current)
	}

	return cues
}

// Tiered gap thresholds for gap-aware batching. See plan-v-ai-translate-dialog-aware
// for the data analysis that chose these values: natural cuts stay above 1.5s
// across both sitcom (dense dialog) and movie (sparse dialog) samples, and
// larger thresholds apply earlier when the batch is less full so that the
// final batches don't waste a 3s scene break just because they only hold 10 cues.
const (
	gapSceneBreak       = 3.0 // likely scene break / commercial break
	gapDialogPause      = 2.0 // clear dialog pause
	gapSoftPause        = 1.5 // soft pause, only acceptable when batch nearly full

	minSizeSceneBreak  = 20
	minSizeDialogPause = 30
	minSizeSoftPause   = 40

	// contextOverlap is the number of cues attached as PRIOR/FOLLOWING context
	// at each forced-cut boundary. 3 is enough to preserve pronouns and tone
	// across sitcom dialog while keeping token overhead below ~15% per batch.
	contextOverlap = 3
)

// overlapTexts returns cues[start:end].Text as a slice, clamped to valid
// bounds and returning nil when the range is empty. Used to build PRIOR /
// FOLLOWING context snippets around forced-cut boundaries.
func overlapTexts(cues []SRTCue, start, end int) []string {
	if start < 0 {
		start = 0
	}
	if end > len(cues) {
		end = len(cues)
	}
	if start >= end {
		return nil
	}
	out := make([]string, end-start)
	for i := start; i < end; i++ {
		out[i-start] = cues[i].Text
	}
	return out
}

// Batch describes one chunk of cues produced by chunkCuesByGap.
// LeftForced/RightForced are true when the corresponding boundary was created
// because the chunker hit maxBatch (not a natural gap). Phase 2 consumes these
// to decide whether to attach context-overlap cues on a given side.
type Batch struct {
	Start, End  int
	LeftForced  bool
	RightForced bool
	LeftGap     float64
	RightGap    float64
}

// ChunkResult is the return type of chunkCuesByGap.
// TimingValid is false when the chunker fell back to fixed-size chunking
// because at least one cue had malformed timing. When false, callers must
// ignore the per-boundary Forced flags (they're all false in this mode) and
// treat the entire batch set as "timing unknown".
type ChunkResult struct {
	Batches     []Batch
	TimingValid bool
}

// chunkCuesByGap splits cues into batches that respect dialog structure.
// When all cues have valid timing, it applies tiered gap thresholds so batch
// boundaries land on natural dialog pauses whenever possible. When any cue has
// invalid timing (StartSec<0), it falls back to fixed-size chunking and sets
// TimingValid=false so Phase 2 can disable context overlap entirely.
func chunkCuesByGap(cues []SRTCue, maxBatch int) ChunkResult {
	if len(cues) == 0 {
		return ChunkResult{TimingValid: true}
	}
	if maxBatch <= 0 {
		maxBatch = 50
	}

	for _, c := range cues {
		if c.StartSec < 0 || c.EndSec < 0 {
			return ChunkResult{
				Batches:     fixedSizeChunks(cues, maxBatch),
				TimingValid: false,
			}
		}
	}

	var batches []Batch
	start := 0

	closeBatch := func(end int, rightGap float64, rightForced bool) {
		var leftGap float64
		var leftForced bool
		if len(batches) > 0 {
			prev := batches[len(batches)-1]
			leftGap = prev.RightGap
			leftForced = prev.RightForced
		}
		batches = append(batches, Batch{
			Start:       start,
			End:         end,
			LeftForced:  leftForced,
			RightForced: rightForced,
			LeftGap:     leftGap,
			RightGap:    rightGap,
		})
		start = end
	}

	for i := 0; i < len(cues)-1; i++ {
		curSize := i - start + 1
		gap := cues[i+1].StartSec - cues[i].EndSec

		switch {
		case curSize >= maxBatch:
			closeBatch(i+1, gap, true)
		case gap >= gapSceneBreak && curSize >= minSizeSceneBreak:
			closeBatch(i+1, gap, false)
		case gap >= gapDialogPause && curSize >= minSizeDialogPause:
			closeBatch(i+1, gap, false)
		case gap >= gapSoftPause && curSize >= minSizeSoftPause:
			closeBatch(i+1, gap, false)
		}
	}

	// Final batch — right boundary is end-of-stream, not a forced cut.
	closeBatch(len(cues), 0, false)

	return ChunkResult{Batches: batches, TimingValid: true}
}

// fixedSizeChunks is the malformed-timing fallback. All boundaries have
// Forced=false because the chunker no longer trusts the timing to classify
// them; the caller uses ChunkResult.TimingValid=false as the global signal.
func fixedSizeChunks(cues []SRTCue, maxBatch int) []Batch {
	if len(cues) == 0 {
		return nil
	}
	if maxBatch <= 0 {
		maxBatch = 50
	}
	var batches []Batch
	for i := 0; i < len(cues); i += maxBatch {
		end := i + maxBatch
		if end > len(cues) {
			end = len(cues)
		}
		batches = append(batches, Batch{Start: i, End: end})
	}
	return batches
}

// BuildSRT reconstructs an SRT file from cues.
func BuildSRT(cues []SRTCue) string {
	var sb strings.Builder
	for i, cue := range cues {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(cue.Index)
		sb.WriteString("\n")
		sb.WriteString(cue.Timing)
		sb.WriteString("\n")
		sb.WriteString(cue.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// TranslateSRT translates an SRT file content to the target language.
// Batches text to minimize API calls and processes them in parallel.
func TranslateSRT(ctx context.Context, translator Translator, srtContent string, targetLang string) (string, error) {
	cues := ParseSRT(srtContent)
	if len(cues) == 0 {
		return "", fmt.Errorf("no subtitle cues found")
	}

	slog.Debug("translate_srt_start",
		"translator", translator.Name(),
		"total_cues", len(cues),
		"target_lang", targetLang,
	)

	// Batch translate in provider-sized chunks so LLM-based translators keep enough context
	// without forcing extremely large prompts.
	batchSize := 50
	if sized, ok := translator.(batchSizer); ok && sized.MaxBatchSize() > 0 {
		batchSize = sized.MaxBatchSize()
	}

	// Gap-aware chunking: batch boundaries land on natural dialog pauses when
	// timing is valid, fixed-size fallback otherwise. Per-boundary Forced flags
	// and TimingValid drive context overlap: overlap cues attach only to forced
	// boundaries, and only when timing is trustworthy.
	chunked := chunkCuesByGap(cues, batchSize)

	type batch struct {
		start      int
		end        int
		texts      []string
		batchCues  []SRTCue
		prior      []string // cues to send as PRIOR context (empty when no overlap)
		following  []string // cues to send as FOLLOWING context (empty when no overlap)
		result     []string
		err        error
	}
	batches := make([]batch, 0, len(chunked.Batches))
	for _, b := range chunked.Batches {
		texts := make([]string, b.End-b.Start)
		for j := b.Start; j < b.End; j++ {
			texts[j-b.Start] = cues[j].Text
		}
		var prior, following []string
		if chunked.TimingValid {
			if b.LeftForced {
				prior = overlapTexts(cues, b.Start-contextOverlap, b.Start)
			}
			if b.RightForced {
				following = overlapTexts(cues, b.End, b.End+contextOverlap)
			}
		}
		batches = append(batches, batch{
			start:     b.Start,
			end:       b.End,
			texts:     texts,
			batchCues: cues[b.Start:b.End],
			prior:     prior,
			following: following,
		})
	}

	// Limit concurrent API calls to avoid rate limiting
	maxConcurrency := 5
	if translator.Name() == "google" {
		maxConcurrency = 10 // Google is more lenient
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for i := range batches {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			b := &batches[idx]
			slog.Debug("translate_srt_batch",
				"batch_start", b.start,
				"batch_end", b.end,
				"batch_size", len(b.texts),
				"translator", translator.Name(),
			)

			translated, err := translateBatchWithRetry(ctx, translator, b.texts, b.batchCues, b.prior, b.following, targetLang)
			if err != nil {
				slog.Error("translate_srt_batch_failed",
					"batch_start", b.start,
					"batch_end", b.end,
					"translator", translator.Name(),
					"error", err,
				)
				mu.Lock()
				b.err = err
				mu.Unlock()
				return
			}

			mu.Lock()
			b.result = translated
			mu.Unlock()

			slog.Debug("translate_srt_batch_done",
				"batch_start", b.start,
				"batch_end", b.end,
				"translator", translator.Name(),
			)
		}(i)
	}

	wg.Wait()

	// Check for errors and apply results
	for _, b := range batches {
		if b.err != nil {
			return "", fmt.Errorf("translating batch %d-%d: %w", b.start, b.end, b.err)
		}
		for j, t := range b.result {
			cues[b.start+j].Text = t
		}
	}

	return BuildSRT(cues), nil
}

// callTranslator dispatches to TranslateWithContext when the translator
// supports it and at least one side has overlap; otherwise falls back to
// the plain Translate path. Keeps all callers (including retry sub-paths)
// routed through one decision point.
func callTranslator(ctx context.Context, translator Translator, texts, prior, following []string, targetLang string) ([]string, error) {
	if len(prior) > 0 || len(following) > 0 {
		if ct, ok := translator.(ContextualTranslator); ok {
			return ct.TranslateWithContext(ctx, texts, prior, following, targetLang)
		}
	}
	return translator.Translate(ctx, texts, targetLang)
}

// translateBatchWithRetry handles a single batch call with shrink-on-failure
// retry. cues is the subset of SRTCue values corresponding to texts, used for
// gap-aware subdivision during downsize retry. Pass cues=nil for the partial-
// retry path where indices are non-contiguous and gap information is meaningless.
// prior/following are the optional PRIOR/FOLLOWING context cues (already the
// text values); they flow only to the sub-batch that owns that boundary during
// retry subdivision. Interior retry boundaries get nil/nil because the cues
// on either side are already part of another sub-call in the same retry.
func translateBatchWithRetry(ctx context.Context, translator Translator, texts []string, cues []SRTCue, prior, following []string, targetLang string) ([]string, error) {
	translated, err := callTranslator(ctx, translator, texts, prior, following, targetLang)
	if err == nil {
		return translated, nil
	}
	if !shouldRetrySmallerBatch(translator, err, len(texts)) {
		slog.Warn("translate_batch_no_retry",
			"translator", translator.Name(),
			"batch_size", len(texts),
			"error", err,
		)
		return nil, err
	}

	// Check if this is a partial translation error (some cues translated, some missing)
	partialInfo := extractPartialTranslations(err, translated)
	if partialInfo != nil && len(partialInfo.missing) > 0 {
		slog.Info("translate_batch_partial_use",
			"translator", translator.Name(),
			"got", len(partialInfo.translations),
			"missing_count", len(partialInfo.missing),
			"missing_indexes", partialInfo.missing,
		)
		// Build texts for only the missing indexes
		missingTexts := make([]string, len(partialInfo.missing))
		for i, idx := range partialInfo.missing {
			missingTexts[i] = texts[idx]
		}
		// Retry only the missing ones. Pass cues=nil because the missing set is
		// non-contiguous in the original timeline, so gap-aware subdivision would
		// measure "fake gaps" between unrelated cues and possibly oversplit.
		// prior/following=nil because the neighboring cues in `missingTexts` are
		// unrelated to the original boundaries, so context overlap is meaningless.
		missingTranslated, retryErr := translateBatchWithRetry(ctx, translator, missingTexts, nil, nil, nil, targetLang)
		if retryErr == nil && len(missingTranslated) == len(partialInfo.missing) {
			// Success - merge results in correct order
			result := make([]string, len(texts))
			// Fill in what we had
			origIdx := 0
			for i := 0; i < len(texts); i++ {
				if contains(partialInfo.missing, i) {
					result[i] = missingTranslated[indexOf(partialInfo.missing, i)]
				} else {
					result[i] = partialInfo.translations[origIdx]
					origIdx++
				}
			}
			return result, nil
		}
		slog.Warn("translate_batch_partial_retry_failed",
			"translator", translator.Name(),
			"using_partial", len(partialInfo.translations),
			"error", retryErr,
		)
		// Fall through to smaller batch retry
	}

	nextBatchSize := nextRetryBatchSize(len(texts))
	if nextBatchSize >= len(texts) {
		slog.Warn("translate_batch_retry_skipped",
			"translator", translator.Name(),
			"batch_size", len(texts),
			"next_batch_size", nextBatchSize,
			"error", err,
		)
		return nil, err
	}

	slog.Info("translate_batch_retry",
		"translator", translator.Name(),
		"original_size", len(texts),
		"retry_batch_size", nextBatchSize,
		"error", err,
	)

	out := make([]string, 0, len(texts))
	ranges := splitForRetry(texts, cues, nextBatchSize)
	for i, r := range ranges {
		var subCues []SRTCue
		if cues != nil {
			subCues = cues[r[0]:r[1]]
		}
		// Only the first sub-batch inherits PRIOR context from the original
		// batch boundary; only the last inherits FOLLOWING. Interior sub-
		// batch boundaries don't need overlap because the neighboring cues
		// are already part of another sub-call in this same retry.
		var subPrior, subFollowing []string
		if i == 0 {
			subPrior = prior
		}
		if i == len(ranges)-1 {
			subFollowing = following
		}
		partial, partialErr := translateBatchWithRetry(ctx, translator, texts[r[0]:r[1]], subCues, subPrior, subFollowing, targetLang)
		if partialErr != nil {
			return nil, partialErr
		}
		out = append(out, partial...)
	}

	return out, nil
}

// splitForRetry chooses how to subdivide a failing batch during shrink-retry.
// When cues match texts 1:1, it uses gap-aware chunking so sub-batches still
// respect dialog structure. When cues is nil (partial-retry path with
// non-contiguous missing indexes) or mismatches length, it falls back to
// fixed-size slicing — gap-awareness is meaningless on a discontiguous slice.
func splitForRetry(texts []string, cues []SRTCue, size int) [][2]int {
	if size <= 0 {
		size = 1
	}
	if cues == nil || len(cues) != len(texts) {
		ranges := make([][2]int, 0, (len(texts)+size-1)/size)
		for i := 0; i < len(texts); i += size {
			end := i + size
			if end > len(texts) {
				end = len(texts)
			}
			ranges = append(ranges, [2]int{i, end})
		}
		return ranges
	}
	result := chunkCuesByGap(cues, size)
	ranges := make([][2]int, 0, len(result.Batches))
	for _, b := range result.Batches {
		ranges = append(ranges, [2]int{b.Start, b.End})
	}
	return ranges
}

// partialResult holds partial translation results with info about what's missing
type partialResult struct {
	translations []string // translations we did receive, in order
	missing      []int    // indexes of missing translations
}

// extractPartialTranslations checks if we got usable partial results despite the error.
// Returns nil if no partial results available.
func extractPartialTranslations(err error, rawResult []string) *partialResult {
	if err == nil || len(rawResult) == 0 {
		return nil
	}

	// Check if this is our partialTranslationError with missing indexes
	partialErr, ok := err.(*partialTranslationError)
	if ok && len(partialErr.translations) > 0 {
		return &partialResult{
			translations: partialErr.translations,
			missing:      partialErr.missing,
		}
	}

	// Fallback: check if error is a count mismatch
	errMsg := err.Error()
	if !strings.Contains(errMsg, "expected") || !strings.Contains(errMsg, "got") {
		return nil
	}
	// We got partial results but don't know which are missing - assume first N are present
	return &partialResult{
		translations: rawResult,
		missing:      nil, // will retry from the received count
	}
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func indexOf(slice []int, item int) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func shouldRetrySmallerBatch(translator Translator, err error, size int) bool {
	if err == nil || size <= 1 {
		return false
	}

	message := err.Error()
	if strings.Contains(message, "parse llm translations: expected") {
		return true
	}
	if strings.Contains(message, "llm response does not contain a JSON object") {
		return true
	}
	if strings.Contains(message, "no text content returned") {
		return true
	}

	name := translator.Name()
	return strings.HasSuffix(name, "_compatible") &&
		(strings.Contains(message, "parse llm translations") ||
			strings.Contains(message, "llm response") ||
			strings.Contains(message, "no text content"))
}

func nextRetryBatchSize(size int) int {
	switch {
	case size > 35:
		return 35
	case size > 25:
		return 25
	case size > 15:
		return 15
	case size > 10:
		return 10
	case size > 5:
		return 5
	case size > 1:
		return 1
	default:
		return 1
	}
}
