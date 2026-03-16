package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

// WriteMovie writes a Movie NFO to the given path.
// If a file already exists, it's backed up to .nfo.bak before overwrite.
func WriteMovie(m *Movie, path string) error {
	return writeNFO(m, path)
}

// WriteTVShow writes a TVShow NFO to the given path.
func WriteTVShow(s *TVShow, path string) error {
	return writeNFO(s, path)
}

// WriteEpisode writes an Episode NFO to the given path.
func WriteEpisode(e *EpisodeNFO, path string) error {
	return writeNFO(e, path)
}

// EpisodeNFO is a dedicated write struct for episode NFO files.
// Separate from the read Episode struct to have cleaner xml output.
type EpisodeNFO struct {
	XMLName   xml.Name `xml:"episodedetails"`
	Title     string   `xml:"title"`
	ShowTitle string   `xml:"showtitle,omitempty"`
	Season    int      `xml:"season"`
	Episode   int      `xml:"episode"`
	Plot      string   `xml:"plot,omitempty"`
	Rating    float64  `xml:"rating,omitempty"`
	Aired     string   `xml:"aired,omitempty"`
	TMDbID    string   `xml:"tmdbid,omitempty"`
	IMDbID    string   `xml:"imdbid,omitempty"`
	Directors []string `xml:"director,omitempty"`
	Credits   []string `xml:"credits,omitempty"`
	Actors    []Actor  `xml:"actor,omitempty"`
}

// MovieNFOPath returns the NFO path for a movie (same dir, same basename).
func MovieNFOPath(videoPath string) string {
	ext := filepath.Ext(videoPath)
	return strings.TrimSuffix(videoPath, ext) + ".nfo"
}

// TVShowNFOPath returns the tvshow.nfo path for a series directory.
func TVShowNFOPath(seriesDir string) string {
	return filepath.Join(seriesDir, "tvshow.nfo")
}

// EpisodeNFOPath returns the NFO path for an episode (same dir, same basename).
func EpisodeNFOPath(videoPath string) string {
	return MovieNFOPath(videoPath) // Same convention
}

func writeNFO(v any, path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating nfo directory: %w", err)
	}

	// Backup existing file
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak"
		_ = os.Rename(path, backupPath)
	}

	// Marshal with indentation
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling nfo xml: %w", err)
	}

	// Write to temp file first for atomic operation
	tmpPath := path + ".tmp"
	content := []byte(xmlHeader)
	content = append(content, data...)
	content = append(content, '\n')

	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return fmt.Errorf("writing nfo temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming nfo file: %w", err)
	}

	return nil
}
