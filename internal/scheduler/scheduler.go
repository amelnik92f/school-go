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
	cron             *cron.Cron
	schoolService    *service.SchoolService
	statisticService *service.StatisticService
	config           *config.Config
	logger           *slog.Logger
}

func New(cfg *config.Config, schoolService *service.SchoolService, statisticService *service.StatisticService) *Scheduler {
	return &Scheduler{
		cron:             cron.New(),
		schoolService:    schoolService,
		statisticService: statisticService,
		config:           cfg,
		logger:           slog.Default(),
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
		s.logger.Error("failed to schedule school data refresh job", slog.String("error", err.Error()))
	}

	// Schedule statistics scraping
	_, err = s.cron.AddFunc(s.config.FetchSchedule, func() {
		s.logger.Info("running scheduled statistics scrape")

		// Create context with timeout for the scrape operation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := s.statisticService.RefreshStatisticsData(ctx); err != nil {
			s.logger.Error("scheduled statistics scrape failed", slog.String("error", err.Error()))
		} else {
			s.logger.Info("statistics scrape completed successfully")
		}
	})
	if err != nil {
		s.logger.Error("failed to schedule statistics scrape job", slog.String("error", err.Error()))
	}

	s.cron.Start()
	s.logger.Info("scheduler started",
		slog.String("school_refresh_schedule", s.config.FetchSchedule),
		slog.String("statistics_scrape_schedule", s.config.FetchSchedule),
	)
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
