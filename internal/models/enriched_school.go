package models

// EnrichedSchool contains a school with all related data from other tables
type EnrichedSchool struct {
	// Base school data
	School School `json:"school"`

	// Detailed school information
	Details *SchoolDetail `json:"details,omitempty"`

	// Statistical data
	CitizenshipStats []SchoolCitizenshipStat `json:"citizenship_stats,omitempty"`
	LanguageStat     *SchoolLanguageStat     `json:"language_stat,omitempty"`
	ResidenceStats   []SchoolResidenceStat   `json:"residence_stats,omitempty"`
	AbsenceStat      *SchoolAbsenceStat      `json:"absence_stat,omitempty"`

	// Construction projects related to this school
	ConstructionProjects []ConstructionProject `json:"construction_projects,omitempty"`
}
