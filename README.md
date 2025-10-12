# Schools Backend (Go)

A Go backend service that fetches school data from external sources, stores it locally, and exposes REST API endpoints.

## 🏗️ Project Structure

```
school-go/
├── cmd/
│   ├── api/                    # Main application entry point
│   ├── migrate/                # Database migration tool
│   ├── scrape-statistics/      # Statistics scraper
│   └── scrape-school-details/  # School details scraper (NEW!)
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # Database connection and migrations
│   ├── models/         # Data models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   ├── scraper/        # Web scrapers for Berlin school data
│   ├── fetcher/        # External data fetchers
│   ├── handler/        # HTTP handlers
│   ├── scheduler/      # Scheduled jobs (cron)
│   └── server/         # HTTP server setup
├── data/               # Database files (gitignored)
├── cache/              # Scraper cache (gitignored)
├── .env.example        # Example environment variables
├── go.mod              # Go module definition
└── Makefile           # Build automation
```

## 🚀 Getting Started

### Prerequisites

- Go 1.25.2 or higher
- Make (optional, but recommended)

### Installation

1. Clone the repository and navigate to the project:
```bash
cd school-go
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
make install-deps
# or
go mod download
```

4. Run the application:
```bash
make run
# or
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

## 📋 Available Make Commands

```bash
make help                  # Show all available commands
make install-deps          # Install Go dependencies
make build                 # Build the application binary
make run                   # Run the application
make dev                   # Run with hot reload (requires air)
make test                  # Run tests
make test-coverage         # Run tests with coverage report
make clean                 # Clean build artifacts
make migrate               # Run database migrations
make scrape-statistics     # Scrape school statistics from Berlin education website
make scrape-school-details # Scrape detailed school information (takes several hours)
make build-scrapers        # Build all scraper binaries
```

## 🔌 API Endpoints

### Health Check
- `GET /health` - Health check endpoint

### Schools
- `GET /api/v1/schools` - Get all schools
- `GET /api/v1/schools?type=Gymnasium` - Get schools by type
- `GET /api/v1/schools/:id` - Get a specific school
- `POST /api/v1/schools` - Create a new school
- `PUT /api/v1/schools/:id` - Update a school
- `DELETE /api/v1/schools/:id` - Delete a school

### Admin
- `POST /api/v1/refresh` - Manually trigger data refresh

## 📦 Core Libraries Used

- **chi** - Lightweight, idiomatic HTTP router
- **sqlx** - Extensions to database/sql
- **sqlite3** - SQLite database driver
- **cron** - Cron job scheduler
- **godotenv** - Load environment variables from .env

## 🔄 Scheduled Jobs

The application includes a scheduler that runs periodic tasks:
- **Data Refresh**: Runs daily at 2 AM (configurable via `FETCH_SCHEDULE`)

## 🗄️ Database

The application uses SQLite for local storage. The database file is created automatically in the `data/` directory.

### Migrations

Migrations run automatically on application startup. You can also run them manually:
```bash
make migrate
# or
go run cmd/migrate/main.go
```

## 🛠️ Development Tips

### Hot Reload (Optional)

Install Air for automatic reloading during development:
```bash
go install github.com/air-verse/air@latest
```

Create `.air.toml` in the project root:
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/api/main.go"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor", "data"]
```

Then run:
```bash
make dev
```

### Testing

Write tests alongside your code:
```go
// internal/service/school_service_test.go
package service

import "testing"

func TestGetAllSchools(t *testing.T) {
    // Test implementation
}
```

Run tests:
```bash
make test
```

## 🔧 Configuration

Configuration is managed through environment variables. See `.env.example` for available options:

- `PORT` - Server port (default: 8080)
- `ENV` - Environment (development/production)
- `DB_PATH` - Database file path
- `FETCH_SCHEDULE` - Cron schedule for data fetching
- `API_TIMEOUT` - API request timeout

## 🕷️ Web Scrapers

This project includes specialized scrapers for the Berlin education website:

### School Statistics Scraper
Scrapes basic statistics (students, teachers, classes) from the Berlin education statistics website.

```bash
make scrape-statistics
```

### School Details Scraper (NEW!)
Comprehensive scraper that extracts detailed information for each school:
- Languages offered (Sprachen)
- Advanced courses (Leistungskurse)  
- Programs and offerings (Angebote)
- Availability after 4th grade
- Student statistics (citizenship, languages, residence, absences)
- **File-based caching** for fast subsequent runs

```bash
# First run (2-4 hours)
make scrape-school-details

## 📝 Next Steps

1. **Web Scrapers**: Already implemented for Berlin school data
2. **Add More Models**: Create additional models in `internal/models/`
3. **Extend Repositories**: Add more data access methods in `internal/repository/`
4. **Add Authentication**: Implement JWT or API key authentication
5. **API Integration**: Expose school details through REST API
6. **Add Tests**: Write unit and integration tests for scrapers
7. **Frontend Integration**: Connect with the Next.js frontend

## 📚 Learning Resources

- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Chi Documentation](https://github.com/go-chi/chi)
- [SQLX Documentation](https://jmoiron.github.io/sqlx/)

## 📄 License

This project is open source and available under the MIT License.


