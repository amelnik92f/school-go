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
	"github.com/go-playground/validator/v10"
)

type SchoolHandler struct {
	service  *service.SchoolService
	validate *validator.Validate
	logger   *slog.Logger
}

func NewSchoolHandler(service *service.SchoolService) *SchoolHandler {
	return &SchoolHandler{
		service:  service,
		validate: validator.New(),
		logger:   slog.Default(),
	}
}

// RefreshData manually triggers a data refresh
func (h *SchoolHandler) RefreshData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.service.RefreshSchoolsData(ctx); err != nil {
		h.logger.Error("failed to refresh data", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "failed to refresh data")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "data refresh completed"})
}

// GetSchoolsEnriched returns all schools with enriched data from all related tables
func (h *SchoolHandler) GetSchoolsEnriched(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	schools, err := h.service.GetAllSchoolsEnriched(ctx)
	if err != nil {
		h.logger.Error("failed to get enriched schools", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve enriched schools")
		return
	}

	h.respondJSON(w, http.StatusOK, schools)
}

// GetSchoolEnriched returns a single enriched school by ID
func (h *SchoolHandler) GetSchoolEnriched(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	school, err := h.service.GetSchoolByIDEnriched(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "school not found")
			return
		}
		h.logger.Error("failed to get enriched school",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve enriched school")
		return
	}

	h.respondJSON(w, http.StatusOK, school)
}

// Helper functions for JSON responses
func (h *SchoolHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *SchoolHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
