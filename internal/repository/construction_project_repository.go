package repository

import (
	"context"
	"time"

	"schools-be/internal/errors"
	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
)

type ConstructionProjectRepository struct {
	db *sqlx.DB
}

func NewConstructionProjectRepository(db *sqlx.DB) *ConstructionProjectRepository {
	return &ConstructionProjectRepository{db: db}
}

// Create creates a new construction project
func (r *ConstructionProjectRepository) Create(ctx context.Context, input models.CreateConstructionProjectInput) (*models.ConstructionProject, error) {
	query := `
		INSERT INTO construction_projects (
			project_id, school_number, school_name, district, school_type,
			construction_measure, description, built_school_places, places_after_construction,
			class_tracks_after_construction, handover_date, total_costs, street,
			postal_code, city, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		input.ProjectID, input.SchoolNumber, input.SchoolName, input.District, input.SchoolType,
		input.ConstructionMeasure, input.Description, input.BuiltSchoolPlaces, input.PlacesAfterConstruction,
		input.ClassTracksAfterConstruction, input.HandoverDate, input.TotalCosts, input.Street,
		input.PostalCode, input.City, now, now)
	if err != nil {
		return nil, errors.NewDatabaseError("create construction project", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.NewDatabaseError("get last insert id", err)
	}

	return r.GetByID(ctx, id)
}

// GetByID retrieves a construction project by ID
func (r *ConstructionProjectRepository) GetByID(ctx context.Context, id int64) (*models.ConstructionProject, error) {
	var project models.ConstructionProject
	query := `SELECT * FROM construction_projects WHERE id = ?`

	err := r.db.GetContext(ctx, &project, query, id)
	if err != nil {
		return nil, errors.NewDatabaseError("get construction project by id", err)
	}

	return &project, nil
}

// GetAll retrieves all construction projects
func (r *ConstructionProjectRepository) GetAll(ctx context.Context) ([]models.ConstructionProject, error) {
	var projects []models.ConstructionProject
	query := `SELECT * FROM construction_projects ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query)
	if err != nil {
		return nil, errors.NewDatabaseError("get all construction projects", err)
	}

	return projects, nil
}

// GetBySchoolNumber retrieves construction projects for a specific school
func (r *ConstructionProjectRepository) GetBySchoolNumber(ctx context.Context, schoolNumber string) ([]models.ConstructionProject, error) {
	var projects []models.ConstructionProject
	query := `SELECT * FROM construction_projects WHERE school_number = ? ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get construction projects by school number", err)
	}

	return projects, nil
}

// DeleteAll deletes all construction projects
func (r *ConstructionProjectRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM construction_projects`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.NewDatabaseError("delete all construction projects", err)
	}
	return nil
}
