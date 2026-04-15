package translate

import (
	"strings"
	"testing"
)

func TestParseLLMTranslationsExtractsStrictJSON(t *testing.T) {
	raw := "```json\n{\"translations\":[\"Xin chao\",\"Tam biet\"]}\n```"

	got, err := parseLLMTranslations(raw, 2)
	if err != nil {
		t.Fatalf("parseLLMTranslations returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0] != "Xin chao" || got[1] != "Tam biet" {
		t.Fatalf("unexpected translations: %#v", got)
	}
}

func TestParseLLMTranslationsRejectsWrongCount(t *testing.T) {
	raw := "{\"translations\":[\"Xin chao\"]}"

	_, err := parseLLMTranslations(raw, 2)
	if err == nil {
		t.Fatal("expected error for mismatched translation count")
	}
}

func TestParserIndexed_ValidCount(t *testing.T) {
	raw := `{"translations":[{"index":0,"text":"Xin chao"},{"index":1,"text":"Tam biet"},{"index":2,"text":"Hen gap lai"}]}`

	got, err := parseLLMTranslations(raw, 3)
	if err != nil {
		t.Fatalf("parseLLMTranslations returned error: %v", err)
	}
	want := []string{"Xin chao", "Tam biet", "Hen gap lai"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestParserIndexed_RejectOverflow(t *testing.T) {
	// LLM returns 6 items for expected=3 (e.g. leaked context translation).
	// Previously accepted silently and caller would overwrite next batch's cues.
	raw := `{"translations":[{"index":0,"text":"a"},{"index":1,"text":"b"},{"index":2,"text":"c"},{"index":3,"text":"d"},{"index":4,"text":"e"},{"index":5,"text":"f"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected error for index overflow, got nil")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error should mention 'out of range', got: %v", err)
	}
}

func TestParserIndexed_RejectIndexOutOfRange(t *testing.T) {
	// Sparse indexes with one beyond expected (0, 1, 5 for expected=3).
	raw := `{"translations":[{"index":0,"text":"a"},{"index":1,"text":"b"},{"index":5,"text":"f"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected error for out-of-range index, got nil")
	}
	if !strings.Contains(err.Error(), "5") || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error should mention index 5 out of range, got: %v", err)
	}
}

func TestParserIndexed_RejectNegativeIndex(t *testing.T) {
	// Previously maxIndex:=-1 initial + index:-1 → loop didn't run → empty success.
	raw := `{"translations":[{"index":-1,"text":"a"},{"index":0,"text":"b"},{"index":1,"text":"c"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected error for negative index, got nil")
	}
	if !strings.Contains(err.Error(), "-1") || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error should mention index -1 out of range, got: %v", err)
	}
}

func TestParserIndexed_RejectDuplicateIndex(t *testing.T) {
	// Previously byIndex[t.Index] = t.Text silently overwrote duplicates,
	// so [0,1,1,2] collapsed to 3 items and bypassed count validation.
	raw := `{"translations":[{"index":0,"text":"a"},{"index":1,"text":"b"},{"index":1,"text":"b2"},{"index":2,"text":"c"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected error for duplicate index, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error should mention 'duplicate', got: %v", err)
	}
}

func TestParserIndexed_PartialMissingStillWorks(t *testing.T) {
	// Legitimate partial response: LLM skipped index 1.
	// Parser must surface this via partialTranslationError so retry can fill gaps.
	raw := `{"translations":[{"index":0,"text":"a"},{"index":2,"text":"c"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected partial error, got nil")
	}
	partial, ok := err.(*partialTranslationError)
	if !ok {
		t.Fatalf("expected *partialTranslationError, got %T: %v", err, err)
	}
	if len(partial.missing) != 1 || partial.missing[0] != 1 {
		t.Fatalf("missing should be [1], got %v", partial.missing)
	}
	if len(partial.translations) != 2 {
		t.Fatalf("should have 2 translations present, got %d", len(partial.translations))
	}
}

func TestParserIndexed_PartialMissingTrailing(t *testing.T) {
	// LLM only returned index 0 for expected=3; previously accepted as complete
	// (maxIndex=0 loop) — now must surface missing=[1,2].
	raw := `{"translations":[{"index":0,"text":"a"}]}`

	_, err := parseLLMTranslations(raw, 3)
	if err == nil {
		t.Fatal("expected partial error for trailing missing indexes")
	}
	partial, ok := err.(*partialTranslationError)
	if !ok {
		t.Fatalf("expected *partialTranslationError, got %T: %v", err, err)
	}
	if len(partial.missing) != 2 || partial.missing[0] != 1 || partial.missing[1] != 2 {
		t.Fatalf("missing should be [1, 2], got %v", partial.missing)
	}
}

func TestParseLLMTranslationsAcceptsDataWrapperInCodeFence(t *testing.T) {
	raw := "Day la noi dung translate cua em\n\n```json\n{\"data\":[\"Xin chao\",\"Tam biet\"]}\n```"

	got, err := parseLLMTranslations(raw, 2)
	if err != nil {
		t.Fatalf("parseLLMTranslations returned error: %v", err)
	}
	if got[0] != "Xin chao" || got[1] != "Tam biet" {
		t.Fatalf("unexpected translations: %#v", got)
	}
}

func TestParseLLMTranslationsAcceptsNestedDataTranslations(t *testing.T) {
	raw := "{\"data\":{\"translations\":[\"Xin chao\",\"Tam biet\"]}}"

	got, err := parseLLMTranslations(raw, 2)
	if err != nil {
		t.Fatalf("parseLLMTranslations returned error: %v", err)
	}
	if got[0] != "Xin chao" || got[1] != "Tam biet" {
		t.Fatalf("unexpected translations: %#v", got)
	}
}

func TestExtractOpenAIMessageTextSupportsPartArrays(t *testing.T) {
	content := []any{
		map[string]any{"text": "{\"translations\":[\"hello\"]}"},
	}

	got, err := extractOpenAIMessageText(content)
	if err != nil {
		t.Fatalf("extractOpenAIMessageText returned error: %v", err)
	}
	if got != "{\"translations\":[\"hello\"]}" {
		t.Fatalf("got %q, want JSON text", got)
	}
}

func TestProviderEndpointBuilders(t *testing.T) {
	if got := openAIChatCompletionsURL(""); got != "https://api.openai.com/v1/chat/completions" {
		t.Fatalf("openAIChatCompletionsURL() = %q", got)
	}
	if got := geminiGenerateContentURL("", "gemini-2.5-flash"); got != "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent" {
		t.Fatalf("geminiGenerateContentURL() = %q", got)
	}
	if got := anthropicMessagesURL(""); got != "https://api.anthropic.com/v1/messages" {
		t.Fatalf("anthropicMessagesURL() = %q", got)
	}
}

func TestBuildLLMUserPromptIncludesMediaContext(t *testing.T) {
	prompt, err := buildLLMUserPrompt([]string{"Hello there"}, "vi", &AIContext{
		Title:     "Malcolm in the Middle",
		MediaType: "episode",
		Genres:    []string{"Comedy", "Family"},
		Overview:  "A gifted boy navigates a chaotic family.",
		Tagline:   "Life is unfair.",
	})
	if err != nil {
		t.Fatalf("buildLLMUserPrompt returned error: %v", err)
	}

	if !strings.Contains(prompt, "Media context:") {
		t.Fatalf("prompt missing media context block: %q", prompt)
	}
	if !strings.Contains(prompt, "Title: Malcolm in the Middle") {
		t.Fatalf("prompt missing title: %q", prompt)
	}
	if !strings.Contains(prompt, "target language: vi") {
		t.Fatalf("prompt missing target language: %q", prompt)
	}
}
