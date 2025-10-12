package repository

import (
	"context"
	"time"

	"schools-be/internal/errors"
	"schools-be/internal/models"

	"github.com/jmoiron/sqlx"
)

type SchoolStatisticsRepository struct {
	db *sqlx.DB
}

func NewSchoolStatisticsRepository(db *sqlx.DB) *SchoolStatisticsRepository {
	return &SchoolStatisticsRepository{db: db}
}

// SaveCitizenshipStats saves citizenship statistics (replaces existing data for the school)
func (r *SchoolStatisticsRepository) SaveCitizenshipStats(ctx context.Context, stats []models.SchoolCitizenshipStat) error {
	if len(stats) == 0 {
		return nil
	}

	// Delete existing stats for this school
	schoolNumber := stats[0].SchoolNumber
	_, err := r.db.ExecContext(ctx, `DELETE FROM school_citizenship_stats WHERE school_number = ?`, schoolNumber)
	if err != nil {
		return errors.NewDatabaseError("delete old citizenship stats", err)
	}

	// Insert new stats
	query := `INSERT INTO school_citizenship_stats (school_number, citizenship, female_students, male_students, total, scraped_at, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	for _, stat := range stats {
		_, err := r.db.ExecContext(ctx, query,
			stat.SchoolNumber, stat.Citizenship, stat.FemaleStudents, stat.MaleStudents, stat.Total,
			stat.ScrapedAt, time.Now())
		if err != nil {
			return errors.NewDatabaseError("insert citizenship stat", err)
		}
	}

	return nil
}

// SaveLanguageStat saves language statistics (replaces existing data for the school)
func (r *SchoolStatisticsRepository) SaveLanguageStat(ctx context.Context, stat models.SchoolLanguageStat) error {
	query := `INSERT OR REPLACE INTO school_language_stats 
	          (school_number, total_students, ndh_female_students, ndh_male_students, ndh_total, ndh_percentage, scraped_at, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		stat.SchoolNumber, stat.TotalStudents, stat.NDHFemaleStudents, stat.NDHMaleStudents,
		stat.NDHTotal, stat.NDHPercentage, stat.ScrapedAt, time.Now())

	if err != nil {
		return errors.NewDatabaseError("save language stat", err)
	}

	return nil
}

// SaveResidenceStats saves residence statistics (replaces existing data for the school)
func (r *SchoolStatisticsRepository) SaveResidenceStats(ctx context.Context, stats []models.SchoolResidenceStat) error {
	if len(stats) == 0 {
		return nil
	}

	// Delete existing stats for this school
	schoolNumber := stats[0].SchoolNumber
	_, err := r.db.ExecContext(ctx, `DELETE FROM school_residence_stats WHERE school_number = ?`, schoolNumber)
	if err != nil {
		return errors.NewDatabaseError("delete old residence stats", err)
	}

	// Insert new stats
	query := `INSERT INTO school_residence_stats (school_number, district, student_count, scraped_at, created_at) 
	          VALUES (?, ?, ?, ?, ?)`

	for _, stat := range stats {
		_, err := r.db.ExecContext(ctx, query,
			stat.SchoolNumber, stat.District, stat.StudentCount, stat.ScrapedAt, time.Now())
		if err != nil {
			return errors.NewDatabaseError("insert residence stat", err)
		}
	}

	return nil
}

// SaveAbsenceStat saves absence statistics (replaces existing data for the school)
func (r *SchoolStatisticsRepository) SaveAbsenceStat(ctx context.Context, stat models.SchoolAbsenceStat) error {
	query := `INSERT OR REPLACE INTO school_absence_stats 
	          (school_number, school_absence_rate, school_unexcused_rate, school_type_absence_rate, school_type_unexcused_rate,
	           region_absence_rate, region_unexcused_rate, berlin_absence_rate, berlin_unexcused_rate, scraped_at, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		stat.SchoolNumber, stat.SchoolAbsenceRate, stat.SchoolUnexcusedRate,
		stat.SchoolTypeAbsenceRate, stat.SchoolTypeUnexcusedRate,
		stat.RegionAbsenceRate, stat.RegionUnexcusedRate,
		stat.BerlinAbsenceRate, stat.BerlinUnexcusedRate,
		stat.ScrapedAt, time.Now())

	if err != nil {
		return errors.NewDatabaseError("save absence stat", err)
	}

	return nil
}

// GetCitizenshipStats retrieves citizenship statistics for a school
func (r *SchoolStatisticsRepository) GetCitizenshipStats(ctx context.Context, schoolNumber string) ([]models.SchoolCitizenshipStat, error) {
	var stats []models.SchoolCitizenshipStat
	query := `SELECT * FROM school_citizenship_stats WHERE school_number = ? ORDER BY citizenship`

	err := r.db.SelectContext(ctx, &stats, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get citizenship stats", err)
	}

	return stats, nil
}

// GetLanguageStat retrieves language statistics for a school
func (r *SchoolStatisticsRepository) GetLanguageStat(ctx context.Context, schoolNumber string) (*models.SchoolLanguageStat, error) {
	var stat models.SchoolLanguageStat
	query := `SELECT * FROM school_language_stats WHERE school_number = ?`

	err := r.db.GetContext(ctx, &stat, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get language stat", err)
	}

	return &stat, nil
}

// GetResidenceStats retrieves residence statistics for a school
func (r *SchoolStatisticsRepository) GetResidenceStats(ctx context.Context, schoolNumber string) ([]models.SchoolResidenceStat, error) {
	var stats []models.SchoolResidenceStat
	query := `SELECT * FROM school_residence_stats WHERE school_number = ? ORDER BY student_count DESC`

	err := r.db.SelectContext(ctx, &stats, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get residence stats", err)
	}

	return stats, nil
}

// GetAbsenceStat retrieves absence statistics for a school
func (r *SchoolStatisticsRepository) GetAbsenceStat(ctx context.Context, schoolNumber string) (*models.SchoolAbsenceStat, error) {
	var stat models.SchoolAbsenceStat
	query := `SELECT * FROM school_absence_stats WHERE school_number = ?`

	err := r.db.GetContext(ctx, &stat, query, schoolNumber)
	if err != nil {
		return nil, errors.NewDatabaseError("get absence stat", err)
	}

	return &stat, nil
}
