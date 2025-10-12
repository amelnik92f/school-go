package models

import "time"

// SchoolDetail contains detailed information about a school scraped from the Berlin school directory
type SchoolDetail struct {
	ID                     int64     `json:"id" db:"id"`
	SchoolNumber           string    `json:"school_number" db:"school_number"`                         // BSN - Link to schools table
	SchoolName             string    `json:"school_name" db:"school_name"`                             // Name of the school
	Languages              string    `json:"languages" db:"languages"`                                 // Sprachen - Languages offered
	Courses                string    `json:"courses" db:"courses"`                                     // Leistungskurse - Advanced courses
	Offerings              string    `json:"offerings" db:"offerings"`                                 // Angebote - Programs and offerings
	AvailableAfter4thGrade bool      `json:"available_after_4th_grade" db:"available_after_4th_grade"` // True if "ab Jahrgangsstufe 5 beginnende"
	AdditionalInfo         string    `json:"additional_info" db:"additional_info"`                     // Bemerkungen - Additional information
	Equipment              string    `json:"equipment" db:"equipment"`                                 // Ausstattung - Equipment and facilities
	WorkingGroups          string    `json:"working_groups" db:"working_groups"`                       // AGs - Working groups/extracurricular activities
	Partners               string    `json:"partners" db:"partners"`                                   // Partner - External partners
	Differentiation        string    `json:"differentiation" db:"differentiation"`                     // Differenzierung - Differentiation methods
	LunchInfo              string    `json:"lunch_info" db:"lunch_info"`                               // Mittagessen - Lunch information
	DualLearning           string    `json:"dual_learning" db:"dual_learning"`                         // Duales Lernen - Dual learning programs
	CitizenshipData        string    `json:"citizenship_data" db:"citizenship_data"`                   // JSON: Staatsangehörigkeit statistics
	LanguageData           string    `json:"language_data" db:"language_data"`                         // JSON: Nichtdeutsche Herkunftssprache statistics
	ResidenceData          string    `json:"residence_data" db:"residence_data"`                       // JSON: Wohnorte statistics
	AbsenceData            string    `json:"absence_data" db:"absence_data"`                           // JSON: Fehlzeiten statistics
	ScrapedAt              time.Time `json:"scraped_at" db:"scraped_at"`                               // When this data was scraped
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// StatisticTable represents a generic statistics table with headers and rows
type StatisticTable struct {
	Headers []string          `json:"headers"`
	Rows    [][]string        `json:"rows"`
	Data    map[string]string `json:"data,omitempty"` // Key-value pairs for simple tables
}

// SchoolDetailData is used for data collection during scraping
type SchoolDetailData struct {
	SchoolNumber           string          `json:"school_number"`
	SchoolName             string          `json:"school_name"`
	SchoolURL              string          `json:"school_url"`
	Languages              string          `json:"languages"`
	Courses                string          `json:"courses"`
	Offerings              string          `json:"offerings"`
	AvailableAfter4thGrade bool            `json:"available_after_4th_grade"`
	AdditionalInfo         string          `json:"additional_info"`
	Equipment              string          `json:"equipment"`
	WorkingGroups          string          `json:"working_groups"`
	Partners               string          `json:"partners"`
	Differentiation        string          `json:"differentiation"`
	LunchInfo              string          `json:"lunch_info"`
	DualLearning           string          `json:"dual_learning"`
	CitizenshipTable       *StatisticTable `json:"citizenship_table,omitempty"`
	LanguageTable          *StatisticTable `json:"language_table,omitempty"`
	ResidenceTable         *StatisticTable `json:"residence_table,omitempty"`
	AbsenceTable           *StatisticTable `json:"absence_table,omitempty"`
	ScrapedAt              time.Time       `json:"scraped_at"`
}
