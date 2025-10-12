package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	apperrors "schools-be/internal/errors"
	"schools-be/internal/service"

	"github.com/go-chi/chi/v5"
)

type ConstructionProjectHandler struct {
	service *service.ConstructionProjectService
	logger  *slog.Logger
}

func NewConstructionProjectHandler(service *service.ConstructionProjectService) *ConstructionProjectHandler {
	return &ConstructionProjectHandler{
		service: service,
		logger:  slog.Default(),
	}
}

// GetAll returns all construction projects
func (h *ConstructionProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.service.GetAll(ctx)
	if err != nil {
		h.logger.Error("failed to get construction projects", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve construction projects")
		return
	}

	h.respondJSON(w, http.StatusOK, projects)
}

// GetByID returns a single construction project by ID
func (h *ConstructionProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "construction project not found")
			return
		}
		h.logger.Error("failed to get construction project",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve construction project")
		return
	}

	h.respondJSON(w, http.StatusOK, project)
}

// GetStandalone returns valid construction projects that are not assigned to any existing school
// Only includes orphaned projects with meaningful data (excludes meta entries and legends)
func (h *ConstructionProjectHandler) GetStandalone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.service.GetStandalone(ctx)
	if err != nil {
		h.logger.Error("failed to get standalone construction projects", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve standalone construction projects")
		return
	}

	h.logger.Info("retrieved standalone construction projects", slog.Int("count", len(projects)))
	h.respondJSON(w, http.StatusOK, projects)
}

// respondJSON sends a JSON response
func (h *ConstructionProjectHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

// respondError sends an error JSON response
func (h *ConstructionProjectHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
