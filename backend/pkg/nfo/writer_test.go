package nfo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMovie(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{
		Title:     "Test Movie",
		Year:      2024,
		Plot:      "A test movie plot",
		Tagline:   "Test tagline",
		Rating:    8.5,
		Premiered: "2024-01-15",
		Genres:    []string{"Action", "Drama"},
		Directors: []string{"Director 1"},
		Credits:   []string{"Writer 1"},
		IMDbID:    "tt1234567",
		TMDbID:    "12345",
	}

	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	// Verify file exists and has content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<title>Test Movie</title>")
	assert.Contains(t, string(data), "<year>2024</year>")
	assert.Contains(t, string(data), "<plot>A test movie plot</plot>")
	assert.Contains(t, string(data), "<tagline>Test tagline</tagline>")
	assert.Contains(t, string(data), "<rating>8.5</rating>")
	assert.Contains(t, string(data), "<genre>Action</genre>")
	assert.Contains(t, string(data), "<genre>Drama</genre>")
	assert.Contains(t, string(data), "<director>Director 1</director>")
	assert.Contains(t, string(data), "<credits>Writer 1</credits>")
	assert.Contains(t, string(data), "<imdbid>tt1234567</imdbid>")
	assert.Contains(t, string(data), "<tmdbid>12345</tmdbid>")
}

func TestWriteMovie_BackupExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	// Create initial file
	err := os.WriteFile(outputPath, []byte("old content"), 0644)
	require.NoError(t, err)

	// Write new movie
	movie := &Movie{Title: "New Movie"}
	err = WriteMovie(movie, outputPath)
	require.NoError(t, err)

	// Verify .bak was created with old content
	bakPath := outputPath + ".bak"
	data, err := os.ReadFile(bakPath)
	require.NoError(t, err)
	assert.Equal(t, "old content", string(data))

	// Verify new file has new content
	newData, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(newData), "<title>New Movie</title>")
}

func TestWriteMovie_XMLHeader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{Title: "Test"}
	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "<?xml version=\"1.0\" encoding=\"UTF-8\""))
}

func TestWriteTVShow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "tvshow.nfo")

	tvshow := &TVShow{
		Title:     "Test TV Show",
		Year:      2024,
		Plot:      "A test TV show",
		Status:    "Returning Series",
		Premiered: "2024-01-15",
		Genres:    []string{"Drama"},
		Studios:   []string{"Netflix"},
		IMDbID:    "tt1234567",
		TMDbID:    "12345",
	}

	err := WriteTVShow(tvshow, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<title>Test TV Show</title>")
	assert.Contains(t, string(data), "<plot>A test TV show</plot>")
	assert.Contains(t, string(data), "<status>Returning Series</status>")
	assert.Contains(t, string(data), "<premiered>2024-01-15</premiered>")
	assert.Contains(t, string(data), "<genre>Drama</genre>")
	assert.Contains(t, string(data), "<studio>Netflix</studio>")
}

func TestWriteTVShow_BackupExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "tvshow.nfo")

	// Create initial file
	err := os.WriteFile(outputPath, []byte("old content"), 0644)
	require.NoError(t, err)

	// Write new TV show
	tvshow := &TVShow{Title: "New Show"}
	err = WriteTVShow(tvshow, outputPath)
	require.NoError(t, err)

	// Verify .bak was created
	bakPath := outputPath + ".bak"
	_, err = os.Stat(bakPath)
	assert.NoError(t, err)
}

func TestWriteEpisode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "episode.nfo")

	episode := &EpisodeNFO{
		Title:     "Episode Title",
		ShowTitle: "Test Show",
		Season:    1,
		Episode:   5,
		Plot:      "Episode plot",
		Rating:    9.0,
		Aired:     "2024-01-15",
		TMDbID:    "12345",
		Directors: []string{"Director 1"},
		Credits:   []string{"Writer 1"},
	}

	err := WriteEpisode(episode, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<title>Episode Title</title>")
	assert.Contains(t, string(data), "<showtitle>Test Show</showtitle>")
	assert.Contains(t, string(data), "<season>1</season>")
	assert.Contains(t, string(data), "<episode>5</episode>")
	assert.Contains(t, string(data), "<plot>Episode plot</plot>")
	assert.Contains(t, string(data), "<rating>9</rating>")
	assert.Contains(t, string(data), "<aired>2024-01-15</aired>")
	assert.Contains(t, string(data), "<tmdbid>12345</tmdbid>")
}

func TestWriteEpisode_WithActors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "episode.nfo")

	episode := &EpisodeNFO{
		Title:   "Episode Title",
		Season:  1,
		Episode: 1,
		Actors: []Actor{
			{Name: "Actor One", Role: "Character 1", Order: 0},
			{Name: "Actor Two", Role: "Character 2", Order: 1},
		},
	}

	err := WriteEpisode(episode, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<name>Actor One</name>")
	assert.Contains(t, string(data), "<role>Character 1</role>")
	assert.Contains(t, string(data), "<name>Actor Two</name>")
	assert.Contains(t, string(data), "<role>Character 2</role>")
}

func TestWriteEpisode_BackupExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "episode.nfo")

	// Create initial file
	err := os.WriteFile(outputPath, []byte("old content"), 0644)
	require.NoError(t, err)

	// Write new episode
	episode := &EpisodeNFO{Title: "New Episode", Season: 1, Episode: 1}
	err = WriteEpisode(episode, outputPath)
	require.NoError(t, err)

	// Verify .bak was created
	bakPath := outputPath + ".bak"
	_, err = os.Stat(bakPath)
	assert.NoError(t, err)
}

func TestWriteMovie_CreatesDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "subdir", "movie.nfo")

	movie := &Movie{Title: "Test"}
	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	_, err = os.Stat(outputPath)
	assert.NoError(t, err)
}

func TestWriteMovie_AtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{Title: "Atomic Test"}
	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	// File should exist and be complete
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<title>Atomic Test</title>")

	// Temp file should be cleaned up
	_, err = os.Stat(outputPath + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestMovieNFOPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		videoPath string
		expected  string
	}{
		{"/movies/Movie.mkv", "/movies/Movie.nfo"},
		{"/movies/Movie.mp4", "/movies/Movie.nfo"},
		{"/movies/Movie.avi", "/movies/Movie.nfo"},
		{"/movies/Movie (2024).mkv", "/movies/Movie (2024).nfo"},
	}

	for _, tt := range tests {
		got := MovieNFOPath(tt.videoPath)
		assert.Equal(t, tt.expected, got)
	}
}

func TestTVShowNFOPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		seriesDir string
		expected  string
	}{
		{"/tv/Show", "/tv/Show/tvshow.nfo"},
		{"/tv/Show/", "/tv/Show/tvshow.nfo"},
	}

	for _, tt := range tests {
		got := TVShowNFOPath(tt.seriesDir)
		assert.Equal(t, tt.expected, got)
	}
}

func TestEpisodeNFOPath(t *testing.T) {
	t.Parallel()

	videoPath := "/tv/Show/S01E01.mkv"
	expected := "/tv/Show/S01E01.nfo"

	got := EpisodeNFOPath(videoPath)
	assert.Equal(t, expected, got)
}

func TestWriteMovie_WithFanart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{
		Title:  "Movie with Fanart",
		Fanart: []Fanart{{URL: "https://example.com/fanart.jpg"}},
	}

	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "https://example.com/fanart.jpg")
}

func TestWriteMovie_WithActors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{
		Title: "Movie with Cast",
		Actors: []Actor{
			{Name: "Actor One", Role: "Hero", Order: 0, Thumb: "https://example.com/actor1.jpg"},
			{Name: "Actor Two", Role: "Villain", Order: 1},
		},
	}

	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<name>Actor One</name>")
	assert.Contains(t, string(data), "<role>Hero</role>")
	assert.Contains(t, string(data), "<thumb>https://example.com/actor1.jpg</thumb>")
	assert.Contains(t, string(data), "<name>Actor Two</name>")
	assert.Contains(t, string(data), "<role>Villain</role>")
}

func TestWriteTVShow_WithFanart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "tvshow.nfo")

	tvshow := &TVShow{
		Title:  "TV Show with Fanart",
		Fanart: []Fanart{{URL: "https://example.com/backdrop.jpg"}},
	}

	err := WriteTVShow(tvshow, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "https://example.com/backdrop.jpg")
}

func TestWriteTVShow_NoBackupIfFileDoesNotExist(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "tvshow.nfo")

	tvshow := &TVShow{Title: "New Show"}
	err := WriteTVShow(tvshow, outputPath)
	require.NoError(t, err)

	// No .bak file should exist for new files
	bakPath := outputPath + ".bak"
	_, err = os.Stat(bakPath)
	assert.True(t, os.IsNotExist(err))
}

func TestWriteMovie_NoBackupIfFileDoesNotExist(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "movie.nfo")

	movie := &Movie{Title: "New Movie"}
	err := WriteMovie(movie, outputPath)
	require.NoError(t, err)

	// No .bak file should exist for new files
	bakPath := outputPath + ".bak"
	_, err = os.Stat(bakPath)
	assert.True(t, os.IsNotExist(err))
}
