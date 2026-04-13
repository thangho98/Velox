package handler

import (
	"net/http"
	"time"

	"github.com/thawng/velox/internal/service"
)

// SchedulerHandler handles scheduled task endpoints.
type SchedulerHandler struct {
	scheduler *service.Scheduler
}

func NewSchedulerHandler(scheduler *service.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{scheduler: scheduler}
}

// ListTasks returns all registered tasks with their status.
// GET /api/admin/tasks
func (h *SchedulerHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.scheduler.ListTasks()
	respondJSON(w, http.StatusOK, tasks)
}

// RunTask triggers a task to run immediately.
// POST /api/admin/tasks/{name}/run
func (h *SchedulerHandler) RunTask(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "task name is required")
		return
	}

	if err := h.scheduler.RunNow(name); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "task": name})
}

// UpdateTask updates a scheduled task's settings.
// PATCH /api/admin/tasks/{name}
func (h *SchedulerHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "task name is required")
		return
	}

	var req struct {
		Interval string `json:"interval"`
	}
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Interval == "" {
		respondError(w, http.StatusBadRequest, "interval is required")
		return
	}

	dur, err := time.ParseDuration(req.Interval)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid duration format (e.g. 24h, 30m)")
		return
	}

	if err := h.scheduler.UpdateInterval(r.Context(), name, dur); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated", "task": name})
}
