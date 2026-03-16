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

const (
	deeplFreeAPI = "https://api-free.deepl.com/v2/translate"
	deeplProAPI  = "https://api.deepl.com/v2/translate"
)

// DeepLTranslator translates text using the DeepL API.
type DeepLTranslator struct {
	apiKey string
	isPro  bool
	http   *http.Client
}

// NewDeepL creates a DeepL translator.
// Free API keys end with ":fx", pro keys don't.
func NewDeepL(apiKey string) *DeepLTranslator {
	isPro := !strings.HasSuffix(apiKey, ":fx")
	return &DeepLTranslator{
		apiKey: apiKey,
		isPro:  isPro,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *DeepLTranslator) Name() string { return "deepl" }

func (d *DeepLTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	apiURL := deeplFreeAPI
	if d.isPro {
		apiURL = deeplProAPI
	}

	// DeepL uses uppercase language codes: EN, DE, FR, VI, etc.
	// Some have variants: EN-US, EN-GB, PT-BR, PT-PT
	target := deeplLangCode(targetLang)

	// Build form data — DeepL supports multiple "text" parameters
	form := url.Values{}
	form.Set("target_lang", target)
	for _, t := range texts {
		form.Add("text", t)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "DeepL-Auth-Key "+d.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deepl request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 456 {
		return nil, fmt.Errorf("deepl: quota exceeded (456)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepl: status %d: %s", resp.StatusCode, string(body))
	}

	var result deeplResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("deepl: decode: %w", err)
	}

	translated := make([]string, len(result.Translations))
	for i, t := range result.Translations {
		translated[i] = t.Text
	}

	return translated, nil
}

type deeplResponse struct {
	Translations []deeplTranslation `json:"translations"`
}

type deeplTranslation struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}

// deeplLangCode converts ISO 639-1 to DeepL target language codes.
func deeplLangCode(lang string) string {
	m := map[string]string{
		"en": "EN", "vi": "VI", "fr": "FR", "de": "DE", "es": "ES",
		"pt": "PT-PT", "it": "IT", "nl": "NL", "sv": "SV", "no": "NB",
		"da": "DA", "fi": "FI", "ja": "JA", "ko": "KO", "zh": "ZH",
		"ar": "AR", "pl": "PL", "ru": "RU", "tr": "TR", "cs": "CS",
		"hu": "HU", "ro": "RO", "el": "EL", "id": "ID", "sk": "SK",
		"sl": "SL", "uk": "UK", "bg": "BG", "lt": "LT", "lv": "LV",
		"et": "ET",
	}
	if v, ok := m[strings.ToLower(lang)]; ok {
		return v
	}
	return strings.ToUpper(lang)
}
