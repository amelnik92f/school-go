package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"schools-be/internal/config"
	"schools-be/internal/handler"
	appmiddleware "schools-be/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	router *chi.Mux
	config *config.Config
	server *http.Server
}

func New(cfg *config.Config, schoolHandler *handler.SchoolHandler, constructionProjectHandler *handler.ConstructionProjectHandler) *Server {
	s := &Server{
		router: chi.NewRouter(),
		config: cfg,
	}

	// Setup middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes(schoolHandler, constructionProjectHandler)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) setupMiddleware() {
	// Basic middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Timeout middleware
	s.router.Use(middleware.Timeout(s.config.APITimeout))

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes(schoolHandler *handler.SchoolHandler, constructionProjectHandler *handler.ConstructionProjectHandler) {
	// Health check (no authentication required)
	healthHandler := handler.NewHealthHandler()
	s.router.Get("/health", healthHandler.HealthCheck)

	// API routes (with authentication)
	s.router.Route("/api/v1", func(r chi.Router) {
		// Apply API key authentication middleware to all API routes
		r.Use(appmiddleware.APIKeyAuth(s.config))

		// Schools endpoints
		r.Route("/schools", func(r chi.Router) {
			r.Get("/", schoolHandler.GetSchoolsEnriched)
			r.Get("/{id}", schoolHandler.GetSchoolEnriched)
			r.Get("/{id}/summary", schoolHandler.GetSchoolSummary)
			r.Post("/{id}/routes", schoolHandler.CalculateRoutes)
		})

		// Construction projects endpoints
		r.Route("/construction-projects", func(r chi.Router) {
			r.Get("/", constructionProjectHandler.GetAll)
			r.Get("/standalone", constructionProjectHandler.GetStandalone)
			r.Get("/{id}", constructionProjectHandler.GetByID)
		})
	})

	// 404 handler
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "route not found"}`)
	})
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
