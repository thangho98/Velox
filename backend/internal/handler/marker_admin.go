package handler

import (
	"encoding/json"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// MarkerAdminHandler provides admin endpoints for marker management
type MarkerAdminHandler struct {
	markerSvc *service.MarkerService
}

// NewMarkerAdminHandler creates a new marker admin handler
func NewMarkerAdminHandler(markerSvc *service.MarkerService) *MarkerAdminHandler {
	return &MarkerAdminHandler{markerSvc: markerSvc}
}

// BackfillMarkersRequest represents a request to backfill markers
type BackfillMarkersRequest struct {
	FileIDs   []int64 `json:"file_ids,omitempty"`   // Specific file IDs to process
	SeasonID  int64   `json:"season_id,omitempty"`  // Process all episodes in a season
	LibraryID int64   `json:"library_id,omitempty"` // Process entire library
}

// BackfillMarkersResponse represents the result of a backfill operation
type BackfillMarkersResponse struct {
	Processed int      `json:"processed"`        // Number of files processed
	Skipped   int      `json:"skipped"`          // Number of files skipped (already have markers)
	Errors    []string `json:"errors,omitempty"` // Any errors encountered
}

// DetectRequest represents a request to run a specific detector
type DetectRequest struct {
	FileID       int64  `json:"file_id"`       // Required: file to detect on
	DetectorName string `json:"detector_name"` // Required: detector to use (e.g., "fingerprint")
}

// DetectResponse represents the result of a detection operation
type DetectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// BackfillMarkers runs fingerprint detection on files without existing markers
// POST /api/admin/markers/backfill
func (h *MarkerAdminHandler) BackfillMarkers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BackfillMarkersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var processed, skipped int
	var err error

	if len(req.FileIDs) > 0 {
		processed, skipped, err = h.markerSvc.BackfillMarkers(ctx, req.FileIDs)
	} else if req.SeasonID > 0 {
		processed, skipped, err = h.markerSvc.DetectSeason(ctx, req.SeasonID)
	} else {
		respondError(w, http.StatusBadRequest, "file_ids or season_id required")
		return
	}

	resp := BackfillMarkersResponse{
		Processed: processed,
		Skipped:   skipped,
	}

	if err != nil {
		resp.Errors = append(resp.Errors, err.Error())
	}

	respondJSON(w, http.StatusOK, resp)
}

// DetectWithDetector runs a specific detector on a media file
// POST /api/admin/markers/detect
func (h *MarkerAdminHandler) DetectWithDetector(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FileID <= 0 {
		respondError(w, http.StatusBadRequest, "file_id required")
		return
	}
	if req.DetectorName == "" {
		respondError(w, http.StatusBadRequest, "detector_name required")
		return
	}

	// Run detection
	if err := h.markerSvc.DetectWithDetector(ctx, req.FileID, req.DetectorName); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, DetectResponse{
		Success: true,
		Message: "detection completed",
	})
}

// ListDetectors returns available marker detectors
// GET /api/admin/markers/detectors
func (h *MarkerAdminHandler) ListDetectors(w http.ResponseWriter, r *http.Request) {
	detectors := h.markerSvc.GetAvailableDetectors()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"detectors": detectors,
	})
}
