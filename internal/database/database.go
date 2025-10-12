package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// New creates a new database connection
func New(dbPath string) (*sqlx.DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works better with single connection
	db.SetMaxIdleConns(1)

	return db, nil
}

// RunMigrations runs all database migrations
func RunMigrations(db *sqlx.DB) error {
	migrations := []string{
		// Create schools table with all fields
		`CREATE TABLE IF NOT EXISTS schools (
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
		)`,
		// Create indexes for schools
		`CREATE INDEX IF NOT EXISTS idx_schools_school_number ON schools(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_school_type ON schools(school_type)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_district ON schools(district)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_created_at ON schools(created_at)`,

		// Create construction_projects table
		`CREATE TABLE IF NOT EXISTS construction_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL UNIQUE,
			school_number TEXT NOT NULL,
			school_name TEXT NOT NULL,
			district TEXT DEFAULT '',
			school_type TEXT DEFAULT '',
			construction_measure TEXT DEFAULT '',
			description TEXT DEFAULT '',
			built_school_places TEXT DEFAULT '',
			places_after_construction TEXT DEFAULT '',
			class_tracks_after_construction TEXT DEFAULT '',
			handover_date TEXT DEFAULT '',
			total_costs TEXT DEFAULT '',
			street TEXT DEFAULT '',
			postal_code TEXT DEFAULT '',
			city TEXT DEFAULT '',
			latitude REAL,
			longitude REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Create indexes for construction_projects
		`CREATE INDEX IF NOT EXISTS idx_construction_projects_project_id ON construction_projects(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_construction_projects_school_number ON construction_projects(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_construction_projects_district ON construction_projects(district)`,
		`CREATE INDEX IF NOT EXISTS idx_construction_projects_created_at ON construction_projects(created_at)`,

		// Create school_statistics table
		`CREATE TABLE IF NOT EXISTS school_statistics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT,
			school_name TEXT,
			district TEXT,
			school_type TEXT,
			school_year TEXT,
			students TEXT,
			teachers TEXT,
			classes TEXT,
			metadata TEXT,
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(school_number, school_year)
		)`,
		// Create indexes for school_statistics
		`CREATE INDEX IF NOT EXISTS idx_statistics_school_number ON school_statistics(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_statistics_school_year ON school_statistics(school_year)`,
		`CREATE INDEX IF NOT EXISTS idx_statistics_scraped_at ON school_statistics(scraped_at)`,

		// Create school_details table for detailed school information
		`CREATE TABLE IF NOT EXISTS school_details (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL,
			school_name TEXT NOT NULL,
			languages TEXT DEFAULT '',
			courses TEXT DEFAULT '',
			offerings TEXT DEFAULT '',
			available_after_4th_grade BOOLEAN DEFAULT 0,
			additional_info TEXT DEFAULT '',
			equipment TEXT DEFAULT '',
			working_groups TEXT DEFAULT '',
			partners TEXT DEFAULT '',
			differentiation TEXT DEFAULT '',
			lunch_info TEXT DEFAULT '',
			dual_learning TEXT DEFAULT '',
			citizenship_data TEXT DEFAULT '',
			language_data TEXT DEFAULT '',
			residence_data TEXT DEFAULT '',
			absence_data TEXT DEFAULT '',
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(school_number)
		)`,
		// Create indexes for school_details
		`CREATE INDEX IF NOT EXISTS idx_school_details_school_number ON school_details(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_school_details_scraped_at ON school_details(scraped_at)`,
		`CREATE INDEX IF NOT EXISTS idx_school_details_available_after_4th ON school_details(available_after_4th_grade)`,

		// Create school_citizenship_stats table for normalized citizenship data
		`CREATE TABLE IF NOT EXISTS school_citizenship_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL,
			citizenship TEXT NOT NULL,
			female_students INTEGER DEFAULT 0,
			male_students INTEGER DEFAULT 0,
			total INTEGER DEFAULT 0,
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(school_number, citizenship)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_citizenship_school_number ON school_citizenship_stats(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_citizenship_scraped_at ON school_citizenship_stats(scraped_at)`,

		// Create school_language_stats table for normalized language data
		`CREATE TABLE IF NOT EXISTS school_language_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL UNIQUE,
			total_students INTEGER DEFAULT 0,
			ndh_female_students INTEGER DEFAULT 0,
			ndh_male_students INTEGER DEFAULT 0,
			ndh_total INTEGER DEFAULT 0,
			ndh_percentage REAL DEFAULT 0.0,
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_language_school_number ON school_language_stats(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_language_scraped_at ON school_language_stats(scraped_at)`,

		// Create school_residence_stats table for normalized residence data
		`CREATE TABLE IF NOT EXISTS school_residence_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL,
			district TEXT NOT NULL,
			student_count INTEGER DEFAULT 0,
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(school_number, district)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_residence_school_number ON school_residence_stats(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_residence_district ON school_residence_stats(district)`,
		`CREATE INDEX IF NOT EXISTS idx_residence_scraped_at ON school_residence_stats(scraped_at)`,

		// Create school_absence_stats table for normalized absence data
		`CREATE TABLE IF NOT EXISTS school_absence_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			school_number TEXT NOT NULL UNIQUE,
			school_absence_rate REAL DEFAULT 0.0,
			school_unexcused_rate REAL DEFAULT 0.0,
			school_type_absence_rate REAL DEFAULT 0.0,
			school_type_unexcused_rate REAL DEFAULT 0.0,
			region_absence_rate REAL DEFAULT 0.0,
			region_unexcused_rate REAL DEFAULT 0.0,
			berlin_absence_rate REAL DEFAULT 0.0,
			berlin_unexcused_rate REAL DEFAULT 0.0,
			scraped_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_absence_school_number ON school_absence_stats(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_absence_scraped_at ON school_absence_stats(scraped_at)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	return nil
}
