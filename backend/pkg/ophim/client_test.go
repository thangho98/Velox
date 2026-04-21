package ophim

import (
	"context"
	"testing"
)

func TestGetRecentMovies(t *testing.T) {
	client := New()
	ctx := context.Background()

	res, err := client.GetRecentMovies(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get recent movies: %v", err)
	}

	if len(res.Data.Items) == 0 {
		t.Fatal("expected more than 0 items")
	}

	t.Logf("Found %d movies on page 1", len(res.Data.Items))
	t.Logf("First movie: %s (slug: %s)", res.Data.Items[0].Name, res.Data.Items[0].Slug)
}

func TestGetMovieDetails(t *testing.T) {
	client := New()
	ctx := context.Background()

	// Using a known slug
	slug := "cuoc-chien-trong-chung-ta"
	res, err := client.GetMovieDetails(ctx, slug)
	if err != nil {
		t.Fatalf("failed to get movie details: %v", err)
	}

	if res.Movie.Name == "" {
		t.Fatal("expected movie name")
	}

	t.Logf("Movie: %s (%d)", res.Movie.Name, res.Movie.Year)

	if len(res.Episodes) > 0 && len(res.Episodes[0].ServerData) > 0 {
		t.Logf("First episode M3U8 link: %s", res.Episodes[0].ServerData[0].LinkM3U8)
	} else {
		t.Fatal("expected at least one episode and stream link")
	}
}
