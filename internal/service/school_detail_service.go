package service

import (
	"context"
	"fmt"
	"log/slog"

	"schools-be/internal/models"
	"schools-be/internal/repository"
	"schools-be/internal/scraper"
)

type SchoolDetailService struct {
	repo      *repository.SchoolDetailRepository
	statsRepo *repository.SchoolStatisticsRepository
	scraper   *scraper.SchoolDetailsScraper
	logger    *slog.Logger
}

func NewSchoolDetailService(repo *repository.SchoolDetailRepository, statsRepo *repository.SchoolStatisticsRepository, scraper *scraper.SchoolDetailsScraper) *SchoolDetailService {
	return &SchoolDetailService{
		repo:      repo,
		statsRepo: statsRepo,
		scraper:   scraper,
		logger:    slog.Default(),
	}
}

// ScrapeAndStoreDetails scrapes school details and stores them in the database
func (s *SchoolDetailService) ScrapeAndStoreDetails(ctx context.Context) error {
	s.logger.Info("starting school details scrape and store")

	// Scrape details from website
	details, err := s.scraper.ScrapeSchoolDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to scrape school details: %w", err)
	}

	s.logger.Info("scraped school details", slog.Int("count", len(details)))

	// Store each detail in database
	successCount := 0
	errorCount := 0

	for i, detail := range details {
		s.logger.Info("storing school detail",
			slog.Int("index", i+1),
			slog.Int("total", len(details)),
			slog.String("school", detail.SchoolName),
		)

		err := s.repo.Upsert(ctx, &detail)
		if err != nil {
			s.logger.Error("failed to store school detail",
				slog.String("school", detail.SchoolName),
				slog.String("error", err.Error()),
			)
			errorCount++
			continue
		}

		// Store normalized statistics
		if err := s.saveNormalizedStatistics(ctx, &detail); err != nil {
			s.logger.Warn("failed to store normalized statistics",
				slog.String("school", detail.SchoolName),
				slog.String("error", err.Error()),
			)
			// Don't fail the whole operation, just log the warning
		}

		successCount++
	}

	s.logger.Info("finished storing school details",
		slog.Int("success", successCount),
		slog.Int("errors", errorCount),
	)

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors out of %d schools", errorCount, len(details))
	}

	return nil
}

// GetAll retrieves all school details
func (s *SchoolDetailService) GetAll(ctx context.Context) ([]models.SchoolDetail, error) {
	return s.repo.GetAll(ctx)
}

// GetBySchoolNumber retrieves school details by school number
func (s *SchoolDetailService) GetBySchoolNumber(ctx context.Context, schoolNumber string) (*models.SchoolDetail, error) {
	return s.repo.GetBySchoolNumber(ctx, schoolNumber)
}

// GetAvailableAfter4thGrade retrieves schools available after 4th grade
func (s *SchoolDetailService) GetAvailableAfter4thGrade(ctx context.Context) ([]models.SchoolDetail, error) {
	return s.repo.GetAvailableAfter4thGrade(ctx)
}

// GetSummary returns a summary of school details in the database
func (s *SchoolDetailService) GetSummary(ctx context.Context) (map[string]interface{}, error) {
	totalCount, err := s.repo.GetCount(ctx)
	if err != nil {
		return nil, err
	}

	availableAfter4th, err := s.repo.GetAvailableAfter4thGrade(ctx)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_schools":             totalCount,
		"available_after_4th_grade": len(availableAfter4th),
		"not_available_after_4th":   totalCount - len(availableAfter4th),
	}

	return summary, nil
}

// DeleteAll deletes all school details
func (s *SchoolDetailService) DeleteAll(ctx context.Context) error {
	return s.repo.DeleteAll(ctx)
}

// ClearCache clears the scraper cache
func (s *SchoolDetailService) ClearCache() error {
	return s.scraper.ClearCache()
}

// saveNormalizedStatistics saves the normalized statistics tables
func (s *SchoolDetailService) saveNormalizedStatistics(ctx context.Context, detail *models.SchoolDetailData) error {
	if detail.SchoolNumber == "" {
		return nil
	}

	// Normalize and save citizenship stats
	if detail.CitizenshipTable != nil {
		citizenshipStats := scraper.NormalizeCitizenshipTable(detail.SchoolNumber, detail.CitizenshipTable, detail.ScrapedAt)
		if len(citizenshipStats) > 0 {
			if err := s.statsRepo.SaveCitizenshipStats(ctx, citizenshipStats); err != nil {
				s.logger.Warn("failed to save citizenship stats",
					slog.String("school", detail.SchoolNumber),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	// Normalize and save language stats
	if detail.LanguageTable != nil {
		languageStat := scraper.NormalizeLanguageTable(detail.SchoolNumber, detail.LanguageTable, detail.ScrapedAt)
		if languageStat != nil {
			if err := s.statsRepo.SaveLanguageStat(ctx, *languageStat); err != nil {
				s.logger.Warn("failed to save language stats",
					slog.String("school", detail.SchoolNumber),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	// Normalize and save residence stats
	if detail.ResidenceTable != nil {
		residenceStats := scraper.NormalizeResidenceTable(detail.SchoolNumber, detail.ResidenceTable, detail.ScrapedAt)
		if len(residenceStats) > 0 {
			if err := s.statsRepo.SaveResidenceStats(ctx, residenceStats); err != nil {
				s.logger.Warn("failed to save residence stats",
					slog.String("school", detail.SchoolNumber),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	// Normalize and save absence stats
	if detail.AbsenceTable != nil {
		absenceStat := scraper.NormalizeAbsenceTable(detail.SchoolNumber, detail.AbsenceTable, detail.ScrapedAt)
		if absenceStat != nil {
			if err := s.statsRepo.SaveAbsenceStat(ctx, *absenceStat); err != nil {
				s.logger.Warn("failed to save absence stats",
					slog.String("school", detail.SchoolNumber),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return nil
}
