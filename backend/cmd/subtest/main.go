// subtest is a CLI tool to test subtitle providers individually.
// Usage: go run ./cmd/subtest -provider=subdl -query="Inception" -lang=en
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/thawng/velox/pkg/bsplayer"
	"github.com/thawng/velox/pkg/podnapisi"
	"github.com/thawng/velox/pkg/subdl"
	"github.com/thawng/velox/pkg/subprovider"
	"github.com/thawng/velox/pkg/subscene"
)

func main() {
	provider := flag.String("provider", "all", "Provider to test: subdl, podnapisi, bsplayer, subscene, all")
	query := flag.String("query", "Inception", "Search query (movie/series title)")
	lang := flag.String("lang", "en", "Language code (ISO 639-1)")
	imdbID := flag.String("imdb", "", "IMDB ID for BSPlayer (e.g. tt1375666)")
	season := flag.Int("season", 0, "Season number for Subscene TV show search")
	timeout := flag.Int("timeout", 60, "Timeout in seconds")

	// Batch mode flags
	batch := flag.Bool("batch", false, "Batch auto-download subtitles for all media in DB")
	dbPath := flag.String("db", "", "Path to velox.db (required for -batch)")
	subDir := flag.String("subdir", "", "Directory to save downloaded subtitles (required for -batch)")
	dryRun := flag.Bool("dry-run", false, "Only show what would be downloaded (use with -batch)")
	titleFilter := flag.String("title", "", "Filter media by title (case-insensitive, use with -batch)")
	limitN := flag.Int("limit", 0, "Limit number of media to process (use with -batch)")
	flag.Parse()

	// Batch mode
	if *batch {
		if *dbPath == "" || *subDir == "" {
			fmt.Println("Usage: subtest -batch -db /path/to/velox.db -subdir /path/to/subtitles [-dry-run] [-lang en,vi] [-title friends] [-limit 5]")
			os.Exit(1)
		}
		langs := strings.Split(*lang, ",")
		batchAutoDownload(*dbPath, *subDir, langs, *dryRun, *titleFilter, *limitN)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	fmt.Printf("=== Subtitle Provider Test ===\n")
	fmt.Printf("Query:    %s\n", *query)
	fmt.Printf("Language: %s\n", *lang)
	fmt.Printf("Provider: %s\n", *provider)
	fmt.Println()

	providers := []string{*provider}
	if *provider == "all" {
		providers = []string{"subdl", "podnapisi", "bsplayer", "subscene"}
	}

	for _, p := range providers {
		fmt.Printf("--- %s ---\n", strings.ToUpper(p))
		start := time.Now()

		var results []subprovider.Result
		var err error

		switch p {
		case "subdl":
			results, err = testSubdl(ctx, *query, *lang)
		case "podnapisi":
			results, err = testPodnapisi(ctx, *query, *lang)
		case "bsplayer":
			results, err = testBSPlayer(ctx, *imdbID, *lang)
		case "subscene":
			results, err = testSubscene(ctx, *query, *lang, *season)
		default:
			fmt.Printf("Unknown provider: %s\n", p)
			continue
		}

		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("Found %d results (took %s)\n", len(results), elapsed.Round(time.Millisecond))
			for i, r := range results {
				if i >= 10 {
					fmt.Printf("  ... and %d more\n", len(results)-10)
					break
				}
				fmt.Printf("  [%d] %-12s | %-4s | %-6s | DL:%-5d | %s\n",
					i+1, r.Provider, r.Language, r.Format, r.Downloads, truncate(r.Title, 60))
				if r.ExternalID != "" {
					fmt.Printf("      ID: %s\n", truncate(r.ExternalID, 80))
				}
			}
		}
		fmt.Println()
	}
}

func testSubdl(ctx context.Context, query, lang string) ([]subprovider.Result, error) {
	apiKey := os.Getenv("VELOX_SUBDL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SUBDL_API_KEY not set")
	}
	client := subdl.New(apiKey)
	return client.Search(ctx, subdl.SearchParams{
		FilmName: query,
		Language: lang,
	})
}

func testPodnapisi(ctx context.Context, query, lang string) ([]subprovider.Result, error) {
	client := podnapisi.New()
	return client.Search(ctx, podnapisi.SearchParams{
		Keywords: query,
		Language: lang,
	})
}

func testBSPlayer(ctx context.Context, imdbID, lang string) ([]subprovider.Result, error) {
	if imdbID == "" {
		return nil, fmt.Errorf("BSPlayer requires -imdb flag (e.g. -imdb=tt1375666)")
	}
	client := bsplayer.New()
	return client.Search(ctx, bsplayer.SearchParams{
		ImdbID:   imdbID,
		Language: lang,
	})
}

func testSubscene(ctx context.Context, query, lang string, season int) ([]subprovider.Result, error) {
	scraper := subscene.New()
	if depErr := subscene.CheckDeps(); depErr != "" {
		return nil, fmt.Errorf("%s", depErr)
	}
	log.Println("Subscene uses DrissionPage (Python) — this may take 15-30s...")
	return scraper.Search(ctx, subscene.SearchParams{
		Query:    query,
		Language: lang,
		Season:   season,
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	log.SetOutput(os.Stderr)
}
