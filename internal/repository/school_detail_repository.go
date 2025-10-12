package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"schools-be/internal/errors"
	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
)

type SchoolDetailRepository struct {
	db *sqlx.DB
}

func NewSchoolDetailRepository(db *sqlx.DB) *SchoolDetailRepository {
	return &SchoolDetailRepository{db: db}
}

// Create creates a new school detail record
func (r *SchoolDetailRepository) Create(ctx context.Context, detail *models.SchoolDetailData) error {
	query := `
		INSERT INTO school_details (
			school_number, school_name, languages, courses, offerings,
			available_after_4th_grade, additional_info,
			equipment, working_groups, partners, differentiation, lunch_info, dual_learning,
			citizenship_data, language_data, residence_data, absence_data,
			scraped_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	// Convert tables to JSON
	citizenshipJSON, _ := r.tableToJSON(detail.CitizenshipTable)
	languageJSON, _ := r.tableToJSON(detail.LanguageTable)
	residenceJSON, _ := r.tableToJSON(detail.ResidenceTable)
	absenceJSON, _ := r.tableToJSON(detail.AbsenceTable)

	_, err := r.db.ExecContext(ctx, query,
		detail.SchoolNumber,
		detail.SchoolName,
		detail.Languages,
		detail.Courses,
		detail.Offerings,
		detail.AvailableAfter4thGrade,
		detail.AdditionalInfo,
		detail.Equipment,
		detail.WorkingGroups,
		detail.Partners,
		detail.Differentiation,
		detail.LunchInfo,
		detail.DualLearning,
		citizenshipJSON,
		languageJSON,
		residenceJSON,
		absenceJSON,
		detail.ScrapedAt,
		now,
		now,
	)

	if err != nil {
		return errors.NewDatabaseError("create school detail", err)
	}

	return nil
}

// Upsert inserts or updates a school detail record
func (r *SchoolDetailRepository) Upsert(ctx context.Context, detail *models.SchoolDetailData) error {
	query := `
		INSERT INTO school_details (
			school_number, school_name, languages, courses, offerings,
			available_after_4th_grade, additional_info,
			equipment, working_groups, partners, differentiation, lunch_info, dual_learning,
			citizenship_data, language_data, residence_data, absence_data,
			scraped_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(school_number) DO UPDATE SET
			school_name = excluded.school_name,
			languages = excluded.languages,
			courses = excluded.courses,
			offerings = excluded.offerings,
			available_after_4th_grade = excluded.available_after_4th_grade,
			additional_info = excluded.additional_info,
			equipment = excluded.equipment,
			working_groups = excluded.working_groups,
			partners = excluded.partners,
			differentiation = excluded.differentiation,
			lunch_info = excluded.lunch_info,
			dual_learning = excluded.dual_learning,
			citizenship_data = excluded.citizenship_data,
			language_data = excluded.language_data,
			residence_data = excluded.residence_data,
			absence_data = excluded.absence_data,
			scraped_at = excluded.scraped_at,
			updated_at = excluded.updated_at
	`

	now := time.Now()

	// Convert tables to JSON
	citizenshipJSON, _ := r.tableToJSON(detail.CitizenshipTable)
	languageJSON, _ := r.tableToJSON(detail.LanguageTable)
	residenceJSON, _ := r.tableToJSON(detail.ResidenceTable)
	absenceJSON, _ := r.tableToJSON(detail.AbsenceTable)

	_, err := r.db.ExecContext(ctx, query,
		detail.SchoolNumber,
		detail.SchoolName,
		detail.Languages,
		detail.Courses,
		detail.Offerings,
		detail.AvailableAfter4thGrade,
		detail.AdditionalInfo,
		detail.Equipment,
		detail.WorkingGroups,
		detail.Partners,
		detail.Differentiation,
		detail.LunchInfo,
		detail.DualLearning,
		citizenshipJSON,
		languageJSON,
		residenceJSON,
		absenceJSON,
		detail.ScrapedAt,
		now,
		now,
	)

	if err != nil {
		return errors.NewDatabaseError("upsert school detail", err)
	}

	return nil
}

// GetBySchoolNumber retrieves school details by school number
func (r *SchoolDetailRepository) GetBySchoolNumber(ctx context.Context, schoolNumber string) (*models.SchoolDetail, error) {
	var detail models.SchoolDetail
	query := `SELECT * FROM school_details WHERE school_number = ?`

	err := r.db.GetContext(ctx, &detail, query, schoolNumber)
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("school detail", schoolNumber)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("get school detail by number", err)
	}

	return &detail, nil
}

// GetAll retrieves all school details
func (r *SchoolDetailRepository) GetAll(ctx context.Context) ([]models.SchoolDetail, error) {
	var details []models.SchoolDetail
	query := `SELECT * FROM school_details ORDER BY school_name`

	err := r.db.SelectContext(ctx, &details, query)
	if err != nil {
		return nil, errors.NewDatabaseError("get all school details", err)
	}

	return details, nil
}

// GetAvailableAfter4thGrade retrieves schools available after 4th grade
func (r *SchoolDetailRepository) GetAvailableAfter4thGrade(ctx context.Context) ([]models.SchoolDetail, error) {
	var details []models.SchoolDetail
	query := `SELECT * FROM school_details WHERE available_after_4th_grade = 1 ORDER BY school_name`

	err := r.db.SelectContext(ctx, &details, query)
	if err != nil {
		return nil, errors.NewDatabaseError("get schools available after 4th grade", err)
	}

	return details, nil
}

// Delete deletes a school detail by school number
func (r *SchoolDetailRepository) Delete(ctx context.Context, schoolNumber string) error {
	query := `DELETE FROM school_details WHERE school_number = ?`

	_, err := r.db.ExecContext(ctx, query, schoolNumber)
	if err != nil {
		return errors.NewDatabaseError("delete school detail", err)
	}

	return nil
}

// DeleteAll deletes all school details
func (r *SchoolDetailRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM school_details`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.NewDatabaseError("delete all school details", err)
	}

	return nil
}

// GetCount returns the count of school details
func (r *SchoolDetailRepository) GetCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM school_details`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, errors.NewDatabaseError("get school details count", err)
	}

	return count, nil
}

// tableToJSON converts a StatisticTable to JSON string
func (r *SchoolDetailRepository) tableToJSON(table *models.StatisticTable) (string, error) {
	if table == nil {
		return "{}", nil
	}

	data, err := json.Marshal(table)
	if err != nil {
		return "{}", err
	}

	return string(data), nil
}
