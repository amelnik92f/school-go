package service

import (
	"context"
	"fmt"
	"log/slog"

	"schools-be/internal/models"
	"schools-be/internal/repository"
	"schools-be/internal/scraper"
)

type StatisticService struct {
	repo    *repository.StatisticRepository
	scraper *scraper.StatisticsScraper
	logger  *slog.Logger
}

func NewStatisticService(repo *repository.StatisticRepository, scraper *scraper.StatisticsScraper) *StatisticService {
	return &StatisticService{
		repo:    repo,
		scraper: scraper,
		logger:  slog.Default(),
	}
}

// GetAllStatistics returns all statistics from the database
func (s *StatisticService) GetAllStatistics(ctx context.Context) ([]models.SchoolStatistic, error) {
	return s.repo.GetAll(ctx)
}

// GetStatisticByID returns a statistic by its ID
func (s *StatisticService) GetStatisticByID(ctx context.Context, id int64) (*models.SchoolStatistic, error) {
	return s.repo.GetByID(ctx, id)
}

// GetStatisticsBySchoolNumber returns all statistics for a specific school
func (s *StatisticService) GetStatisticsBySchoolNumber(ctx context.Context, schoolNumber string) ([]models.SchoolStatistic, error) {
	return s.repo.GetBySchoolNumber(ctx, schoolNumber)
}

// GetStatisticsBySchoolYear returns all statistics for a specific school year
func (s *StatisticService) GetStatisticsBySchoolYear(ctx context.Context, schoolYear string) ([]models.SchoolStatistic, error) {
	return s.repo.GetBySchoolYear(ctx, schoolYear)
}

// GetStatisticsSummary returns summary information about statistics
func (s *StatisticService) GetStatisticsSummary(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStatisticsSummary(ctx)
}

// ScrapeAndStoreStatistics scrapes statistics from the website and stores them in the database
func (s *StatisticService) ScrapeAndStoreStatistics(ctx context.Context) error {
	s.logger.Info("starting statistics scrape and store")

	// Scrape the data
	statistics, err := s.scraper.ScrapeStatistics(ctx)
	if err != nil {
		s.logger.Error("failed to scrape statistics", slog.String("error", err.Error()))
		return fmt.Errorf("scrape statistics: %w", err)
	}

	if len(statistics) == 0 {
		s.logger.Warn("no statistics scraped")
		return fmt.Errorf("no statistics found")
	}

	s.logger.Info("scraped statistics", slog.Int("count", len(statistics)))

	// Save to database using bulk insert
	saved, err := s.repo.BulkCreateOrUpdate(ctx, statistics)
	if err != nil {
		s.logger.Error("failed to save statistics", slog.String("error", err.Error()))
		return fmt.Errorf("save statistics: %w", err)
	}

	s.logger.Info("statistics saved successfully",
		slog.Int("saved", saved),
		slog.Int("total", len(statistics)),
	)

	return nil
}

// RefreshStatisticsData is an alias for ScrapeAndStoreStatistics for consistency with SchoolService
func (s *StatisticService) RefreshStatisticsData(ctx context.Context) error {
	return s.ScrapeAndStoreStatistics(ctx)
}
