package service

import (
	"context"
	"log/slog"

	"schools-be/internal/models"
	"schools-be/internal/repository"
)

type ConstructionProjectService struct {
	repo   *repository.ConstructionProjectRepository
	logger *slog.Logger
}

func NewConstructionProjectService(repo *repository.ConstructionProjectRepository) *ConstructionProjectService {
	return &ConstructionProjectService{
		repo:   repo,
		logger: slog.Default(),
	}
}

// GetAll returns all construction projects
func (s *ConstructionProjectService) GetAll(ctx context.Context) ([]models.ConstructionProject, error) {
	return s.repo.GetAll(ctx)
}

// GetByID returns a single construction project by ID
func (s *ConstructionProjectService) GetByID(ctx context.Context, id int64) (*models.ConstructionProject, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySchoolNumber returns construction projects for a specific school
func (s *ConstructionProjectService) GetBySchoolNumber(ctx context.Context, schoolNumber string) ([]models.ConstructionProject, error) {
	return s.repo.GetBySchoolNumber(ctx, schoolNumber)
}

// GetStandalone returns valid construction projects that are not assigned to any existing school
// Only includes orphaned projects where school_number doesn't exist in the schools table
// Excludes meta entries, legends, and projects with no meaningful data (empty school_name)
func (s *ConstructionProjectService) GetStandalone(ctx context.Context) ([]models.ConstructionProject, error) {
	return s.repo.GetStandalone(ctx)
}
