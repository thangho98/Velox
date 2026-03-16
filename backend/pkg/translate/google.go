package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const googleTranslateURL = "https://translate.googleapis.com/translate_a/single"

// batchSeparator is used to join/split multiple texts in a single request.
// Using a unique separator that won't appear in subtitle text.
const batchSeparator = "\n|||VELOX_SEP|||\n"

// maxBatchChars is the max characters per Google Translate request (~5000 is safe).
const maxBatchChars = 4500

// GoogleTranslator translates text using the unofficial Google Translate API.
// No API key needed. Batches multiple texts per request for efficiency.
type GoogleTranslator struct {
	http *http.Client
}

// NewGoogle creates a Google Translate translator (unofficial, free).
func NewGoogle() *GoogleTranslator {
	return &GoogleTranslator{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GoogleTranslator) Name() string { return "google" }

func (g *GoogleTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	results := make([]string, len(texts))

	// Group texts into batches that fit within maxBatchChars
	i := 0
	for i < len(texts) {
		var batch []int // indices into texts
		charCount := 0

		for j := i; j < len(texts); j++ {
			addLen := len(texts[j]) + len(batchSeparator)
			if charCount+addLen > maxBatchChars && len(batch) > 0 {
				break
			}
			batch = append(batch, j)
			charCount += addLen
		}

		// Join batch texts with separator
		batchTexts := make([]string, len(batch))
		for k, idx := range batch {
			batchTexts[k] = texts[idx]
		}
		joined := strings.Join(batchTexts, batchSeparator)

		// Translate the batch
		translated, err := g.translateSingle(ctx, joined, targetLang)
		if err != nil {
			return nil, fmt.Errorf("translating batch at index %d: %w", i, err)
		}

		// Split back into individual results
		parts := strings.Split(translated, batchSeparator)
		// Google sometimes modifies the separator slightly, try fallback splits
		if len(parts) != len(batch) {
			parts = strings.Split(translated, "|||VELOX_SEP|||")
		}
		if len(parts) != len(batch) {
			// Last resort: split by any remaining separator fragment
			parts = splitFlexible(translated, len(batch))
		}

		for k, idx := range batch {
			if k < len(parts) {
				results[idx] = strings.TrimSpace(parts[k])
			} else {
				results[idx] = texts[idx] // fallback to original
			}
		}

		i += len(batch)
	}

	return results, nil
}

// splitFlexible attempts to split translated text into n parts by looking for
// separator remnants that Google Translate may have altered.
func splitFlexible(text string, n int) []string {
	// Try common variations Google might produce
	for _, sep := range []string{
		"|||velox_sep|||",
		"|||VELOX_SEP|||",
		"|||Velox_Sep|||",
		"||| VELOX_SEP |||",
	} {
		parts := strings.Split(text, sep)
		if len(parts) == n {
			return parts
		}
	}
	// Give up — return as single part
	return []string{text}
}

func (g *GoogleTranslator) translateSingle(ctx context.Context, text, targetLang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	params := url.Values{
		"client": {"gtx"},
		"sl":     {"auto"},
		"tl":     {targetLang},
		"dt":     {"t"},
		"q":      {text},
	}

	reqURL := googleTranslateURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := g.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("google translate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google translate: status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Response format: [[["translated text","source text",null,null,10]],null,"en",...]
	var raw []interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("google translate: decode: %w", err)
	}

	if len(raw) == 0 {
		return text, nil
	}

	sentences, ok := raw[0].([]interface{})
	if !ok {
		return text, nil
	}

	var sb strings.Builder
	for _, s := range sentences {
		parts, ok := s.([]interface{})
		if !ok || len(parts) == 0 {
			continue
		}
		if translated, ok := parts[0].(string); ok {
			sb.WriteString(translated)
		}
	}

	result := sb.String()
	if result == "" {
		return text, nil
	}
	return result, nil
}
