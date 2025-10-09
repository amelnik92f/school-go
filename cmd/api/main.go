package main

import (
	"context"
	"log"
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
	"schools-be/internal/server"
	"schools-be/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	schoolRepo := repository.NewSchoolRepository(db)

	// Initialize fetchers
	schoolFetcher := fetcher.NewSchoolFetcher()

	// Initialize services
	schoolService := service.NewSchoolService(schoolRepo, schoolFetcher)

	// Initialize handlers
	schoolHandler := handler.NewSchoolHandler(schoolService)

	// Initialize HTTP server
	srv := server.New(cfg, schoolHandler)

	// Initialize and start scheduler
	sched := scheduler.New(cfg, schoolService)
	sched.Start()
	defer sched.Stop()

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

