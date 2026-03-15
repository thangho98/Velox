package subscene

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/thawng/velox/pkg/subprovider"
)

// Scraper is a Subscene subtitle scraper that delegates to a Python script
// using DrissionPage for Cloudflare bypass. Designed for background use only.
type Scraper struct {
	mu         sync.Mutex
	scriptPath string // path to subscene_search.py
	pythonPath string // path to venv python
}

// SearchParams configures a Subscene search.
type SearchParams struct {
	Query    string // movie/series title
	Language string // ISO 639-1 ("en", "vi")
	Season   int    // season number for TV shows (0 = movie)
}

// New creates a new Subscene scraper. Automatically locates the Python script
// relative to the running binary or via VELOX_SUBSCENE_SCRIPT env var.
func New() *Scraper {
	s := &Scraper{}
	s.scriptPath, s.pythonPath = locatePaths()
	return s
}

// Available returns true if the Python script and venv are properly set up.
func (s *Scraper) Available() bool {
	if s.scriptPath == "" || s.pythonPath == "" {
		return false
	}
	if _, err := os.Stat(s.scriptPath); err != nil {
		return false
	}
	if _, err := os.Stat(s.pythonPath); err != nil {
		return false
	}
	return true
}

// Search scrapes Subscene for subtitles by calling the Python script.
func (s *Scraper) Search(ctx context.Context, params SearchParams) ([]subprovider.Result, error) {
	if !s.Available() {
		return nil, fmt.Errorf("subscene: python script or venv not found (script=%s, python=%s)", s.scriptPath, s.pythonPath)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("subscene: searching %q lang=%s season=%d via python script", params.Query, params.Language, params.Season)

	args := []string{s.scriptPath, "search", "--query", params.Query, "--lang", params.Language}
	if params.Season > 0 {
		args = append(args, "--season", fmt.Sprintf("%d", params.Season))
	}
	cmd := exec.CommandContext(ctx, s.pythonPath, args...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("subscene: python script failed: %w", err)
	}

	// Check if output is an error object
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(output, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("subscene: %s", errResp.Error)
	}

	var results []subprovider.Result
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("subscene: parsing results: %w (output: %s)", err, truncate(string(output), 200))
	}

	log.Printf("subscene: found %d subtitles", len(results))
	return results, nil
}

// Download fetches a subtitle by calling the Python script.
func (s *Scraper) Download(ctx context.Context, downloadURL string) ([]byte, string, error) {
	if !s.Available() {
		return nil, "", fmt.Errorf("subscene: python script or venv not found")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create temp file for the download
	tmpDir, err := os.MkdirTemp("", "subscene-dl-*")
	if err != nil {
		return nil, "", fmt.Errorf("subscene: creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "subtitle.srt")

	cmd := exec.CommandContext(ctx, s.pythonPath, s.scriptPath,
		"download", "--url", downloadURL, "--output", outputPath)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("subscene: download script failed: %w", err)
	}

	var resp struct {
		File         string `json:"file"`
		OriginalName string `json:"original_name"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, "", fmt.Errorf("subscene: parsing download response: %w", err)
	}
	if resp.Error != "" {
		return nil, "", fmt.Errorf("subscene: %s", resp.Error)
	}

	data, err := os.ReadFile(resp.File)
	if err != nil {
		return nil, "", fmt.Errorf("subscene: reading downloaded file: %w", err)
	}

	filename := resp.OriginalName
	if filename == "" {
		filename = "subtitle.srt"
	}

	return data, filename, nil
}

// locatePaths finds the Python script and venv python binary.
func locatePaths() (scriptPath, pythonPath string) {
	// Check env var override
	if envScript := os.Getenv("VELOX_SUBSCENE_SCRIPT"); envScript != "" {
		scriptPath = envScript
	}

	// Try relative to executable
	if scriptPath == "" {
		if exe, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "..", "scripts", "subscene_search.py")
			if _, err := os.Stat(candidate); err == nil {
				scriptPath = candidate
			}
		}
	}

	// Try relative to working directory
	if scriptPath == "" {
		candidates := []string{
			"scripts/subscene_search.py",
			"backend/scripts/subscene_search.py",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				scriptPath = c
				break
			}
		}
	}

	if scriptPath != "" {
		scriptPath, _ = filepath.Abs(scriptPath)
	}

	// Find venv python
	if scriptPath != "" {
		venvPython := filepath.Join(filepath.Dir(scriptPath), ".venv", "bin", "python3")
		if _, err := os.Stat(venvPython); err == nil {
			pythonPath = venvPython
		}
	}

	// Fallback to system python
	if pythonPath == "" {
		if p, err := exec.LookPath("python3"); err == nil {
			pythonPath = p
		}
	}

	return scriptPath, pythonPath
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CheckDeps verifies that Python dependencies are installed.
// Returns a user-friendly message if something is missing.
func CheckDeps() string {
	s := New()
	if s.scriptPath == "" {
		return "subscene: script not found (scripts/subscene_search.py)"
	}
	if s.pythonPath == "" {
		return "subscene: python3 not found"
	}

	// Check DrissionPage is importable
	cmd := exec.Command(s.pythonPath, "-c", "from DrissionPage import ChromiumPage; print('ok')")
	if out, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(msg, "ModuleNotFoundError") {
			return fmt.Sprintf("subscene: DrissionPage not installed. Run: %s -m pip install DrissionPage", s.pythonPath)
		}
		return fmt.Sprintf("subscene: dependency check failed: %s", msg)
	}

	return ""
}
