package scheduler

import (
	"context"
	"log/slog"
	"time"

	"schools-be/internal/config"
	"schools-be/internal/service"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron          *cron.Cron
	schoolService *service.SchoolService
	config        *config.Config
	logger        *slog.Logger
}

func New(cfg *config.Config, schoolService *service.SchoolService) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		schoolService: schoolService,
		config:        cfg,
		logger:        slog.Default(),
	}
}

func (s *Scheduler) Start() {
	// Schedule school data refresh
	_, err := s.cron.AddFunc(s.config.FetchSchedule, func() {
		s.logger.Info("running scheduled school data refresh")

		// Create context with timeout for the refresh operation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := s.schoolService.RefreshSchoolsData(ctx); err != nil {
			s.logger.Error("scheduled refresh failed", slog.String("error", err.Error()))
		}
	})
	if err != nil {
		s.logger.Error("failed to schedule job", slog.String("error", err.Error()))
	}

	// Add more scheduled jobs here as needed
	// Example:
	// s.cron.AddFunc("0 3 * * *", s.cleanupOldData)

	s.cron.Start()
	s.logger.Info("scheduler started", slog.String("schedule", s.config.FetchSchedule))
}

func (s *Scheduler) Stop() {
	s.logger.Info("stopping scheduler")
	s.cron.Stop()
}

// Add more scheduled job functions here
// func (s *Scheduler) cleanupOldData() {
//     log.Println("Running cleanup job...")
//     // Implementation
// }
