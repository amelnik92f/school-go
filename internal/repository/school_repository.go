package repository

import (
	"context"
	"database/sql"
	"time"

	"schools-be/internal/errors"
	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
)

type SchoolRepository struct {
	db *sqlx.DB
}

func NewSchoolRepository(db *sqlx.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) GetAll(ctx context.Context) ([]models.School, error) {
	var schools []models.School
	query := `SELECT * FROM schools ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &schools, query)
	if err != nil {
		return nil, errors.NewDatabaseError("get all schools", err)
	}

	return schools, nil
}

func (r *SchoolRepository) GetByID(ctx context.Context, id int64) (*models.School, error) {
	var school models.School
	query := `SELECT * FROM schools WHERE id = ?`

	err := r.db.GetContext(ctx, &school, query, id)
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("school", id)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("get school by id", err)
	}

	return &school, nil
}

func (r *SchoolRepository) GetByType(ctx context.Context, schoolType string) ([]models.School, error) {
	var schools []models.School
	query := `SELECT * FROM schools WHERE school_type = ? ORDER BY name`

	err := r.db.SelectContext(ctx, &schools, query, schoolType)
	if err != nil {
		return nil, errors.NewDatabaseError("get schools by type", err)
	}

	return schools, nil
}

func (r *SchoolRepository) Create(ctx context.Context, input models.CreateSchoolInput) (*models.School, error) {
	query := `
		INSERT INTO schools (
			school_number, name, school_type, operator, school_category,
			district, neighborhood, postal_code, street, house_number,
			phone, fax, email, website, school_year,
			latitude, longitude, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		input.SchoolNumber, input.Name, input.SchoolType, input.Operator, input.SchoolCategory,
		input.District, input.Neighborhood, input.PostalCode, input.Street, input.HouseNumber,
		input.Phone, input.Fax, input.Email, input.Website, input.SchoolYear,
		input.Latitude, input.Longitude, now, now)
	if err != nil {
		return nil, errors.NewDatabaseError("create school", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.NewDatabaseError("get last insert id", err)
	}

	return r.GetByID(ctx, id)
}

func (r *SchoolRepository) Update(ctx context.Context, id int64, input models.UpdateSchoolInput) (*models.School, error) {
	// Build dynamic update query
	query := `UPDATE schools SET updated_at = ?`
	args := []interface{}{time.Now()}

	if input.SchoolNumber != nil {
		query += `, school_number = ?`
		args = append(args, *input.SchoolNumber)
	}
	if input.Name != nil {
		query += `, name = ?`
		args = append(args, *input.Name)
	}
	if input.SchoolType != nil {
		query += `, school_type = ?`
		args = append(args, *input.SchoolType)
	}
	if input.Operator != nil {
		query += `, operator = ?`
		args = append(args, *input.Operator)
	}
	if input.SchoolCategory != nil {
		query += `, school_category = ?`
		args = append(args, *input.SchoolCategory)
	}
	if input.District != nil {
		query += `, district = ?`
		args = append(args, *input.District)
	}
	if input.Neighborhood != nil {
		query += `, neighborhood = ?`
		args = append(args, *input.Neighborhood)
	}
	if input.PostalCode != nil {
		query += `, postal_code = ?`
		args = append(args, *input.PostalCode)
	}
	if input.Street != nil {
		query += `, street = ?`
		args = append(args, *input.Street)
	}
	if input.HouseNumber != nil {
		query += `, house_number = ?`
		args = append(args, *input.HouseNumber)
	}
	if input.Phone != nil {
		query += `, phone = ?`
		args = append(args, *input.Phone)
	}
	if input.Fax != nil {
		query += `, fax = ?`
		args = append(args, *input.Fax)
	}
	if input.Email != nil {
		query += `, email = ?`
		args = append(args, *input.Email)
	}
	if input.Website != nil {
		query += `, website = ?`
		args = append(args, *input.Website)
	}
	if input.SchoolYear != nil {
		query += `, school_year = ?`
		args = append(args, *input.SchoolYear)
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

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError("update school", err)
	}

	return r.GetByID(ctx, id)
}

func (r *SchoolRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM schools WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("delete school", err)
	}

	return nil
}

func (r *SchoolRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM schools`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.NewDatabaseError("delete all schools", err)
	}

	return nil
}
