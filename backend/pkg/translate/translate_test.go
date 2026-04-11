package translate

import (
	"context"
	"fmt"
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
