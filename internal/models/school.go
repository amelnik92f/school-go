package models

import "time"

type School struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	Type      string    `json:"type" db:"type"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSchoolInput struct {
	Name      string  `json:"name" validate:"required,min=1,max=200"`
	Address   string  `json:"address" validate:"required,min=1,max=500"`
	Type      string  `json:"type" validate:"required,min=1,max=100"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type UpdateSchoolInput struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Address   *string  `json:"address,omitempty" validate:"omitempty,min=1,max=500"`
	Type      *string  `json:"type,omitempty" validate:"omitempty,min=1,max=100"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
}
