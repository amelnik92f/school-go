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
	cron                *cron.Cron
	schoolService       *service.SchoolService
	statisticService    *service.StatisticService
	schoolDetailService *service.SchoolDetailService
	config              *config.Config
	logger              *slog.Logger
}

func New(cfg *config.Config, schoolService *service.SchoolService, statisticService *service.StatisticService, schoolDetailService *service.SchoolDetailService) *Scheduler {
	return &Scheduler{
		cron:                cron.New(),
		schoolService:       schoolService,
		statisticService:    statisticService,
		schoolDetailService: schoolDetailService,
		config:              cfg,
		logger:              slog.Default(),
	}
}

func (s *Scheduler) Start() {
	// Schedule full data refresh (all tasks run sequentially)
	_, err := s.cron.AddFunc(s.config.FetchSchedule, func() {
		s.runFullDataRefresh()
	})
	if err != nil {
		s.logger.Error("failed to schedule data refresh job", slog.String("error", err.Error()))
		return
	}

	s.cron.Start()
	s.logger.Info("scheduler started",
		slog.String("refresh_schedule", s.config.FetchSchedule),
	)
}

// runFullDataRefresh executes all data refresh tasks sequentially
func (s *Scheduler) runFullDataRefresh() {
	startTime := time.Now()
	s.logger.Info("starting full data refresh cycle")

	// Step 1: Fetch schools and construction projects
	s.logger.Info("step 1/3: fetching school data")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel1()

	if err := s.schoolService.FetchAndStoreSchools(ctx1); err != nil {
		s.logger.Error("schools fetch failed", slog.String("error", err.Error()))
	} else {
		s.logger.Info("schools fetch completed")
	}

	if err := s.schoolService.FetchAndStoreConstructionProjects(ctx1); err != nil {
		s.logger.Error("construction projects fetch failed", slog.String("error", err.Error()))
	} else {
		s.logger.Info("construction projects fetch completed")
	}

	// Step 2: Scrape statistics
	s.logger.Info("step 2/3: scraping statistics")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel2()

	if err := s.statisticService.ScrapeAndStoreStatistics(ctx2); err != nil {
		s.logger.Error("statistics scrape failed", slog.String("error", err.Error()))
	} else {
		s.logger.Info("statistics scrape completed")
	}

	// Step 3: Scrape school details (longest operation)
	s.logger.Info("step 3/3: scraping school details (this may take several hours)")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel3()

	if err := s.schoolDetailService.ScrapeAndStoreDetails(ctx3); err != nil {
		s.logger.Error("school details scrape failed", slog.String("error", err.Error()))
	} else {
		s.logger.Info("school details scrape completed")
	}

	duration := time.Since(startTime)
	s.logger.Info("full data refresh cycle completed",
		slog.String("duration", duration.String()),
	)
}

func (s *Scheduler) Stop() {
	s.logger.Info("stopping scheduler")
	s.cron.Stop()
}
