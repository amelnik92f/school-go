# API Authentication

## Overview

All API endpoints (except `/health`) are protected with API key authentication to secure access to the school data API.

## Configuration

### Backend Setup

Add the following environment variable to your `.env` file:

```bash
# Authentication
# IMPORTANT: Set a strong API key to secure your API endpoints
# Generate a secure key: openssl rand -hex 32
API_KEY=your-secure-api-key-here
```

### Generating a Secure API Key

You can generate a secure API key using OpenSSL:

```bash
openssl rand -hex 32
```

This will generate a 64-character hexadecimal string that can be used as your API key.

### Frontend Setup

Add the following environment variable to your frontend's `.env.local` file:

```bash
# API Authentication
# IMPORTANT: This must match the API_KEY set in your backend .env file
NEXT_PUBLIC_API_KEY=your-secure-api-key-here
```

## How It Works

### Backend Implementation

The backend uses a middleware (`internal/middleware/auth.go`) that:

1. Checks for an API key in the `X-API-Key` header
2. Also supports the `Authorization: Bearer <token>` format
3. Validates the provided key against the configured `API_KEY`
4. Returns `401 Unauthorized` if the key is missing or invalid
5. **Development Mode**: If no `API_KEY` is configured, authentication is disabled (logs a warning)

### Protected Endpoints

All endpoints under `/api/v1` require authentication:

- `GET /api/v1/schools` - List all schools
- `GET /api/v1/schools/{id}` - Get a specific school
- `GET /api/v1/schools/{id}/summary` - Get AI summary for a school
- `POST /api/v1/schools/{id}/routes` - Calculate travel times
- `GET /api/v1/construction-projects` - List construction projects
- `GET /api/v1/construction-projects/standalone` - List standalone projects
- `GET /api/v1/construction-projects/{id}` - Get a specific project
- `POST /api/v1/refresh` - Manually refresh data

### Unprotected Endpoints

The health check endpoint does not require authentication:

- `GET /health` - Health check (no authentication required)

## Making Authenticated Requests

### Using X-API-Key Header

```bash
curl -H "X-API-Key: your-api-key-here" http://localhost:8080/api/v1/schools
```

### Using Authorization Header

```bash
curl -H "Authorization: Bearer your-api-key-here" http://localhost:8080/api/v1/schools
```

### Frontend Integration

The frontend automatically includes the API key in all requests through centralized header functions in:

- `/lib/api/index.ts` - Main API client
- `/lib/utils/travel-time.ts` - Travel time calculations
- `/lib/store/ai-summary-store.ts` - AI summary fetching

No additional configuration is needed once the environment variable is set.

## Security Best Practices

1. **Never commit** `.env` files to version control
2. **Use different keys** for development and production environments
3. **Rotate keys** periodically for enhanced security
4. **Keep keys secret** - never expose them in client-side code (use `NEXT_PUBLIC_` prefix only for the frontend key)
5. **Use HTTPS** in production to prevent key interception

## CORS Configuration

The API has been configured to accept the `X-API-Key` header through CORS:

```go
AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-API-Key"}
```

Allowed origins by default:
- `http://localhost:3000` (Next.js frontend)
- `http://localhost:8080` (API server)

Update `internal/server/server.go` to add production origins as needed.

## Troubleshooting

### 401 Unauthorized Error

**Cause**: Missing or invalid API key

**Solutions**:
1. Verify the `API_KEY` is set in the backend `.env` file
2. Verify the `NEXT_PUBLIC_API_KEY` matches in the frontend `.env.local` file
3. Restart both backend and frontend servers after changing environment variables
4. Check that the API key doesn't have extra whitespace or newlines

### CORS Errors

**Cause**: Frontend origin not allowed or headers not permitted

**Solution**: Update the CORS configuration in `internal/server/server.go` to include your frontend's origin.

### Development Mode

If you see the warning `"API key authentication is disabled - no API_KEY configured"` in the logs, authentication is bypassed. This is useful for development but should **never** be used in production.

