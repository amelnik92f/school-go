package models

import "time"

// SchoolStatistic represents school statistics data
type SchoolStatistic struct {
	ID           int64     `json:"id" db:"id"`
	SchoolNumber string    `json:"school_number" db:"school_number"`
	SchoolName   string    `json:"school_name" db:"school_name"`
	District     string    `json:"district" db:"district"`
	SchoolType   string    `json:"school_type" db:"school_type"`
	SchoolYear   string    `json:"school_year" db:"school_year"`
	Students     string    `json:"students" db:"students"`
	Teachers     string    `json:"teachers" db:"teachers"`
	Classes      string    `json:"classes" db:"classes"`
	Metadata     string    `json:"metadata" db:"metadata"` // JSON string
	ScrapedAt    time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// StatisticData represents scraped statistics data before saving to database
type StatisticData struct {
	SchoolNumber string
	SchoolName   string
	District     string
	SchoolType   string
	SchoolYear   string
	Students     string
	Teachers     string
	Classes      string
	Metadata     map[string]string
	ScrapedAt    time.Time
}
