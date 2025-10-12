package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"schools-be/internal/config"
	"schools-be/internal/database"
	"schools-be/internal/fetcher"
	"schools-be/internal/handler"
	"schools-be/internal/repository"
	"schools-be/internal/scheduler"
	"schools-be/internal/scraper"
	"schools-be/internal/server"
	"schools-be/internal/service"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("starting application",
		slog.String("port", cfg.Port),
		slog.String("env", cfg.Env),
	)

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
	logger.Info("database migrations completed")

	// Initialize repositories
	schoolRepo := repository.NewSchoolRepository(db)
	constructionRepo := repository.NewConstructionProjectRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)
	schoolDetailRepo := repository.NewSchoolDetailRepository(db)
	schoolStatsRepo := repository.NewSchoolStatisticsRepository(db)

	// Initialize fetchers and scrapers
	schoolFetcher := fetcher.NewSchoolFetcher()
	statisticsScraper := scraper.NewStatisticsScraper()

	// Initialize services
	schoolService := service.NewSchoolService(schoolRepo, constructionRepo, schoolDetailRepo, schoolStatsRepo, schoolFetcher)
	statisticService := service.NewStatisticService(statisticRepo, statisticsScraper)
	constructionProjectService := service.NewConstructionProjectService(constructionRepo)

	// Initialize handlers
	schoolHandler := handler.NewSchoolHandler(schoolService)
	constructionProjectHandler := handler.NewConstructionProjectHandler(constructionProjectService)

	// Initialize HTTP server
	srv := server.New(cfg, schoolHandler, constructionProjectHandler)

	// Initialize and start scheduler
	sched := scheduler.New(cfg, schoolService, statisticService)
	sched.Start()
	defer sched.Stop()

	// Start server in a goroutine
	go func() {
		logger.Info("starting http server", slog.String("port", cfg.Port))
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
