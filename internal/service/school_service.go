package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "schools-be/internal/errors"
	"schools-be/internal/fetcher"
	"schools-be/internal/models"
	"schools-be/internal/repository"
	"schools-be/internal/utils"
)

type SchoolService struct {
	repo             *repository.SchoolRepository
	constructionRepo *repository.ConstructionProjectRepository
	fetcher          *fetcher.SchoolFetcher
	geocoder         *utils.Geocoder
	logger           *slog.Logger
}

func NewSchoolService(repo *repository.SchoolRepository, constructionRepo *repository.ConstructionProjectRepository, fetcher *fetcher.SchoolFetcher) *SchoolService {
	return &SchoolService{
		repo:             repo,
		constructionRepo: constructionRepo,
		fetcher:          fetcher,
		geocoder:         utils.NewGeocoder(),
		logger:           slog.Default(),
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

// FetchAndStoreSchools fetches schools from WFS API and stores them in the database
func (s *SchoolService) FetchAndStoreSchools(ctx context.Context) error {
	s.logger.Info("starting school data fetch")

	// Fetch GeoJSON from WFS
	geoJSON, err := s.fetcher.FetchBerlinSchools()
	if err != nil {
		s.logger.Error("failed to fetch schools", slog.String("error", err.Error()))
		return apperrors.NewDatabaseError("fetch schools", err)
	}

	// Convert to CreateSchoolInput
	schools := make([]models.CreateSchoolInput, 0, len(geoJSON.Features))
	for _, feature := range geoJSON.Features {
		props := feature.Properties
		school := models.CreateSchoolInput{
			SchoolNumber:   props.BSN,
			Name:           props.Schulname,
			SchoolType:     props.Schulart,
			Operator:       props.Traeger,
			SchoolCategory: props.Schultyp,
			District:       props.Bezirk,
			Neighborhood:   props.Ortsteil,
			PostalCode:     props.PLZ,
			Street:         props.Strasse,
			HouseNumber:    props.Hausnr,
			Phone:          props.Telefon,
			Fax:            props.Fax,
			Email:          props.Email,
			Website:        props.Internet,
			SchoolYear:     props.Schuljahr,
			Longitude:      feature.Geometry.Coordinates[0],
			Latitude:       feature.Geometry.Coordinates[1],
		}
		schools = append(schools, school)
	}

	// Clear existing data
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

	s.logger.Info("school data fetch completed",
		slog.Int("success_count", successCount),
		slog.Int("total_count", len(schools)),
	)
	return nil
}

// FetchAndStoreConstructionProjects fetches construction projects from Berlin API and stores them in the database
func (s *SchoolService) FetchAndStoreConstructionProjects(ctx context.Context) error {
	s.logger.Info("starting construction projects data fetch")

	// Fetch construction projects
	response, err := s.fetcher.FetchConstructionProjects()
	if err != nil {
		s.logger.Error("failed to fetch construction projects", slog.String("error", err.Error()))
		return apperrors.NewDatabaseError("fetch construction projects", err)
	}

	// Separate projects into those with and without school numbers
	projectsWithSchools := 0
	standaloneProjects := 0
	for _, proj := range response.Index {
		if proj.SchoolNumber != "" {
			projectsWithSchools++
		} else {
			standaloneProjects++
		}
	}

	s.logger.Info("processing construction projects",
		slog.Int("total", len(response.Index)),
		slog.Int("with_school_number", projectsWithSchools),
		slog.Int("standalone_to_geocode", standaloneProjects),
	)

	// Convert to CreateConstructionProjectInput and geocode only standalone projects
	projects := make([]models.CreateConstructionProjectInput, 0, len(response.Index))
	geocodedCount := 0
	skippedCount := 0

	for _, proj := range response.Index {
		var lat, lon float64

		// Only geocode projects without a school number (standalone projects)
		if proj.SchoolNumber == "" || proj.SchoolNumber == " " {
			// Build address string for geocoding
			address := fmt.Sprintf("%s, %s %s", proj.Street, proj.PostalCode, proj.City)

			// Geocode the address
			coords := s.geocoder.GeocodeAddressSafe(address)

			if coords != nil {
				lat = coords.Latitude
				lon = coords.Longitude
				geocodedCount++
				s.logger.Debug("geocoded standalone construction project",
					slog.Int("project_id", proj.ID),
					slog.String("address", address),
					slog.Float64("lat", lat),
					slog.Float64("lon", lon),
				)
			} else {
				s.logger.Warn("failed to geocode construction project",
					slog.Int("project_id", proj.ID),
					slog.String("address", address),
				)
			}
		} else {
			// Skip geocoding for projects with school numbers
			skippedCount++
		}

		project := models.CreateConstructionProjectInput{
			ProjectID:                    proj.ID,
			SchoolNumber:                 proj.SchoolNumber,
			SchoolName:                   proj.SchoolName,
			District:                     proj.District,
			SchoolType:                   proj.SchoolType,
			ConstructionMeasure:          proj.ConstructionMeasure,
			Description:                  proj.Description,
			BuiltSchoolPlaces:            proj.BuiltSchoolPlaces,
			PlacesAfterConstruction:      proj.PlacesAfterConstruction,
			ClassTracksAfterConstruction: proj.ClassTracksAfterConstruction,
			HandoverDate:                 proj.HandoverDate,
			TotalCosts:                   proj.TotalCosts,
			Street:                       proj.Street,
			PostalCode:                   proj.PostalCode,
			City:                         proj.City,
			Latitude:                     lat,
			Longitude:                    lon,
		}
		projects = append(projects, project)

		// Log progress every 10 geocoding operations (not every project)
		if standaloneProjects > 0 && geocodedCount > 0 && geocodedCount%10 == 0 {
			s.logger.Info("geocoding progress",
				slog.Int("geocoded", geocodedCount),
				slog.Int("standalone_total", standaloneProjects),
			)
		}
	}

	s.logger.Info("construction projects processing completed",
		slog.Int("total_projects", len(response.Index)),
		slog.Int("skipped_with_school_number", skippedCount),
		slog.Int("standalone_geocoded", geocodedCount),
		slog.Int("standalone_failed", standaloneProjects-geocodedCount),
	)

	// Clear existing data
	if err := s.constructionRepo.DeleteAll(ctx); err != nil {
		s.logger.Error("failed to clear existing construction projects", slog.String("error", err.Error()))
		return err
	}

	// Insert new data
	successCount := 0
	for _, project := range projects {
		_, err := s.constructionRepo.Create(ctx, project)
		if err != nil {
			s.logger.Warn("failed to create construction project",
				slog.String("school_name", project.SchoolName),
				slog.String("error", err.Error()),
			)
			continue
		}
		successCount++
	}

	s.logger.Info("construction projects data fetch completed",
		slog.Int("success_count", successCount),
		slog.Int("total_count", len(projects)),
	)
	return nil
}

// RefreshSchoolsData is deprecated, use FetchAndStoreSchools instead
// This is kept for backward compatibility with scheduler
func (s *SchoolService) RefreshSchoolsData(ctx context.Context) error {
	if err := s.FetchAndStoreSchools(ctx); err != nil {
		return err
	}
	if err := s.FetchAndStoreConstructionProjects(ctx); err != nil {
		return err
	}
	return nil
}
