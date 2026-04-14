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
Return JSON only using this exact shape: {"translations":["..."]}.
Rules:
- Keep the same number of items and preserve their order exactly.
- Treat each input item as one atomic subtitle cue. Never split one cue into multiple output items.
- Count the input items first, then count your output items again before answering. The counts must match exactly.
- Translate naturally for cinematic subtitles, preserving tone, subtext, humor, and character intent.
- Use the optional media context only as soft background to improve tone, terminology, and character voice.
- Because SRT does not label speakers, you MUST infer speaker gender, relationships, and context by analyzing the conversational flow across the entire batch of provided cues.
- For languages with complex pronoun systems (like Vietnamese), maintain consistent relative pronouns between speakers across adjacent dialog cues.
- Never invent plot details, relationships, or facts that are not supported by the cue text or the supplied media context.
- Keep proper nouns, names, and franchise-specific terms consistent when they should remain unchanged.
- DO NOT generate any <thinking>, reasoning, or explanation blocks. Output the raw JSON directly and immediately.
- Preserve line breaks when they improve subtitle readability.
- Preserve lyric and music markers such as ♪ when they appear in the cue.
- If a cue contains song lyrics, repeated lines, multiple short lines, or chant-like phrasing, keep all of it inside the same single output item.
- If two neighboring cues are both lyrics, they are still two separate JSON strings because they were two separate input items.
- Never merge adjacent cues together, and never split one cue apart, even if punctuation, lyrics, or line breaks look unusual.
- Do not turn lyric lines into separate list entries, verses, bullet points, or extra JSON items.
- Do not add explanations, markdown, speaker labels, timestamps, or extra fields.
- If a cue should stay unchanged, return it unchanged.`
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
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":           t.cfg.Model,
		"temperature":     0.2,
		"max_tokens":      estimateLLMMaxTokens(texts),
		"response_format": map[string]string{"type": "json_object"},
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
			"error", parseErr,
			"content_preview", truncate(content, 300),
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
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context)
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
			"error", parseErr,
			"content_preview", truncate(result.Candidates[0].Content.Parts[0].Text, 300),
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
	if len(texts) == 0 {
		return nil, nil
	}

	userPrompt, err := buildLLMUserPrompt(texts, targetLang, t.cfg.Context)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":       t.cfg.Model,
		"temperature": 0.2,
		"max_tokens":  estimateLLMMaxTokens(texts),
		"system":      defaultAIModelPrompt,
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
		return nil, fmt.Errorf("anthropic compatible: no text content returned")
	}

	translations, parseErr := parseLLMTranslations(sb.String(), len(texts))
	if parseErr != nil {
		slog.Warn("ai_translate_anthropic_parse_failed",
			"provider", t.Name(),
			"error", parseErr,
			"content_preview", truncate(sb.String(), 300),
		)
	}
	return translations, parseErr
}

func buildLLMUserPrompt(texts []string, targetLang string, mediaCtx *AIContext) (string, error) {
	var sb strings.Builder
	if contextBlock := buildMediaContextBlock(mediaCtx); contextBlock != "" {
		sb.WriteString(contextBlock)
		sb.WriteString("\n\n")
	}
	sb.WriteString(fmt.Sprintf("Translate these subtitle cues to target language: %s.\n", targetLang))
	sb.WriteString("IMPORTANT: You MUST return EXACTLY the same number of items in the same order.\n")
	sb.WriteString("Each input item has an index (0, 1, 2...). Preserve this index in your response.\n")
	sb.WriteString("Return JSON with exact shape: {\"translations\":[{\"index\":0,\"text\":\"translated\"},{\"index\":1,\"text\":\"translated\"}]}\n")
	sb.WriteString("Input items:\n")
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i, text))
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
	if translations, missing, err := tryDecodeIndexedTranslations(jsonBlock); err == nil {
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

// tryDecodeIndexedTranslations attempts to parse indexed response format
// Returns translations in order (by index), missing indexes if any, and error if parsing failed
func tryDecodeIndexedTranslations(jsonBlock string) ([]string, []int, error) {
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

	// Sort by index and reconstruct
	byIndex := make(map[int]string)
	maxIndex := -1
	for _, t := range indexed {
		byIndex[t.Index] = t.Text
		if t.Index > maxIndex {
			maxIndex = t.Index
		}
	}

	// Check for missing indexes (between 0 and maxIndex)
	var missing []int
	translations := make([]string, 0, len(indexed))
	for i := 0; i <= maxIndex; i++ {
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
