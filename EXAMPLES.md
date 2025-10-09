# Code Examples & Recipes

Practical examples for common tasks in your Go backend.

## 🌐 Fetching Data from External APIs

### HTTP GET Request with JSON Response

```go
// internal/fetcher/school_fetcher.go

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type ExternalSchool struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Address   string  `json:"address"`
    SchoolType string  `json:"school_type"`
    Lat       float64 `json:"lat"`
    Lng       float64 `json:"lng"`
}

func (f *SchoolFetcher) FetchFromAPI() ([]models.CreateSchoolInput, error) {
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    resp, err := client.Get("https://api.example.com/schools")
    if err != nil {
        return nil, fmt.Errorf("failed to fetch: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read body: %w", err)
    }
    
    var externalSchools []ExternalSchool
    if err := json.Unmarshal(body, &externalSchools); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }
    
    // Convert to internal format
    schools := make([]models.CreateSchoolInput, 0, len(externalSchools))
    for _, ext := range externalSchools {
        schools = append(schools, models.CreateSchoolInput{
            Name:      ext.Name,
            Address:   ext.Address,
            Type:      ext.SchoolType,
            Latitude:  ext.Lat,
            Longitude: ext.Lng,
        })
    }
    
    return schools, nil
}
```

### HTTP POST Request with Headers

```go
func (f *SchoolFetcher) FetchWithAuth(apiKey string) error {
    req, err := http.NewRequest("POST", "https://api.example.com/schools", nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // Process response...
    return nil
}
```

## 📄 Reading CSV Files

```go
// internal/fetcher/csv_fetcher.go

import (
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
)

func (f *SchoolFetcher) FetchFromCSV(filepath string) ([]models.CreateSchoolInput, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to open CSV: %w", err)
    }
    defer file.Close()
    
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV: %w", err)
    }
    
    schools := make([]models.CreateSchoolInput, 0)
    
    // Skip header row
    for i, record := range records {
        if i == 0 {
            continue
        }
        
        // CSV format: name,address,type,latitude,longitude
        lat, _ := strconv.ParseFloat(record[3], 64)
        lng, _ := strconv.ParseFloat(record[4], 64)
        
        schools = append(schools, models.CreateSchoolInput{
            Name:      record[0],
            Address:   record[1],
            Type:      record[2],
            Latitude:  lat,
            Longitude: lng,
        })
    }
    
    return schools, nil
}
```

## 🗄️ Database Queries

### Find with WHERE Clause

```go
// internal/repository/school_repository.go

func (r *SchoolRepository) FindByName(name string) ([]models.School, error) {
    var schools []models.School
    query := `SELECT * FROM schools WHERE name LIKE ? ORDER BY name`
    
    err := r.db.Select(&schools, query, "%"+name+"%")
    return schools, err
}
```

### Find with Multiple Conditions

```go
func (r *SchoolRepository) FindByFilters(schoolType string, minLat, maxLat float64) ([]models.School, error) {
    var schools []models.School
    query := `
        SELECT * FROM schools 
        WHERE type = ? 
        AND latitude BETWEEN ? AND ?
        ORDER BY name
    `
    
    err := r.db.Select(&schools, query, schoolType, minLat, maxLat)
    return schools, err
}
```

### Count Records

```go
func (r *SchoolRepository) Count() (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM schools`
    
    err := r.db.Get(&count, query)
    return count, err
}
```

### Upsert (Insert or Update)

```go
func (r *SchoolRepository) Upsert(input models.CreateSchoolInput) (*models.School, error) {
    // Check if school exists by name
    var existing models.School
    err := r.db.Get(&existing, "SELECT * FROM schools WHERE name = ?", input.Name)
    
    if err == sql.ErrNoRows {
        // Doesn't exist, create new
        return r.Create(input)
    }
    
    if err != nil {
        return nil, err
    }
    
    // Exists, update
    updateInput := models.UpdateSchoolInput{
        Address:   &input.Address,
        Type:      &input.Type,
        Latitude:  &input.Latitude,
        Longitude: &input.Longitude,
    }
    
    return r.Update(existing.ID, updateInput)
}
```

## 📊 Adding a New Entity (Complete Example)

Let's add a "Teacher" entity:

### 1. Create Model

```go
// internal/models/teacher.go

package models

import "time"

type Teacher struct {
    ID        int64     `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    Subject   string    `json:"subject" db:"subject"`
    SchoolID  int64     `json:"school_id" db:"school_id"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateTeacherInput struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Subject  string `json:"subject"`
    SchoolID int64  `json:"school_id"`
}
```

### 2. Add Migration

```go
// internal/database/database.go

func RunMigrations(db *sqlx.DB) error {
    migrations := []string{
        // ... existing school migrations ...
        
        `CREATE TABLE IF NOT EXISTS teachers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            subject TEXT,
            school_id INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (school_id) REFERENCES schools(id) ON DELETE CASCADE
        )`,
        `CREATE INDEX IF NOT EXISTS idx_teachers_school_id ON teachers(school_id)`,
        `CREATE INDEX IF NOT EXISTS idx_teachers_email ON teachers(email)`,
    }
    // ... rest of migration code
}
```

### 3. Create Repository

```go
// internal/repository/teacher_repository.go

package repository

import (
    "github.com/jmoiron/sqlx"
    "schools-be/internal/models"
    "time"
)

type TeacherRepository struct {
    db *sqlx.DB
}

func NewTeacherRepository(db *sqlx.DB) *TeacherRepository {
    return &TeacherRepository{db: db}
}

func (r *TeacherRepository) GetAll() ([]models.Teacher, error) {
    var teachers []models.Teacher
    query := `SELECT * FROM teachers ORDER BY name`
    err := r.db.Select(&teachers, query)
    return teachers, err
}

func (r *TeacherRepository) GetBySchoolID(schoolID int64) ([]models.Teacher, error) {
    var teachers []models.Teacher
    query := `SELECT * FROM teachers WHERE school_id = ? ORDER BY name`
    err := r.db.Select(&teachers, query, schoolID)
    return teachers, err
}

func (r *TeacherRepository) Create(input models.CreateTeacherInput) (*models.Teacher, error) {
    query := `
        INSERT INTO teachers (name, email, subject, school_id, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
    
    result, err := r.db.Exec(query, input.Name, input.Email, input.Subject, 
        input.SchoolID, time.Now())
    if err != nil {
        return nil, err
    }
    
    id, _ := result.LastInsertId()
    
    var teacher models.Teacher
    err = r.db.Get(&teacher, "SELECT * FROM teachers WHERE id = ?", id)
    return &teacher, err
}
```

### 4. Create Service

```go
// internal/service/teacher_service.go

package service

import (
    "schools-be/internal/models"
    "schools-be/internal/repository"
)

type TeacherService struct {
    repo *repository.TeacherRepository
}

func NewTeacherService(repo *repository.TeacherRepository) *TeacherService {
    return &TeacherService{repo: repo}
}

func (s *TeacherService) GetAllTeachers() ([]models.Teacher, error) {
    return s.repo.GetAll()
}

func (s *TeacherService) GetTeachersBySchool(schoolID int64) ([]models.Teacher, error) {
    return s.repo.GetBySchoolID(schoolID)
}

func (s *TeacherService) CreateTeacher(input models.CreateTeacherInput) (*models.Teacher, error) {
    return s.repo.Create(input)
}
```

### 5. Create Handler

```go
// internal/handler/teacher_handler.go

package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
    "schools-be/internal/models"
    "schools-be/internal/service"
)

type TeacherHandler struct {
    service *service.TeacherService
}

func NewTeacherHandler(service *service.TeacherService) *TeacherHandler {
    return &TeacherHandler{service: service}
}

func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
    schoolIDStr := r.URL.Query().Get("school_id")
    
    var teachers []models.Teacher
    var err error
    
    if schoolIDStr != "" {
        schoolID, _ := strconv.ParseInt(schoolIDStr, 10, 64)
        teachers, err = h.service.GetTeachersBySchool(schoolID)
    } else {
        teachers, err = h.service.GetAllTeachers()
    }
    
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, teachers)
}

func (h *TeacherHandler) CreateTeacher(w http.ResponseWriter, r *http.Request) {
    var input models.CreateTeacherInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request")
        return
    }
    
    teacher, err := h.service.CreateTeacher(input)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusCreated, teacher)
}
```

### 6. Wire It Up in main.go

```go
// cmd/api/main.go

// Add to main():
teacherRepo := repository.NewTeacherRepository(db)
teacherService := service.NewTeacherService(teacherRepo)
teacherHandler := handler.NewTeacherHandler(teacherService)

// Pass to server:
srv := server.New(cfg, schoolHandler, teacherHandler)
```

### 7. Add Routes

```go
// internal/server/server.go

func (s *Server) setupRoutes(schoolHandler *handler.SchoolHandler, teacherHandler *handler.TeacherHandler) {
    // ... existing routes ...
    
    s.router.Route("/api/v1", func(r chi.Router) {
        // ... existing school routes ...
        
        // Teacher routes
        r.Route("/teachers", func(r chi.Router) {
            r.Get("/", teacherHandler.GetTeachers)
            r.Post("/", teacherHandler.CreateTeacher)
        })
    })
}
```

## ⏰ Adding More Scheduled Jobs

```go
// internal/scheduler/scheduler.go

func (s *Scheduler) Start() {
    // Daily at 2 AM - refresh schools
    s.cron.AddFunc("0 2 * * *", func() {
        log.Println("Running school refresh...")
        s.schoolService.RefreshSchoolsData()
    })
    
    // Every hour - cleanup old data
    s.cron.AddFunc("0 * * * *", func() {
        log.Println("Running cleanup...")
        s.cleanupOldData()
    })
    
    // Every Monday at 9 AM - send weekly report
    s.cron.AddFunc("0 9 * * 1", func() {
        log.Println("Sending weekly report...")
        s.sendWeeklyReport()
    })
    
    s.cron.Start()
}

func (s *Scheduler) cleanupOldData() {
    // Implementation
}

func (s *Scheduler) sendWeeklyReport() {
    // Implementation
}
```

## 🔐 Adding API Key Authentication

```go
// internal/middleware/auth.go

package middleware

import (
    "net/http"
    "os"
)

func APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        expectedKey := os.Getenv("API_KEY")
        
        if apiKey == "" || apiKey != expectedKey {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

Use in server:

```go
// internal/server/server.go

import "schools-be/internal/middleware"

// Protected routes
r.Group(func(r chi.Router) {
    r.Use(middleware.APIKeyAuth)
    r.Post("/refresh", schoolHandler.RefreshData)
    r.Delete("/schools/{id}", schoolHandler.DeleteSchool)
})
```

## 📝 Writing Tests

```go
// internal/service/school_service_test.go

package service

import (
    "testing"
)

func TestGetAllSchools(t *testing.T) {
    // Setup
    // ... create test database, repository, service
    
    // Execute
    schools, err := service.GetAllSchools()
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    
    if len(schools) != 2 {
        t.Errorf("expected 2 schools, got %d", len(schools))
    }
}
```

Run tests:
```bash
go test ./...
```

## 🌍 Environment-Based Config

```go
// internal/config/config.go

func (c *Config) GetDatabasePath() string {
    if c.Env == "test" {
        return ":memory:" // In-memory SQLite for tests
    }
    if c.Env == "production" {
        return "/var/data/schools.db"
    }
    return c.DBPath // development
}
```

---

These examples should cover most common scenarios! Copy and adapt them as needed. 🚀


