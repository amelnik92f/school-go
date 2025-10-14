# Docker Setup Guide

This guide explains how to run the Schools Backend application using Docker.

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)

## Quick Start

1. **Ensure you have your `.env` file in the project root**
   ```bash
   # If you don't have one, create it with your configuration
   # See the configuration section below for required variables
   ```

2. **Build and run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

3. **Check the logs**
   ```bash
   docker-compose logs -f schools-api
   ```

4. **Stop the application**
   ```bash
   docker-compose down
   ```

## Data Persistence

The Docker setup uses volumes to persist your data:

- **`./data`**: SQLite database files
- **`./cache`**: Scraper cache files

If you already have existing `data/` and `cache/` directories with data, they will be automatically mounted and reused. This means you won't lose your scraped data or database when using Docker.

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Server Configuration
PORT=8080
ENV=production

# Database
DB_PATH=/app/data/schools.db

# Scheduling (cron format)
FETCH_SCHEDULE=0 2 * * 0

# API Configuration
API_TIMEOUT=30s

# Optional API Keys
GEMINI_API_KEY=your_gemini_api_key_here
OPENROUTESERVICE_API_KEY=your_openroute_api_key_here
API_KEY=your_api_key_here
```

## Docker Commands

### Build the image
```bash
docker-compose build
```

### Start the service
```bash
docker-compose up -d
```

### View logs
```bash
docker-compose logs -f
```

### Stop the service
```bash
docker-compose down
```

### Restart the service
```bash
docker-compose restart
```

### Rebuild and restart
```bash
docker-compose up -d --build
```

### Remove volumes (⚠️ This will delete your data!)
```bash
docker-compose down -v
```

## Health Check

The Docker container includes a health check that monitors the `/health` endpoint. You can check the health status with:

```bash
docker-compose ps
```

## Accessing the API

Once running, the API will be available at:
- `http://localhost:8080` (or the port specified in your `.env` file)

Test it:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/schools
```

## Troubleshooting

### Container fails to start
Check the logs:
```bash
docker-compose logs schools-api
```

### Database issues
Ensure the `data/` directory has proper permissions:
```bash
chmod -R 755 data/
```

### Cache issues
Ensure the `cache/` directory has proper permissions:
```bash
chmod -R 755 cache/
```

### Port already in use
Change the port in your `.env` file or in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Use port 8081 on host instead
```

### Rebuild from scratch
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Development with Docker

For development, you can use Docker with volume mounting for live code changes. However, for hot reload, it's recommended to use the native Go development setup with `make dev` instead.

If you want to run commands inside the container:
```bash
# Get a shell in the running container
docker-compose exec schools-api sh

# Run tests (if container is running)
docker-compose exec schools-api go test ./...
```

## Production Deployment

For production deployment:

1. Set `ENV=production` in your `.env` file
2. Ensure all API keys are properly configured
3. Set up appropriate backup for the `data/` directory
4. Consider using Docker secrets for sensitive data
5. Set up a reverse proxy (like Nginx) for HTTPS
6. Configure proper logging and monitoring

## Notes

- The Dockerfile uses a multi-stage build to keep the final image small
- SQLite3 support requires CGO, which is enabled in the build
- Timezone data is included for proper cron scheduling
- The container runs as a non-root user for security
- Health checks ensure the service is running properly

