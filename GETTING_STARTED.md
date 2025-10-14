# Getting Started with Your Go Backend

## 🎯 What I've Built For You

A complete, production-ready Go backend application with:

✅ **Clean Architecture** - Separation of concerns with proper layers
✅ **REST API** - Full CRUD endpoints for schools data  
✅ **Database** - SQLite with automatic migrations
✅ **Scheduled Jobs** - Daily data fetching (configurable)
✅ **Popular Libraries** - Industry-standard tools
✅ **Best Practices** - Idiomatic Go patterns

## 📁 Project Structure Explained

```
school-go/
├── cmd/
│   └── api/main.go        # 🚀 Main application entry point - START HERE
│
├── internal/              # Private application code
│   ├── config/           # ⚙️  Configuration from .env files
│   ├── database/         # 💾 Database connection & migrations
│   ├── models/           # 📦 Data structures (School, etc.)
│   ├── repository/       # 🗄️  Database queries (CRUD operations)
│   ├── service/          # 💼 Business logic layer
│   ├── fetcher/          # 🌐 External data fetching (TODO: customize)
│   ├── handler/          # 🎯 HTTP request handlers
│   ├── scheduler/        # ⏰ Cron jobs for periodic tasks
│   └── server/           # 🌍 HTTP server & routing setup
│
├── .env.example          # Example configuration
├── go.mod               # Go dependencies
├── Makefile             # Build automation
└── README.md            # Full documentation
```

## 🚀 Quick Start (3 steps)

### 1. Install Dependencies
```bash
cd school-go
go mod download
```

### 2. Create Configuration
```bash
cp .env.example .env
# Edit .env if needed (defaults work fine)
```

### 3. Run the Server
```bash
go run cmd/api/main.go
```

That's it! Server runs on `http://localhost:8080`

## 🧪 Test the API

### Check if it's running:
```bash
curl http://localhost:8080/health
```

### Get all schools:
```bash
curl http://localhost:8080/api/v1/schools
```

### Create a school:
```bash
curl -X POST http://localhost:8080/api/v1/schools \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test School",
    "address": "123 Main St",
    "type": "Gymnasium",
    "latitude": 52.5200,
    "longitude": 13.4050
  }'
```

### Manually trigger data refresh:
```bash
curl -X POST http://localhost:8080/api/v1/refresh
```

## 🔄 How It Works

### Data Flow
```
External APIs/Sources
        ↓
    Fetcher (you'll customize this)
        ↓
    Service (business logic)
        ↓
    Repository (database)
        ↓
    SQLite Database
```

### API Flow
```
HTTP Request
    ↓
Handler (parses request)
    ↓
Service (business logic)
    ↓
Repository (database query)
    ↓
HTTP Response (JSON)
```

### Scheduled Jobs
```
Cron Scheduler (runs daily at 2 AM)
    ↓
Service.RefreshSchoolsData()
    ↓
Fetcher.FetchSchools()
    ↓
Repository.Create() for each school
```

## 📝 Next Steps - Customize for Your Project

### 1. Update the Fetcher (IMPORTANT!)
The fetcher currently has placeholder data. Update it to fetch from your actual sources:

**File:** `internal/fetcher/school_fetcher.go`

```go
func (f *SchoolFetcher) FetchSchools() ([]models.CreateSchoolInput, error) {
    // Replace this with:
    // - HTTP API calls
    // - CSV/JSON file parsing
    // - Web scraping
    // - Database queries from other sources
    
    // Example with HTTP:
    resp, err := http.Get("https://api.example.com/schools")
    // ... parse response
    
    return schools, nil
}
```

### 2. Add More Data Models
Create new files in `internal/models/` for other entities:
- `teacher.go`
- `student.go`
- `course.go`

### 3. Add Corresponding Repositories
In `internal/repository/`, create:
- `teacher_repository.go`
- `student_repository.go`

### 4. Add API Endpoints
Update `internal/server/server.go` to add new routes:
```go
r.Route("/teachers", func(r chi.Router) {
    r.Get("/", teacherHandler.GetTeachers)
    r.Post("/", teacherHandler.CreateTeacher)
})
```

## 🛠️ Useful Make Commands

```bash
make help            # See all available commands
make run             # Run the app
make build           # Build binary to bin/schools-be
make test            # Run tests (write tests as you go!)
make clean           # Clean build artifacts
```

## 📚 Go Libraries Used

| Library | Purpose | Why? |
|---------|---------|------|
| **chi** | HTTP router | Lightweight, fast, idiomatic Go |
| **sqlx** | Database | Better than database/sql, easier queries |
| **sqlite3** | Database driver | Simple, no setup, perfect for local |
| **cron** | Scheduling | Simple, reliable cron jobs |
| **godotenv** | Config | Load .env files easily |
| **cors** | CORS handling | Easy cross-origin requests |

## 🎓 Learning Tips

### Start Here (in order):
1. `cmd/api/main.go` - See how everything connects
2. `internal/models/school.go` - Understand data structures
3. `internal/handler/school_handler.go` - See HTTP handling
4. `internal/service/school_service.go` - Business logic
5. `internal/repository/school_repository.go` - Database queries

### Key Go Concepts in This Project:
- **Interfaces** - Abstraction between layers
- **Structs** - Data models
- **Pointers** - Passing data efficiently
- **Error handling** - Go's explicit error returns
- **Goroutines** - Concurrent server & scheduler
- **Channels** - Signal handling for graceful shutdown

### Common Patterns You'll See:
```go
// Constructor pattern
func NewSchoolService(repo *Repository) *SchoolService {
    return &SchoolService{repo: repo}
}

// Error handling
if err != nil {
    return nil, fmt.Errorf("failed: %w", err)
}

// Pointer receivers
func (s *SchoolService) GetSchools() ([]School, error) {
    // ...
}
```

## 🐛 Troubleshooting

### Port already in use?
Change PORT in `.env` or kill the process:
```bash
lsof -ti:8080 | xargs kill
```

### Database locked?
SQLite allows only one writer. The app uses `SetMaxOpenConns(1)` to prevent this.

### Import errors?
```bash
go mod tidy
```

## 📖 Next Learning Resources

Once comfortable with this structure:
1. Add **validation** with `go-playground/validator`
2. Add **structured logging** with `zap` or `zerolog`
3. Add **authentication** with JWT
4. Switch to **PostgreSQL** (just change driver!)
5. Add **Docker** for containerization
6. Add **tests** - Go makes testing easy!

## 💡 Pro Tips

1. **Run with hot reload** (install air first):
   ```bash
   go install github.com/air-verse/air@latest
   make dev
   ```

2. **Use VS Code with Go extension** - excellent autocomplete & debugging

3. **Read errors carefully** - Go's compiler is very helpful

4. **Start simple** - Get one endpoint working, then expand

5. **Write tests as you go** - Future you will thank present you!

---

**You're all set! 🎉**

This structure will scale from a simple pet project to a production application. Enjoy learning Go!


