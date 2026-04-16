// Command srtchunk inspects how the subtitle chunker slices an SRT file.
// By default it runs TranslateSRT through a mock ContextualTranslator that
// records every call so we can see batch sizes, boundary kinds, and context
// overlap without exercising any real AI provider. With -real it calls the
// configured Anthropic-compatible endpoint and prints a few translated cues
// so quality can be spot-checked end-to-end.
//
// Usage:
//
//	srtchunk [-max N] [-v] <file.srt> [<file.srt> ...]
//	srtchunk -real -model MiniMax-M2.7 [-limit N] <file.srt>
//
// When -real is set the tool reads ANTHROPIC_AUTH_TOKEN (required) and
// ANTHROPIC_BASE_URL (optional) from the environment. -limit N keeps only
// the first N cues to bound API cost during smoke tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/thawng/velox/pkg/translate"
)

type recordedCall struct {
	size      int
	prior     int
	following int
	firstText string // used to recover chunker order after parallel dispatch
}

type captureTranslator struct {
	maxBatch int
	mu       sync.Mutex
	calls    []recordedCall
}

// wrappedTranslator forwards all calls to a real translator while still
// recording each one via the capture shim, so -real mode produces the same
// metrics as mock mode plus actual translated output.
type wrappedTranslator struct {
	inner translate.Translator
	cap   *captureTranslator
}

func (w *wrappedTranslator) Name() string      { return w.inner.Name() }
func (w *wrappedTranslator) MaxBatchSize() int { return w.cap.maxBatch }

func (w *wrappedTranslator) Translate(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	return w.TranslateWithContext(ctx, texts, nil, nil, targetLang)
}

func (w *wrappedTranslator) TranslateWithContext(ctx context.Context, texts, prior, following []string, targetLang string) ([]string, error) {
	first := ""
	if len(texts) > 0 {
		first = texts[0]
	}
	w.cap.mu.Lock()
	w.cap.calls = append(w.cap.calls, recordedCall{size: len(texts), prior: len(prior), following: len(following), firstText: first})
	w.cap.mu.Unlock()
	if ct, ok := w.inner.(translate.ContextualTranslator); ok {
		return ct.TranslateWithContext(ctx, texts, prior, following, targetLang)
	}
	return w.inner.Translate(ctx, texts, targetLang)
}

func trimTo(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (t *captureTranslator) Name() string      { return "capture" }
func (t *captureTranslator) MaxBatchSize() int { return t.maxBatch }

func (t *captureTranslator) Translate(_ context.Context, texts []string, _ string) ([]string, error) {
	return t.TranslateWithContext(context.Background(), texts, nil, nil, "vi")
}

func (t *captureTranslator) TranslateWithContext(_ context.Context, texts, prior, following []string, _ string) ([]string, error) {
	first := ""
	if len(texts) > 0 {
		first = texts[0]
	}
	t.mu.Lock()
	t.calls = append(t.calls, recordedCall{size: len(texts), prior: len(prior), following: len(following), firstText: first})
	t.mu.Unlock()
	out := make([]string, len(texts))
	for i, tx := range texts {
		out[i] = tx
	}
	return out, nil
}

func main() {
	maxBatch := flag.Int("max", 50, "max batch size for chunker")
	verbose := flag.Bool("v", false, "print per-batch detail")
	real := flag.Bool("real", false, "call real Anthropic-compatible API (reads ANTHROPIC_AUTH_TOKEN)")
	model := flag.String("model", "", "model name for -real (defaults to $ANTHROPIC_MODEL)")
	limit := flag.Int("limit", 0, "keep only the first N cues (0 = all) to bound API cost")
	lang := flag.String("lang", "vi", "target language")
	preview := flag.Int("preview", 6, "number of translated cues to print after -real")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: srtchunk [-max N] [-v] [-real [-model M] [-limit N]] <file.srt>...")
		os.Exit(1)
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("skip %s: %v\n\n", path, err)
			continue
		}

		cues := translate.ParseSRT(string(data))
		if *limit > 0 && len(cues) > *limit {
			cues = cues[:*limit]
			data = []byte(translate.BuildSRT(cues))
		}

		var translator translate.Translator
		cap := &captureTranslator{maxBatch: *maxBatch}
		if *real {
			token := strings.TrimSpace(os.Getenv("ANTHROPIC_AUTH_TOKEN"))
			if token == "" {
				token = strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
			}
			if token == "" {
				fmt.Fprintln(os.Stderr, "-real requires ANTHROPIC_AUTH_TOKEN or ANTHROPIC_API_KEY")
				os.Exit(1)
			}
			resolvedModel := *model
			if resolvedModel == "" {
				resolvedModel = strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL"))
			}
			if resolvedModel == "" {
				fmt.Fprintln(os.Stderr, "-real requires -model or ANTHROPIC_MODEL")
				os.Exit(1)
			}
			trans, err := translate.NewAI(translate.AIConfig{
				Provider: translate.ProviderAnthropicCompatible,
				APIKey:   token,
				BaseURL:  strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")),
				Model:    resolvedModel,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "build translator: %v\n", err)
				os.Exit(1)
			}
			translator = &wrappedTranslator{inner: trans, cap: cap}
		} else {
			translator = cap
		}

		translated, err := translate.TranslateSRT(context.Background(), translator, string(data), *lang)
		if err != nil {
			fmt.Printf("%s: translate error: %v\n\n", path, err)
			continue
		}

		// Restore chunker order: goroutines record in completion order, so we
		// re-sort calls by the position of each firstText in the original cue
		// stream. This gives the -v output a stable, meaningful ordering.
		pos := make(map[string]int, len(cues))
		for i, c := range cues {
			if _, exists := pos[c.Text]; !exists {
				pos[c.Text] = i
			}
		}
		sort.SliceStable(cap.calls, func(i, j int) bool {
			return pos[cap.calls[i].firstText] < pos[cap.calls[j].firstText]
		})

		totalCalls := len(cap.calls)
		overlapCalls := 0
		totalSize := 0
		overheadCues := 0
		minSize, maxSize := 1<<30, 0
		for _, c := range cap.calls {
			if c.prior > 0 || c.following > 0 {
				overlapCalls++
			}
			totalSize += c.size
			overheadCues += c.prior + c.following
			if c.size < minSize {
				minSize = c.size
			}
			if c.size > maxSize {
				maxSize = c.size
			}
		}
		avg := 0.0
		if totalCalls > 0 {
			avg = float64(totalSize) / float64(totalCalls)
		}
		overlapPct := 0.0
		if totalCalls > 0 {
			overlapPct = 100 * float64(overlapCalls) / float64(totalCalls)
		}
		overheadPct := 0.0
		if totalSize > 0 {
			overheadPct = 100 * float64(overheadCues) / float64(totalSize)
		}

		fmt.Printf("=== %s ===\n", path)
		fmt.Printf("  Cues:               %d\n", len(cues))
		fmt.Printf("  Batches:            %d\n", totalCalls)
		fmt.Printf("  Batch size:         min=%d avg=%.1f max=%d\n", minSize, avg, maxSize)
		fmt.Printf("  Batches w/ overlap: %d / %d (%.1f%%)\n", overlapCalls, totalCalls, overlapPct)
		fmt.Printf("  Context overhead:   %d cues (%.1f%% of total input)\n", overheadCues, overheadPct)

		if *verbose {
			fmt.Println()
			for i, c := range cap.calls {
				marker := "  "
				switch {
				case c.prior > 0 && c.following > 0:
					marker = "◀▶"
				case c.prior > 0:
					marker = "◀ "
				case c.following > 0:
					marker = " ▶"
				}
				fmt.Printf("    batch %2d: size=%2d  %s  prior=%d following=%d\n", i, c.size, marker, c.prior, c.following)
			}
		}

		if *real && *preview > 0 {
			fmt.Println()
			fmt.Printf("  Preview (%s, first %d cues):\n", *lang, *preview)
			translatedCues := translate.ParseSRT(translated)
			n := *preview
			if n > len(translatedCues) {
				n = len(translatedCues)
			}
			for i := 0; i < n; i++ {
				orig := strings.ReplaceAll(cues[i].Text, "\n", " / ")
				trans := strings.ReplaceAll(translatedCues[i].Text, "\n", " / ")
				fmt.Printf("    [%d] %s\n        -> %s\n", i, trimTo(orig, 90), trimTo(trans, 90))
			}
		}
		fmt.Println()
	}
}
