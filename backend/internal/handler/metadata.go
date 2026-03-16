package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/storage"
)

// MetadataHandler handles metadata identify/refresh/edit/image endpoints.
type MetadataHandler struct {
	mediaSvc    *service.MediaService
	metadataSvc *service.MetadataService
	imgStorage  *storage.ImageStorage
}

// NewMetadataHandler creates a new metadata handler. Returns nil if metadataSvc is nil.
func NewMetadataHandler(mediaSvc *service.MediaService, metadataSvc *service.MetadataService, imgStorage *storage.ImageStorage) *MetadataHandler {
	if metadataSvc == nil {
		return nil
	}
	return &MetadataHandler{mediaSvc: mediaSvc, metadataSvc: metadataSvc, imgStorage: imgStorage}
}

// identifyRequest is the body for PUT /api/media/{id}/identify
type identifyRequest struct {
	TmdbID    int    `json:"tmdb_id"`
	MediaType string `json:"media_type"` // "movie" or "tv"
}

// Identify overrides auto-match with a specific TMDb ID.
// PUT /api/media/{id}/identify
func (h *MetadataHandler) Identify(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req identifyRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.TmdbID <= 0 {
		respondError(w, http.StatusBadRequest, "tmdb_id is required")
		return
	}

	media, err := h.mediaSvc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.metadataSvc.IdentifyByTmdbID(r.Context(), media, req.TmdbID, req.MediaType); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Re-fetch updated media
	updated, _ := h.mediaSvc.Get(r.Context(), id)
	respondJSON(w, http.StatusOK, updated)
}

// Refresh re-fetches metadata from TMDb for a media item.
// POST /api/media/{id}/refresh
func (h *MetadataHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	media, err := h.mediaSvc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if media.TmdbID == nil {
		// No TMDb match yet — try auto-matching from file name
		if err := h.metadataSvc.AutoMatchAndRefresh(r.Context(), media); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := h.metadataSvc.RefreshMetadata(r.Context(), media); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	updated, _ := h.mediaSvc.Get(r.Context(), id)
	respondJSON(w, http.StatusOK, updated)
}

// BulkRefreshRatings auto-matches all unmatched media and fetches OMDb ratings.
// POST /api/admin/metadata/refresh-ratings
func (h *MetadataHandler) BulkRefreshRatings(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting bulk metadata refresh (TMDb match + OMDb ratings)...")
	updated, err := h.metadataSvc.BulkRefreshAllMetadata(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("Bulk metadata refresh complete: %d items updated", updated)
	respondJSON(w, http.StatusOK, map[string]int{"updated": updated})
}

// EditMediaMetadata partially updates metadata for a media item.
// PATCH /api/media/{id}/metadata
func (h *MetadataHandler) EditMediaMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.MetadataEditRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate: title must not be empty if provided
	if req.Title != nil && *req.Title == "" {
		respondError(w, http.StatusBadRequest, "title must not be empty")
		return
	}

	// Verify media exists
	if _, err := h.mediaSvc.Get(r.Context(), id); errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.metadataSvc.EditMediaMetadata(r.Context(), id, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, _ := h.mediaSvc.Get(r.Context(), id)
	respondJSON(w, http.StatusOK, updated)
}

// EditSeriesMetadata partially updates metadata for a series.
// PATCH /api/series/{id}/metadata
func (h *MetadataHandler) EditSeriesMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.SeriesMetadataEditRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title != nil && *req.Title == "" {
		respondError(w, http.StatusBadRequest, "title must not be empty")
		return
	}

	if err := h.metadataSvc.EditSeriesMetadata(r.Context(), id, req); errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "series not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// UnlockMediaMetadata removes the metadata lock for a media item.
// DELETE /api/media/{id}/metadata/lock
func (h *MetadataHandler) UnlockMediaMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.metadataSvc.UnlockMediaMetadata(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"metadata_locked": false})
}

// UnlockSeriesMetadata removes the metadata lock for a series.
// DELETE /api/series/{id}/metadata/lock
func (h *MetadataHandler) UnlockSeriesMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.metadataSvc.UnlockSeriesMetadata(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"metadata_locked": false})
}

const maxUploadSize = 10 << 20 // 10MB

// UploadMediaImage handles image upload for a media item.
// POST /api/media/{id}/images (multipart/form-data: image_type + file)
func (h *MetadataHandler) UploadMediaImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "media")
}

// UploadSeriesImage handles image upload for a series.
// POST /api/series/{id}/images
func (h *MetadataHandler) UploadSeriesImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "series")
}

func (h *MetadataHandler) uploadImage(w http.ResponseWriter, r *http.Request, entityType string) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Verify entity exists before processing upload
	if entityType == "media" {
		if _, err := h.mediaSvc.Get(r.Context(), id); errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "media not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if _, err := h.metadataSvc.GetSeries(r.Context(), id); errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "series not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	imageType := r.FormValue("image_type")
	if imageType != "poster" && imageType != "backdrop" {
		respondError(w, http.StatusBadRequest, "image_type must be 'poster' or 'backdrop'")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	localPath, err := h.imgStorage.Save(entityType, id, imageType, data)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update DB image path + auto-lock metadata
	if entityType == "media" {
		if err := h.metadataSvc.UpdateMediaImagePath(r.Context(), id, imageType, localPath); err != nil {
			// Clean up orphan file on DB failure
			_ = h.imgStorage.Delete(entityType, id, imageType)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := h.metadataSvc.UpdateSeriesImagePath(r.Context(), id, imageType, localPath); err != nil {
			_ = h.imgStorage.Delete(entityType, id, imageType)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"path": localPath, "image_type": imageType})
}

// DeleteMediaImage removes a custom uploaded image for a media item.
// DELETE /api/media/{id}/images/{imageType}
func (h *MetadataHandler) DeleteMediaImage(w http.ResponseWriter, r *http.Request) {
	h.deleteImage(w, r, "media")
}

// DeleteSeriesImage removes a custom uploaded image for a series.
// DELETE /api/series/{id}/images/{imageType}
func (h *MetadataHandler) DeleteSeriesImage(w http.ResponseWriter, r *http.Request) {
	h.deleteImage(w, r, "series")
}

func (h *MetadataHandler) deleteImage(w http.ResponseWriter, r *http.Request, entityType string) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	imageType := r.PathValue("imageType")
	if imageType != "poster" && imageType != "backdrop" {
		respondError(w, http.StatusBadRequest, "imageType must be 'poster' or 'backdrop'")
		return
	}

	if err := h.imgStorage.Delete(entityType, id, imageType); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if entityType == "media" {
		if err := h.metadataSvc.UpdateMediaImagePath(r.Context(), id, imageType, ""); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := h.metadataSvc.UpdateSeriesImagePath(r.Context(), id, imageType, ""); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ServeLocalImage serves a locally uploaded image.
// GET /api/images/local/{type}/{id}/{filename}
func (h *MetadataHandler) ServeLocalImage(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	if entityType != "media" && entityType != "series" {
		respondError(w, http.StatusBadRequest, "invalid type")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	filename := r.PathValue("filename")
	if filename == "" || strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
		respondError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	absPath := h.imgStorage.AbsPath(entityType, id, filename)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		respondError(w, http.StatusNotFound, "image not found")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=2592000") // 30 days
	http.ServeFile(w, r, absPath)
}

// WriteMediaNFO generates and writes an NFO file for a media item.
// POST /api/media/{id}/nfo
func (h *MetadataHandler) WriteMediaNFO(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.metadataSvc.WriteMediaNFO(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "nfo_written"})
}

// WriteSeriesNFO generates and writes a tvshow.nfo for a series.
// POST /api/series/{id}/nfo
func (h *MetadataHandler) WriteSeriesNFO(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.metadataSvc.WriteSeriesNFO(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "nfo_written"})
}
