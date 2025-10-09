package service

import (
	"context"
	"testing"

	"schools-be/internal/errors"
	"schools-be/internal/fetcher"
	"schools-be/internal/models"
	"schools-be/internal/repository"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) (*SchoolService, *sqlx.DB) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create schema
	schema := `
		CREATE TABLE schools (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			address TEXT,
			type TEXT,
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
	service := NewSchoolService(repo, fetcher)

	return service, db
}

func TestSchoolService_CreateSchool(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()
	input := models.CreateSchoolInput{
		Name:      "Test School",
		Address:   "123 Test St",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}

	school, err := service.CreateSchool(ctx, input)
	require.NoError(t, err)
	assert.NotZero(t, school.ID)
	assert.Equal(t, input.Name, school.Name)
}

func TestSchoolService_GetSchoolByID(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Create a school
	input := models.CreateSchoolInput{
		Name:      "Test School",
		Address:   "123 Test St",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	created, err := service.CreateSchool(ctx, input)
	require.NoError(t, err)

	// Get the school
	school, err := service.GetSchoolByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, school.ID)
	assert.Equal(t, created.Name, school.Name)

	// Try to get non-existent school
	_, err = service.GetSchoolByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestSchoolService_GetAllSchools(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Initially empty
	schools, err := service.GetAllSchools(ctx)
	require.NoError(t, err)
	assert.Empty(t, schools)

	// Create schools
	for i := 1; i <= 3; i++ {
		input := models.CreateSchoolInput{
			Name:      "Test School",
			Address:   "Address",
			Type:      "Gymnasium",
			Latitude:  52.5200,
			Longitude: 13.4050,
		}
		_, err := service.CreateSchool(ctx, input)
		require.NoError(t, err)
	}

	// Get all
	schools, err = service.GetAllSchools(ctx)
	require.NoError(t, err)
	assert.Len(t, schools, 3)
}

func TestSchoolService_GetSchoolsByType(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Create schools of different types
	gymnasium := models.CreateSchoolInput{
		Name:      "Gymnasium",
		Address:   "Gym St",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	_, err := service.CreateSchool(ctx, gymnasium)
	require.NoError(t, err)

	grundschule := models.CreateSchoolInput{
		Name:      "Grundschule",
		Address:   "Grund St",
		Type:      "Grundschule",
		Latitude:  52.5167,
		Longitude: 13.3833,
	}
	_, err = service.CreateSchool(ctx, grundschule)
	require.NoError(t, err)

	// Get by type
	gymnasiums, err := service.GetSchoolsByType(ctx, "Gymnasium")
	require.NoError(t, err)
	assert.Len(t, gymnasiums, 1)

	grundschules, err := service.GetSchoolsByType(ctx, "Grundschule")
	require.NoError(t, err)
	assert.Len(t, grundschules, 1)
}

func TestSchoolService_UpdateSchool(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Create a school
	input := models.CreateSchoolInput{
		Name:      "Original Name",
		Address:   "Original Address",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	created, err := service.CreateSchool(ctx, input)
	require.NoError(t, err)

	// Update it
	newName := "Updated Name"
	updateInput := models.UpdateSchoolInput{
		Name: &newName,
	}

	updated, err := service.UpdateSchool(ctx, created.ID, updateInput)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, created.Address, updated.Address) // Unchanged

	// Try to update non-existent school
	_, err = service.UpdateSchool(ctx, 99999, updateInput)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestSchoolService_DeleteSchool(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Create a school
	input := models.CreateSchoolInput{
		Name:      "Test School",
		Address:   "123 Test St",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	created, err := service.CreateSchool(ctx, input)
	require.NoError(t, err)

	// Delete it
	err = service.DeleteSchool(ctx, created.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = service.GetSchoolByID(ctx, created.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)

	// Try to delete non-existent school
	err = service.DeleteSchool(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestSchoolService_RefreshSchoolsData(t *testing.T) {
	service, db := setupTestService(t)
	defer db.Close()

	ctx := context.Background()

	// Create some existing schools
	input := models.CreateSchoolInput{
		Name:      "Old School",
		Address:   "Old Address",
		Type:      "Gymnasium",
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	_, err := service.CreateSchool(ctx, input)
	require.NoError(t, err)

	// Refresh data (this will delete old and insert new from fetcher)
	err = service.RefreshSchoolsData(ctx)
	require.NoError(t, err)

	// Verify new data is there
	schools, err := service.GetAllSchools(ctx)
	require.NoError(t, err)
	// The fetcher has 2 example schools
	assert.Len(t, schools, 2)
}
