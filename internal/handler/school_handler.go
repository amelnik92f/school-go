package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"schools-be/internal/models"
	"schools-be/internal/service"

	"github.com/go-chi/chi/v5"
)

type SchoolHandler struct {
	service *service.SchoolService
}

func NewSchoolHandler(service *service.SchoolService) *SchoolHandler {
	return &SchoolHandler{service: service}
}

// GetSchools returns all schools
func (h *SchoolHandler) GetSchools(w http.ResponseWriter, r *http.Request) {
	// Check for type filter
	schoolType := r.URL.Query().Get("type")

	var schools []models.School
	var err error

	if schoolType != "" {
		schools, err = h.service.GetSchoolsByType(schoolType)
	} else {
		schools, err = h.service.GetAllSchools()
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, schools)
}

// GetSchool returns a single school by ID
func (h *SchoolHandler) GetSchool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	school, err := h.service.GetSchoolByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, school)
}

// CreateSchool creates a new school
func (h *SchoolHandler) CreateSchool(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSchoolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	school, err := h.service.CreateSchool(input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, school)
}

// UpdateSchool updates an existing school
func (h *SchoolHandler) UpdateSchool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	var input models.UpdateSchoolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	school, err := h.service.UpdateSchool(id, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, school)
}

// DeleteSchool deletes a school
func (h *SchoolHandler) DeleteSchool(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid school id")
		return
	}

	if err := h.service.DeleteSchool(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "school deleted"})
}

// RefreshData manually triggers a data refresh
func (h *SchoolHandler) RefreshData(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RefreshSchoolsData(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "data refresh completed"})
}

// Helper functions for JSON responses
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

