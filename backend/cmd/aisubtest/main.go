package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thawng/velox/pkg/translate"
)

func main() {
	provider := flag.String("provider", "", "AI provider: openai_compatible, gemini_compatible, anthropic_compatible")
	apiKey := flag.String("api-key", "", "Provider API key (or set VELOX_AI_API_KEY)")
	baseURL := flag.String("base-url", "", "Provider base URL (optional)")
	model := flag.String("model", "", "Model name")
	inputPath := flag.String("file", "cmd/aisubtest/testdata/example.srt", "Path to input .srt file")
	targetLang := flag.String("target-lang", "vi", "Target language code")
	timeoutSeconds := flag.Int("timeout", 180, "Request timeout in seconds")
	batchSize := flag.Int("batch-size", 0, "Optional override batch size for debugging")
	outputPath := flag.String("output", "", "Optional output file path for translated .srt")
	title := flag.String("title", "", "Optional media title context for the AI prompt")
	mediaType := flag.String("media-type", "", "Optional media type context, e.g. movie or episode")
	genres := flag.String("genres", "", "Optional comma-separated genres context")
	overview := flag.String("overview", "", "Optional short overview context")
	tagline := flag.String("tagline", "", "Optional tagline context")
	flag.Parse()

	if err := run(
		*provider,
		*apiKey,
		*baseURL,
		*model,
		*inputPath,
		*targetLang,
		*timeoutSeconds,
		*batchSize,
		*outputPath,
		translate.AIContext{
			Title:     *title,
			MediaType: *mediaType,
			Genres:    splitCSV(*genres),
			Overview:  *overview,
			Tagline:   *tagline,
		},
	); err != nil {
		log.Fatalf("aisubtest failed: %v", err)
	}
}

func run(
	provider,
	apiKey,
	baseURL,
	model,
	inputPath,
	targetLang string,
	timeoutSeconds,
	batchSize int,
	outputPath string,
	mediaContext translate.AIContext,
) error {
	if strings.TrimSpace(provider) == "" {
		return errors.New("provider is required")
	}
	if strings.TrimSpace(model) == "" {
		return errors.New("model is required")
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("VELOX_AI_API_KEY"))
	}
	if apiKey == "" {
		return errors.New("api key is required via -api-key or VELOX_AI_API_KEY")
	}

	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	translator, err := translate.NewAI(translate.AIConfig{
		Provider: translate.AIProvider(strings.TrimSpace(provider)),
		APIKey:   apiKey,
		BaseURL:  strings.TrimSpace(baseURL),
		Model:    strings.TrimSpace(model),
		Context:  optionalContext(mediaContext),
	})
	if err != nil {
		return err
	}

	cues := translate.ParseSRT(string(content))
	if len(cues) == 0 {
		return errors.New("no subtitle cues found in input file")
	}

	if batchSize <= 0 {
		if sized, ok := translator.(interface{ MaxBatchSize() int }); ok && sized.MaxBatchSize() > 0 {
			batchSize = sized.MaxBatchSize()
		} else {
			batchSize = 50
		}
	}

	if outputPath == "" {
		outputPath = defaultOutputPath(inputPath, targetLang)
	}

	fmt.Printf("=== AI Subtitle Test ===\n")
	fmt.Printf("Provider:    %s\n", translator.Name())
	fmt.Printf("Model:       %s\n", model)
	fmt.Printf("Input:       %s\n", inputPath)
	fmt.Printf("Target lang: %s\n", targetLang)
	fmt.Printf("Total cues:  %d\n", len(cues))
	fmt.Printf("Batch size:  %d\n", batchSize)
	fmt.Printf("Output:      %s\n\n", outputPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	startedAt := time.Now()

	for i := 0; i < len(cues); i += batchSize {
		end := i + batchSize
		if end > len(cues) {
			end = len(cues)
		}

		texts := make([]string, end-i)
		for j := i; j < end; j++ {
			texts[j-i] = cues[j].Text
		}

		batchStart := time.Now()
		fmt.Printf("Batch %d-%d ... ", i, end)

		translated, err := translator.Translate(ctx, texts, targetLang)
		if err != nil {
			fmt.Printf("FAILED (%s)\n", time.Since(batchStart).Round(time.Millisecond))
			fmt.Printf("Error: %v\n\n", err)
			printBatchPreview(cues[i:end])
			return err
		}

		for j, text := range translated {
			cues[i+j].Text = text
		}

		fmt.Printf("OK (%s)\n", time.Since(batchStart).Round(time.Millisecond))
	}

	translatedSRT := translate.BuildSRT(cues)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(translatedSRT), 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	fmt.Printf("\nDone in %s\n", time.Since(startedAt).Round(time.Millisecond))
	fmt.Printf("Translated subtitle written to %s\n", outputPath)
	return nil
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func optionalContext(ctx translate.AIContext) *translate.AIContext {
	if strings.TrimSpace(ctx.Title) == "" &&
		strings.TrimSpace(ctx.MediaType) == "" &&
		strings.TrimSpace(ctx.Overview) == "" &&
		strings.TrimSpace(ctx.Tagline) == "" &&
		len(ctx.Genres) == 0 {
		return nil
	}
	return &ctx
}

func printBatchPreview(cues []translate.SRTCue) {
	fmt.Println("Failing batch preview:")
	for i, cue := range cues {
		text := strings.ReplaceAll(cue.Text, "\n", " / ")
		if len(text) > 90 {
			text = text[:87] + "..."
		}
		fmt.Printf("  [%d] #%s %s\n", i, cue.Index, text)
	}
}

func defaultOutputPath(inputPath, targetLang string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	dir := filepath.Dir(inputPath)
	return filepath.Join(dir, base+"."+targetLang+".translated"+ext)
}
