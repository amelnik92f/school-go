package fetcher

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"schools-be/internal/models"
	"time"
)

const (
	wfsBaseURL         = "https://gdi.berlin.de/services/wfs/schulen"
	wfsVersion         = "2.0.0"
	defaultTypenames   = "fis:schulen"
	constructionAPIURL = "https://www.berlin.de/sen/bildung/schule/bauen-und-sanieren/schulbaukarte/index.php/index/all.json?q="
)

// SchoolFetcher fetches school data from external sources
type SchoolFetcher struct {
	httpClient *http.Client
	typenames  string
}

func NewSchoolFetcher() *SchoolFetcher {
	typenames := os.Getenv("WFS_TYPENAMES")
	if typenames == "" {
		typenames = defaultTypenames
	}

	// Create HTTP client with custom transport to disable HTTP/2 (force HTTP/1.1)
	// This fixes issues with some Berlin APIs that don't handle HTTP/2 properly
	transport := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	return &SchoolFetcher{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		typenames: typenames,
	}
}

// GeoJSON response structures for Berlin WFS service

// SchoolProperties represents the properties of a school from the WFS API
type SchoolProperties struct {
	BSN       string `json:"bsn"`       // School number
	Schulname string `json:"schulname"` // School name
	Schulart  string `json:"schulart"`  // School type
	Traeger   string `json:"traeger"`   // Operator (public/private)
	Schultyp  string `json:"schultyp"`  // School category
	Bezirk    string `json:"bezirk"`    // District
	Ortsteil  string `json:"ortsteil"`  // Neighborhood
	PLZ       string `json:"plz"`       // Postal code
	Strasse   string `json:"strasse"`   // Street name
	Hausnr    string `json:"hausnr"`    // House number
	Telefon   string `json:"telefon"`   // Phone number
	Fax       string `json:"fax"`       // Fax number
	Email     string `json:"email"`     // Email address
	Internet  string `json:"internet"`  // Website URL
	Schuljahr string `json:"schuljahr"` // School year
}

// SchoolGeometry represents the geometry of a school feature
type SchoolGeometry struct {
	Type        string     `json:"type"`        // Should be "Point"
	Coordinates [2]float64 `json:"coordinates"` // [longitude, latitude]
}

// SchoolFeature represents a single school feature in the GeoJSON
type SchoolFeature struct {
	Type         string           `json:"type"`          // Should be "Feature"
	ID           string           `json:"id"`            // Feature ID
	Geometry     SchoolGeometry   `json:"geometry"`      // Geometry data
	GeometryName string           `json:"geometry_name"` // Geometry field name
	Properties   SchoolProperties `json:"properties"`    // School properties
	BBox         [4]float64       `json:"bbox"`          // Bounding box
}

// SchoolsGeoJSON represents the complete GeoJSON response from WFS
type SchoolsGeoJSON struct {
	Type           string          `json:"type"`           // Should be "FeatureCollection"
	Features       []SchoolFeature `json:"features"`       // Array of school features
	TotalFeatures  int             `json:"totalFeatures"`  // Total number of features
	NumberMatched  int             `json:"numberMatched"`  // Number matching query
	NumberReturned int             `json:"numberReturned"` // Number in this response
	TimeStamp      string          `json:"timeStamp"`      // ISO timestamp
	CRS            struct {
		Type       string `json:"type"`
		Properties struct {
			Name string `json:"name"`
		} `json:"properties"`
	} `json:"crs"` // Coordinate reference system
	BBox [4]float64 `json:"bbox"` // Overall bounding box
}

// FetchBerlinSchools fetches all schools data from the Berlin WFS service
func (f *SchoolFetcher) FetchBerlinSchools() (*SchoolsGeoJSON, error) {
	// Build the WFS URL
	params := url.Values{}
	params.Set("SERVICE", "WFS")
	params.Set("VERSION", wfsVersion)
	params.Set("REQUEST", "GetFeature")
	params.Set("TYPENAMES", f.typenames)
	params.Set("SRSNAME", "EPSG:4326")
	params.Set("OUTPUTFORMAT", "application/json")

	requestURL := fmt.Sprintf("%s?%s", wfsBaseURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	log.Println("Fetching schools from Berlin WFS service...")
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schools: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch schools: %d %s", resp.StatusCode, resp.Status)
	}

	// Parse JSON response
	var geoJSON SchoolsGeoJSON
	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate response structure
	if geoJSON.Features == nil {
		return nil, fmt.Errorf("invalid response format: missing features array")
	}

	log.Printf("Successfully fetched %d schools from WFS service", len(geoJSON.Features))
	return &geoJSON, nil
}

// FetchSchools fetches schools from external sources and converts them to CreateSchoolInput
func (f *SchoolFetcher) FetchSchools() ([]models.CreateSchoolInput, error) {
	geoJSON, err := f.FetchBerlinSchools()
	if err != nil {
		return nil, err
	}

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

	log.Printf("Converted %d schools to CreateSchoolInput", len(schools))
	return schools, nil
}

// FetchSchoolDetails fetches detailed information for a specific school
func (f *SchoolFetcher) FetchSchoolDetails(schoolID string) (*models.CreateSchoolInput, error) {
	// TODO: Implement fetching school details from external API
	return nil, fmt.Errorf("not implemented")
}

// Construction Projects API structures

// ConstructionProject represents a single construction project from the Berlin API
type ConstructionProject struct {
	ID                           int    `json:"id"`
	SchoolNumber                 string `json:"schulnummer"`                    // School number (BSN)
	SchoolName                   string `json:"schulname"`                      // School name
	District                     string `json:"bezirk"`                         // District
	SchoolType                   string `json:"schulart"`                       // School type
	ConstructionMeasure          string `json:"baumassnahme"`                   // Construction measure (e.g., "Sanierung; Erweiterung")
	Description                  string `json:"beschreibung"`                   // Description of the construction
	BuiltSchoolPlaces            string `json:"gebaute_schulplaetze"`           // Built school places
	PlacesAfterConstruction      string `json:"schulplaetze_nach_baumassnahme"` // School places after construction
	ClassTracksAfterConstruction string `json:"zuegigkeit_nach_baumassnahme"`   // Class tracks after construction
	HandoverDate                 string `json:"nutzungsuebergabe"`              // Handover date (e.g., "2027/2028")
	TotalCosts                   string `json:"gesamtkosten"`                   // Total costs
	Street                       string `json:"strasse"`                        // Street
	PostalCode                   string `json:"plz"`                            // Postal code
	City                         string `json:"ort"`                            // City
}

// ConstructionProjectsResponse represents the full response from the construction API
type ConstructionProjectsResponse struct {
	Messages struct {
		Messages []string `json:"messages"`
		Success  bool     `json:"success"`
	} `json:"messages"`
	Results struct {
		Count        int `json:"count"`
		ItemsPerPage int `json:"items_per_page"`
	} `json:"results"`
	Index []ConstructionProject `json:"index"`
}

// FetchConstructionProjects fetches all construction projects from the Berlin school construction API
func (f *SchoolFetcher) FetchConstructionProjects() (*ConstructionProjectsResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", constructionAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	log.Println("Fetching construction projects from Berlin API...")
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch construction projects: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch construction projects: %d %s", resp.StatusCode, resp.Status)
	}

	// Parse JSON response
	var data ConstructionProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate response structure
	if data.Index == nil {
		return nil, fmt.Errorf("invalid response format: missing index array")
	}

	log.Printf("Successfully fetched %d construction projects", len(data.Index))
	return &data, nil
}
