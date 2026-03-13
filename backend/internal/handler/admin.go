package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// AdminHandler handles admin dashboard endpoints.
type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
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
