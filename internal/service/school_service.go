package service

import (
	"context"
	"log/slog"

	apperrors "schools-be/internal/errors"
	"schools-be/internal/fetcher"
	"schools-be/internal/models"
	"schools-be/internal/repository"
)

type SchoolService struct {
	repo    *repository.SchoolRepository
	fetcher *fetcher.SchoolFetcher
	logger  *slog.Logger
}

func NewSchoolService(repo *repository.SchoolRepository, fetcher *fetcher.SchoolFetcher) *SchoolService {
	return &SchoolService{
		repo:    repo,
		fetcher: fetcher,
		logger:  slog.Default(),
	}
}

// GetAllSchools returns all schools from the database
func (s *SchoolService) GetAllSchools(ctx context.Context) ([]models.School, error) {
	return s.repo.GetAll(ctx)
}

// GetSchoolByID returns a school by its ID
func (s *SchoolService) GetSchoolByID(ctx context.Context, id int64) (*models.School, error) {
	school, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return school, nil
}

// GetSchoolsByType returns schools filtered by type
func (s *SchoolService) GetSchoolsByType(ctx context.Context, schoolType string) ([]models.School, error) {
	return s.repo.GetByType(ctx, schoolType)
}

// CreateSchool creates a new school
func (s *SchoolService) CreateSchool(ctx context.Context, input models.CreateSchoolInput) (*models.School, error) {
	return s.repo.Create(ctx, input)
}

// UpdateSchool updates an existing school
func (s *SchoolService) UpdateSchool(ctx context.Context, id int64, input models.UpdateSchoolInput) (*models.School, error) {
	// Check if school exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, input)
}

// DeleteSchool deletes a school
func (s *SchoolService) DeleteSchool(ctx context.Context, id int64) error {
	// Check if school exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// RefreshSchoolsData fetches fresh data and updates the database
// This is typically called by the scheduler
func (s *SchoolService) RefreshSchoolsData(ctx context.Context) error {
	s.logger.Info("starting school data refresh")

	// Fetch schools from external sources
	schools, err := s.fetcher.FetchSchools()
	if err != nil {
		s.logger.Error("failed to fetch schools", slog.String("error", err.Error()))
		return apperrors.NewDatabaseError("fetch schools", err)
	}

	// Clear existing data (or implement upsert logic)
	if err := s.repo.DeleteAll(ctx); err != nil {
		s.logger.Error("failed to clear existing schools", slog.String("error", err.Error()))
		return err
	}

	// Insert new data
	successCount := 0
	for _, school := range schools {
		_, err := s.repo.Create(ctx, school)
		if err != nil {
			s.logger.Warn("failed to create school",
				slog.String("school_name", school.Name),
				slog.String("error", err.Error()),
			)
			continue
		}
		successCount++
	}

	s.logger.Info("school data refresh completed",
		slog.Int("success_count", successCount),
		slog.Int("total_count", len(schools)),
	)
	return nil
}
