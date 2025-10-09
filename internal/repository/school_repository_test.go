package repository

import (
	"context"
	"testing"
	"time"

	"schools-be/internal/errors"
	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
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

	return db
}

func TestSchoolRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	input := models.CreateSchoolInput{
		SchoolNumber:   "01B01",
		Name:           "Test School",
		SchoolType:     "Gymnasium",
		Operator:       "öffentlich",
		SchoolCategory: "Gymnasium",
		District:       "Mitte",
		Neighborhood:   "Wedding",
		PostalCode:     "13347",
		Street:         "Musterstraße",
		HouseNumber:    "123",
		Phone:          "030-12345678",
		Fax:            "030-12345679",
		Email:          "test@school.de",
		Website:        "https://test-school.de",
		SchoolYear:     "2025/26",
		Latitude:       52.5200,
		Longitude:      13.4050,
	}

	school, err := repo.Create(ctx, input)
	require.NoError(t, err)
	assert.NotZero(t, school.ID)
	assert.Equal(t, input.SchoolNumber, school.SchoolNumber)
	assert.Equal(t, input.Name, school.Name)
	assert.Equal(t, input.SchoolType, school.SchoolType)
	assert.Equal(t, input.District, school.District)
	assert.Equal(t, input.Email, school.Email)
	assert.Equal(t, input.Latitude, school.Latitude)
	assert.Equal(t, input.Longitude, school.Longitude)
	assert.False(t, school.CreatedAt.IsZero())
	assert.False(t, school.UpdatedAt.IsZero())
}

func TestSchoolRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Create a school first
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Test School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Test successful retrieval
	school, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, school.ID)
	assert.Equal(t, created.Name, school.Name)

	// Test not found
	_, err = repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestSchoolRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Initially empty
	schools, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, schools)

	// Create multiple schools
	for i := 1; i <= 3; i++ {
		input := models.CreateSchoolInput{
			SchoolNumber: "01B0" + string(rune('0'+i)),
			Name:         "Test School " + string(rune('A'+i-1)),
			SchoolType:   "Gymnasium",
			District:     "Mitte",
			Latitude:     52.5200,
			Longitude:    13.4050,
		}
		_, err := repo.Create(ctx, input)
		require.NoError(t, err)
	}

	// Get all schools
	schools, err = repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, schools, 3)
}

func TestSchoolRepository_GetByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Create schools of different types
	gymnasium := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Gymnasium School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	_, err := repo.Create(ctx, gymnasium)
	require.NoError(t, err)

	grundschule := models.CreateSchoolInput{
		SchoolNumber: "01B02",
		Name:         "Grundschule School",
		SchoolType:   "Grundschule",
		District:     "Mitte",
		Latitude:     52.5167,
		Longitude:    13.3833,
	}
	_, err = repo.Create(ctx, grundschule)
	require.NoError(t, err)

	// Get by type
	gymnasiums, err := repo.GetByType(ctx, "Gymnasium")
	require.NoError(t, err)
	assert.Len(t, gymnasiums, 1)
	assert.Equal(t, "Gymnasium School", gymnasiums[0].Name)

	grundschules, err := repo.GetByType(ctx, "Grundschule")
	require.NoError(t, err)
	assert.Len(t, grundschules, 1)
	assert.Equal(t, "Grundschule School", grundschules[0].Name)
}

func TestSchoolRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Create a school
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Original Name",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Email:        "original@school.de",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Wait a bit to ensure updated_at changes
	time.Sleep(10 * time.Millisecond)

	// Update the school
	newName := "Updated Name"
	newDistrict := "Charlottenburg"
	newEmail := "updated@school.de"
	updateInput := models.UpdateSchoolInput{
		Name:     &newName,
		District: &newDistrict,
		Email:    &newEmail,
	}

	updated, err := repo.Update(ctx, created.ID, updateInput)
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDistrict, updated.District)
	assert.Equal(t, newEmail, updated.Email)
	assert.Equal(t, created.SchoolType, updated.SchoolType) // Unchanged
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt))
}

func TestSchoolRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Create a school
	input := models.CreateSchoolInput{
		SchoolNumber: "01B01",
		Name:         "Test School",
		SchoolType:   "Gymnasium",
		District:     "Mitte",
		Latitude:     52.5200,
		Longitude:    13.4050,
	}
	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Delete it
	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = repo.GetByID(ctx, created.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestSchoolRepository_DeleteAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)
	ctx := context.Background()

	// Create multiple schools
	for i := 1; i <= 3; i++ {
		input := models.CreateSchoolInput{
			SchoolNumber: "01B0" + string(rune('0'+i)),
			Name:         "Test School",
			SchoolType:   "Gymnasium",
			District:     "Mitte",
			Latitude:     52.5200,
			Longitude:    13.4050,
		}
		_, err := repo.Create(ctx, input)
		require.NoError(t, err)
	}

	// Verify they exist
	schools, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, schools, 3)

	// Delete all
	err = repo.DeleteAll(ctx)
	require.NoError(t, err)

	// Verify all deleted
	schools, err = repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, schools)
}

func TestSchoolRepository_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSchoolRepository(db)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to query with cancelled context
	_, err := repo.GetAll(ctx)
	assert.Error(t, err)
}
