# Construction Projects API

## Overview

The Construction Projects API provides access to school construction and renovation projects in Berlin. These projects include information about ongoing and planned construction work, including costs, timelines, and capacity changes.

## Database Status

- **Total construction projects**: 540
- **Valid standalone/orphaned projects**: 19
  - These are projects with a `school_number` that doesn't exist in the schools table
  - Could be new schools under construction, data quality issues, or schools that were removed
  - **Excluded**: Meta entries, legends, and projects with empty `school_name` (2 invalid projects)
  - Most projects (~519) are properly linked to existing schools via `school_number`

## Endpoints

### Get All Construction Projects

```
GET /api/v1/construction-projects
```

Returns all construction projects, both standalone and school-linked.

**Response**: Array of `ConstructionProject` objects

**Example**:
```bash
curl http://localhost:8080/api/v1/construction-projects
```

### Get Standalone Construction Projects

```
GET /api/v1/construction-projects/standalone
```

Returns only **valid** construction projects that are **not assigned to any existing school**. 

**Includes**: 
- Orphaned projects where `school_number` doesn't exist in the schools table
- Only projects with meaningful data (non-empty `school_name`)

**Excludes**:
- Projects with empty `school_number` and empty `school_name` (invalid entries)
- Meta entries and legends (e.g., "Legende: Orange bedeutet...")

**Response**: Array of `ConstructionProject` objects

**Example**:
```bash
curl http://localhost:8080/api/v1/construction-projects/standalone
```

**Use Case**: This endpoint is useful for:
- Finding construction projects for **new schools** not yet in the system (e.g., "Gymnasium Schulstraße" with number "01Yn01")
- Identifying **orphaned data** where the school was removed but the construction project remains
- Discovering **data quality issues** (projects that should be linked but aren't)

**Technical Implementation**: 
```sql
-- Uses LEFT JOIN to check if school_number exists in schools table
-- Filters out entries with empty school_name or legend entries
WHERE school_number != '' 
  AND s.school_number IS NULL
  AND school_name != ''
  AND school_name NOT LIKE 'Legende:%'
```

### Get Single Construction Project

```
GET /api/v1/construction-projects/{id}
```

Returns a single construction project by its database ID.

**Parameters**:
- `id` (path parameter): The project's database ID (not the `project_id`)

**Response**: Single `ConstructionProject` object

**Example**:
```bash
curl http://localhost:8080/api/v1/construction-projects/42
```

## Response Structure

### ConstructionProject Object

```json
{
  "id": 123,
  "project_id": 456,
  "school_number": "01G01",
  "school_name": "Example Gymnasium",
  "district": "Mitte",
  "school_type": "Gymnasium",
  "construction_measure": "Sanierung",
  "description": "Dachsanierung und Erneuerung der Fenster",
  "built_school_places": "600",
  "places_after_construction": "600",
  "class_tracks_after_construction": "4",
  "handover_date": "Q4 2025",
  "total_costs": "2.5 Mio. EUR",
  "street": "Musterstraße 1",
  "postal_code": "13353",
  "city": "Berlin",
  "latitude": 52.5200,
  "longitude": 13.4050,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Database primary key |
| `project_id` | integer | Original project ID from Berlin's construction API |
| `school_number` | string | School number (BSN) - empty for standalone projects |
| `school_name` | string | Name of the school |
| `district` | string | Berlin district (Bezirk) |
| `school_type` | string | Type of school (e.g., "Gymnasium", "Grundschule") |
| `construction_measure` | string | Type of construction measure |
| `description` | string | Detailed description of the construction work |
| `built_school_places` | string | Number of school places built |
| `places_after_construction` | string | Total school places after construction |
| `class_tracks_after_construction` | string | Number of class tracks (Züge) after construction |
| `handover_date` | string | Expected completion/handover date |
| `total_costs` | string | Total project costs |
| `street` | string | Street address |
| `postal_code` | string | Postal code |
| `city` | string | City (typically "Berlin") |
| `latitude` | float | Latitude coordinate (geocoded for standalone projects) |
| `longitude` | float | Longitude coordinate (geocoded for standalone projects) |
| `created_at` | timestamp | Record creation time |
| `updated_at` | timestamp | Record last update time |

## Geocoding Notes

- Projects **linked to schools** inherit coordinates from the school
- **Standalone projects** are geocoded based on their address (`street`, `postal_code`, `city`)
- If geocoding fails, `latitude` and `longitude` will be `0.0`

## Integration with Schools API

Construction projects are included in the enriched schools endpoint:

```
GET /api/v1/schools
```

Each school in the response includes a `construction_projects` array with all projects linked to that school via `school_number`.

## Error Responses

### 404 Not Found
```json
{
  "error": "construction project not found"
}
```

### 400 Bad Request
```json
{
  "error": "invalid project id"
}
```

### 500 Internal Server Error
```json
{
  "error": "failed to retrieve construction projects"
}
```

## Example Usage

### Find all standalone construction projects
```bash
curl http://localhost:8080/api/v1/construction-projects/standalone | jq '.'
```

### Find projects in a specific district
```bash
curl http://localhost:8080/api/v1/construction-projects | jq '.[] | select(.district == "Mitte")'
```

### Find expensive projects (requires jq with string parsing)
```bash
curl http://localhost:8080/api/v1/construction-projects | jq '.[] | select(.total_costs != "")'
```

### Check geocoding success rate for standalone projects
```bash
curl http://localhost:8080/api/v1/construction-projects/standalone | \
  jq '[.[] | select(.latitude != 0 and .longitude != 0)] | length'
```

## Data Updates

Construction project data is refreshed when you call:

```bash
POST /api/v1/refresh
```

This endpoint:
1. Fetches latest construction data from Berlin's API
2. Geocodes standalone projects
3. Updates the database

The refresh happens automatically on a schedule (configured in the application), but can be triggered manually for testing or immediate updates.

