# API Migration Summary

## Overview

This document describes the migration of Gemini AI and OpenRouteService API calls from the Next.js frontend to the Go backend.

## Changes Made

### 1. Go Backend (school-go)

#### New Configuration
- Added `GeminiAPIKey` to config struct
- Added `OpenRouteServiceAPIKey` to config struct
- Both keys are loaded from environment variables: `GEMINI_API_KEY` and `OPENROUTESERVICE_API_KEY`

#### New Services

**AI Service** (`internal/service/ai_service.go`)
- Handles Gemini AI integration using Google's Generative AI SDK
- Generates comprehensive school summaries based on enriched school data
- Uses `gemini-2.0-flash-exp` model
- Includes detailed prompts with school information, statistics, languages, facilities, etc.

**Routes Service** (`internal/service/routes_service.go`)
- Handles OpenRouteService API integration
- Calculates travel times for multiple transportation modes (walking, bicycle, car)
- Returns duration in minutes and distance in kilometers
- Supports parallel requests for multiple modes

#### New API Endpoints

**GET `/api/v1/schools/{id}/summary`**
- Generates an AI summary for a specific school
- Returns JSON with `success`, `summary`, and `schoolName` fields
- Uses Gemini AI to create a comprehensive profile

**POST `/api/v1/schools/{id}/routes`**
- Calculates travel times between two coordinates
- Request body:
  ```json
  {
    "start": [longitude, latitude],
    "end": [longitude, latitude],
    "modes": ["walking", "bicycle", "car"]
  }
  ```
- Returns JSON with `results` array containing travel time information

#### Handler Updates
- Updated `SchoolHandler` to include AI and Routes services
- Added `GetSchoolSummary` method
- Added `CalculateRoutes` method

#### Main Application Updates
- Initialize AI service in `cmd/api/main.go`
- Initialize Routes service in `cmd/api/main.go`
- Gracefully handle missing API keys (AI service will be `nil` if not configured)
- Ensure AI service is closed on shutdown

### 2. Frontend (schools-info)

#### Updated API Calls

**AI Summary Store** (`lib/store/ai-summary-store.ts`)
- Changed from POST `/api/summarize-school` to GET `/api/v1/schools/{id}/summary`
- Now calls the Go backend directly
- Uses `API_URL` environment variable

**Travel Time Utility** (`lib/utils/travel-time.ts`)
- Changed from POST `/api/travel-time` to POST `/api/v1/schools/{id}/routes`
- Now calls the Go backend directly
- Accepts optional `schoolId` parameter (defaults to "0" for generic routes)

#### Removed Files
- ❌ `app/api/summarize-school/route.ts`
- ❌ `app/api/summarize-school/prompts.ts`
- ❌ `app/api/travel-time/route.ts`
- ❌ Removed empty directories: `app/api/summarize-school/` and `app/api/travel-time/`

## Environment Variables Required

### Go Backend (.env)
```bash
GEMINI_API_KEY=your_gemini_api_key_here
OPENROUTESERVICE_API_KEY=your_openroute_service_api_key_here
```

### Frontend (.env.local)
```bash
API_URL=http://localhost:8080
```

## Dependencies Added

### Go Backend
- `github.com/google/generative-ai-go` - Google Generative AI SDK
- Related dependencies for Google AI (see `go.mod` for full list)

### Frontend
No new dependencies added (removed dependency on `@google/generative-ai`)

## Benefits of This Migration

1. **Centralized API Management**: All external API calls are now managed in one place
2. **Better Security**: API keys are stored server-side only, not exposed to frontend
3. **Improved Performance**: Backend can cache results more effectively
4. **Reduced Frontend Bundle Size**: Removed heavy dependencies from frontend
5. **Better Error Handling**: Centralized error handling and logging in backend
6. **Easier Monitoring**: All API calls can be monitored from backend logs
7. **Cost Control**: Better tracking and control of API usage

## Testing

### Manual Testing Steps

1. **Test AI Summary**
   ```bash
   curl http://localhost:8080/api/v1/schools/1/summary
   ```

2. **Test Travel Routes**
   ```bash
   curl -X POST http://localhost:8080/api/v1/schools/1/routes \
     -H "Content-Type: application/json" \
     -d '{
       "start": [13.404954, 52.520008],
       "end": [13.387978, 52.517037],
       "modes": ["walking", "bicycle", "car"]
     }'
   ```

3. **Test Frontend Integration**
   - Open the frontend application
   - Navigate to a school details page
   - Verify that AI summary loads correctly
   - Set a home location and verify travel times are calculated

## Migration Notes

- The AI summary endpoint is now a GET request (previously POST) since it doesn't require a request body
- The travel routes endpoint remains POST since it requires start/end coordinates in the body
- Both endpoints maintain backward compatibility with the frontend
- Error handling has been improved with proper HTTP status codes

## Rollback Plan

If issues arise, you can:
1. Restore the deleted frontend API route files from git history
2. Revert the changes in `ai-summary-store.ts` and `travel-time.ts`
3. Continue using the old API structure until issues are resolved

## Future Improvements

1. Add caching layer in Go backend for AI summaries
2. Implement rate limiting for API calls
3. Add metrics and monitoring for API usage
4. Consider adding authentication for sensitive endpoints
5. Add request validation middleware

