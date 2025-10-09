package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	apperrors "schools-be/internal/errors"
	"schools-be/internal/models"
	"schools-be/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

const maxRequestBodySize = 1 << 20 // 1MB

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

// GetSchools returns all schools
func (h *SchoolHandler) GetSchools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check for type filter
	schoolType := r.URL.Query().Get("type")

	var schools []models.School
	var err error

	if schoolType != "" {
		schools, err = h.service.GetSchoolsByType(ctx, schoolType)
	} else {
		schools, err = h.service.GetAllSchools(ctx)
	}

	if err != nil {
		h.logger.Error("failed to get schools",
			slog.String("type", schoolType),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve schools")
		return
	}

	h.respondJSON(w, http.StatusOK, schools)
}

// GetSchool returns a single school by ID
func (h *SchoolHandler) GetSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	school, err := h.service.GetSchoolByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "school not found")
			return
		}
		h.logger.Error("failed to get school",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve school")
		return
	}

	h.respondJSON(w, http.StatusOK, school)
}

// CreateSchool creates a new school
func (h *SchoolHandler) CreateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var input models.CreateSchoolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if err := h.validate.Struct(input); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		h.respondValidationError(w, err)
		return
	}

	school, err := h.service.CreateSchool(ctx, input)
	if err != nil {
		h.logger.Error("failed to create school",
			slog.String("name", input.Name),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to create school")
		return
	}

	h.respondJSON(w, http.StatusCreated, school)
}

// UpdateSchool updates an existing school
func (h *SchoolHandler) UpdateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var input models.UpdateSchoolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if err := h.validate.Struct(input); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		h.respondValidationError(w, err)
		return
	}

	school, err := h.service.UpdateSchool(ctx, id, input)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "school not found")
			return
		}
		h.logger.Error("failed to update school",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to update school")
		return
	}

	h.respondJSON(w, http.StatusOK, school)
}

// DeleteSchool deletes a school
func (h *SchoolHandler) DeleteSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	if err := h.service.DeleteSchool(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "school not found")
			return
		}
		h.logger.Error("failed to delete school",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to delete school")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "school deleted"})
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

func (h *SchoolHandler) respondValidationError(w http.ResponseWriter, err error) {
	validationErrors := make(map[string]string)

	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range verrs {
			field := e.Field()
			switch e.Tag() {
			case "required":
				validationErrors[field] = "This field is required"
			case "min":
				validationErrors[field] = "Value is too short"
			case "max":
				validationErrors[field] = "Value is too long"
			case "latitude":
				validationErrors[field] = "Invalid latitude value"
			case "longitude":
				validationErrors[field] = "Invalid longitude value"
			default:
				validationErrors[field] = "Invalid value"
			}
		}
	}

	h.respondJSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":  "validation failed",
		"fields": validationErrors,
	})
}
