package models

import "time"

type School struct {
	ID             int64     `json:"id" db:"id"`
	SchoolNumber   string    `json:"school_number" db:"school_number"`     // BSN - School number (e.g., "01B01")
	Name           string    `json:"name" db:"name"`                       // Schulname - School name
	SchoolType     string    `json:"school_type" db:"school_type"`         // Schulart - School type (e.g., "Gymnasium", "Grundschule")
	Operator       string    `json:"operator" db:"operator"`               // Traeger - Operator (e.g., "öffentlich", "privat")
	SchoolCategory string    `json:"school_category" db:"school_category"` // Schultyp - School category
	District       string    `json:"district" db:"district"`               // Bezirk - District (e.g., "Mitte")
	Neighborhood   string    `json:"neighborhood" db:"neighborhood"`       // Ortsteil - Neighborhood
	PostalCode     string    `json:"postal_code" db:"postal_code"`         // PLZ - Postal code
	Street         string    `json:"street" db:"street"`                   // Strasse - Street name
	HouseNumber    string    `json:"house_number" db:"house_number"`       // Hausnr - House number
	Phone          string    `json:"phone" db:"phone"`                     // Telefon - Phone number
	Fax            string    `json:"fax" db:"fax"`                         // Fax - Fax number
	Email          string    `json:"email" db:"email"`                     // Email - Email address
	Website        string    `json:"website" db:"website"`                 // Internet - Website URL
	SchoolYear     string    `json:"school_year" db:"school_year"`         // Schuljahr - School year (e.g., "2025/26")
	Latitude       float64   `json:"latitude" db:"latitude"`               // Geographic coordinate
	Longitude      float64   `json:"longitude" db:"longitude"`             // Geographic coordinate
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSchoolInput struct {
	SchoolNumber   string  `json:"school_number" validate:"required,min=1,max=50"`
	Name           string  `json:"name" validate:"required,min=1,max=300"`
	SchoolType     string  `json:"school_type" validate:"required,min=1,max=100"`
	Operator       string  `json:"operator" validate:"omitempty,max=100"`
	SchoolCategory string  `json:"school_category" validate:"omitempty,max=100"`
	District       string  `json:"district" validate:"omitempty,max=100"`
	Neighborhood   string  `json:"neighborhood" validate:"omitempty,max=100"`
	PostalCode     string  `json:"postal_code" validate:"omitempty,max=10"`
	Street         string  `json:"street" validate:"omitempty,max=200"`
	HouseNumber    string  `json:"house_number" validate:"omitempty,max=20"`
	Phone          string  `json:"phone" validate:"omitempty,max=50"`
	Fax            string  `json:"fax" validate:"omitempty,max=50"`
	Email          string  `json:"email" validate:"omitempty,email,max=200"`
	Website        string  `json:"website" validate:"omitempty,url,max=500"`
	SchoolYear     string  `json:"school_year" validate:"omitempty,max=20"`
	Latitude       float64 `json:"latitude" validate:"required,latitude"`
	Longitude      float64 `json:"longitude" validate:"required,longitude"`
}

type UpdateSchoolInput struct {
	SchoolNumber   *string  `json:"school_number,omitempty" validate:"omitempty,min=1,max=50"`
	Name           *string  `json:"name,omitempty" validate:"omitempty,min=1,max=300"`
	SchoolType     *string  `json:"school_type,omitempty" validate:"omitempty,min=1,max=100"`
	Operator       *string  `json:"operator,omitempty" validate:"omitempty,max=100"`
	SchoolCategory *string  `json:"school_category,omitempty" validate:"omitempty,max=100"`
	District       *string  `json:"district,omitempty" validate:"omitempty,max=100"`
	Neighborhood   *string  `json:"neighborhood,omitempty" validate:"omitempty,max=100"`
	PostalCode     *string  `json:"postal_code,omitempty" validate:"omitempty,max=10"`
	Street         *string  `json:"street,omitempty" validate:"omitempty,max=200"`
	HouseNumber    *string  `json:"house_number,omitempty" validate:"omitempty,max=20"`
	Phone          *string  `json:"phone,omitempty" validate:"omitempty,max=50"`
	Fax            *string  `json:"fax,omitempty" validate:"omitempty,max=50"`
	Email          *string  `json:"email,omitempty" validate:"omitempty,email,max=200"`
	Website        *string  `json:"website,omitempty" validate:"omitempty,url,max=500"`
	SchoolYear     *string  `json:"school_year,omitempty" validate:"omitempty,max=20"`
	Latitude       *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude      *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
}
