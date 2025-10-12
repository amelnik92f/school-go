# Enriched Schools API

## Overview

The enriched schools API endpoints provide complete school data by combining information from multiple database tables:

- **Base school data**: Basic information like name, location, contact details
- **School details**: Languages, courses, offerings, equipment, etc.
- **Citizenship statistics**: Student citizenship breakdown by region
- **Language statistics**: Non-German heritage language speakers
- **Residence statistics**: Student residence distribution by district
- **Absence statistics**: Absence and unexcused absence rates
- **School statistics**: Historical data on students, teachers, and classes by school year
- **Construction projects**: Related construction/renovation projects

## Endpoints

### Get All Enriched Schools

```
GET /api/v1/schools
```

Returns all schools with complete enriched data from all related tables.

**Response**: Array of `EnrichedSchool` objects

**Example**:
```bash
curl http://localhost:8080/api/v1/schools
```

### Get Single Enriched School

```
GET /api/v1/schools/{id}
```

Returns a single school by ID with complete enriched data.

**Parameters**:
- `id` (path parameter): The school's database ID

**Response**: Single `EnrichedSchool` object

**Example**:
```bash
curl http://localhost:8080/api/v1/schools/42
```

## Response Structure

### EnrichedSchool Object

```json
{
  "school": {
    "id": 1,
    "school_number": "01G01",
    "name": "Example Gymnasium",
    "school_type": "Gymnasium",
    "operator": "öffentlich",
    "school_category": "Gymnasium",
    "district": "Mitte",
    "neighborhood": "Wedding",
    "postal_code": "13353",
    "street": "Musterstraße",
    "house_number": "1",
    "phone": "030-12345678",
    "fax": "030-12345679",
    "email": "info@example-schule.de",
    "website": "https://example-schule.de",
    "school_year": "2025/26",
    "latitude": 52.5200,
    "longitude": 13.4050,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "details": {
    "id": 1,
    "school_number": "01G01",
    "school_name": "Example Gymnasium",
    "languages": "Englisch, Französisch, Spanisch",
    "courses": "Mathematik, Physik, Chemie, Biologie",
    "offerings": "Musik, Kunst, Sport",
    "available_after_4th_grade": true,
    "additional_info": "MINT-freundliche Schule",
    "equipment": "Smartboards, Computerräume, Sporthalle",
    "working_groups": "Schach, Robotik, Theater",
    "partners": "TU Berlin, Max-Planck-Institut",
    "differentiation": "Begabtenförderung, Nachhilfe",
    "lunch_info": "Mensa mit Bio-Essen",
    "dual_learning": "Praktika ab Klasse 9",
    "citizenship_data": "{}",
    "language_data": "{}",
    "residence_data": "{}",
    "absence_data": "{}",
    "scraped_at": "2025-01-15T12:00:00Z",
    "created_at": "2025-01-15T12:00:00Z",
    "updated_at": "2025-01-15T12:00:00Z"
  },
  "citizenship_stats": [
    {
      "id": 1,
      "school_number": "01G01",
      "citizenship": "Deutschland",
      "female_students": 120,
      "male_students": 115,
      "total": 235,
      "scraped_at": "2025-01-15T12:00:00Z",
      "created_at": "2025-01-15T12:00:00Z"
    }
  ],
  "language_stat": {
    "id": 1,
    "school_number": "01G01",
    "total_students": 500,
    "ndh_female_students": 85,
    "ndh_male_students": 78,
    "ndh_total": 163,
    "ndh_percentage": 32.6,
    "scraped_at": "2025-01-15T12:00:00Z",
    "created_at": "2025-01-15T12:00:00Z"
  },
  "residence_stats": [
    {
      "id": 1,
      "school_number": "01G01",
      "district": "Mitte",
      "student_count": 350,
      "scraped_at": "2025-01-15T12:00:00Z",
      "created_at": "2025-01-15T12:00:00Z"
    }
  ],
  "absence_stat": {
    "id": 1,
    "school_number": "01G01",
    "school_absence_rate": 5.2,
    "school_unexcused_rate": 0.8,
    "school_type_absence_rate": 4.9,
    "school_type_unexcused_rate": 0.7,
    "region_absence_rate": 5.5,
    "region_unexcused_rate": 1.0,
    "berlin_absence_rate": 5.8,
    "berlin_unexcused_rate": 1.2,
    "scraped_at": "2025-01-15T12:00:00Z",
    "created_at": "2025-01-15T12:00:00Z"
  },
  "statistics": [
    {
      "id": 1,
      "school_number": "01G01",
      "school_name": "Example Gymnasium",
      "district": "Mitte",
      "school_type": "Gymnasium",
      "school_year": "2024/25",
      "students": "512",
      "teachers": "45",
      "classes": "24",
      "metadata": "{}",
      "scraped_at": "2025-01-15T12:00:00Z",
      "created_at": "2025-01-15T12:00:00Z"
    },
    {
      "id": 2,
      "school_number": "01G01",
      "school_name": "Example Gymnasium",
      "district": "Mitte",
      "school_type": "Gymnasium",
      "school_year": "2023/24",
      "students": "498",
      "teachers": "43",
      "classes": "23",
      "metadata": "{}",
      "scraped_at": "2025-01-15T12:00:00Z",
      "created_at": "2025-01-15T12:00:00Z"
    }
  ],
  "construction_projects": [
    {
      "id": 1,
      "project_id": 123,
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
  ]
}
```

## Notes

- Fields that contain no data will be `null` or an empty array `[]`
- The endpoint gracefully handles missing data - if a school has no details or statistics, those fields will simply be omitted or null
- All timestamps are in UTC
- The endpoint uses the school's `school_number` to link related data from different tables

## Performance Considerations

- The enriched endpoints perform multiple database queries per school
- For large datasets (e.g., all schools in Berlin), response times may be longer
- Consider implementing pagination if you need to fetch all schools regularly
- Use the single school endpoint (`/schools/{id}`) when you only need data for one school

## Comparison with Basic Endpoints

| Endpoint | Use Case | Data Included |
|----------|----------|---------------|
| `/api/v1/schools` | Complete school data for all schools | All related data |
| `/api/v1/schools/{id}` | Complete data for one school | All related data |

