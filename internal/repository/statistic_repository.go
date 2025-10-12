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

type StatisticRepository struct {
	db *sqlx.DB
}

func NewStatisticRepository(db *sqlx.DB) *StatisticRepository {
	return &StatisticRepository{db: db}
}

// GetAll returns all statistics ordered by school year desc
func (r *StatisticRepository) GetAll(ctx context.Context) ([]models.SchoolStatistic, error) {
	var statistics []models.SchoolStatistic
	query := `SELECT * FROM school_statistics ORDER BY school_year DESC, school_name`

	err := r.db.SelectContext(ctx, &statistics, query)
	if err != nil {
		return nil, errors.NewDatabaseError("get all statistics", err)
	}

	return statistics, nil
}

// GetByID returns a statistic by its ID
func (r *StatisticRepository) GetByID(ctx context.Context, id int64) (*models.SchoolStatistic, error) {
	var statistic models.SchoolStatistic
	query := `SELECT * FROM school_statistics WHERE id = ?`

	err := r.db.GetContext(ctx, &statistic, query, id)
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("statistic", id)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("get statistic by id", err)
	}

	return &statistic, nil
}

// GetBySchoolNumber returns all statistics for a school
func (r *StatisticRepository) GetBySchoolNumber(ctx context.Context, schoolNumber string) ([]models.SchoolStatistic, error) {
	var statistics []models.SchoolStatistic
	query := `SELECT * FROM school_statistics WHERE school_number = ? ORDER BY school_year DESC`

	err := r.db.SelectContext(ctx, &statistics, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get statistics by school number", err)
	}

	return statistics, nil
}

// GetBySchoolYear returns all statistics for a specific school year
func (r *StatisticRepository) GetBySchoolYear(ctx context.Context, schoolYear string) ([]models.SchoolStatistic, error) {
	var statistics []models.SchoolStatistic
	query := `SELECT * FROM school_statistics WHERE school_year = ? ORDER BY school_name`

	err := r.db.SelectContext(ctx, &statistics, query, schoolYear)
	if err != nil {
		return nil, errors.NewDatabaseError("get statistics by school year", err)
	}

	return statistics, nil
}

// Create creates a new statistic record
func (r *StatisticRepository) Create(ctx context.Context, data models.StatisticData) (*models.SchoolStatistic, error) {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(data.Metadata)
	if err != nil {
		return nil, errors.NewValidationError("metadata", "invalid metadata: "+err.Error())
	}

	query := `
		INSERT INTO school_statistics (
			school_number, school_name, district, school_type, school_year,
			students, teachers, classes, metadata, scraped_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		data.SchoolNumber,
		data.SchoolName,
		data.District,
		data.SchoolType,
		data.SchoolYear,
		data.Students,
		data.Teachers,
		data.Classes,
		string(metadataJSON),
		data.ScrapedAt,
	)
	if err != nil {
		return nil, errors.NewDatabaseError("create statistic", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.NewDatabaseError("get last insert id", err)
	}

	return r.GetByID(ctx, id)
}

// CreateOrUpdate creates a new statistic or updates if exists (based on school_number + school_year)
func (r *StatisticRepository) CreateOrUpdate(ctx context.Context, data models.StatisticData) error {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(data.Metadata)
	if err != nil {
		return errors.NewValidationError("metadata", "invalid metadata: "+err.Error())
	}

	query := `
		INSERT INTO school_statistics (
			school_number, school_name, district, school_type, school_year,
			students, teachers, classes, metadata, scraped_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(school_number, school_year) DO UPDATE SET
			school_name = excluded.school_name,
			district = excluded.district,
			school_type = excluded.school_type,
			students = excluded.students,
			teachers = excluded.teachers,
			classes = excluded.classes,
			metadata = excluded.metadata,
			scraped_at = excluded.scraped_at
	`

	_, err = r.db.ExecContext(ctx, query,
		data.SchoolNumber,
		data.SchoolName,
		data.District,
		data.SchoolType,
		data.SchoolYear,
		data.Students,
		data.Teachers,
		data.Classes,
		string(metadataJSON),
		data.ScrapedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("create or update statistic", err)
	}

	return nil
}

// BulkCreateOrUpdate creates or updates multiple statistics in a transaction
func (r *StatisticRepository) BulkCreateOrUpdate(ctx context.Context, statistics []models.StatisticData) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, errors.NewDatabaseError("begin transaction", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO school_statistics (
			school_number, school_name, district, school_type, school_year,
			students, teachers, classes, metadata, scraped_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(school_number, school_year) DO UPDATE SET
			school_name = excluded.school_name,
			district = excluded.district,
			school_type = excluded.school_type,
			students = excluded.students,
			teachers = excluded.teachers,
			classes = excluded.classes,
			metadata = excluded.metadata,
			scraped_at = excluded.scraped_at
	`)
	if err != nil {
		return 0, errors.NewDatabaseError("prepare statement", err)
	}
	defer stmt.Close()

	saved := 0
	for _, data := range statistics {
		// Convert metadata to JSON
		metadataJSON, err := json.Marshal(data.Metadata)
		if err != nil {
			continue // Skip invalid records
		}

		_, err = stmt.ExecContext(ctx,
			data.SchoolNumber,
			data.SchoolName,
			data.District,
			data.SchoolType,
			data.SchoolYear,
			data.Students,
			data.Teachers,
			data.Classes,
			string(metadataJSON),
			data.ScrapedAt,
		)
		if err != nil {
			continue // Skip failed records
		}
		saved++
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.NewDatabaseError("commit transaction", err)
	}

	return saved, nil
}

// DeleteBySchoolYear deletes all statistics for a specific school year
func (r *StatisticRepository) DeleteBySchoolYear(ctx context.Context, schoolYear string) error {
	query := `DELETE FROM school_statistics WHERE school_year = ?`

	_, err := r.db.ExecContext(ctx, query, schoolYear)
	if err != nil {
		return errors.NewDatabaseError("delete statistics by school year", err)
	}

	return nil
}

// GetLatestScrapedAt returns the most recent scrape timestamp
func (r *StatisticRepository) GetLatestScrapedAt(ctx context.Context) (*time.Time, error) {
	var scrapedAt sql.NullTime
	query := `SELECT MAX(scraped_at) FROM school_statistics`

	err := r.db.QueryRowContext(ctx, query).Scan(&scrapedAt)
	if err != nil {
		return nil, errors.NewDatabaseError("get latest scraped at", err)
	}

	if !scrapedAt.Valid {
		return nil, nil
	}

	return &scrapedAt.Time, nil
}

// GetStatisticsSummary returns summary statistics
func (r *StatisticRepository) GetStatisticsSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM school_statistics`).Scan(&totalCount)
	if err != nil {
		return nil, errors.NewDatabaseError("get total count", err)
	}
	summary["total_count"] = totalCount

	// Count by school year
	type yearCount struct {
		SchoolYear string `db:"school_year"`
		Count      int    `db:"count"`
	}
	var yearCounts []yearCount
	err = r.db.SelectContext(ctx, &yearCounts, `
		SELECT school_year, COUNT(*) as count 
		FROM school_statistics 
		GROUP BY school_year 
		ORDER BY school_year DESC
	`)
	if err != nil {
		return nil, errors.NewDatabaseError("get year counts", err)
	}
	summary["by_year"] = yearCounts

	// Latest scrape
	latestScrape, _ := r.GetLatestScrapedAt(ctx)
	if latestScrape != nil {
		summary["latest_scrape"] = latestScrape.Format(time.RFC3339)
	}

	return summary, nil
}
