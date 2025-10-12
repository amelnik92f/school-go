# Geocoding Implementation for Construction Projects

## Overview

Construction projects are now enriched with geographic coordinates (latitude and longitude) by geocoding their addresses using the Nominatim OpenStreetMap API.

## Features

### 1. Geocoder Utility (`internal/utils/geocode.go`)

- **Rate Limiting**: Enforces 1 request per 1.1 seconds to respect Nominatim's usage policy
- **Error Handling**: Two methods available:
  - `GeocodeAddress()`: Returns error for explicit error handling
  - `GeocodeAddressSafe()`: Logs errors and returns nil (used in batch operations)
- **Address Format**: `{Street}, {PostalCode} {City}` (e.g., "Alexanderplatz, 10178 Berlin")
- **Country Restriction**: Searches limited to Germany (`countrycodes=de`)

### 2. Database Schema

Added fields to `construction_projects` table:
```sql
latitude REAL,
longitude REAL
```

### 3. Service Integration

When `FetchAndStoreConstructionProjects()` is called:
1. Fetches construction projects from Berlin API
2. Fetches all existing schools to build a lookup set of valid school numbers
3. Categorizes projects:
   - **Projects with existing school numbers**: Skipped (belong to existing schools with coordinates)
   - **Standalone/orphaned projects**: Geocoded (projects without school numbers OR with school numbers that don't exist in the schools table)
4. For each standalone/orphaned project:
   - Builds address string from street, postal code, and city
   - Geocodes the address (with rate limiting)
   - Stores coordinates along with other project data
5. Logs progress every 10 geocoded projects
6. Logs final statistics:
   - Total projects
   - Projects skipped (with existing school numbers)
   - Standalone projects successfully geocoded
   - Failed geocoding attempts

### 4. Logging

Detailed logging at multiple levels:
- **INFO**: Start, progress (every 10 projects), completion with statistics
- **DEBUG**: Each successful geocoding with coordinates
- **WARN**: Failed geocoding attempts with address

## API Used

**Nominatim OpenStreetMap API**
- Endpoint: `https://nominatim.openstreetmap.org/search`
- Format: JSON
- Limit: 1 result per query
- Country: Germany only
- User-Agent: "Berlin Schools Go Backend" (required by Nominatim)

## Rate Limiting

The geocoder enforces a 1.1 second delay between requests using Go's `time.Tick()`:
```go
rateLimiter: time.Tick(1100 * time.Millisecond)
```

This ensures compliance with Nominatim's usage policy while being slightly conservative to account for network latency.

## Usage Example

```go
// Service automatically geocodes when fetching
err := schoolService.FetchAndStoreConstructionProjects(ctx)

// Manual geocoding
geocoder := utils.NewGeocoder()
coords, err := geocoder.GeocodeAddress("Alexanderplatz, 10178 Berlin")
if err != nil {
    log.Printf("Failed to geocode: %v", err)
}
fmt.Printf("Lat: %f, Lon: %f\n", coords.Latitude, coords.Longitude)
```

## Performance Considerations

- **Selective Geocoding**: Only standalone/orphaned projects are geocoded:
  - Projects without school numbers
  - Projects with school numbers that don't exist in the schools table (orphaned)
  - Projects with existing school numbers are skipped instantly
- **Time**: Depends on the number of standalone/orphaned projects
  - ~10-20 standalone projects: ~20-40 seconds
  - ~50-100 standalone projects: ~1-2 minutes
- **Rate Limiting**: Cannot be parallelized due to Nominatim's rate limits (1 req/1.1s)
- **Failures**: Projects that fail geocoding are still stored (with 0,0 coordinates) and logged

## Testing

Run geocoding tests:
```bash
go test ./internal/utils/... -v
```

Note: Tests make actual API calls and will take several seconds due to rate limiting.

