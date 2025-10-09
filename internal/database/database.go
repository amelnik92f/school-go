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
		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_schools_school_number ON schools(school_number)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_school_type ON schools(school_type)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_district ON schools(district)`,
		`CREATE INDEX IF NOT EXISTS idx_schools_created_at ON schools(created_at)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	return nil
}
