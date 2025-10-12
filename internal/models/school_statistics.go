package models

import "time"

// SchoolCitizenshipStat represents citizenship statistics for a school
type SchoolCitizenshipStat struct {
	ID             int64     `json:"id" db:"id"`
	SchoolNumber   string    `json:"school_number" db:"school_number"`
	Citizenship    string    `json:"citizenship" db:"citizenship"`         // e.g., "Europa (ohne Deutschland)", "Afrika"
	FemaleStudents int       `json:"female_students" db:"female_students"` // Schülerinnen
	MaleStudents   int       `json:"male_students" db:"male_students"`     // Schüler
	Total          int       `json:"total" db:"total"`                     // Insgesamt
	ScrapedAt      time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// SchoolLanguageStat represents non-German heritage language statistics for a school
type SchoolLanguageStat struct {
	ID                int64     `json:"id" db:"id"`
	SchoolNumber      string    `json:"school_number" db:"school_number"`
	TotalStudents     int       `json:"total_students" db:"total_students"`           // Total students
	NDHFemaleStudents int       `json:"ndh_female_students" db:"ndh_female_students"` // Non-German heritage female
	NDHMaleStudents   int       `json:"ndh_male_students" db:"ndh_male_students"`     // Non-German heritage male
	NDHTotal          int       `json:"ndh_total" db:"ndh_total"`                     // Non-German heritage total
	NDHPercentage     float64   `json:"ndh_percentage" db:"ndh_percentage"`           // Percentage
	ScrapedAt         time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// SchoolResidenceStat represents student residence statistics for a school
type SchoolResidenceStat struct {
	ID           int64     `json:"id" db:"id"`
	SchoolNumber string    `json:"school_number" db:"school_number"`
	District     string    `json:"district" db:"district"`           // Wohnort (e.g., "Steglitz-Zehlendorf")
	StudentCount int       `json:"student_count" db:"student_count"` // Number of students from this district
	ScrapedAt    time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// SchoolAbsenceStat represents absence statistics for a school
type SchoolAbsenceStat struct {
	ID                      int64     `json:"id" db:"id"`
	SchoolNumber            string    `json:"school_number" db:"school_number"`
	SchoolAbsenceRate       float64   `json:"school_absence_rate" db:"school_absence_rate"`               // der Schule
	SchoolUnexcusedRate     float64   `json:"school_unexcused_rate" db:"school_unexcused_rate"`           // der Schule unentschuldigt
	SchoolTypeAbsenceRate   float64   `json:"school_type_absence_rate" db:"school_type_absence_rate"`     // der Schulart
	SchoolTypeUnexcusedRate float64   `json:"school_type_unexcused_rate" db:"school_type_unexcused_rate"` // der Schulart unentschuldigt
	RegionAbsenceRate       float64   `json:"region_absence_rate" db:"region_absence_rate"`               // der Region
	RegionUnexcusedRate     float64   `json:"region_unexcused_rate" db:"region_unexcused_rate"`           // der Region unentschuldigt
	BerlinAbsenceRate       float64   `json:"berlin_absence_rate" db:"berlin_absence_rate"`               // in Berlin
	BerlinUnexcusedRate     float64   `json:"berlin_unexcused_rate" db:"berlin_unexcused_rate"`           // in Berlin unentschuldigt
	ScrapedAt               time.Time `json:"scraped_at" db:"scraped_at"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
}
