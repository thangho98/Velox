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
