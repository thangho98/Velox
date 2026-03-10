package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/thawng/velox/internal/service"
)

type LibraryHandler struct {
	svc *service.LibraryService
}

func NewLibraryHandler(svc *service.LibraryService) *LibraryHandler {
	return &LibraryHandler{svc: svc}
}

func (h *LibraryHandler) List(w http.ResponseWriter, r *http.Request) {
	libs, err := h.svc.List()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, libs)
}

type createLibraryReq struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (h *LibraryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLibraryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Path == "" {
		respondError(w, http.StatusBadRequest, "name and path are required")
		return
	}

	// Verify the path exists
	info, err := os.Stat(req.Path)
	if err != nil || !info.IsDir() {
		respondError(w, http.StatusBadRequest, "path does not exist or is not a directory")
		return
	}

	lib, err := h.svc.Create(req.Name, req.Path)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, lib)
}

func (h *LibraryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) Scan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Run scan in background
	go func() {
		if err := h.svc.Scan(id); err != nil {
			// Log error - in production use structured logger
			println("scan error:", err.Error())
		}
	}()

	respondJSON(w, http.StatusAccepted, map[string]string{"status": "scanning"})
}
