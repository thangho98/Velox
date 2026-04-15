package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/thawng/velox/internal/model"
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

func resolveTMDbSize(imageType, widthParam string) (string, bool) {
	buckets, ok := model.TMDbSizes[imageType]
	if !ok {
		return "", false
	}
	if widthParam == "original" {
		return "original", true
	}
	n, err := strconv.Atoi(widthParam)
	if err != nil || n <= 0 {
		n = model.TMDbDefaultWidth[imageType]
	}
	for _, b := range buckets {
		if n <= b {
			return fmt.Sprintf("w%d", b), true
		}
	}
	return "original", true
}

// Serve proxies a TMDb image.
// GET /api/images/tmdb/{type}/{path...}?width=N
//
// The backend picks the nearest TMDb size variant for the requested width and
// proxies the upstream file. `width` is optional; default is 500px for posters.
// Callers that want the full-resolution asset can pass a large value (e.g. width=9999)
// or "original" explicitly.
func (h *ImageHandler) Serve(w http.ResponseWriter, r *http.Request) {
	imgPath := r.PathValue("path")
	if imgPath == "" || containsDotDot(imgPath) {
		respondError(w, http.StatusBadRequest, "invalid image path")
		return
	}

	imageType := r.PathValue("type")
	size, ok := resolveTMDbSize(imageType, r.URL.Query().Get("width"))
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid image type")
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

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	io.Copy(w, resp.Body)
}

func containsDotDot(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '.' && s[i+1] == '.' {
			return true
		}
	}
	return false
}
