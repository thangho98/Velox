// Package translate provides subtitle translation using DeepL (primary) and
// Google Translate (fallback). On-demand translation for SRT/VTT subtitles.
package translate

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
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

type batchSizer interface {
	MaxBatchSize() int
}

// SRTCue represents a single subtitle cue from an SRT file.
type SRTCue struct {
	Index  string // "1", "2", etc.
	Timing string // "00:01:23,456 --> 00:01:25,789"
	Text   string // Can be multi-line
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

	// Build batches
	type batch struct {
		start  int
		end    int
		texts  []string
		result []string
		err    error
	}
	var batches []batch
	for i := 0; i < len(cues); i += batchSize {
		end := i + batchSize
		if end > len(cues) {
			end = len(cues)
		}
		texts := make([]string, end-i)
		for j := i; j < end; j++ {
			texts[j-i] = cues[j].Text
		}
		batches = append(batches, batch{start: i, end: end, texts: texts})
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

			translated, err := translateBatchWithRetry(ctx, translator, b.texts, targetLang)
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

func translateBatchWithRetry(ctx context.Context, translator Translator, texts []string, targetLang string) ([]string, error) {
	translated, err := translator.Translate(ctx, texts, targetLang)
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
		// Retry only the missing ones
		missingTranslated, retryErr := translateBatchWithRetry(ctx, translator, missingTexts, targetLang)
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
	for i := 0; i < len(texts); i += nextBatchSize {
		end := i + nextBatchSize
		if end > len(texts) {
			end = len(texts)
		}

		partial, partialErr := translateBatchWithRetry(ctx, translator, texts[i:end], targetLang)
		if partialErr != nil {
			return nil, partialErr
		}
		out = append(out, partial...)
	}

	return out, nil
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

	name := translator.Name()
	return strings.HasSuffix(name, "_compatible") &&
		(strings.Contains(message, "parse llm translations") || strings.Contains(message, "llm response"))
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
