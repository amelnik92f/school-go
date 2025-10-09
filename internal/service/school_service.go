package service

import (
	"fmt"
	"log"

	"schools-be/internal/fetcher"
	"schools-be/internal/models"
	"schools-be/internal/repository"
)

type SchoolService struct {
	repo    *repository.SchoolRepository
	fetcher *fetcher.SchoolFetcher
}

func NewSchoolService(repo *repository.SchoolRepository, fetcher *fetcher.SchoolFetcher) *SchoolService {
	return &SchoolService{
		repo:    repo,
		fetcher: fetcher,
	}
}

// GetAllSchools returns all schools from the database
func (s *SchoolService) GetAllSchools() ([]models.School, error) {
	return s.repo.GetAll()
}

// GetSchoolByID returns a school by its ID
func (s *SchoolService) GetSchoolByID(id int64) (*models.School, error) {
	school, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if school == nil {
		return nil, fmt.Errorf("school not found")
	}
	return school, nil
}

// GetSchoolsByType returns schools filtered by type
func (s *SchoolService) GetSchoolsByType(schoolType string) ([]models.School, error) {
	return s.repo.GetByType(schoolType)
}

// CreateSchool creates a new school
func (s *SchoolService) CreateSchool(input models.CreateSchoolInput) (*models.School, error) {
	return s.repo.Create(input)
}

// UpdateSchool updates an existing school
func (s *SchoolService) UpdateSchool(id int64, input models.UpdateSchoolInput) (*models.School, error) {
	// Check if school exists
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("school not found")
	}

	return s.repo.Update(id, input)
}

// DeleteSchool deletes a school
func (s *SchoolService) DeleteSchool(id int64) error {
	// Check if school exists
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("school not found")
	}

	return s.repo.Delete(id)
}

// RefreshSchoolsData fetches fresh data and updates the database
// This is typically called by the scheduler
func (s *SchoolService) RefreshSchoolsData() error {
	log.Println("Starting school data refresh...")

	// Fetch schools from external sources
	schools, err := s.fetcher.FetchSchools()
	if err != nil {
		return fmt.Errorf("failed to fetch schools: %w", err)
	}

	// Clear existing data (or implement upsert logic)
	if err := s.repo.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear existing schools: %w", err)
	}

	// Insert new data
	successCount := 0
	for _, school := range schools {
		_, err := s.repo.Create(school)
		if err != nil {
			log.Printf("Failed to create school %s: %v", school.Name, err)
			continue
		}
		successCount++
	}

	log.Printf("School data refresh completed: %d/%d schools imported", successCount, len(schools))
	return nil
}

