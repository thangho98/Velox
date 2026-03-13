package handler

import (
	"net/http"

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
