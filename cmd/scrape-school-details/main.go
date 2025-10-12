package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"schools-be/internal/config"
	"schools-be/internal/database"
	"schools-be/internal/repository"
	"schools-be/internal/scraper"
	"schools-be/internal/service"
)

func main() {
	// Parse command-line flags
	clearCache := flag.Bool("clear-cache", false, "Clear the cache before scraping")
	noCache := flag.Bool("no-cache", false, "Disable cache (force re-scrape all schools)")
	flag.Parse()

	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("=== Berlin School Details Scraper ===")
	logger.Info("This will scrape detailed information from the Berlin school directory")
	logger.Info("including languages, courses, offerings, and student statistics")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		logger.Error("failed to initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize repositories, scraper, and service
	detailRepo := repository.NewSchoolDetailRepository(db)
	statsRepo := repository.NewSchoolStatisticsRepository(db)
	detailScraper := scraper.NewSchoolDetailsScraperWithCache(!*noCache)
	detailService := service.NewSchoolDetailService(detailRepo, statsRepo, detailScraper)

	// Clear cache if requested
	if *clearCache {
		logger.Info("clearing cache...")
		if err := detailService.ClearCache(); err != nil {
			logger.Error("failed to clear cache", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("✓ cache cleared successfully")
	}

	// Run the scraper with context (longer timeout for many schools)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel()

	logger.Info("starting school details scrape and store...")
	if *noCache {
		logger.Info("⚠️  Cache disabled - will re-scrape all schools")
	} else {
		logger.Info("💾  Cache enabled - will use cached data when available")
	}
	logger.Info("⚠️  This may take several hours as we need to visit each school's page")
	logger.Info("⚠️  The scraper is rate-limited to be respectful to the server")

	if err := detailService.ScrapeAndStoreDetails(ctx); err != nil {
		logger.Error("scraping failed", slog.String("error", err.Error()))
		// Don't exit with error if we got partial results
		logger.Warn("scraping completed with errors, but some data may have been saved")
	}

	// Get summary
	summary, err := detailService.GetSummary(ctx)
	if err != nil {
		logger.Warn("failed to get summary", slog.String("error", err.Error()))
	} else {
		logger.Info("school details summary", slog.Any("summary", summary))
	}

	logger.Info("✓ All done! School details have been scraped and saved to database")
	logger.Info("💾 Data saved to:", slog.String("database", cfg.DBPath))
}
