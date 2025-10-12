package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"schools-be/internal/models"
	"schools-be/internal/scraper"

	"github.com/chromedp/chromedp"
)

func main() {
	// Parse command-line flags
	headless := flag.Bool("headless", true, "Run browser in headless mode (set to false to see the browser)")
	flag.Parse()

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// The school URL to test - Hermann-Ehlers-Gymnasium
	schoolURL := "https://www.bildung.berlin.de/Schulverzeichnis/Schulportrait.aspx?IDSchulzweig=28722"

	fmt.Println("=== School Details Scraper - Debug Mode ===")
	fmt.Printf("Testing school: Hermann-Ehlers-Gymnasium - 06Y08\n")
	fmt.Printf("URL: %s\n", schoolURL)
	if !*headless {
		fmt.Println("🔍 Running in VISIBLE mode - you'll see the browser window")
	}
	fmt.Println()

	// Create chrome context with optional headless mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", *headless),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// Create context with timeout
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Add overall timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer timeoutCancel()

	fmt.Println("🔍 Scraping school details...")
	fmt.Println("-----------------------------------")

	// Create scraper (disable cache for testing to ensure fresh data)
	detailScraper := scraper.NewSchoolDetailsScraperWithCache(false)

	// Scrape the school using the same function as the main scraper
	details, err := detailScraper.ScrapeSchoolDetail(ctx, schoolURL)
	if err != nil {
		log.Fatalf("❌ Error scraping school: %v", err)
	}

	// Print results in a readable format
	fmt.Println("\n✅ Scraping completed successfully!")
	fmt.Println("\n📋 EXTRACTED DATA:")
	fmt.Println("===================")

	fmt.Printf("\n🏫 School Name: %s\n", details.SchoolName)
	fmt.Printf("🔢 School Number: %s\n", details.SchoolNumber)
	fmt.Printf("🌐 URL: %s\n", details.SchoolURL)

	fmt.Printf("\n📚 Languages:\n%s\n", formatField(details.Languages))
	fmt.Printf("\n📖 Courses (Leistungskurse):\n%s\n", formatField(details.Courses))
	fmt.Printf("\n🎯 Offerings (Angebote):\n%s\n", formatField(details.Offerings))
	fmt.Printf("\n✅ Available after 4th grade: %v\n", details.AvailableAfter4thGrade)
	fmt.Printf("\n📝 Additional Info:\n%s\n", formatField(details.AdditionalInfo))
	fmt.Printf("\n🏗️  Equipment (Ausstattung):\n%s\n", formatField(details.Equipment))
	fmt.Printf("\n👥 Working Groups (AGs):\n%s\n", formatField(details.WorkingGroups))
	fmt.Printf("\n🤝 Partners:\n%s\n", formatField(details.Partners))
	fmt.Printf("\n📊 Differentiation:\n%s\n", formatField(details.Differentiation))
	fmt.Printf("\n🍽️  Lunch Info:\n%s\n", formatField(details.LunchInfo))
	fmt.Printf("\n🎓 Dual Learning:\n%s\n", formatField(details.DualLearning))

	// Print raw statistics tables
	fmt.Println("\n📊 RAW STATISTICS TABLES:")
	fmt.Println("=========================")

	if details.CitizenshipTable != nil {
		fmt.Println("\n🌍 Citizenship Data (Staatsangehörigkeit) - Raw:")
		printTable(details.CitizenshipTable)
	} else {
		fmt.Println("\n🌍 Citizenship Data: Not found")
	}

	if details.LanguageTable != nil {
		fmt.Println("\n🗣️  Language Data (Nichtdeutsche Herkunftssprache) - Raw:")
		printTable(details.LanguageTable)
	} else {
		fmt.Println("\n🗣️  Language Data: Not found")
	}

	if details.ResidenceTable != nil {
		fmt.Println("\n🏘️  Residence Data (Wohnorte) - Raw:")
		printTable(details.ResidenceTable)
	} else {
		fmt.Println("\n🏘️  Residence Data: Not found")
	}

	if details.AbsenceTable != nil {
		fmt.Println("\n📅 Absence Data (Fehlzeiten) - Raw:")
		printTable(details.AbsenceTable)
	} else {
		fmt.Println("\n📅 Absence Data: Not found")
	}

	// Show normalized data (as it will be stored in DB)
	fmt.Println("\n\n💾 NORMALIZED DATA (As stored in DB):")
	fmt.Println("======================================")

	// Citizenship stats
	if details.CitizenshipTable != nil {
		citizenshipStats := scraper.NormalizeCitizenshipTable(details.SchoolNumber, details.CitizenshipTable, details.ScrapedAt)
		fmt.Printf("\n🌍 Citizenship Stats (%d records for table school_citizenship_stats):\n", len(citizenshipStats))
		for i, stat := range citizenshipStats {
			if i < 10 { // Show first 10
				fmt.Printf("  %2d. %-30s | Female: %3d | Male: %3d | Total: %3d\n",
					i+1, stat.Citizenship, stat.FemaleStudents, stat.MaleStudents, stat.Total)
			}
		}
		if len(citizenshipStats) > 10 {
			fmt.Printf("  ... and %d more records\n", len(citizenshipStats)-10)
		}
	}

	// Language stats
	if details.LanguageTable != nil {
		languageStat := scraper.NormalizeLanguageTable(details.SchoolNumber, details.LanguageTable, details.ScrapedAt)
		if languageStat != nil {
			fmt.Println("\n🗣️  Language Stats (1 record for table school_language_stats):")
			fmt.Printf("  Total Students: %d\n", languageStat.TotalStudents)
			fmt.Printf("  NDH Female: %d, NDH Male: %d, NDH Total: %d\n",
				languageStat.NDHFemaleStudents, languageStat.NDHMaleStudents, languageStat.NDHTotal)
			fmt.Printf("  NDH Percentage: %.1f%%\n", languageStat.NDHPercentage)
		}
	}

	// Residence stats
	if details.ResidenceTable != nil {
		residenceStats := scraper.NormalizeResidenceTable(details.SchoolNumber, details.ResidenceTable, details.ScrapedAt)
		fmt.Printf("\n🏘️  Residence Stats (%d records for table school_residence_stats):\n", len(residenceStats))
		for i, stat := range residenceStats {
			if i < 10 { // Show first 10
				fmt.Printf("  %2d. %-35s | Students: %4d\n", i+1, stat.District, stat.StudentCount)
			}
		}
		if len(residenceStats) > 10 {
			fmt.Printf("  ... and %d more records\n", len(residenceStats)-10)
		}
	}

	// Absence stats
	if details.AbsenceTable != nil {
		absenceStat := scraper.NormalizeAbsenceTable(details.SchoolNumber, details.AbsenceTable, details.ScrapedAt)
		if absenceStat != nil {
			fmt.Println("\n📅 Absence Stats (1 record for table school_absence_stats):")
			fmt.Printf("  School:      Absence: %.1f%% | Unexcused: %.1f%%\n",
				absenceStat.SchoolAbsenceRate, absenceStat.SchoolUnexcusedRate)
			fmt.Printf("  School Type: Absence: %.1f%% | Unexcused: %.1f%%\n",
				absenceStat.SchoolTypeAbsenceRate, absenceStat.SchoolTypeUnexcusedRate)
			fmt.Printf("  Region:      Absence: %.1f%% | Unexcused: %.1f%%\n",
				absenceStat.RegionAbsenceRate, absenceStat.RegionUnexcusedRate)
			fmt.Printf("  Berlin:      Absence: %.1f%% | Unexcused: %.1f%%\n",
				absenceStat.BerlinAbsenceRate, absenceStat.BerlinUnexcusedRate)
		}
	}

	// Print full JSON for verification
	fmt.Println("\n🔍 FULL JSON OUTPUT:")
	fmt.Println("====================")
	jsonData, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		log.Printf("Error marshaling to JSON: %v", err)
	} else {
		fmt.Println(string(jsonData))
	}

	fmt.Println("\n✨ Debug complete!")
}

func formatField(field string) string {
	if field == "" {
		return "  (empty)"
	}
	return "  " + field
}

func printTable(table *models.StatisticTable) {
	if len(table.Headers) > 0 {
		fmt.Printf("  Headers: %v\n", table.Headers)
	}
	if len(table.Rows) > 0 {
		fmt.Printf("  Rows (%d):\n", len(table.Rows))
		for i, row := range table.Rows {
			if i < 10 { // Print first 10 rows to avoid clutter
				fmt.Printf("    %d: %v\n", i+1, row)
			}
		}
		if len(table.Rows) > 10 {
			fmt.Printf("    ... and %d more rows\n", len(table.Rows)-10)
		}
	}
	if len(table.Data) > 0 {
		fmt.Printf("  Data: %v\n", table.Data)
	}
}
