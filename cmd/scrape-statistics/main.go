package main

import (
	"context"
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
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("=== Berlin School Statistics Scraper ===")

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

	// Initialize repository, scraper, and service
	statisticRepo := repository.NewStatisticRepository(db)
	statisticsScraper := scraper.NewStatisticsScraper()
	statisticService := service.NewStatisticService(statisticRepo, statisticsScraper)

	// Run the scraper with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("starting statistics scrape and store...")
	if err := statisticService.ScrapeAndStoreStatistics(ctx); err != nil {
		logger.Error("scraping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Get summary
	summary, err := statisticService.GetStatisticsSummary(ctx)
	if err != nil {
		logger.Warn("failed to get summary", slog.String("error", err.Error()))
	} else {
		logger.Info("statistics summary", slog.Any("summary", summary))
	}

	logger.Info("✓ All done! Data has been scraped and saved to database")
}
