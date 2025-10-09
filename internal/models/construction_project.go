package models

import "time"

// ConstructionProject represents a construction project from the Berlin API
type ConstructionProject struct {
	ID                           int64     `json:"id" db:"id"`
	ProjectID                    int       `json:"project_id" db:"project_id"`
	SchoolNumber                 string    `json:"school_number" db:"school_number"`
	SchoolName                   string    `json:"school_name" db:"school_name"`
	District                     string    `json:"district" db:"district"`
	SchoolType                   string    `json:"school_type" db:"school_type"`
	ConstructionMeasure          string    `json:"construction_measure" db:"construction_measure"`
	Description                  string    `json:"description" db:"description"`
	BuiltSchoolPlaces            string    `json:"built_school_places" db:"built_school_places"`
	PlacesAfterConstruction      string    `json:"places_after_construction" db:"places_after_construction"`
	ClassTracksAfterConstruction string    `json:"class_tracks_after_construction" db:"class_tracks_after_construction"`
	HandoverDate                 string    `json:"handover_date" db:"handover_date"`
	TotalCosts                   string    `json:"total_costs" db:"total_costs"`
	Street                       string    `json:"street" db:"street"`
	PostalCode                   string    `json:"postal_code" db:"postal_code"`
	City                         string    `json:"city" db:"city"`
	Latitude                     float64   `json:"latitude" db:"latitude"`
	Longitude                    float64   `json:"longitude" db:"longitude"`
	CreatedAt                    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateConstructionProjectInput represents the input for creating a construction project
type CreateConstructionProjectInput struct {
	ProjectID                    int     `json:"project_id"`
	SchoolNumber                 string  `json:"school_number"`
	SchoolName                   string  `json:"school_name"`
	District                     string  `json:"district"`
	SchoolType                   string  `json:"school_type"`
	ConstructionMeasure          string  `json:"construction_measure"`
	Description                  string  `json:"description"`
	BuiltSchoolPlaces            string  `json:"built_school_places"`
	PlacesAfterConstruction      string  `json:"places_after_construction"`
	ClassTracksAfterConstruction string  `json:"class_tracks_after_construction"`
	HandoverDate                 string  `json:"handover_date"`
	TotalCosts                   string  `json:"total_costs"`
	Street                       string  `json:"street"`
	PostalCode                   string  `json:"postal_code"`
	City                         string  `json:"city"`
	Latitude                     float64 `json:"latitude"`
	Longitude                    float64 `json:"longitude"`
}
