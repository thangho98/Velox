package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL   = "https://api.openai.com/v1"
	defaultGeminiBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultAnthropicAPIURL = "https://api.anthropic.com"
	defaultAIModelPrompt   = `You are an expert subtitle localizer for films and TV series.

PRIMARY INVARIANT (violating this breaks the entire system):
Output is a JSON object {"translations":[{"index":N,"text":"..."},...]}. For every input cue [N] you produce EXACTLY ONE object with index N. Input count == output count. No merging, no splitting, no extras, no gaps.

Indexing rules:
- If "TRANSLATE THESE" lists M items [0]..[M-1], return exactly M objects with indexes 0..M-1 — unique, contiguous, no duplicates, no index >= M.
- PRIOR CONTEXT / FOLLOWING CONTEXT items are read-only. Never translate them. Never emit objects for them.

Per-cue rules:
- Treat each [N] as one atomic cue. Preserve its internal line breaks inside its own "text" field (use \n). Never pull lines out of a cue into a separate object. Never fold two cues into one object, even if their content looks like one continuous sentence or song verse.
- Preserve lyric/music markers such as ♪ and repeated chant lines inside the cue that contains them.
- Keep proper nouns, character names, and franchise-specific terms consistent.
- If a cue should stay unchanged (e.g. numeric-only, standalone name), return it unchanged.

Translation quality:
- Translate naturally for cinematic subtitles, preserving tone, subtext, humor, and character intent.
- Use media context only as soft background to improve tone, terminology, and character voice.
- Infer speaker gender, relationships, and context by analyzing the conversational flow across the batch. For languages with complex pronoun systems (like Vietnamese), keep relative pronouns consistent between speakers across adjacent cues.
- Never invent plot details, relationships, or facts not supported by the cue text or media context.

Output format:
- Output the raw JSON object immediately. No <thinking> blocks, no explanations, no markdown, no speaker labels, no timestamps, no extra fields.`
)

type AIProvider string

const (
	ProviderOpenAICompatible    AIProvider = "openai_compatible"
	ProviderGeminiCompatible    AIProvider = "gemini_compatible"
	ProviderAnthropicCompatible AIProvider = "anthropic_compatible"
)

// AIConfig configures an AI-backed subtitle translator.
type AIConfig struct {
	Provider AIProvider
	APIKey   string
	BaseURL  string
	Model    string
	Context  *AIContext
}

type AIContext struct {
	Title     string
	MediaType string
	Genres    []string
	Overview  string
	Tagline   string
}

// NewAI creates a translator backed by a compatible LLM provider.
func NewAI(cfg AIConfig) (Translator, error) {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.Model = strings.TrimSpace(cfg.Model)

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("ai translator: api key is required")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("ai translator: model is required")
	}

	client := &http.Client{Timeout: 180 * time.Second}
	switch cfg.Provider {
	case ProviderOpenAICompatible:
		return &openAICompatibleTranslator{cfg: cfg, http: client}, nil
	case ProviderGeminiCompatible:
		return &geminiTranslator{cfg: cfg, http: client}, nil
	case ProviderAnthropicCompatible:
		return &anthropicTranslator{cfg: cfg, http: client}, nil
	default:
		return nil, fmt.Errorf("ai translator: unsupported provider %q", cfg.Provider)
	}
}

type indexedTranslation struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type openAICompatibleTranslator struct {
	cfg  AIConfig
	http *http.Client
}

func (t *openAICompatibleTranslator) Name() string      { return string(t.cfg.Provider) }
func (t *openAICompatibleTranslator) MaxBatchSize() int { return 50 }

func (t *openAICompatibleTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	return t.TranslateWithContext(ctx, texts, nil, nil, targetLang)
}

func (t *openAICompatibleTranslator) TranslateWithContext(ctx context.Context, texts, prior, following []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context, prior, following)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":            t.cfg.Model,
		"temperature":      0.2,
		"max_tokens":       estimateLLMMaxTokens(texts),
		"response_format":  map[string]string{"type": "json_object"},
		"reasoning_effort": "low",
		"messages": []map[string]string{
			{"role": "system", "content": defaultAIModelPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	slog.Debug("ai_translate_openai_request",
		"provider", t.Name(),
		"model", t.cfg.Model,
		"base_url", t.cfg.BaseURL,
		"batch_size", len(texts),
		"target_lang", targetLang,
		"prompt_preview", truncate(userPrompt, 200),
	)

	raw, err := t.postJSON(ctx, openAIChatCompletionsURL(t.cfg.BaseURL), map[string]string{
		"Authorization": "Bearer " + t.cfg.APIKey,
	}, body)
	if err != nil {
		slog.Error("ai_translate_openai_request_failed",
			"provider", t.Name(),
			"error", err,
		)
		return nil, err
	}

	slog.Debug("ai_translate_openai_response",
		"provider", t.Name(),
		"response_len", len(raw),
		"response_preview", truncate(string(raw), 300),
	)

	var resp struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("openai compatible: decode: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai compatible: no choices returned")
	}

	content, err := extractOpenAIMessageText(resp.Choices[0].Message.Content)
	if err != nil {
		slog.Warn("ai_translate_openai_extract_failed",
			"provider", t.Name(),
			"error", err,
		)
		return nil, err
	}
	translations, parseErr := parseLLMTranslations(content, len(texts))
	if parseErr != nil {
		slog.Warn("ai_translate_openai_parse_failed",
			"provider", t.Name(),
			"model", t.cfg.Model,
			"batch_size", len(texts),
			"error", parseErr,
			"content_full", content,
			"raw_response", string(raw),
		)
	}
	return translations, parseErr
}

func (t *openAICompatibleTranslator) postJSON(
	ctx context.Context,
	endpoint string,
	extraHeaders map[string]string,
	payload any,
) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: encode request: %w", t.Name(), err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s: build request: %w", t.Name(), err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := t.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", t.Name(), err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("%s: read response: %w", t.Name(), readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: status %d: %s", t.Name(), resp.StatusCode, string(body))
	}
	return body, nil
}

type geminiTranslator struct {
	cfg  AIConfig
	http *http.Client
}

func (t *geminiTranslator) Name() string      { return string(t.cfg.Provider) }
func (t *geminiTranslator) MaxBatchSize() int { return 50 }

func (t *geminiTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	return t.TranslateWithContext(ctx, texts, nil, nil, targetLang)
}

func (t *geminiTranslator) TranslateWithContext(ctx context.Context, texts, prior, following []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context, prior, following)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{
				{"text": defaultAIModelPrompt},
			},
		},
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": userPrompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":      0.2,
			"responseMimeType": "application/json",
			"thinkingConfig": map[string]any{
				"thinkingBudget": -1,
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini compatible: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, geminiGenerateContentURL(t.cfg.BaseURL, t.cfg.Model), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gemini compatible: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", t.cfg.APIKey)

	slog.Debug("ai_translate_gemini_request",
		"provider", t.Name(),
		"model", t.cfg.Model,
		"base_url", t.cfg.BaseURL,
		"batch_size", len(texts),
		"target_lang", targetLang,
		"prompt_preview", truncate(userPrompt, 200),
	)

	resp, err := t.http.Do(req)
	if err != nil {
		slog.Error("ai_translate_gemini_request_failed",
			"provider", t.Name(),
			"error", err,
		)
		return nil, fmt.Errorf("gemini compatible: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("gemini compatible: read response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("ai_translate_gemini_response_error",
			"provider", t.Name(),
			"status", resp.StatusCode,
			"response_preview", truncate(string(raw), 300),
		)
		return nil, fmt.Errorf("gemini compatible: status %d: %s", resp.StatusCode, string(raw))
	}

	slog.Debug("ai_translate_gemini_response",
		"provider", t.Name(),
		"response_len", len(raw),
		"response_preview", truncate(string(raw), 300),
	)

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("gemini compatible: decode: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini compatible: no content returned")
	}

	translations, parseErr := parseLLMTranslations(result.Candidates[0].Content.Parts[0].Text, len(texts))
	if parseErr != nil {
		slog.Warn("ai_translate_gemini_parse_failed",
			"provider", t.Name(),
			"model", t.cfg.Model,
			"batch_size", len(texts),
			"error", parseErr,
			"content_full", result.Candidates[0].Content.Parts[0].Text,
			"raw_response", string(raw),
		)
	}
	return translations, parseErr
}

type anthropicTranslator struct {
	cfg  AIConfig
	http *http.Client
}

func (t *anthropicTranslator) Name() string      { return string(t.cfg.Provider) }
func (t *anthropicTranslator) MaxBatchSize() int { return 50 }

func (t *anthropicTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	return t.TranslateWithContext(ctx, texts, nil, nil, targetLang)
}

func (t *anthropicTranslator) TranslateWithContext(ctx context.Context, texts, prior, following []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context, prior, following)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":       t.cfg.Model,
		"temperature": 0.2,
		"max_tokens":  estimateLLMMaxTokens(texts),
		"system":      defaultAIModelPrompt,
		"thinking":    map[string]any{"type": "disabled"},
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic compatible: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicMessagesURL(t.cfg.BaseURL), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("anthropic compatible: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", t.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	slog.Debug("ai_translate_anthropic_request",
		"provider", t.Name(),
		"model", t.cfg.Model,
		"base_url", t.cfg.BaseURL,
		"batch_size", len(texts),
		"target_lang", targetLang,
		"prompt_preview", truncate(userPrompt, 200),
	)

	resp, err := t.http.Do(req)
	if err != nil {
		slog.Error("ai_translate_anthropic_request_failed",
			"provider", t.Name(),
			"error", err,
		)
		return nil, fmt.Errorf("anthropic compatible: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("anthropic compatible: read response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("ai_translate_anthropic_response_error",
			"provider", t.Name(),
			"status", resp.StatusCode,
			"response_preview", truncate(string(raw), 300),
		)
		return nil, fmt.Errorf("anthropic compatible: status %d: %s", resp.StatusCode, string(raw))
	}

	slog.Debug("ai_translate_anthropic_response",
		"provider", t.Name(),
		"response_len", len(raw),
		"response_preview", truncate(string(raw), 300),
	)

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("anthropic compatible: decode: %w", err)
	}

	var sb strings.Builder
	for _, block := range result.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if sb.Len() == 0 {
		slog.Warn("ai_translate_anthropic_no_text_content",
			"provider", t.Name(),
			"model", t.cfg.Model,
			"batch_size", len(texts),
			"raw_response", string(raw),
		)
		return nil, fmt.Errorf("anthropic compatible: no text content returned")
	}

	translations, parseErr := parseLLMTranslations(sb.String(), len(texts))
	if parseErr != nil {
		slog.Warn("ai_translate_anthropic_parse_failed",
			"provider", t.Name(),
			"model", t.cfg.Model,
			"batch_size", len(texts),
			"error", parseErr,
			"content_full", sb.String(),
			"raw_response", string(raw),
		)
	}
	return translations, parseErr
}

func buildLLMUserPrompt(texts []string, targetLang string, mediaCtx *AIContext, prior, following []string) (string, error) {
	var sb strings.Builder
	if contextBlock := buildMediaContextBlock(mediaCtx); contextBlock != "" {
		sb.WriteString(contextBlock)
		sb.WriteString("\n\n")
	}
	if len(prior) > 0 {
		sb.WriteString("--- PRIOR CONTEXT (do NOT translate, for flow only) ---\n")
		for i, text := range prior {
			sb.WriteString(fmt.Sprintf("[prior-%d] %s\n", len(prior)-i, text))
		}
		sb.WriteString("--- END PRIOR CONTEXT ---\n\n")
	}
	n := len(texts)
	sb.WriteString("--- TRANSLATE THESE ---\n")
	sb.WriteString(fmt.Sprintf("Target language: %s\n", targetLang))
	sb.WriteString(fmt.Sprintf("Required output: translations array of length EXACTLY %d with every index 0..%d present, unique, in order. No skips. No duplicates. No index %d or higher.\n", n, n-1, n))
	sb.WriteString(fmt.Sprintf("Shape: {\"translations\":[{\"index\":0,\"text\":\"...\"},{\"index\":1,\"text\":\"...\"},...,{\"index\":%d,\"text\":\"...\"}]}\n", n-1))
	sb.WriteString("Input items:\n")
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i, text))
	}
	sb.WriteString("--- END TRANSLATE THESE ---\n")
	if len(following) > 0 {
		sb.WriteString("\n--- FOLLOWING CONTEXT (do NOT translate, for flow only) ---\n")
		for i, text := range following {
			sb.WriteString(fmt.Sprintf("[next-%d] %s\n", i+1, text))
		}
		sb.WriteString("--- END FOLLOWING CONTEXT ---\n")
	}
	return sb.String(), nil
}

func buildMediaContextBlock(mediaCtx *AIContext) string {
	if mediaCtx == nil {
		return ""
	}

	title := normalizePromptText(mediaCtx.Title, 120)
	mediaType := normalizePromptText(mediaCtx.MediaType, 40)
	overview := normalizePromptText(mediaCtx.Overview, 420)
	tagline := normalizePromptText(mediaCtx.Tagline, 160)

	genres := make([]string, 0, len(mediaCtx.Genres))
	for _, genre := range mediaCtx.Genres {
		if normalized := normalizePromptText(genre, 40); normalized != "" {
			genres = append(genres, normalized)
		}
	}

	var parts []string
	if title != "" {
		parts = append(parts, "Title: "+title)
	}
	if mediaType != "" {
		parts = append(parts, "Type: "+mediaType)
	}
	if len(genres) > 0 {
		parts = append(parts, "Genres: "+strings.Join(genres, ", "))
	}
	if tagline != "" {
		parts = append(parts, "Tagline: "+tagline)
	}
	if overview != "" {
		parts = append(parts, "Short premise: "+overview)
	}
	if len(parts) == 0 {
		return ""
	}

	return "Media context:\n" + strings.Join(parts, "\n")
}

func normalizePromptText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	if maxLen > 0 && len(text) > maxLen {
		text = strings.TrimSpace(text[:maxLen])
		text = strings.TrimRight(text, ",;:- ")
		text += "..."
	}
	return text
}

func parseLLMTranslations(raw string, expected int) ([]string, error) {
	candidates := extractJSONCandidates(raw)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("llm response does not contain a JSON object")
	}

	var lastErr error
	for _, candidate := range candidates {
		translations, err := decodeLLMTranslations(candidate, expected)
		if err == nil {
			return translations, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("llm response does not contain a valid translations payload")
}

func extractJSONObject(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("llm response does not contain a JSON object")
	}

	// Find JSON object start - look for {" (JSON object notation)
	// This avoids picking up braces in thinking/reasoning content
	jsonStart := strings.Index(raw, `{"`)
	if jsonStart < 0 {
		// Fallback: find first {
		jsonStart = strings.Index(raw, "{")
	}
	if jsonStart < 0 {
		return "", fmt.Errorf("llm response does not contain a JSON object")
	}

	// Find the matching closing brace by counting brace levels
	// Start from jsonStart and find the } that closes this {
	depth := 0
	inString := false
	jsonEnd := -1

	for i := jsonStart; i < len(raw); i++ {
		c := raw[i]
		if c == '"' && (i == 0 || raw[i-1] != '\\') {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				jsonEnd = i
				break
			}
		}
	}

	if jsonEnd < 0 {
		return "", fmt.Errorf("llm response does not contain a JSON object")
	}
	return raw[jsonStart : jsonEnd+1], nil
}

func extractJSONCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	seen := map[string]struct{}{}
	var candidates []string
	add := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	if jsonBlock, err := extractJSONObject(raw); err == nil {
		add(jsonBlock)
	}

	for _, block := range extractCodeFenceBlocks(raw) {
		if jsonBlock, err := extractJSONObject(block); err == nil {
			add(jsonBlock)
			continue
		}
		add(block)
	}

	return candidates
}

func extractCodeFenceBlocks(raw string) []string {
	var blocks []string
	parts := strings.Split(raw, "```")
	for i := 1; i < len(parts); i += 2 {
		block := strings.TrimSpace(parts[i])
		if block == "" {
			continue
		}
		if newline := strings.Index(block, "\n"); newline >= 0 {
			lang := strings.TrimSpace(block[:newline])
			if strings.EqualFold(lang, "json") || strings.EqualFold(lang, "javascript") || strings.EqualFold(lang, "js") {
				block = strings.TrimSpace(block[newline+1:])
			}
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func decodeLLMTranslations(jsonBlock string, expected int) ([]string, error) {
	// Try indexed format first
	translations, missing, err := tryDecodeIndexedTranslations(jsonBlock, expected)
	if err == nil {
		if len(missing) == 0 {
			return translations, nil
		}
		// Some indexes are missing - return partial results with error
		return translations, &partialTranslationError{
			translations: translations,
			missing:      missing,
			expected:     expected,
		}
	}
	// Validation errors (out-of-range, duplicate, etc.) must not fall through
	// to the non-indexed path — the payload IS indexed, just malformed.
	// Shape-mismatch errors (no translations key, empty, etc.) may fall through
	// because the payload might actually use a different shape.
	if _, isValidation := err.(*indexValidationError); isValidation {
		return nil, err
	}

	// Fallback to non-indexed format
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonBlock), &payload); err != nil {
		return nil, fmt.Errorf("parse llm translations: %w", err)
	}

	for _, key := range []string{"translations", "data", "result", "output", "response"} {
		rawValue, ok := payload[key]
		if !ok {
			continue
		}
		translations, err := decodeTranslationArray(rawValue)
		if err == nil {
			return finalizeTranslations(translations, expected)
		}
	}

	return nil, fmt.Errorf("parse llm translations: missing translations array")
}

// partialTranslationError indicates some translations are missing but we have partial results
type partialTranslationError struct {
	translations []string
	missing      []int // indexes of missing translations
	expected     int
}

func (e *partialTranslationError) Error() string {
	return fmt.Sprintf("parse llm translations: expected %d items, got %d (missing indexes: %v)", e.expected, len(e.translations), e.missing)
}

// indexValidationError signals that the indexed payload parsed but violated
// integrity rules (out-of-range, duplicate, negative index). decodeLLMTranslations
// must treat this as a hard reject rather than falling through to non-indexed
// shape decoding — the payload IS indexed, just malformed.
type indexValidationError struct {
	msg string
}

func (e *indexValidationError) Error() string {
	return "parse llm translations: " + e.msg
}

// tryDecodeIndexedTranslations attempts to parse indexed response format.
// When expected >= 0, every index must be unique and lie in [0, expected).
// Returns translations in order (by index), missing indexes if any, and error if parsing failed.
func tryDecodeIndexedTranslations(jsonBlock string, expected int) ([]string, []int, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonBlock), &payload); err != nil {
		return nil, nil, fmt.Errorf("decode indexed: %w", err)
	}

	rawValue, ok := payload["translations"]
	if !ok {
		return nil, nil, fmt.Errorf("no translations key")
	}

	var indexed []indexedTranslation
	if err := json.Unmarshal(rawValue, &indexed); err != nil {
		return nil, nil, fmt.Errorf("decode indexed array: %w", err)
	}

	if len(indexed) == 0 {
		return nil, nil, fmt.Errorf("empty translations")
	}

	// Strict validation when caller knows expected count: every index must be
	// unique and within [0, expected). This blocks three silent-corruption bugs:
	//   1. LLM returns extra items (e.g. 6 for expected=3) — caller would
	//      overwrite cues of the next batch.
	//   2. Duplicate indexes — map writes silently drop earlier values.
	//   3. Negative indexes — prior reconstruction loop ignored them.
	if expected >= 0 {
		seen := make(map[int]bool, len(indexed))
		for _, t := range indexed {
			if t.Index < 0 || t.Index >= expected {
				return nil, nil, &indexValidationError{msg: fmt.Sprintf("index %d out of range [0, %d)", t.Index, expected)}
			}
			if seen[t.Index] {
				return nil, nil, &indexValidationError{msg: fmt.Sprintf("duplicate index %d", t.Index)}
			}
			seen[t.Index] = true
		}
	}

	byIndex := make(map[int]string, len(indexed))
	maxIndex := -1
	for _, t := range indexed {
		byIndex[t.Index] = t.Text
		if t.Index > maxIndex {
			maxIndex = t.Index
		}
	}

	// When expected is known, always reconstruct against [0, expected) so
	// missing detection covers trailing gaps (e.g. LLM returns only [0] for
	// expected=3). When unknown (legacy nested-decode path, expected=-1),
	// fall back to maxIndex-based reconstruction.
	bound := expected
	if bound < 0 {
		bound = maxIndex + 1
	}

	var missing []int
	translations := make([]string, 0, bound)
	for i := 0; i < bound; i++ {
		if text, ok := byIndex[i]; ok {
			translations = append(translations, text)
		} else {
			missing = append(missing, i)
		}
	}

	return translations, missing, nil
}

func decodeTranslationArray(rawValue json.RawMessage) ([]string, error) {
	var arr []string
	if err := json.Unmarshal(rawValue, &arr); err == nil {
		return arr, nil
	}

	var nested struct {
		Translations []string `json:"translations"`
		Data         []string `json:"data"`
	}
	if err := json.Unmarshal(rawValue, &nested); err == nil {
		switch {
		case len(nested.Translations) > 0:
			return nested.Translations, nil
		case len(nested.Data) > 0:
			return nested.Data, nil
		}
	}

	var str string
	if err := json.Unmarshal(rawValue, &str); err == nil {
		candidates := extractJSONCandidates(str)
		for _, candidate := range candidates {
			translations, nestedErr := decodeLLMTranslations(candidate, -1)
			if nestedErr == nil {
				return translations, nil
			}
		}
	}

	return nil, fmt.Errorf("parse llm translations: value is not a translation array")
}

func finalizeTranslations(translations []string, expected int) ([]string, error) {
	if expected >= 0 && len(translations) != expected {
		return nil, fmt.Errorf("parse llm translations: expected %d items, got %d", expected, len(translations))
	}

	out := make([]string, len(translations))
	for i, item := range translations {
		out[i] = strings.TrimSpace(item)
	}
	return out, nil
}

func extractOpenAIMessageText(content any) (string, error) {
	switch v := content.(type) {
	case string:
		return v, nil
	case []any:
		var sb strings.Builder
		for _, part := range v {
			obj, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := obj["text"].(string); ok {
				sb.WriteString(text)
			}
		}
		if sb.Len() == 0 {
			return "", fmt.Errorf("openai compatible: empty content parts")
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("openai compatible: unsupported message content shape")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func estimateLLMMaxTokens(texts []string) int {
	totalChars := 0
	for _, text := range texts {
		totalChars += len(text)
	}
	// Increased buffer: translation output is often ~1.5x input length
	// Plus extra for JSON overhead and HEAVY reasoning/thinking content for reasoning models
	estimate := totalChars/2 + 8192
	if estimate < 8192 {
		return 8192
	}
	if estimate > 262144 {
		return 262144
	}
	return estimate
}

func openAIChatCompletionsURL(base string) string {
	if strings.TrimSpace(base) == "" {
		base = defaultOpenAIBaseURL
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

func geminiGenerateContentURL(base, model string) string {
	if strings.TrimSpace(base) == "" {
		base = defaultGeminiBaseURL
	}
	base = strings.TrimRight(base, "/")
	escapedModel := url.PathEscape(model)
	if strings.Contains(base, ":generateContent") {
		return base
	}
	if strings.Contains(base, "/models/") {
		return base + ":generateContent"
	}
	return base + "/models/" + escapedModel + ":generateContent"
}

func anthropicMessagesURL(base string) string {
	if strings.TrimSpace(base) == "" {
		base = defaultAnthropicAPIURL
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/messages") {
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/messages"
	}
	return base + "/v1/messages"
}
