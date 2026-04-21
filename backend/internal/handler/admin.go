package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/service"
)

// AdminHandler handles admin dashboard endpoints.
type AdminHandler struct {
	svc          *service.AdminService
	ophimScanner *scanner.OphimScanner
	libraryRepo  *repository.LibraryRepo
}

func NewAdminHandler(svc *service.AdminService, ophimScanner *scanner.OphimScanner, libraryRepo *repository.LibraryRepo) *AdminHandler {
	return &AdminHandler{svc: svc, ophimScanner: ophimScanner, libraryRepo: libraryRepo}
}

// ServerInfo returns server status information.
// GET /api/admin/server
func (h *AdminHandler) ServerInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.GetServerInfo(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, info)
}

// LibraryStats returns per-library statistics.
// GET /api/admin/stats/libraries
func (h *AdminHandler) LibraryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetLibraryStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

// SyncOphim triggers a manual sync of movies from Ophim
// POST /api/admin/ophim/sync?start=1&end=5&library_id=1
func (h *AdminHandler) SyncOphim(w http.ResponseWriter, r *http.Request) {
	if h.ophimScanner == nil {
		respondError(w, http.StatusNotImplemented, "OphimScanner not available")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	libraryIDStr := r.URL.Query().Get("library_id")

	start, end := 1, 5 // default to 5 pages
	if startStr != "" {
		if s, err := strconv.Atoi(startStr); err == nil && s > 0 {
			start = s
		}
	}
	if endStr != "" {
		if e, err := strconv.Atoi(endStr); err == nil && e >= start {
			end = e
		}
	}

	var libraryID int64 = 1
	if libraryIDStr != "" {
		if id, err := strconv.ParseInt(libraryIDStr, 10, 64); err == nil {
			libraryID = id
		}
	} else if h.libraryRepo != nil {
		if libs, err := h.libraryRepo.List(r.Context()); err == nil && len(libs) > 0 {
			libraryID = libs[0].ID
			for _, lib := range libs {
				if strings.Contains(strings.ToLower(lib.Name), "ophim") {
					libraryID = lib.ID
					break
				}
			}
		}
	}

	added, err := h.ophimScanner.SyncRange(r.Context(), libraryID, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "added_items": added, "synced_pages": end - start + 1})
}
