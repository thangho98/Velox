package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const tmdbImageBase = "https://image.tmdb.org/t/p/"

// ImageHandler proxies TMDb images with caching.
type ImageHandler struct {
	client *http.Client
}

// NewImageHandler creates a new image handler.
func NewImageHandler() *ImageHandler {
	return &ImageHandler{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Serve proxies a TMDb image.
// GET /api/images/tmdb/{size}/{path...}
// Example: /api/images/tmdb/w500/abc123.jpg
func (h *ImageHandler) Serve(w http.ResponseWriter, r *http.Request) {
	size := r.PathValue("size")
	imgPath := r.PathValue("path")

	// Validate size — only allow known TMDb sizes
	validSizes := map[string]bool{
		"w92": true, "w154": true, "w185": true, "w342": true, "w500": true, "w780": true,
		"original": true,
		// Backdrop sizes
		"w300": true, "w1280": true,
		// Profile sizes
		"w45": true, "w138": true, "h632": true,
	}
	if !validSizes[size] {
		respondError(w, http.StatusBadRequest, "invalid image size")
		return
	}

	// Validate path — must start with / and be a simple filename
	if imgPath == "" || strings.Contains(imgPath, "..") {
		respondError(w, http.StatusBadRequest, "invalid image path")
		return
	}

	url := tmdbImageBase + size + "/" + imgPath

	resp, err := h.client.Get(url)
	if err != nil {
		respondError(w, http.StatusBadGateway, "failed to fetch image")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondError(w, resp.StatusCode, fmt.Sprintf("upstream error: %d", resp.StatusCode))
		return
	}

	// Forward content type
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	// Cache for 7 days (images rarely change)
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")

	io.Copy(w, resp.Body)
}
