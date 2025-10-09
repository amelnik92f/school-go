package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"schools-be/internal/fetcher"
	"schools-be/internal/models"
	"schools-be/internal/repository"
	"schools-be/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*SchoolHandler, *sqlx.DB) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create schema with all fields
	schema := `
		CREATE TABLE schools (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			school_type TEXT NOT NULL DEFAULT '',
			operator TEXT DEFAULT '',
			school_category TEXT DEFAULT '',
			district TEXT DEFAULT '',
			neighborhood TEXT DEFAULT '',
			postal_code TEXT DEFAULT '',
			street TEXT DEFAULT '',
			house_number TEXT DEFAULT '',
			phone TEXT DEFAULT '',
			fax TEXT DEFAULT '',
			email TEXT DEFAULT '',
			website TEXT DEFAULT '',
			school_year TEXT DEFAULT '',
			latitude REAL,
			longitude REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	repo := repository.NewSchoolRepository(db)
	fetcher := fetcher.NewSchoolFetcher()
	svc := service.NewSchoolService(repo, fetcher)
	handler := NewSchoolHandler(svc)

	return handler, db
}

func TestSchoolHandler_GetSchools(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test schools
	ctx := context.Background()
	repo := repository.NewSchoolRepository(db)
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Test School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	_, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	w := httptest.NewRecorder()

	handler.GetSchools(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var schools []models.School
	err = json.NewDecoder(w.Body).Decode(&schools)
	require.NoError(t, err)
	assert.Len(t, schools, 1)
	assert.Equal(t, "Test School", schools[0].Name)
}

func TestSchoolHandler_GetSchools_WithTypeFilter(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test schools of different types
	ctx := context.Background()
	repo := repository.NewSchoolRepository(db)

	gymnasium := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Gymnasium",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	_, err := repo.Create(ctx, gymnasium)
	require.NoError(t, err)

	grundschule := models.CreateSchoolInput{
		SchoolNumber: "01B02",
		Name:         "Grundschule",
		SchoolType:   "Grundschule",
		District:     "Mitte",
		Latitude:     52.5167,
		Longitude:    13.3833,
	}
	_, err = repo.Create(ctx, grundschule)
	require.NoError(t, err)

	// Test request with type filter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schools?type=Gymnasium", nil)
	w := httptest.NewRecorder()

	handler.GetSchools(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var schools []models.School
	err = json.NewDecoder(w.Body).Decode(&schools)
	require.NoError(t, err)
	assert.Len(t, schools, 1)
	assert.Equal(t, "Gymnasium", schools[0].Name)
}

func TestSchoolHandler_GetSchool(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test school
	ctx := context.Background()
	repo := repository.NewSchoolRepository(db)
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Test School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	school, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Test successful request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schools/1", nil)
	w := httptest.NewRecorder()

	// Add URL params using chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetSchool(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.School
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, school.ID, result.ID)
	assert.Equal(t, "Test School", result.Name)
}

func TestSchoolHandler_GetSchool_NotFound(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schools/99999", nil)
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetSchool(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSchoolHandler_GetSchool_InvalidID(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schools/invalid", nil)
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetSchool(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSchoolHandler_CreateSchool(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "New School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}

	body, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schools", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchool(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var result models.School
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "New School", result.Name)
}

func TestSchoolHandler_CreateSchool_ValidationError(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Invalid input: missing required fields
	input := models.CreateSchoolInput{
		Name: "", // Empty name should fail validation
	}

	body, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schools", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchool(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSchoolHandler_CreateSchool_InvalidJSON(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schools", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchool(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSchoolHandler_UpdateSchool(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test school
	ctx := context.Background()
	repo := repository.NewSchoolRepository(db)
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Original Name",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	school, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Update request
	newName := "Updated Name"
	updateInput := models.UpdateSchoolInput{
		Name: &newName,
	}

	body, err := json.Marshal(updateInput)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/schools/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.UpdateSchool(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.School
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, school.ID, result.ID)
	assert.Equal(t, "Updated Name", result.Name)
}

func TestSchoolHandler_UpdateSchool_NotFound(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	newName := "Updated Name"
	updateInput := models.UpdateSchoolInput{
		Name: &newName,
	}

	body, err := json.Marshal(updateInput)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/schools/99999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.UpdateSchool(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSchoolHandler_DeleteSchool(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test school
	ctx := context.Background()
	repo := repository.NewSchoolRepository(db)
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Test School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	_, err := repo.Create(ctx, input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/schools/1", nil)
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.DeleteSchool(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	_, err = repo.GetByID(ctx, 1)
	assert.Error(t, err)
}

func TestSchoolHandler_DeleteSchool_NotFound(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/schools/99999", nil)
	w := httptest.NewRecorder()

	// Add URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.DeleteSchool(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSchoolHandler_RefreshData(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	w := httptest.NewRecorder()

	handler.RefreshData(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]string
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "data refresh completed", result["message"])
}
