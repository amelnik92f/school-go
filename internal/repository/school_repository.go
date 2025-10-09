package repository

import (
	"database/sql"
	"fmt"
	"time"

	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
)

type SchoolRepository struct {
	db *sqlx.DB
}

func NewSchoolRepository(db *sqlx.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) GetAll() ([]models.School, error) {
	var schools []models.School
	query := `SELECT * FROM schools ORDER BY created_at DESC`

	err := r.db.Select(&schools, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schools: %w", err)
	}

	return schools, nil
}

func (r *SchoolRepository) GetByID(id int64) (*models.School, error) {
	var school models.School
	query := `SELECT * FROM schools WHERE id = ?`

	err := r.db.Get(&school, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get school: %w", err)
	}

	return &school, nil
}

func (r *SchoolRepository) GetByType(schoolType string) ([]models.School, error) {
	var schools []models.School
	query := `SELECT * FROM schools WHERE type = ? ORDER BY name`

	err := r.db.Select(&schools, query, schoolType)
	if err != nil {
		return nil, fmt.Errorf("failed to get schools by type: %w", err)
	}

	return schools, nil
}

func (r *SchoolRepository) Create(input models.CreateSchoolInput) (*models.School, error) {
	query := `
		INSERT INTO schools (name, address, type, latitude, longitude, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, input.Name, input.Address, input.Type,
		input.Latitude, input.Longitude, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create school: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(id)
}

func (r *SchoolRepository) Update(id int64, input models.UpdateSchoolInput) (*models.School, error) {
	// Build dynamic update query
	query := `UPDATE schools SET updated_at = ?`
	args := []interface{}{time.Now()}

	if input.Name != nil {
		query += `, name = ?`
		args = append(args, *input.Name)
	}
	if input.Address != nil {
		query += `, address = ?`
		args = append(args, *input.Address)
	}
	if input.Type != nil {
		query += `, type = ?`
		args = append(args, *input.Type)
	}
	if input.Latitude != nil {
		query += `, latitude = ?`
		args = append(args, *input.Latitude)
	}
	if input.Longitude != nil {
		query += `, longitude = ?`
		args = append(args, *input.Longitude)
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update school: %w", err)
	}

	return r.GetByID(id)
}

func (r *SchoolRepository) Delete(id int64) error {
	query := `DELETE FROM schools WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete school: %w", err)
	}

	return nil
}

func (r *SchoolRepository) DeleteAll() error {
	query := `DELETE FROM schools`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all schools: %w", err)
	}

	return nil
}

