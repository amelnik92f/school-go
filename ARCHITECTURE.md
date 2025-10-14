# Architecture Overview

## 🏛️ Complete File Structure

```
school-go/
│
├── 📱 cmd/                          # Application entry points
│   └── api/main.go                  # Main API server (run this!)
│
├── 🔒 internal/                     # Private application code
│   │
│   ├── ⚙️ config/
│   │   └── config.go                # Configuration management (.env loading)
│   │
│   ├── 💾 database/
│   │   └── database.go              # Database connection & migrations
│   │
│   ├── 📦 models/
│   │   └── school.go                # Data structures (School, CreateSchoolInput, etc.)
│   │
│   ├── 🗄️ repository/               # Data Access Layer (DAL)
│   │   └── school_repository.go     # CRUD operations for schools table
│   │
│   ├── 💼 service/                  # Business Logic Layer
│   │   └── school_service.go        # Business rules & orchestration
│   │
│   ├── 🌐 fetcher/                  # External Data Sources
│   │   └── school_fetcher.go        # Fetch data from APIs/files
│   │
│   ├── 🎯 handler/                  # HTTP Request Handlers
│   │   ├── school_handler.go        # School API endpoints
│   │   └── health_handler.go        # Health check endpoint
│   │
│   ├── ⏰ scheduler/
│   │   └── scheduler.go             # Cron jobs (scheduled tasks)
│   │
│   └── 🌍 server/
│       └── server.go                # HTTP server setup & routing
│
├── 📊 data/                         # Created automatically
│   └── schools.db                   # SQLite database file
│
├── 📄 Documentation Files
│   ├── README.md                    # Full project documentation
│   ├── GETTING_STARTED.md          # Quick start guide
│   ├── EXAMPLES.md                 # Code examples & recipes
│   └── ARCHITECTURE.md             # This file
│
├── ⚙️ Configuration Files
│   ├── .env.example                # Example environment variables
│   ├── .env                        # Your config (create from .env.example)
│   ├── .gitignore                  # Git ignore rules
│   ├── go.mod                      # Go module dependencies
│   └── Makefile                    # Build automation
│
└── 🧪 Tests (create as you go)
    └── internal/
        └── service/
            └── school_service_test.go
```

## 🔄 Request Flow Diagram

### GET /api/v1/schools

```
1. HTTP Request
   ↓
2. server.go (Chi Router)
   ↓ routes to
3. school_handler.go → GetSchools()
   ↓ calls
4. school_service.go → GetAllSchools()
   ↓ calls
5. school_repository.go → GetAll()
   ↓ queries
6. SQLite Database
   ↓ returns []School
7. school_repository.go
   ↓ returns
8. school_service.go
   ↓ returns
9. school_handler.go → respondJSON()
   ↓
10. HTTP Response (JSON)
```

### POST /api/v1/schools

```
1. HTTP Request (JSON body)
   ↓
2. server.go (Chi Router)
   ↓
3. school_handler.go → CreateSchool()
   ├─ Parse JSON → CreateSchoolInput
   ↓ calls
4. school_service.go → CreateSchool()
   ├─ Validate business rules
   ↓ calls
5. school_repository.go → Create()
   ↓ executes INSERT
6. SQLite Database
   ↓ returns new School
7. HTTP Response (JSON)
```

## ⏰ Scheduled Job Flow

```
Application Startup
   ↓
scheduler.go → Start()
   ├─ Register cron job: "0 2 * * *" (2 AM daily)
   ↓
Cron Trigger (2 AM)
   ↓
scheduler.go → callback function
   ↓ calls
service.go → RefreshSchoolsData()
   ├─ Step 1: Clear existing data
   │  └─ repository.DeleteAll()
   ├─ Step 2: Fetch new data
   │  └─ fetcher.FetchSchools()
   ├─ Step 3: Insert each school
   │  └─ repository.Create() (in loop)
   ↓
Database Updated
```

## 🧩 Layer Responsibilities

### 1. Handler Layer (`internal/handler/`)
**Purpose:** HTTP request/response handling

**Responsibilities:**
- Parse HTTP requests (URL params, query strings, JSON body)
- Validate input format
- Call service layer
- Format HTTP responses (JSON)
- Set HTTP status codes
- Handle HTTP errors

**Does NOT:**
- Business logic
- Database access
- External API calls

**Example:**
```go
func (h *SchoolHandler) GetSchool(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")                    // Parse request
    school, err := h.service.GetSchoolByID(id)     // Call service
    respondJSON(w, http.StatusOK, school)          // Format response
}
```

---

### 2. Service Layer (`internal/service/`)
**Purpose:** Business logic & orchestration

**Responsibilities:**
- Business rules and validation
- Orchestrate multiple operations
- Transaction management
- Call repositories
- Call external fetchers
- Error handling with context

**Does NOT:**
- HTTP concerns (status codes, headers)
- SQL queries directly

**Example:**
```go
func (s *SchoolService) RefreshSchoolsData() error {
    // Business logic: fetch, validate, update
    schools, err := s.fetcher.FetchSchools()
    if err != nil { return err }
    
    s.repo.DeleteAll()
    for _, school := range schools {
        s.repo.Create(school)
    }
}
```

---

### 3. Repository Layer (`internal/repository/`)
**Purpose:** Data access abstraction

**Responsibilities:**
- SQL queries (CRUD operations)
- Database interactions
- Map between database and models
- Handle database errors

**Does NOT:**
- Business logic
- HTTP handling
- External API calls

**Example:**
```go
func (r *SchoolRepository) GetAll() ([]models.School, error) {
    var schools []models.School
    err := r.db.Select(&schools, "SELECT * FROM schools")
    return schools, err
}
```

---

### 4. Fetcher Layer (`internal/fetcher/`)
**Purpose:** External data retrieval

**Responsibilities:**
- HTTP API calls
- File parsing (CSV, JSON)
- Web scraping
- Data transformation to internal format

**Does NOT:**
- Database operations
- Business logic

**Example:**
```go
func (f *SchoolFetcher) FetchSchools() ([]models.CreateSchoolInput, error) {
    resp, err := http.Get("https://api.example.com/schools")
    // ... parse and return
}
```

---

### 5. Model Layer (`internal/models/`)
**Purpose:** Data structures

**Responsibilities:**
- Define structs
- JSON/DB tags
- Input/Output types

**Example:**
```go
type School struct {
    ID        int64     `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

---

## 🔗 Dependency Flow

```
main.go
  ↓ creates
config → database → repository → service → handler → server
                    ↓
                  fetcher → service
                    ↓
                scheduler → service
```

**Key Principle:** Dependencies flow inward
- Handlers depend on Services
- Services depend on Repositories & Fetchers
- Repositories depend on Database
- **NO** reverse dependencies!

## 🎯 Design Patterns Used

### 1. **Repository Pattern**
Abstracts data access layer. Services don't know about SQL.

### 2. **Dependency Injection**
Dependencies are passed via constructors:
```go
func NewSchoolService(repo *Repository, fetcher *Fetcher) *SchoolService {
    return &SchoolService{repo: repo, fetcher: fetcher}
}
```

### 3. **Constructor Pattern**
Each package has `New...()` functions that create and return struct instances.

### 4. **Separation of Concerns**
Each layer has a single responsibility. Changes in one layer don't affect others.

### 5. **Clean Architecture**
Business logic (service) is independent of frameworks, databases, and UI.

## 🗂️ Data Models

### Database Schema

```sql
CREATE TABLE schools (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    address    TEXT,
    type       TEXT,
    latitude   REAL,
    longitude  REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_schools_type ON schools(type);
CREATE INDEX idx_schools_created_at ON schools(created_at);
```

### Go Struct Mapping

```go
type School struct {
    ID        int64     `json:"id" db:"id"`            // Maps to database column
    Name      string    `json:"name" db:"name"`        // Also serializes to JSON
    Address   string    `json:"address" db:"address"`
    Type      string    `json:"type" db:"type"`
    Latitude  float64   `json:"latitude" db:"latitude"`
    Longitude float64   `json:"longitude" db:"longitude"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

## 🚀 Startup Sequence

```
1. Load Configuration (.env file)
   ↓
2. Connect to Database
   ↓
3. Run Migrations (create tables if needed)
   ↓
4. Initialize Repositories
   ↓
5. Initialize Fetchers
   ↓
6. Initialize Services
   ↓
7. Initialize Handlers
   ↓
8. Setup HTTP Server & Routes
   ↓
9. Start Scheduler (background cron jobs)
   ↓
10. Start HTTP Server (listen on port)
    ↓
11. Wait for Shutdown Signal (Ctrl+C)
    ↓
12. Graceful Shutdown (finish pending requests)
```

## 📊 API Routes Map

```
GET    /health                    → health_handler.HealthCheck()
GET    /api/v1/schools            → school_handler.GetSchools()
POST   /api/v1/schools            → school_handler.CreateSchool()
GET    /api/v1/schools/:id        → school_handler.GetSchool()
PUT    /api/v1/schools/:id        → school_handler.UpdateSchool()
DELETE /api/v1/schools/:id        → school_handler.DeleteSchool()
POST   /api/v1/refresh            → school_handler.RefreshData()
```

## 🔐 Security Features

1. **CORS Configuration** - Control cross-origin requests
2. **Request Timeout** - Prevent hanging requests
3. **Graceful Shutdown** - Finish pending requests before exit
4. **Error Handling** - Never expose internal errors to clients
5. **Input Validation** - Validate all user inputs

## 🎓 Key Go Concepts in This Project

### Structs
```go
type SchoolService struct {
    repo    *repository.SchoolRepository
    fetcher *fetcher.SchoolFetcher
}
```

### Pointers
```go
func (s *SchoolService) GetSchools() ([]School, error) {
    // 's' is a pointer receiver
}
```

### Interfaces (implicit)
Go doesn't require explicit interface implementation. Any type that has matching methods implements the interface.

### Error Handling
```go
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

### Goroutines
```go
go func() {
    // Runs in background (server, scheduler)
}()
```

### Channels
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // Block until signal received
```

## 📈 Scalability Path

This architecture scales from pet project → production:

1. **Current**: SQLite, single file deployment
2. **Next**: PostgreSQL, multiple instances
3. **Later**: Add caching (Redis), message queue (RabbitMQ)
4. **Scale**: Microservices, Kubernetes

The layer separation makes these transitions easier!

---

## 🎯 Quick Reference: Where to Add...

| Want to add... | Edit this file... |
|----------------|-------------------|
| New API endpoint | `internal/handler/` + `internal/server/server.go` |
| New database table | `internal/database/database.go` (migrations) |
| New data model | `internal/models/` |
| New external API | `internal/fetcher/` |
| Business logic | `internal/service/` |
| Database query | `internal/repository/` |
| Scheduled job | `internal/scheduler/scheduler.go` |
| Configuration option | `internal/config/config.go` + `.env` |

---

**This architecture follows industry best practices and will serve you well as you learn Go!** 🚀


