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
	detailRepo       *repository.SchoolDetailRepository
	statsRepo        *repository.SchoolStatisticsRepository
	statisticRepo    *repository.StatisticRepository
	fetcher          *fetcher.SchoolFetcher
	geocoder         *utils.Geocoder
	logger           *slog.Logger
}

func NewSchoolService(
	repo *repository.SchoolRepository,
	constructionRepo *repository.ConstructionProjectRepository,
	detailRepo *repository.SchoolDetailRepository,
	statsRepo *repository.SchoolStatisticsRepository,
	statisticRepo *repository.StatisticRepository,
	fetcher *fetcher.SchoolFetcher,
) *SchoolService {
	return &SchoolService{
		repo:             repo,
		constructionRepo: constructionRepo,
		detailRepo:       detailRepo,
		statsRepo:        statsRepo,
		statisticRepo:    statisticRepo,
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

	// Get all existing school numbers to determine which projects need geocoding
	schools, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to fetch existing schools", slog.String("error", err.Error()))
		return apperrors.NewDatabaseError("fetch existing schools", err)
	}

	// Build a set of existing school numbers for fast lookup
	existingSchoolNumbers := make(map[string]bool)
	for _, school := range schools {
		if school.SchoolNumber != "" {
			existingSchoolNumbers[school.SchoolNumber] = true
		}
	}

	// Separate projects into those that need geocoding vs those that don't
	projectsWithExistingSchools := 0
	standaloneProjects := 0
	for _, proj := range response.Index {
		// A project needs geocoding if:
		// 1. It has no school_number (completely standalone), OR
		// 2. It has a school_number that doesn't exist in the schools table (orphaned)
		if proj.SchoolNumber == "" || proj.SchoolNumber == " " || !existingSchoolNumbers[proj.SchoolNumber] {
			standaloneProjects++
		} else {
			projectsWithExistingSchools++
		}
	}

	s.logger.Info("processing construction projects",
		slog.Int("total", len(response.Index)),
		slog.Int("with_existing_schools", projectsWithExistingSchools),
		slog.Int("standalone_to_geocode", standaloneProjects),
	)

	// Convert to CreateConstructionProjectInput and geocode standalone/orphaned projects
	projects := make([]models.CreateConstructionProjectInput, 0, len(response.Index))
	geocodedCount := 0
	skippedCount := 0

	for _, proj := range response.Index {
		var lat, lon float64

		// Geocode projects that are standalone or have school numbers not in the schools table
		shouldGeocode := proj.SchoolNumber == "" || proj.SchoolNumber == " " || !existingSchoolNumbers[proj.SchoolNumber]

		if shouldGeocode {
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
					slog.String("school_number", proj.SchoolNumber),
					slog.String("address", address),
					slog.Float64("lat", lat),
					slog.Float64("lon", lon),
				)
			} else {
				s.logger.Warn("failed to geocode construction project",
					slog.Int("project_id", proj.ID),
					slog.String("school_number", proj.SchoolNumber),
					slog.String("address", address),
				)
			}
		} else {
			// Skip geocoding for projects with existing school numbers
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

// GetAllSchoolsEnriched returns all schools enriched with details, statistics, and construction projects
func (s *SchoolService) GetAllSchoolsEnriched(ctx context.Context) ([]models.EnrichedSchool, error) {
	// Get all schools
	schools, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich each school with additional data
	enrichedSchools := make([]models.EnrichedSchool, 0, len(schools))
	for _, school := range schools {
		enriched, err := s.enrichSchool(ctx, school)
		if err != nil {
			s.logger.Warn("failed to enrich school",
				slog.String("school_number", school.SchoolNumber),
				slog.String("school_name", school.Name),
				slog.String("error", err.Error()),
			)
			// Include school even if enrichment fails
			enriched = models.EnrichedSchool{
				School: school,
			}
		}
		enrichedSchools = append(enrichedSchools, enriched)
	}

	return enrichedSchools, nil
}

// GetSchoolByIDEnriched returns a single enriched school by its ID
func (s *SchoolService) GetSchoolByIDEnriched(ctx context.Context, id int64) (*models.EnrichedSchool, error) {
	school, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	enriched, err := s.enrichSchool(ctx, *school)
	if err != nil {
		s.logger.Warn("failed to enrich school",
			slog.Int64("school_id", id),
			slog.String("error", err.Error()),
		)
		// Return school without enrichment if enrichment fails
		return &models.EnrichedSchool{School: *school}, nil
	}

	return &enriched, nil
}

// enrichSchool enriches a single school with all related data
func (s *SchoolService) enrichSchool(ctx context.Context, school models.School) (models.EnrichedSchool, error) {
	enriched := models.EnrichedSchool{
		School: school,
	}

	// Fetch school details
	details, err := s.detailRepo.GetBySchoolNumber(ctx, school.SchoolNumber)
	if err != nil {
		// Log but don't fail - some schools might not have details
		s.logger.Debug("no details found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.Details = details
	}

	// Fetch citizenship stats
	citizenshipStats, err := s.statsRepo.GetCitizenshipStats(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no citizenship stats found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.CitizenshipStats = citizenshipStats
	}

	// Fetch language stats
	languageStat, err := s.statsRepo.GetLanguageStat(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no language stats found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.LanguageStat = languageStat
	}

	// Fetch residence stats
	residenceStats, err := s.statsRepo.GetResidenceStats(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no residence stats found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.ResidenceStats = residenceStats
	}

	// Fetch absence stats
	absenceStat, err := s.statsRepo.GetAbsenceStat(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no absence stats found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.AbsenceStat = absenceStat
	}

	// Fetch construction projects
	constructionProjects, err := s.constructionRepo.GetBySchoolNumber(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no construction projects found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.ConstructionProjects = constructionProjects
	}

	// Fetch school statistics
	statistics, err := s.statisticRepo.GetBySchoolNumber(ctx, school.SchoolNumber)
	if err != nil {
		s.logger.Debug("no statistics found for school",
			slog.String("school_number", school.SchoolNumber),
		)
	} else {
		enriched.Statistics = statistics
	}

	return enriched, nil
}
