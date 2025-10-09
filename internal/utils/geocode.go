package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// GeocodeResult represents a single result from Nominatim API
type GeocodeResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Coordinates represents latitude and longitude
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// Geocoder handles geocoding requests with rate limiting
type Geocoder struct {
	httpClient *http.Client
	userAgent  string
	logger     *slog.Logger
	// Rate limiter: channel to enforce 1 request per second
	rateLimiter <-chan time.Time
}

// NewGeocoder creates a new Geocoder instance with rate limiting (1 req/sec)
func NewGeocoder() *Geocoder {
	return &Geocoder{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent:   "Berlin Schools Go Backend",
		logger:      slog.Default(),
		rateLimiter: time.Tick(1100 * time.Millisecond), // 1.1 seconds between requests
	}
}

// GeocodeAddress geocodes a single address using Nominatim API
// Returns coordinates or nil if geocoding fails
func (g *Geocoder) GeocodeAddress(address string) (*Coordinates, error) {
	// Wait for rate limiter
	<-g.rateLimiter

	// Build the API URL
	apiURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?format=json&q=%s&limit=1&countrycodes=de",
		url.QueryEscape(address),
	)

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header (required by Nominatim)
	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var results []GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", address)
	}

	// Parse coordinates
	var lat, lon float64
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &lat); err != nil {
		return nil, fmt.Errorf("failed to parse latitude: %w", err)
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &lon); err != nil {
		return nil, fmt.Errorf("failed to parse longitude: %w", err)
	}

	return &Coordinates{
		Latitude:  lat,
		Longitude: lon,
	}, nil
}

// GeocodeAddressSafe geocodes an address and logs errors instead of returning them
// Returns nil if geocoding fails
func (g *Geocoder) GeocodeAddressSafe(address string) *Coordinates {
	coords, err := g.GeocodeAddress(address)
	if err != nil {
		g.logger.Warn("failed to geocode address",
			slog.String("address", address),
			slog.String("error", err.Error()),
		)
		return nil
	}
	return coords
}
