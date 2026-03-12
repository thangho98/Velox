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
	libs, err := h.svc.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, libs)
}

type createLibraryReq struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
	Type  string   `json:"type"`
}

func (h *LibraryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLibraryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Paths) == 0 {
		respondError(w, http.StatusBadRequest, "at least one path is required")
		return
	}

	// Verify every path exists and is a directory
	for _, p := range req.Paths {
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			respondError(w, http.StatusBadRequest, "path does not exist or is not a directory: "+p)
			return
		}
	}

	lib, err := h.svc.Create(r.Context(), req.Name, req.Type, req.Paths)
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
	if err := h.svc.Delete(r.Context(), id); err != nil {
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

	// Run scan in background, return job immediately
	job, err := h.svc.Scan(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Pipeline.Run blocks — run in goroutine would lose the job reference.
	// Instead, Pipeline creates the job synchronously (status=queued),
	// then runs stages. Caller gets the job ID to poll status.
	respondJSON(w, http.StatusAccepted, job)
}

func (h *LibraryHandler) ScanStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	jobs, err := h.svc.ScanJobs(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, jobs)
}
