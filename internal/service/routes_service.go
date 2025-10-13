package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"schools-be/internal/config"
)

const (
	openRouteServiceBaseURL = "https://api.openrouteservice.org/v2"
)

// Map our internal mode names to OpenRouteService profiles
var profileMap = map[string]string{
	"walking": "foot-walking",
	"bicycle": "cycling-regular",
	"car":     "driving-car",
}

type RoutesService struct {
	config     *config.Config
	httpClient *http.Client
}

func NewRoutesService(config *config.Config) *RoutesService {
	return &RoutesService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type TravelTimeRequest struct {
	Start [2]float64 `json:"start"` // [lng, lat]
	End   [2]float64 `json:"end"`   // [lng, lat]
	Modes []string   `json:"modes"` // Array of mode names
}

type TravelTimeResponse struct {
	Mode            string  `json:"mode"`
	DurationMinutes int     `json:"durationMinutes"`
	DistanceKm      float64 `json:"distanceKm"`
	Error           string  `json:"error,omitempty"`
}

type openRouteServiceResponse struct {
	Features []struct {
		Properties struct {
			Summary struct {
				Duration float64 `json:"duration"`
				Distance float64 `json:"distance"`
			} `json:"summary"`
		} `json:"properties"`
	} `json:"features"`
}

// CalculateTravelTimes calculates travel times for multiple modes from start to end
func (s *RoutesService) CalculateTravelTimes(ctx context.Context, req TravelTimeRequest) ([]TravelTimeResponse, error) {
	if s.config.OpenRouteServiceAPIKey == "" {
		return nil, fmt.Errorf("OpenRouteService API key is not configured")
	}

	if len(req.Start) != 2 || len(req.End) != 2 {
		return nil, fmt.Errorf("invalid coordinates format")
	}

	if len(req.Modes) == 0 {
		return nil, fmt.Errorf("at least one travel mode is required")
	}

	results := make([]TravelTimeResponse, 0, len(req.Modes))

	// Process each mode sequentially (to respect API rate limits)
	for _, mode := range req.Modes {
		result := s.fetchTravelTime(ctx, req.Start, req.End, mode)
		results = append(results, result)
	}

	return results, nil
}

func (s *RoutesService) fetchTravelTime(ctx context.Context, start, end [2]float64, mode string) TravelTimeResponse {
	profile, ok := profileMap[mode]
	if !ok {
		profile = "foot-walking" // default
	}

	url := fmt.Sprintf(
		"%s/directions/%s?start=%f,%f&end=%f,%f",
		openRouteServiceBaseURL,
		profile,
		start[0], start[1],
		end[0], end[1],
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return TravelTimeResponse{
			Mode:  mode,
			Error: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", s.config.OpenRouteServiceAPIKey)
	req.Header.Set("Accept", "application/geo+json;charset=UTF-8")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return TravelTimeResponse{
			Mode:  mode,
			Error: fmt.Sprintf("failed to fetch route: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TravelTimeResponse{
			Mode:  mode,
			Error: fmt.Sprintf("API error: %d - %s", resp.StatusCode, string(body)),
		}
	}

	var orsResp openRouteServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&orsResp); err != nil {
		return TravelTimeResponse{
			Mode:  mode,
			Error: fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	// Extract duration and distance from the response
	if len(orsResp.Features) == 0 {
		return TravelTimeResponse{
			Mode:  mode,
			Error: "no route found",
		}
	}

	durationSeconds := orsResp.Features[0].Properties.Summary.Duration
	distanceMeters := orsResp.Features[0].Properties.Summary.Distance

	return TravelTimeResponse{
		Mode:            mode,
		DurationMinutes: int(durationSeconds / 60),
		DistanceKm:      float64(int(distanceMeters/100)) / 10, // Round to 1 decimal
	}
}
