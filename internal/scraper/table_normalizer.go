package scraper

import (
	"strconv"
	"strings"
	"time"

	"schools-be/internal/models"
)

// NormalizeCitizenshipTable converts citizenship table to normalized records
func NormalizeCitizenshipTable(schoolNumber string, table *models.StatisticTable, scrapedAt time.Time) []models.SchoolCitizenshipStat {
	if table == nil || len(table.Rows) == 0 {
		return nil
	}

	var stats []models.SchoolCitizenshipStat

	for _, row := range table.Rows {
		if len(row) < 4 {
			continue
		}

		citizenship := strings.TrimSpace(row[0])
		// Skip empty rows or total rows
		if citizenship == "" {
			continue
		}

		stat := models.SchoolCitizenshipStat{
			SchoolNumber: schoolNumber,
			Citizenship:  citizenship,
			ScrapedAt:    scrapedAt,
		}

		// Parse numbers (handle "45 %" format)
		stat.FemaleStudents = parseInt(row[1])
		stat.MaleStudents = parseInt(row[2])
		stat.Total = parseInt(row[3])

		stats = append(stats, stat)
	}

	return stats
}

// NormalizeLanguageTable converts language table to normalized record
func NormalizeLanguageTable(schoolNumber string, table *models.StatisticTable, scrapedAt time.Time) *models.SchoolLanguageStat {
	if table == nil || len(table.Rows) < 2 {
		return nil
	}

	// Usually the last row has the actual data
	dataRow := table.Rows[len(table.Rows)-1]
	if len(dataRow) < 5 {
		return nil
	}

	stat := models.SchoolLanguageStat{
		SchoolNumber:      schoolNumber,
		TotalStudents:     parseInt(dataRow[0]),
		NDHFemaleStudents: parseInt(dataRow[1]),
		NDHMaleStudents:   parseInt(dataRow[2]),
		NDHTotal:          parseInt(dataRow[3]),
		NDHPercentage:     parseFloat(dataRow[4]),
		ScrapedAt:         scrapedAt,
	}

	return &stat
}

// NormalizeResidenceTable converts residence table to normalized records
func NormalizeResidenceTable(schoolNumber string, table *models.StatisticTable, scrapedAt time.Time) []models.SchoolResidenceStat {
	if table == nil || len(table.Rows) == 0 {
		return nil
	}

	var stats []models.SchoolResidenceStat

	for _, row := range table.Rows {
		if len(row) < 2 {
			continue
		}

		district := strings.TrimSpace(row[0])
		// Skip empty rows or total rows
		if district == "" || district == "Insgesamt" {
			continue
		}

		stat := models.SchoolResidenceStat{
			SchoolNumber: schoolNumber,
			District:     district,
			StudentCount: parseInt(row[1]),
			ScrapedAt:    scrapedAt,
		}

		stats = append(stats, stat)
	}

	return stats
}

// NormalizeAbsenceTable converts absence table to normalized record
func NormalizeAbsenceTable(schoolNumber string, table *models.StatisticTable, scrapedAt time.Time) *models.SchoolAbsenceStat {
	if table == nil || len(table.Rows) < 4 {
		return nil
	}

	stat := models.SchoolAbsenceStat{
		SchoolNumber: schoolNumber,
		ScrapedAt:    scrapedAt,
	}

	// Parse rows by their labels
	for _, row := range table.Rows {
		if len(row) < 3 {
			continue
		}

		label := strings.TrimSpace(strings.ToLower(row[0]))
		totalRate := parseFloat(row[1])
		unexcusedRate := parseFloat(row[2])

		switch {
		case strings.Contains(label, "schule") && !strings.Contains(label, "schulart"):
			stat.SchoolAbsenceRate = totalRate
			stat.SchoolUnexcusedRate = unexcusedRate
		case strings.Contains(label, "schulart"):
			stat.SchoolTypeAbsenceRate = totalRate
			stat.SchoolTypeUnexcusedRate = unexcusedRate
		case strings.Contains(label, "region"):
			stat.RegionAbsenceRate = totalRate
			stat.RegionUnexcusedRate = unexcusedRate
		case strings.Contains(label, "berlin"):
			stat.BerlinAbsenceRate = totalRate
			stat.BerlinUnexcusedRate = unexcusedRate
		}
	}

	return &stat
}

// parseInt safely parses an integer from string, handling various formats
func parseInt(s string) int {
	// Clean the string (remove spaces, %, etc.)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")

	if s == "" {
		return 0
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return val
}

// parseFloat safely parses a float from string, handling German format (comma as decimal separator)
func parseFloat(s string) float64 {
	// Clean the string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "%", "")

	// Replace comma with dot for decimal separator
	s = strings.ReplaceAll(s, ",", ".")

	if s == "" {
		return 0.0
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}

	return val
}
