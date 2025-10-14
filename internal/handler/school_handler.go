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
	service       *service.SchoolService
	aiService     *service.AIService
	routesService *service.RoutesService
	validate      *validator.Validate
	logger        *slog.Logger
}

func NewSchoolHandler(service *service.SchoolService, aiService *service.AIService, routesService *service.RoutesService) *SchoolHandler {
	return &SchoolHandler{
		service:       service,
		aiService:     aiService,
		routesService: routesService,
		validate:      validator.New(),
		logger:        slog.Default(),
	}
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

// GetSchoolSummary generates an AI summary for a school
func (h *SchoolHandler) GetSchoolSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	// Check if AI service is available
	if h.aiService == nil {
		h.respondError(w, http.StatusServiceUnavailable, "AI service is not available")
		return
	}

	// Get enriched school data
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
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve school data")
		return
	}

	// Generate summary using AI
	summary, err := h.aiService.GenerateSchoolSummary(ctx, school)
	if err != nil {
		h.logger.Error("failed to generate school summary",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to generate summary")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"summary":    summary,
		"schoolName": school.School.Name,
	})
}

// CalculateRoutes calculates travel times from a location to a school
func (h *SchoolHandler) CalculateRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	// Check if routes service is available
	if h.routesService == nil {
		h.respondError(w, http.StatusServiceUnavailable, "routes service is not available")
		return
	}

	// Parse request body
	var req service.TravelTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if len(req.Start) != 2 || len(req.End) != 2 {
		h.respondError(w, http.StatusBadRequest, "invalid coordinates format")
		return
	}

	if len(req.Modes) == 0 {
		h.respondError(w, http.StatusBadRequest, "at least one travel mode is required")
		return
	}

	// Calculate travel times
	results, err := h.routesService.CalculateTravelTimes(ctx, req)
	if err != nil {
		h.logger.Error("failed to calculate travel times",
			slog.Int64("schoolId", id),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to calculate travel times")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
	})
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
