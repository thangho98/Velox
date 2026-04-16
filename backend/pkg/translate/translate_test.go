package translate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

type recordingTranslator struct {
	calls []int
}

func (t *recordingTranslator) Name() string { return "anthropic_compatible" }

func (t *recordingTranslator) Translate(_ context.Context, texts []string, _ string) ([]string, error) {
	t.calls = append(t.calls, len(texts))
	if len(texts) > 1 {
		return nil, fmt.Errorf("parse llm translations: expected %d items, got %d", len(texts), len(texts)+1)
	}

	out := make([]string, len(texts))
	for i, text := range texts {
		out[i] = "vi:" + text
	}
	return out, nil
}

func (t *recordingTranslator) MaxBatchSize() int { return 20 }

type failingTranslator struct{}

func (t *failingTranslator) Name() string { return "google" }

func (t *failingTranslator) Translate(_ context.Context, _ []string, _ string) ([]string, error) {
	return nil, fmt.Errorf("upstream unavailable")
}

func TestTranslateSRTRetriesWithSmallerBatchesOnLLMCountMismatch(t *testing.T) {
	translator := &recordingTranslator{}
	srt := buildTestSRT(20)

	got, err := TranslateSRT(context.Background(), translator, srt, "vi")
	if err != nil {
		t.Fatalf("TranslateSRT returned error: %v", err)
	}

	wantPrefix := []int{20, 15, 10, 5, 1}
	if len(translator.calls) < len(wantPrefix) {
		t.Fatalf("calls = %v, want prefix %v", translator.calls, wantPrefix)
	}
	for i, want := range wantPrefix {
		if translator.calls[i] != want {
			t.Fatalf("calls = %v, want prefix %v", translator.calls, wantPrefix)
		}
	}

	if !strings.Contains(got, "vi:Line 1") || !strings.Contains(got, "vi:Line 20") {
		t.Fatalf("translated SRT missing expected content: %q", got)
	}
}

func TestTranslateSRTDoesNotRetryNonLLMErrors(t *testing.T) {
	translator := &failingTranslator{}
	srt := buildTestSRT(3)

	_, err := TranslateSRT(context.Background(), translator, srt, "vi")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "upstream unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func buildTestSRT(count int) string {
	var sb strings.Builder
	for i := 1; i <= count; i++ {
		sb.WriteString(fmt.Sprintf("%d\n", i))
		sb.WriteString("00:00:01,000 --> 00:00:02,000\n")
		sb.WriteString(fmt.Sprintf("Line %d\n\n", i))
	}
	return sb.String()
}

// makeCues builds a slice of SRTCue with times derived from the given gap
// pattern. startSec is where cue[0] begins, duration is each cue's length,
// and gaps[i] is the silence between cue[i].End and cue[i+1].Start.
func makeCues(startSec, duration float64, gaps []float64) []SRTCue {
	cues := make([]SRTCue, len(gaps)+1)
	t := startSec
	for i := range cues {
		cues[i] = SRTCue{
			Index:    fmt.Sprintf("%d", i+1),
			StartSec: t,
			EndSec:   t + duration,
		}
		if i < len(gaps) {
			t = t + duration + gaps[i]
		}
	}
	return cues
}

func countRightForced(batches []Batch) int {
	// Count forced boundaries, ignoring the final batch's always-false RightForced.
	if len(batches) <= 1 {
		return 0
	}
	n := 0
	for _, b := range batches[:len(batches)-1] {
		if b.RightForced {
			n++
		}
	}
	return n
}

func TestChunker_Empty(t *testing.T) {
	result := chunkCuesByGap(nil, 50)
	if !result.TimingValid {
		t.Fatal("empty input should be TimingValid=true (no invalid cues)")
	}
	if len(result.Batches) != 0 {
		t.Fatalf("expected 0 batches, got %d", len(result.Batches))
	}
}

func TestChunker_SingleCue(t *testing.T) {
	cues := makeCues(0, 1.0, nil)
	result := chunkCuesByGap(cues, 50)
	if !result.TimingValid {
		t.Fatal("valid timing should yield TimingValid=true")
	}
	if len(result.Batches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(result.Batches))
	}
	b := result.Batches[0]
	if b.Start != 0 || b.End != 1 {
		t.Fatalf("batch range = [%d, %d), want [0, 1)", b.Start, b.End)
	}
	if b.LeftForced || b.RightForced {
		t.Fatalf("single batch should have both forced flags false: %+v", b)
	}
}

func TestChunker_MovieLike_NoForcedCuts(t *testing.T) {
	// Sparse dialog pattern: 25 cues of tight 0.2s gaps, then one 3.5s scene
	// break, repeated 4 times. maxBatch=50, so we expect the chunker to prefer
	// the 3.5s breaks over hitting the cap.
	var gaps []float64
	for range 4 {
		for range 24 {
			gaps = append(gaps, 0.2)
		}
		gaps = append(gaps, 3.5)
	}
	// Drop the trailing 3.5s so last cue doesn't dangle
	gaps = gaps[:len(gaps)-1]
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	if !result.TimingValid {
		t.Fatal("valid timing expected")
	}
	forced := countRightForced(result.Batches)
	if forced != 0 {
		t.Fatalf("movie-like content should have 0 forced boundaries, got %d (batches=%d)", forced, len(result.Batches))
	}
	if len(result.Batches) < 2 || len(result.Batches) > 6 {
		t.Fatalf("expected 2-6 batches for movie-like input, got %d", len(result.Batches))
	}
}

func TestChunker_SitcomLike_HasForcedCut(t *testing.T) {
	// 60 cues of 0.2s gaps — no natural breaks, so chunker must force cut
	// at maxBatch=50. Expect exactly 2 batches: first forced at 50, second is 10.
	gaps := make([]float64, 59)
	for i := range gaps {
		gaps[i] = 0.2
	}
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	if len(result.Batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(result.Batches))
	}
	if !result.Batches[0].RightForced {
		t.Fatalf("boundary between batch 0 and 1 should be forced (dense dialog hit max)")
	}
	if !result.Batches[1].LeftForced {
		t.Fatalf("batch 1 LeftForced should mirror batch 0 RightForced")
	}
	if result.Batches[1].RightForced {
		t.Fatalf("final batch RightForced should be false (end of stream)")
	}
}

func TestChunker_DenseNoGap_ForcedAtMax(t *testing.T) {
	// 55 cues, 0.1s gaps — identical pattern to sitcom but smaller sample.
	gaps := make([]float64, 54)
	for i := range gaps {
		gaps[i] = 0.1
	}
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	if len(result.Batches) != 2 {
		t.Fatalf("expected 2 batches (50+5), got %d", len(result.Batches))
	}
	if result.Batches[0].End != 50 || result.Batches[1].Start != 50 {
		t.Fatalf("expected split at index 50, got %+v", result.Batches)
	}
}

func TestChunker_MalformedTiming_Fallback(t *testing.T) {
	cues := []SRTCue{
		{Index: "1", StartSec: 0, EndSec: 1},
		{Index: "2", StartSec: -1, EndSec: -1}, // malformed
		{Index: "3", StartSec: 3, EndSec: 4},
	}

	result := chunkCuesByGap(cues, 50)
	if result.TimingValid {
		t.Fatal("malformed timing should yield TimingValid=false")
	}
	if len(result.Batches) != 1 {
		t.Fatalf("expected 1 fallback batch (3 cues fit in maxBatch=50), got %d", len(result.Batches))
	}
	for _, b := range result.Batches {
		if b.LeftForced || b.RightForced {
			t.Fatalf("fallback batches must have both forced flags false, got %+v", b)
		}
	}
}

func TestChunker_MalformedTiming_LargerFallback(t *testing.T) {
	// 100 cues with one malformed entry — ensure fixed-size fallback splits.
	cues := make([]SRTCue, 100)
	for i := range cues {
		cues[i] = SRTCue{
			Index:    fmt.Sprintf("%d", i+1),
			StartSec: float64(i),
			EndSec:   float64(i) + 0.5,
		}
	}
	cues[42].StartSec = -1
	cues[42].EndSec = -1

	result := chunkCuesByGap(cues, 50)
	if result.TimingValid {
		t.Fatal("expected TimingValid=false")
	}
	if len(result.Batches) != 2 {
		t.Fatalf("expected 2 fallback batches of 50, got %d", len(result.Batches))
	}
	for _, b := range result.Batches {
		if b.LeftForced || b.RightForced {
			t.Fatalf("fallback batches must have forced flags false: %+v", b)
		}
	}
}

func TestChunker_TierOrdering_TooEarlyForCut(t *testing.T) {
	// 50 cues, one 3.5s gap at position 10 (curSize=11 < minSizeSceneBreak=20).
	// Chunker must NOT cut there — must wait for a later boundary or hit max.
	gaps := make([]float64, 49)
	for i := range gaps {
		gaps[i] = 0.2
	}
	gaps[9] = 3.5 // gap after cue index 10 (1-based)
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	// With 50 cues and no other large gap, chunker should produce exactly 1 batch
	// (hit end before maxBatch because len(cues)=50 exactly fits).
	if len(result.Batches) != 1 {
		t.Fatalf("early gap should not split: expected 1 batch, got %d", len(result.Batches))
	}
}

func TestChunker_TierOrdering_AcceptsLateSceneBreak(t *testing.T) {
	// 50 cues, 3.5s gap at position 25 (curSize=26 >= minSizeSceneBreak=20).
	// Chunker should cut there.
	gaps := make([]float64, 49)
	for i := range gaps {
		gaps[i] = 0.2
	}
	gaps[24] = 3.5
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	if len(result.Batches) != 2 {
		t.Fatalf("expected 2 batches split at scene break, got %d", len(result.Batches))
	}
	if result.Batches[0].End != 25 {
		t.Fatalf("expected first batch end=25, got %d", result.Batches[0].End)
	}
	if countRightForced(result.Batches) != 0 {
		t.Fatalf("scene-break cut should be natural, not forced")
	}
}

func TestChunker_ContinuousInvariants(t *testing.T) {
	// Random-ish 80 cues with varied gaps — verify batches are contiguous,
	// cover all cues, and no batch exceeds maxBatch.
	gaps := []float64{
		0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1,
		2.5, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
		0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
		4.0, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
		0.2, 0.2, 0.2, 0.2, 0.2, 1.6, 0.2, 0.2, 0.2, 0.2,
		0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
		0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
		0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2,
	}
	cues := makeCues(0, 1.0, gaps)

	result := chunkCuesByGap(cues, 50)
	if !result.TimingValid {
		t.Fatal("valid timing expected")
	}
	if len(result.Batches) == 0 {
		t.Fatal("expected at least 1 batch")
	}
	if result.Batches[0].Start != 0 {
		t.Fatalf("first batch must start at 0, got %d", result.Batches[0].Start)
	}
	if result.Batches[len(result.Batches)-1].End != len(cues) {
		t.Fatalf("last batch must end at len(cues)=%d, got %d", len(cues), result.Batches[len(result.Batches)-1].End)
	}
	for i, b := range result.Batches {
		size := b.End - b.Start
		if size <= 0 || size > 50 {
			t.Fatalf("batch %d size=%d out of range (1, 50]", i, size)
		}
		if i > 0 && result.Batches[i-1].End != b.Start {
			t.Fatalf("batch %d start=%d doesn't match prev end=%d", i, b.Start, result.Batches[i-1].End)
		}
		if i > 0 && b.LeftForced != result.Batches[i-1].RightForced {
			t.Fatalf("batch %d LeftForced=%v != prev RightForced=%v", i, b.LeftForced, result.Batches[i-1].RightForced)
		}
	}
}

func TestSplitForRetry_NilCuesFallsBackToFixed(t *testing.T) {
	texts := make([]string, 13)
	ranges := splitForRetry(texts, nil, 5)
	want := [][2]int{{0, 5}, {5, 10}, {10, 13}}
	if len(ranges) != len(want) {
		t.Fatalf("expected %d ranges, got %d: %v", len(want), len(ranges), ranges)
	}
	for i, r := range ranges {
		if r != want[i] {
			t.Fatalf("range[%d] = %v, want %v", i, r, want[i])
		}
	}
}

func TestSplitForRetry_LengthMismatchFallsBackToFixed(t *testing.T) {
	texts := make([]string, 10)
	cues := makeCues(0, 1.0, []float64{0.1, 0.1, 0.1}) // only 4 cues
	ranges := splitForRetry(texts, cues, 4)
	// Must ignore cues because len mismatch; fall back to fixed size.
	want := [][2]int{{0, 4}, {4, 8}, {8, 10}}
	if len(ranges) != len(want) {
		t.Fatalf("expected %d ranges, got %d: %v", len(want), len(ranges), ranges)
	}
	for i, r := range ranges {
		if r != want[i] {
			t.Fatalf("range[%d] = %v, want %v", i, r, want[i])
		}
	}
}

func TestParseSRT_PopulatesTiming(t *testing.T) {
	srt := "1\n00:00:01,500 --> 00:00:03,750\nHello\n\n2\n00:00:05,000 --> 00:00:06,200\nWorld\n\n"
	cues := ParseSRT(srt)
	if len(cues) != 2 {
		t.Fatalf("expected 2 cues, got %d", len(cues))
	}
	if cues[0].StartSec != 1.5 || cues[0].EndSec != 3.75 {
		t.Fatalf("cue 0 timing = (%f, %f), want (1.5, 3.75)", cues[0].StartSec, cues[0].EndSec)
	}
	if cues[1].StartSec != 5.0 || cues[1].EndSec != 6.2 {
		t.Fatalf("cue 1 timing = (%f, %f), want (5.0, 6.2)", cues[1].StartSec, cues[1].EndSec)
	}
}

// TestChunker_RealSamples asserts numeric metrics on snippets of real subtitle
// files: sitcom_dense should have at most one forced cut per two batches, and
// movie_sparse should have zero forced cuts. Skipped if testdata files are
// absent (e.g. in a clone that trimmed them).
func TestChunker_RealSamples(t *testing.T) {
	cases := []struct {
		name           string
		path           string
		maxBatch       int
		maxForcedRatio float64 // forcedBoundaries / totalBoundaries
		wantAllNatural bool
	}{
		{"sitcom_dense", "testdata/sitcom_dense.srt", 50, 0.5, false},
		{"movie_sparse", "testdata/movie_sparse.srt", 50, 0.0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.path)
			if err != nil {
				t.Skipf("testdata missing (%v)", err)
			}
			cues := ParseSRT(string(data))
			if len(cues) == 0 {
				t.Fatalf("parsed 0 cues from %s", tc.path)
			}
			result := chunkCuesByGap(cues, tc.maxBatch)
			if !result.TimingValid {
				t.Fatalf("real sample %s should have valid timing", tc.name)
			}
			totalBoundaries := len(result.Batches) - 1
			if totalBoundaries <= 0 {
				t.Fatalf("need at least 2 batches for metric; got %d", len(result.Batches))
			}
			forced := countRightForced(result.Batches)
			ratio := float64(forced) / float64(totalBoundaries)
			t.Logf("%s: cues=%d batches=%d forced=%d ratio=%.2f", tc.name, len(cues), len(result.Batches), forced, ratio)
			if tc.wantAllNatural && forced != 0 {
				t.Fatalf("%s: expected 0 forced boundaries, got %d", tc.name, forced)
			}
			if ratio > tc.maxForcedRatio {
				t.Fatalf("%s: forced ratio %.2f exceeds max %.2f (forced=%d total=%d)", tc.name, ratio, tc.maxForcedRatio, forced, totalBoundaries)
			}
		})
	}
}

// contextCapturingTranslator records every call with its prior/following
// slices so tests can assert exactly which boundaries received overlap.
type contextCapturingTranslator struct {
	name  string
	calls []contextCall
}

type contextCall struct {
	texts     []string
	prior     []string
	following []string
}

func (t *contextCapturingTranslator) Name() string      { return t.name }
func (t *contextCapturingTranslator) MaxBatchSize() int { return 50 }

func (t *contextCapturingTranslator) Translate(_ context.Context, texts []string, _ string) ([]string, error) {
	return t.TranslateWithContext(context.Background(), texts, nil, nil, "vi")
}

func (t *contextCapturingTranslator) TranslateWithContext(_ context.Context, texts, prior, following []string, _ string) ([]string, error) {
	t.calls = append(t.calls, contextCall{texts: append([]string(nil), texts...), prior: append([]string(nil), prior...), following: append([]string(nil), following...)})
	out := make([]string, len(texts))
	for i, tx := range texts {
		out[i] = "vi:" + tx
	}
	return out, nil
}

func TestTranslateSRT_OverlapOnlyForcedBoundaries_Movie(t *testing.T) {
	data, err := os.ReadFile("testdata/movie_sparse.srt")
	if err != nil {
		t.Skipf("testdata missing (%v)", err)
	}
	translator := &contextCapturingTranslator{name: "openai_compatible"}
	if _, err := TranslateSRT(context.Background(), translator, string(data), "vi"); err != nil {
		t.Fatalf("TranslateSRT: %v", err)
	}
	if len(translator.calls) == 0 {
		t.Fatal("expected at least 1 call")
	}
	for i, c := range translator.calls {
		if len(c.prior) != 0 || len(c.following) != 0 {
			t.Fatalf("movie sample has no forced cuts: call %d should have empty prior/following, got prior=%d following=%d", i, len(c.prior), len(c.following))
		}
	}
}

func TestTranslateSRT_OverlapOnlyForcedBoundaries_Sitcom(t *testing.T) {
	data, err := os.ReadFile("testdata/sitcom_dense.srt")
	if err != nil {
		t.Skipf("testdata missing (%v)", err)
	}
	translator := &contextCapturingTranslator{name: "openai_compatible"}
	if _, err := TranslateSRT(context.Background(), translator, string(data), "vi"); err != nil {
		t.Fatalf("TranslateSRT: %v", err)
	}
	// Cross-check against chunker: count forced boundaries, then count calls
	// that received at least one side of overlap. The numbers must match.
	cues := ParseSRT(string(data))
	result := chunkCuesByGap(cues, 50)
	expectedOverlapCalls := 0
	for _, b := range result.Batches {
		if b.LeftForced || b.RightForced {
			expectedOverlapCalls++
		}
	}
	gotOverlapCalls := 0
	for _, c := range translator.calls {
		if len(c.prior) > 0 || len(c.following) > 0 {
			gotOverlapCalls++
		}
	}
	if expectedOverlapCalls == 0 {
		t.Fatal("test data no longer contains forced cuts; pick a denser sample")
	}
	if gotOverlapCalls != expectedOverlapCalls {
		t.Fatalf("overlap calls: got %d, want %d (forced boundaries from chunker)", gotOverlapCalls, expectedOverlapCalls)
	}
}

func TestTranslateSRT_MalformedTiming_NoOverlap(t *testing.T) {
	// Build an SRT where cue #3 has a malformed timing line — forces the
	// chunker into fixed-size fallback. Translator must receive no overlap
	// even if fixed-size forces many interior boundaries in principle.
	var sb strings.Builder
	for i := 1; i <= 80; i++ {
		sb.WriteString(fmt.Sprintf("%d\n", i))
		if i == 3 {
			sb.WriteString("BROKEN TIMING -->\n")
		} else {
			// Stagger timing so cues are sequential and valid.
			start := float64(i)
			end := start + 0.5
			sb.WriteString(fmt.Sprintf("00:00:%02d,%03d --> 00:00:%02d,%03d\n",
				int(start), int((start-float64(int(start)))*1000),
				int(end), int((end-float64(int(end)))*1000)))
		}
		sb.WriteString(fmt.Sprintf("line %d\n\n", i))
	}

	translator := &contextCapturingTranslator{name: "openai_compatible"}
	if _, err := TranslateSRT(context.Background(), translator, sb.String(), "vi"); err != nil {
		t.Fatalf("TranslateSRT: %v", err)
	}
	if len(translator.calls) < 2 {
		t.Fatalf("expected at least 2 calls (80 cues / maxBatch 50), got %d", len(translator.calls))
	}
	for i, c := range translator.calls {
		if len(c.prior) != 0 || len(c.following) != 0 {
			t.Fatalf("malformed timing: call %d must have no overlap, got prior=%d following=%d", i, len(c.prior), len(c.following))
		}
	}
}

func TestParseSRT_MalformedTimingYieldsSentinel(t *testing.T) {
	srt := "1\nNOT A TIMING -->\nHello\n\n"
	cues := ParseSRT(srt)
	// Malformed timing line still has "-->" so parser advances state, but
	// parseSRTTiming fails → StartSec/EndSec should be -1 sentinel.
	if len(cues) != 1 {
		t.Fatalf("expected 1 cue, got %d", len(cues))
	}
	if cues[0].StartSec != -1 || cues[0].EndSec != -1 {
		t.Fatalf("malformed timing should yield -1 sentinels, got (%f, %f)", cues[0].StartSec, cues[0].EndSec)
	}
}
