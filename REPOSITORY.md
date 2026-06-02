# Vedsutra - School ERP Backend API

## Project Overview

**Vedsutra** is a modern **School ERP (Enterprise Resource Planning) Backend** written in **Go**, designed as a RESTful API for managing school organizations and class structures. The project follows a **clean architecture** pattern with clear separation of concerns across multiple layers.

**Repository:** https://github.com/shashisharma307703/vedsutra  
**Language:** Go (1.25.4)  
**Primary Framework:** Chi Router (v5.3.0)  
**Database:** PostgreSQL  

---

## Project Architecture

The application follows a **layered architecture pattern**:

```
vedsutra/
├── cmd/api/                 # Application Entry Point
├── internal/
│   ├── app/                # Core Application & Server Setup
│   ├── handler/            # HTTP Request Handlers (Presentation Layer)
│   ├── service/            # Business Logic Layer
│   └── repository/         # Data Access Layer (Database Operations)
├── config/                 # Configuration Management
├── db/                     # Database Schema & Queries
│   ├── queries/            # SQL Query Definitions
│   ├── types/              # Generated Type Definitions
│   ├── dbgen/              # Auto-generated Database Code
│   ├── vedsutra.sql        # Main Database Schema
│   └── obsolete_schema.sql # Legacy Schema Reference
└── csv/                    # CSV Data Files
```

### Architecture Layers

1. **Presentation Layer (Handler)** - HTTP endpoints that handle requests/responses
2. **Service Layer** - Business logic and orchestration
3. **Repository Layer** - Database interaction and data persistence
4. **Infrastructure Layer** - Database connections, configuration

---

## Core Components

### 1. Entry Point: `cmd/api/main.go`

```go
package main

func main() {
	ctx := context.Background()
	
	// Load configuration from environment
	cfg := config.Load()
	
	// Initialize application (DB, services, handlers)
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application engine: %v", err)
	}
	defer application.Close()
	
	// Start HTTP server
	if err := application.Run(); err != nil {
		log.Fatalf("server crash boundary encountered: %v", err)
	}
}
```

**Purpose:** Bootstraps the application by loading configuration, initializing dependencies, and starting the server.

---

### 2. Application Initialization: `internal/app/app.go`

The `App` struct coordinates all application components:

```go
type App struct {
	Cfg    *config.Config
	Pool   *pgxpool.Pool
	Server *Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// Initialize database connection pool
	pool, err := repository.InitPool(ctx, cfg.Database)
	
	// Instantiate layers:
	// Repository → Service → Handler
	repo := repository.NewRepository(pool)
	
	orgSvc := service.NewOrgService(repo)
	orgHnd := handler.NewOrgHandler(orgSvc)
	
	classSvc := service.NewClassService(repo)
	classHnd := handler.NewClassHandler(classSvc)
	
	// Create HTTP server with routes
	srv := NewServer(cfg.Server.Port, cfg.Server.ReadTimeout, orgHnd, classHnd)
	
	return &App{Cfg: cfg, Pool: pool, Server: srv}, nil
}
```

**Responsibilities:**
- Initialize PostgreSQL connection pool
- Instantiate service & handler layers
- Create HTTP server with routes
- Manage application lifecycle

---

### 3. HTTP Server & Routes: `internal/app/server.go`

Built with **Chi Router** (lightweight, idiomatic Go HTTP router):

```go
type Server struct {
	router *chi.Mux
	port   string
}

func NewServer(port string, timeout time.Duration, 
	orgHrv *handler.OrgHandler, classHrv *handler.ClassHandler) *Server {
	r := chi.NewRouter()
	
	// Middleware stack
	r.Use(middleware.Logger)      // Request logging
	r.Use(middleware.Recoverer)   // Panic recovery
	r.Use(middleware.Timeout)     // Request timeout
	
	// API Routes v1
	r.Route("/api/v1", func(r chi.Router) {
		// Organizations endpoints
		r.Route("/organizations", func(r chi.Router) {
			r.Post("/", orgHrv.Create)
			r.Get("/", orgHrv.List)
			r.Route("/{orgId}", func(r chi.Router) {
				r.Get("/", orgHrv.Get)
				r.Put("/", orgHrv.Update)
				r.Patch("/", orgHrv.Patch)
				r.Delete("/", orgHrv.Delete)
			})
		})
		
		// Classes endpoints
		r.Route("/classes", func(r chi.Router) {
			r.Post("/", classHrv.Create)
			r.Get("/", classHrv.List)
			r.Route("/{classId}", func(r chi.Router) {
				r.Get("/", classHrv.Get)
				r.Put("/", classHrv.Update)
				r.Patch("/", classHrv.Patch)
				r.Delete("/", classHrv.Delete)
			})
		})
	})
	
	return &Server{router: r, port: port}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.port, s.router)
}
```

**API Endpoints:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| **Organizations** |
| POST | `/api/v1/organizations` | Create organization |
| GET | `/api/v1/organizations` | List organizations (with search/filter) |
| GET | `/api/v1/organizations/{orgId}` | Get organization details |
| PUT | `/api/v1/organizations/{orgId}` | Update organization (full) |
| PATCH | `/api/v1/organizations/{orgId}` | Update organization (partial) |
| DELETE | `/api/v1/organizations/{orgId}` | Delete organization |
| **Classes** |
| POST | `/api/v1/classes` | Create class |
| GET | `/api/v1/classes` | List classes (with search/filter) |
| GET | `/api/v1/classes/{classId}` | Get class details |
| PUT | `/api/v1/classes/{classId}` | Update class (full) |
| PATCH | `/api/v1/classes/{classId}` | Update class (partial) |
| DELETE | `/api/v1/classes/{classId}` | Delete class |

---

### 4. Configuration: `config/config.go`

Environment-based configuration management:

```go
type Config struct {
	Server   ServerConfig
	Database PoolConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port            string        // Server bind address (default: :8080)
	Mode            string        // Debug/Release mode
	ReadTimeout     time.Duration // Request read timeout (default: 30s)
	WriteTimeout    time.Duration // Response write timeout (default: 30s)
	ShutdownTimeout time.Duration // Graceful shutdown timeout (default: 5s)
}

type PoolConfig struct {
	Host                 string        // PostgreSQL host
	Port                 string        // PostgreSQL port (default: 5432)
	User                 string        // DB user
	Password             string        // DB password
	Database             string        // Database name
	MaxConnections       int32         // Connection pool max (default: 5)
	MinConnections       int32         // Connection pool min (default: 2)
	ConnMaxLifetime      time.Duration // Connection lifetime (default: 15m)
	ConnMaxIdleTime      time.Duration // Idle timeout (default: 5m)
	AcquireTimeout       time.Duration // Acquire connection timeout (default: 5s)
	ConnectionTimeout    time.Duration // Connection timeout (default: 10s)
}
```

**Environment Variables** (.env file):
```
SERVER_PORT=:8080
SERVER_TIMEOUT_SECONDS=60
PGPOOL_HOST=localhost
PGPOOL_PORT=5432
DB_USER=sms_admin
DB_PASSWORD=123456
DB_NAME=school_erp
```

---

### 5. Database Layer

#### Connection Pool: `internal/repository/db.go`

```go
func InitPool(ctx context.Context, cfg config.PoolConfig) (*pgxpool.Pool, error) {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable...", 
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	
	// Parse and configure connection pool
	poolCfg, _ := pgxpool.ParseConfig(dsn)
	poolCfg.MaxConns = cfg.MaxConnections
	poolCfg.MinConns = cfg.MinConnections
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolCfg.HealthCheckPeriod = 1 * time.Minute
	
	// Create pool with timeout context
	connectCtx, cancel := context.WithTimeout(ctx, cfg.AcquireTimeout)
	defer cancel()
	
	pool, _ := pgxpool.NewWithConfig(connectCtx, poolCfg)
	
	// Verify connection
	pool.Ping(connectCtx)
	return pool, nil
}
```

**Features:**
- Uses `pgx` (PostgreSQL native driver) for type-safe database access
- Implements connection pooling with configurable limits
- Health check monitoring (1-minute intervals)
- Connection lifecycle management

#### Repository Layer: `internal/repository/`

**Base Repository:**
- Manages database connection pool
- Provides access to generated queries

**Organization Repository** (`org_repo.go`):
```go
func (r *Repository) FindOrganizations(ctx context.Context, 
	f OrgListAndSearchFilter) ([]dbgen.Organization, error)

func (r *Repository) PatchOrganization(ctx context.Context, 
	id uuid.UUID, updates map[string]interface{}) (*dbgen.Organization, error)
```

**Features:**
- Uses **Squirrel** query builder for SQL safety
- Full-text search on name, code, email
- Optional city filtering
- Pagination with limit/offset
- Dynamic PATCH updates

**Class Repository** (`class_repo.go`):
- Similar CRUD operations for classes
- Organization-aware queries
- Active status filtering

---

### 6. Handler Layer: `internal/handler/`

**Organization Handler** (`org_handler.go`):
```go
type OrgHandler struct {
	svc *service.OrgService
}

Methods:
- Create(w, r)    // POST - Parse JSON body, validate, call service
- Get(w, r)       // GET - Parse UUID param, fetch, return JSON
- List(w, r)      // GET - Parse query params (limit, offset, search), return list
- Update(w, r)    // PUT - Full replacement
- Patch(w, r)     // PATCH - Partial update
- Delete(w, r)    // DELETE - Remove record
```

**Class Handler** (`class_handler.go`):
- Similar interface for class operations
- Multi-parameter UUID extraction (orgId, classId)
- Filters: active status, search term, organization scoping

**Standard Response Pattern:**
- JSON encoding for success responses
- HTTP error codes for failures
- Proper status codes (201 Created, 204 No Content, 400 Bad Request, etc.)

---

### 7. Service Layer: `internal/service/`

**Organization Service** (`org_service.go`):
```go
type OrgService struct {
	repo *repository.Repository
}

Methods:
- Create(ctx, params) → organization
- Get(ctx, id) → organization
- List(ctx, filter) → []organization
- Update(ctx, params) → organization
- Patch(ctx, id, updates) → organization
- Delete(ctx, id) → error
```

**Class Service** (`class_service.go`):
- Orchestrates class-related business logic
- Delegates to repository layer

**Responsibilities:**
- Validate input parameters
- Apply business rules
- Call repository methods
- Transform data as needed

---

## Dependencies

From `go.mod`:

```
module github.com/shashisharma307703/vedantam

go 1.25.4

require:
  - github.com/Masterminds/squirrel v1.5.4      # SQL query builder
  - github.com/go-chi/chi/v5 v5.3.0             # HTTP router
  - github.com/google/uuid v1.6.0               # UUID generation
  - github.com/jackc/pgx/v5 v5.9.2             # PostgreSQL driver
```

**Key Libraries:**
- **Chi Router**: Lightweight, composable HTTP router middleware
- **pgx/v5**: Type-safe PostgreSQL driver with connection pooling
- **Squirrel**: SQL query builder for safe, parameterized queries
- **UUID**: RFC 4122 compliant UUID generation

---

## Database Schema

The project includes comprehensive database schema:

- **Main Schema:** `db/vedsutra.sql` (201KB+ - production schema)
- **Legacy Schema:** `db/obsolete_schema.sql` (60KB+ - reference/archived)
- **Query Definitions:** `db/queries/` - SQL queries for code generation
- **Generated Types:** `db/dbgen/` - Auto-generated Go types

**Primary Entities:**
- `organizations` - School/institution data
- `class_levels` - Class/grade definitions

**Schema Management:**
- Uses SQLC for type-safe query generation
- Configuration in `sqlc.yaml`
- Supports UUID for distributed data (org isolation)

---

## Configuration Files

### `sqlc.yaml` & `old_sqlc.yaml`
- SQLC configuration for code generation
- Defines query-to-Go mapping
- Output paths for generated types

### `.env`
- Local development environment variables
- PostgreSQL connection details
- Server configuration

### `.vscode/`
- VS Code workspace configuration

---

## Project Workflows

### Startup Flow
1. **main()** calls `config.Load()` → Load environment variables
2. Call `app.NewApp()` → Initialize dependencies
3. Create database connection pool → Verify connectivity
4. Instantiate service & handler layers (dependency injection)
5. Create HTTP server with routes
6. Call `app.Run()` → Start listening on configured port

### Request Flow (Example: GET /api/v1/organizations)
1. HTTP request arrives at Chi router
2. Router matches to `orgHnd.List()`
3. Handler extracts query params (limit, offset, search, city)
4. Calls `orgSvc.List(ctx, filter)`
5. Service calls `repo.FindOrganizations(ctx, filter)`
6. Repository builds SQL query with Squirrel builder
7. Executes query against PostgreSQL pool
8. Rows scanned into `Organization` structs
9. Response marshaled to JSON → sent to client

### Data Flow (Example: POST /api/v1/organizations)
```
JSON Request
    ↓
JSON Decode → OrgHandler.Create()
    ↓
Service.Create() - Business Logic
    ↓
Repository.UpsertOrganization() - DB Call
    ↓
PostgreSQL Row Insert/Update
    ↓
Scan Result → Organization Struct
    ↓
JSON Encode → HTTP Response (201 Created)
```

---

## Database Entities

### Organization
```sql
org_id           UUID PRIMARY KEY
org_name         VARCHAR
org_code         VARCHAR UNIQUE
phone_number     VARCHAR
email            VARCHAR
website          VARCHAR
logo_url         VARCHAR
address          TEXT
city             VARCHAR
state            VARCHAR
country          VARCHAR
postal_code      VARCHAR
established_date DATE
affiliation_number VARCHAR
license_number   VARCHAR
license_expiry   DATE
created_at       TIMESTAMP
updated_at       TIMESTAMP
```

### ClassLevel (Classes)
```sql
class_level_id   UUID PRIMARY KEY
org_id           UUID (FK to organizations)
class_name       VARCHAR
class_code       VARCHAR
is_active        BOOLEAN
description      TEXT
created_at       TIMESTAMP
updated_at       TIMESTAMP
```

---

## Features & Capabilities

✅ **RESTful API** - Clean, versioned endpoints (/api/v1)  
✅ **CRUD Operations** - Full Create, Read, Update, Delete support  
✅ **Advanced Search** - Text search across multiple fields  
✅ **Filtering** - City-based, active status, organization scoping  
✅ **Pagination** - Limit/offset based result sets  
✅ **Partial Updates** - PATCH support for selective field updates  
✅ **Connection Pooling** - Efficient database resource management  
✅ **Error Handling** - Proper HTTP status codes & error messages  
✅ **Middleware** - Logging, panic recovery, request timeout  
✅ **Clean Architecture** - Separation of concerns, testability  
✅ **Environment Configuration** - 12-factor app principles  
✅ **Type Safety** - UUID-based IDs, generated types  

---

## Running the Application

### Prerequisites
- Go 1.25.4+
- PostgreSQL 12+
- Environment variables configured in `.env`

### Startup
```bash
# Navigate to project
cd vedsutra

# Load environment
source .env  # or set environment variables

# Run application
go run cmd/api/main.go

# Server starts on :8080
# Log output shows: "Server booting successfully on :8080..."
```

### Building
```bash
# Build binary
go build -o vedsutra cmd/api/main.go

# Run binary
./vedsutra
```

---

## Code Generation

The project uses **SQLC** for type-safe SQL:

```bash
# Generate types from SQL queries
sqlc generate
```

- Reads queries from `db/queries/*.sql`
- Generates Go code to `db/dbgen/`
- Keeps application code in sync with database schema

---

## Directory Structure Summary

| Directory | Purpose |
|-----------|---------|
| `cmd/api/` | Application entry point |
| `internal/app/` | Core app & server setup |
| `internal/handler/` | HTTP request handlers |
| `internal/service/` | Business logic |
| `internal/repository/` | Database operations |
| `config/` | Configuration management |
| `db/` | Schema, queries, generated types |
| `csv/` | Data files |
| `.vscode/` | IDE settings |

---

## Technology Stack Summary

| Layer | Technology |
|-------|-----------|
| **Framework** | Go 1.25.4 |
| **HTTP Router** | Chi v5.3.0 |
| **Database** | PostgreSQL 12+ |
| **DB Driver** | pgx/v5 9.9.2 |
| **Query Builder** | Squirrel v1.5.4 |
| **UUID** | google/uuid v1.6.0 |
| **Architecture** | Layered/Clean Architecture |
| **API Style** | RESTful |
| **Concurrency** | Context-aware |

---

## Status

🚀 **Active Development** - Project was created on June 2, 2026  
📦 **Early Stage** - Foundation layers in place (app, handlers, repository, config)  
🔄 **Ready for** - Feature development, integration testing, deployment  

---

## Key Observations

1. **Clean Architecture** - Well-separated concerns with clear dependency flow
2. **Type Safety** - UUID-based identifiers, generated SQL types
3. **Scalability** - Connection pooling, pagination support, service layer abstraction
4. **Maintainability** - Structured folder layout, middleware composition
5. **PostgreSQL First** - Full reliance on PG features (JSON, arrays, etc.)
6. **Development Ready** - Environment configuration, logging, panic recovery included

---

**Repository Link:** https://github.com/shashisharma307703/vedsutra  
**Owner:** @shashisharma307703  
**Language Composition:** Go (Primary)
