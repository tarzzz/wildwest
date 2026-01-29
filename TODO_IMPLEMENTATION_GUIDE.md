# Todo API Implementation Guide

## Overview
Step-by-step roadmap for implementing the Todo List API. This guide provides a logical sequence for building the API from foundation to deployment.

---

## Implementation Phases

### Phase 1: Foundation & Setup (Week 1)
MVP with core CRUD operations

### Phase 2: Enhancement & Testing (Week 2)
Advanced features and comprehensive testing

### Phase 3: Production Readiness (Week 3)
Security, monitoring, and deployment

---

## Phase 1: Foundation & Setup

### Step 1: Project Initialization

**Duration:** 2 hours

**Tasks:**
1. Create project structure
   ```bash
   mkdir -p todo-api/{cmd/api,internal/{handler,service,repository/postgres,middleware,domain,config},pkg/{database,logger},migrations,docs,scripts}
   cd todo-api
   go mod init wildwest/todo-api
   ```

2. Initialize Git repository
   ```bash
   git init
   git add .
   git commit -m "Initial project structure"
   ```

3. Create `.gitignore`
   ```
   # Binaries
   *.exe
   *.dll
   *.so
   *.dylib
   todo-api

   # Test files
   *.test
   *.out
   coverage.txt

   # IDE
   .idea/
   .vscode/
   *.swp
   *.swo

   # Environment
   .env
   .env.local

   # OS
   .DS_Store
   Thumbs.db

   # Air
   tmp/
   ```

4. Install core dependencies
   ```bash
   go get github.com/gin-gonic/gin@v1.11.0
   go get github.com/jackc/pgx/v5@v5.8.0
   go get github.com/rs/zerolog@v1.34.0
   go get github.com/spf13/viper@v1.18.2
   go get github.com/google/uuid@v1.6.0
   go get github.com/stretchr/testify@v1.11.1
   ```

**Deliverables:**
- ✅ Project structure created
- ✅ Dependencies installed
- ✅ Git repository initialized

---

### Step 2: Configuration Layer

**Duration:** 3 hours

**Implementation Order:**

1. **Create config struct** (`internal/config/config.go`)
   ```go
   type Config struct {
       App      AppConfig
       Server   ServerConfig
       Database DatabaseConfig
       Log      LogConfig
   }
   ```

2. **Implement config loading** with Viper
   - Support YAML files
   - Environment variable overrides
   - Default values
   - Validation on load

3. **Create `.todo-api.yaml`** with defaults

4. **Create `.env.example`** for developers

**Reference:** `user-management-api/internal/config/config.go`

**Deliverables:**
- ✅ Configuration struct defined
- ✅ Config loading implemented
- ✅ Validation logic added
- ✅ Example files created

---

### Step 3: Logging Setup

**Duration:** 2 hours

**Implementation:**

1. **Create logger package** (`pkg/logger/logger.go`)
   - Initialize zerolog
   - Console output (development)
   - JSON output (production)
   - Log levels from config

2. **Create logger middleware** (`internal/middleware/logger.go`)
   - Log all HTTP requests
   - Include: method, path, status, duration, IP
   - Use structured logging

**Reference:** `user-management-api/pkg/logger/logger.go`

**Deliverables:**
- ✅ Logger package created
- ✅ Logger middleware implemented
- ✅ Pretty console output for dev

---

### Step 4: Database Layer

**Duration:** 4 hours

**Implementation Order:**

1. **Create database package** (`pkg/database/postgres.go`)
   - Connection pool setup
   - Health check method
   - Transaction helper
   - Retry logic on connection failure
   - Context timeout handling

2. **Create migrations** (`migrations/`)
   - `000001_create_todos_table.up.sql`
   - `000001_create_todos_table.down.sql`
   - `000002_create_todo_tags_table.up.sql`
   - `000002_create_todo_tags_table.down.sql`
   - `000003_add_indexes.up.sql`
   - `000003_add_indexes.down.sql`

3. **Install golang-migrate**
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

4. **Create migration script** (`scripts/migrate.sh`)

**Reference:** `user-management-api/pkg/database/postgres.go`

**Deliverables:**
- ✅ Database connection package
- ✅ Migration files created
- ✅ Migration scripts working

---

### Step 5: Domain Models

**Duration:** 2 hours

**Implementation:**

1. **Create domain types** (`internal/domain/todo.go`)
   - Todo struct
   - Priority enum
   - Status enum
   - Validation tags

2. **Create DTOs** (`internal/domain/dto.go`)
   - CreateTodoRequest
   - UpdateTodoRequest
   - BulkUpdateRequest
   - TodoResponse
   - ListResponse

3. **Create domain errors** (`internal/domain/errors.go`)
   - ErrTodoNotFound
   - ErrValidationFailed
   - ErrInternalError

**Deliverables:**
- ✅ Domain models defined
- ✅ DTOs created
- ✅ Custom errors defined

---

### Step 6: Repository Layer

**Duration:** 6 hours

**Implementation Order:**

1. **Define repository interface** (`internal/repository/repository.go`)
   ```go
   type TodoRepository interface {
       Create(ctx context.Context, todo *domain.Todo) error
       GetByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error)
       List(ctx context.Context, filter ListFilter) ([]*domain.Todo, *Pagination, error)
       Update(ctx context.Context, todo *domain.Todo) error
       Delete(ctx context.Context, id uuid.UUID) error
       BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) (int, error)
       GetStats(ctx context.Context) (*domain.TodoStats, error)
   }
   ```

2. **Implement PostgreSQL repository** (`internal/repository/postgres/todo_repository_impl.go`)
   - Implement all interface methods
   - Handle tags in separate transactions
   - Use prepared statements
   - Proper error wrapping

3. **Write repository tests** (`internal/repository/postgres/todo_repository_impl_test.go`)
   - Use testcontainers or mock database
   - Test all CRUD operations
   - Test error cases

**Deliverables:**
- ✅ Repository interface defined
- ✅ PostgreSQL implementation complete
- ✅ Repository tests passing

---

### Step 7: Service Layer

**Duration:** 5 hours

**Implementation Order:**

1. **Define service interface** (`internal/service/service.go`)
   ```go
   type TodoService interface {
       CreateTodo(ctx context.Context, req *domain.CreateTodoRequest) (*domain.Todo, error)
       GetTodo(ctx context.Context, id uuid.UUID) (*domain.Todo, error)
       ListTodos(ctx context.Context, filter repository.ListFilter) ([]*domain.Todo, *repository.Pagination, error)
       UpdateTodo(ctx context.Context, id uuid.UUID, req *domain.UpdateTodoRequest) (*domain.Todo, error)
       DeleteTodo(ctx context.Context, id uuid.UUID) error
       BulkUpdateTodos(ctx context.Context, req *domain.BulkUpdateRequest) (int, error)
       GetStats(ctx context.Context) (*domain.TodoStats, error)
   }
   ```

2. **Implement service** (`internal/service/todo_service.go`)
   - Business logic
   - Input validation
   - DTO conversions
   - Error handling

3. **Write service tests** (`internal/service/todo_service_test.go`)
   - Mock repository
   - Test business logic
   - Test validation rules

**Deliverables:**
- ✅ Service interface defined
- ✅ Service implementation complete
- ✅ Service tests passing

---

### Step 8: Handler Layer

**Duration:** 6 hours

**Implementation Order:**

1. **Create handler struct** (`internal/handler/todo_handler.go`)
   ```go
   type TodoHandler struct {
       service service.TodoService
       logger  zerolog.Logger
   }
   ```

2. **Implement HTTP handlers:**
   - `CreateTodo(c *gin.Context)`
   - `GetTodo(c *gin.Context)`
   - `ListTodos(c *gin.Context)`
   - `UpdateTodo(c *gin.Context)`
   - `DeleteTodo(c *gin.Context)`
   - `BulkUpdateTodos(c *gin.Context)`
   - `GetStats(c *gin.Context)`

3. **Implement health handler** (`internal/handler/health_handler.go`)
   - GET /health (liveness)
   - GET /health/ready (readiness)
   - GET /metrics (database metrics)

4. **Write handler tests** (`internal/handler/todo_handler_test.go`)
   - Mock service
   - Test HTTP responses
   - Test status codes
   - Test error responses

**Reference:** `user-management-api/internal/handler/health_handler.go`

**Deliverables:**
- ✅ All handlers implemented
- ✅ Health endpoints working
- ✅ Handler tests passing

---

### Step 9: Middleware Setup

**Duration:** 3 hours

**Implementation:**

1. **Recovery middleware** (`internal/middleware/recovery.go`)
   - Catch panics
   - Log stack trace
   - Return 500 error

2. **CORS middleware** (`internal/middleware/cors.go`)
   - Configure allowed origins
   - Allow credentials
   - Set max age

3. **Request ID middleware** (`internal/middleware/request_id.go`)
   - Generate UUID for each request
   - Add to response headers
   - Include in logs

4. **Rate limiting middleware** (`internal/middleware/ratelimit.go`)
   - Per-IP limiting
   - 100 requests per minute
   - Return 429 when exceeded

**Deliverables:**
- ✅ All middleware implemented
- ✅ Middleware tests passing

---

### Step 10: Main Application & Routing

**Duration:** 4 hours

**Implementation:**

1. **Create main.go** (`cmd/api/main.go`)
   - Load configuration
   - Initialize logger
   - Connect to database
   - Setup Gin router
   - Register middleware
   - Register routes
   - Start server
   - Graceful shutdown

2. **Setup routing:**
   ```go
   v1 := router.Group("/api/v1")
   {
       todos := v1.Group("/todos")
       {
           todos.POST("", todoHandler.CreateTodo)
           todos.GET("", todoHandler.ListTodos)
           todos.GET("/:id", todoHandler.GetTodo)
           todos.PUT("/:id", todoHandler.UpdateTodo)
           todos.DELETE("/:id", todoHandler.DeleteTodo)
           todos.PATCH("/bulk", todoHandler.BulkUpdateTodos)
           todos.GET("/stats", todoHandler.GetStats)
       }

       v1.GET("/health", healthHandler.Health)
       v1.GET("/health/ready", healthHandler.Ready)
       v1.GET("/metrics", healthHandler.Metrics)
   }
   ```

3. **Graceful shutdown:**
   - Handle SIGINT and SIGTERM
   - Close database connections
   - Wait for in-flight requests

**Deliverables:**
- ✅ Main application working
- ✅ All routes registered
- ✅ Graceful shutdown implemented

---

### Step 11: Makefile & Scripts

**Duration:** 2 hours

**Create:**

1. **Makefile:**
   ```makefile
   .PHONY: build run test migrate-up migrate-down seed clean

   build:
       go build -o bin/todo-api cmd/api/main.go

   run:
       go run cmd/api/main.go

   test:
       go test -v -race -coverprofile=coverage.txt ./...

   test-unit:
       go test -v -race -short ./...

   migrate-up:
       migrate -path migrations -database "$(DB_URL)" up

   migrate-down:
       migrate -path migrations -database "$(DB_URL)" down

   seed:
       psql $(DB_URL) < scripts/seed.sql

   clean:
       rm -rf bin/ tmp/
   ```

2. **Seed script** (`scripts/seed.sql`)
   - Sample todos
   - Sample tags

**Deliverables:**
- ✅ Makefile created
- ✅ Scripts working

---

### Step 12: Local Testing

**Duration:** 3 hours

**Tasks:**

1. Start PostgreSQL locally
   ```bash
   docker run -d \
     --name postgres-todo \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=todos \
     -p 5432:5432 \
     postgres:15-alpine
   ```

2. Run migrations
   ```bash
   make migrate-up
   ```

3. Seed database
   ```bash
   make seed
   ```

4. Run application
   ```bash
   make run
   ```

5. Test endpoints with curl
   ```bash
   # Health check
   curl http://localhost:8080/api/v1/health

   # Create todo
   curl -X POST http://localhost:8080/api/v1/todos \
     -H "Content-Type: application/json" \
     -d '{"title": "Test todo", "priority": "high"}'

   # List todos
   curl http://localhost:8080/api/v1/todos
   ```

6. Run tests
   ```bash
   make test
   ```

**Deliverables:**
- ✅ Database running locally
- ✅ Application running
- ✅ All endpoints working
- ✅ All tests passing

---

## Phase 2: Enhancement & Testing

### Step 13: API Documentation (Swagger)

**Duration:** 4 hours

**Implementation:**

1. Install Swaggo
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   go get github.com/swaggo/gin-swagger
   go get github.com/swaggo/files
   ```

2. Add Swagger comments to handlers
   ```go
   // @Summary Create a new todo
   // @Description Create a new todo item
   // @Tags todos
   // @Accept json
   // @Produce json
   // @Param todo body CreateTodoRequest true "Todo to create"
   // @Success 201 {object} TodoResponse
   // @Failure 400 {object} ErrorResponse
   // @Router /todos [post]
   func (h *TodoHandler) CreateTodo(c *gin.Context) { ... }
   ```

3. Add main documentation
   ```go
   // @title Todo API
   // @version 1.0
   // @description RESTful API for managing todos
   // @host localhost:8080
   // @BasePath /api/v1
   func main() { ... }
   ```

4. Generate docs
   ```bash
   swag init -g cmd/api/main.go -o docs
   ```

5. Register Swagger route
   ```go
   router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   ```

**Deliverables:**
- ✅ Swagger annotations added
- ✅ Docs generated
- ✅ Swagger UI accessible at `/swagger/index.html`

---

### Step 14: Comprehensive Testing

**Duration:** 8 hours

**Test Coverage Goals:**
- Unit tests: 80%+
- Integration tests: Critical paths
- API tests: All endpoints

**Implementation:**

1. **Unit tests:**
   - Domain model validation
   - Service business logic
   - Utility functions

2. **Integration tests:**
   - Database operations (use testcontainers)
   - Transaction handling
   - Error scenarios

3. **API tests:**
   - HTTP endpoint testing
   - Request validation
   - Response formats
   - Error responses

4. **Test helpers:**
   - Mock factories
   - Test data builders
   - Database setup/teardown

5. **Run coverage report:**
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

**Deliverables:**
- ✅ 80%+ test coverage
- ✅ All critical paths tested
- ✅ Integration tests passing

---

### Step 15: Performance Testing

**Duration:** 3 hours

**Tools:**
- **wrk** or **hey** for load testing
- **pprof** for profiling

**Implementation:**

1. Load test endpoints
   ```bash
   hey -n 10000 -c 100 http://localhost:8080/api/v1/todos
   ```

2. Profile application
   ```go
   import _ "net/http/pprof"
   go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
   }()
   ```

3. Analyze bottlenecks
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile
   ```

4. Optimize:
   - Database queries
   - JSON serialization
   - Memory allocations

**Deliverables:**
- ✅ Load test results documented
- ✅ Performance bottlenecks identified
- ✅ Optimizations applied

---

### Step 16: Error Handling Improvements

**Duration:** 3 hours

**Implementation:**

1. Create consistent error response format
2. Add error codes for all errors
3. Improve validation error messages
4. Add context to internal errors
5. Create error mapping helper

**Deliverables:**
- ✅ Consistent error responses
- ✅ User-friendly error messages
- ✅ Proper HTTP status codes

---

## Phase 3: Production Readiness

### Step 17: Dockerfile & Docker Compose

**Duration:** 3 hours

**Implementation:**

1. **Create Dockerfile:**
   - Multi-stage build
   - Alpine base image
   - Non-root user
   - Health check

2. **Create docker-compose.yml:**
   ```yaml
   version: '3.8'
   services:
     api:
       build: .
       ports:
         - "8080:8080"
       depends_on:
         - db
       environment:
         - DB_HOST=db

     db:
       image: postgres:15-alpine
       environment:
         - POSTGRES_DB=todos
         - POSTGRES_PASSWORD=password
       volumes:
         - postgres-data:/var/lib/postgresql/data

   volumes:
     postgres-data:
   ```

3. **Test Docker setup:**
   ```bash
   docker-compose up
   ```

**Deliverables:**
- ✅ Dockerfile created
- ✅ Docker Compose working
- ✅ Application runs in container

---

### Step 18: CI/CD Pipeline

**Duration:** 4 hours

**Implementation:**

1. **Create GitHub Actions workflow** (`.github/workflows/ci.yml`):
   ```yaml
   name: CI
   on: [push, pull_request]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-go@v4
         - run: go test -race ./...

     build:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: docker/build-push-action@v4
   ```

2. **Add lint job:**
   - Install golangci-lint
   - Run linter

3. **Add build job:**
   - Build Docker image
   - Push to registry (if main branch)

**Deliverables:**
- ✅ CI pipeline working
- ✅ Tests run on PR
- ✅ Docker image built automatically

---

### Step 19: Monitoring & Observability

**Duration:** 5 hours

**Implementation:**

1. **Add Prometheus metrics:**
   ```bash
   go get github.com/prometheus/client_golang
   ```

2. **Expose metrics:**
   - HTTP request duration histogram
   - HTTP request counter
   - Database pool stats
   - Error counter by type

3. **Create Grafana dashboard:**
   - Request rate
   - Error rate
   - Response time (p50, p95, p99)
   - Database connections

4. **Add structured logging:**
   - Request ID in all logs
   - User ID (when auth added)
   - Error context

**Deliverables:**
- ✅ Prometheus metrics exposed
- ✅ Grafana dashboard created
- ✅ Structured logging improved

---

### Step 20: Documentation

**Duration:** 4 hours

**Create:**

1. **README.md:**
   - Project overview
   - Features
   - Prerequisites
   - Installation
   - Configuration
   - Running locally
   - API documentation link
   - Contributing guidelines

2. **API.md:**
   - Detailed API documentation
   - Examples for each endpoint
   - Error codes reference

3. **DEPLOYMENT.md:**
   - Deployment instructions
   - Environment variables
   - Database setup
   - Kubernetes manifests

4. **DEVELOPMENT.md:**
   - Development setup
   - Project structure
   - Coding conventions
   - Testing guidelines

**Deliverables:**
- ✅ Complete documentation
- ✅ Examples provided
- ✅ Deployment guide ready

---

## Optional Enhancements (Phase 2+)

### Authentication (JWT)

**Duration:** 6 hours

**Implementation:**
1. Install JWT library
2. Create auth middleware
3. Generate/validate tokens
4. Add user context to requests
5. Protect endpoints

---

### Caching (Redis)

**Duration:** 4 hours

**Implementation:**
1. Install Redis client
2. Cache GET requests
3. Invalidate on updates
4. Add cache metrics

---

### Rate Limiting (Redis-backed)

**Duration:** 3 hours

**Implementation:**
1. Replace in-memory rate limiter
2. Use Redis for distributed limiting
3. Per-user rate limits (after auth)

---

### Cursor-based Pagination

**Duration:** 4 hours

**Implementation:**
1. Add cursor encoding/decoding
2. Update repository queries
3. Maintain backward compatibility

---

### Event Publishing

**Duration:** 6 hours

**Implementation:**
1. Add event publisher interface
2. Publish todo events (created, updated, deleted)
3. Integrate with message queue (RabbitMQ, Kafka)

---

## Deployment Checklist

### Pre-Deployment
- [ ] All tests passing (unit, integration, API)
- [ ] Linter passing (no warnings)
- [ ] Code reviewed
- [ ] Documentation updated
- [ ] Environment variables documented
- [ ] Database migrations tested
- [ ] Load testing completed
- [ ] Security audit done

### Deployment
- [ ] Database migrations applied
- [ ] Configuration reviewed
- [ ] Secrets stored securely (Kubernetes secrets)
- [ ] Health checks configured
- [ ] Monitoring alerts setup
- [ ] Log aggregation configured
- [ ] Backup strategy in place

### Post-Deployment
- [ ] Health check passing
- [ ] Smoke tests passing
- [ ] Monitoring dashboards showing data
- [ ] Logs flowing to aggregator
- [ ] Performance metrics baseline recorded
- [ ] Documentation updated with production URLs

---

## Team Roles & Responsibilities

### Solutions Architect (This Role)
- ✅ API design
- ✅ Data model design
- ✅ Technology decisions
- ✅ Implementation guide

### Software Engineers
- [ ] Implement repository layer
- [ ] Implement service layer
- [ ] Implement handler layer
- [ ] Write tests
- [ ] Code reviews

### QA Engineer
- [ ] Test plan creation
- [ ] Manual testing
- [ ] Automated testing
- [ ] Performance testing
- [ ] Security testing

### DevOps/SRE
- [ ] Infrastructure setup
- [ ] CI/CD pipeline
- [ ] Monitoring setup
- [ ] Deployment automation
- [ ] Database management

---

## Risk Mitigation

### Technical Risks

1. **Database Performance**
   - Mitigation: Proper indexing, connection pooling, query optimization
   - Monitor: Query execution time, connection pool stats

2. **API Scalability**
   - Mitigation: Stateless design, horizontal scaling, caching
   - Monitor: Response time, error rate, throughput

3. **Data Loss**
   - Mitigation: Database backups, soft deletes, transaction safety
   - Monitor: Backup success, data integrity checks

### Process Risks

1. **Timeline Slippage**
   - Mitigation: Prioritize MVP features, defer nice-to-haves
   - Monitor: Task completion, velocity

2. **Quality Issues**
   - Mitigation: Test coverage requirements, code reviews, CI checks
   - Monitor: Test coverage, bug count, code quality metrics

---

## Success Criteria

### Functional
- ✅ All CRUD endpoints working
- ✅ Health checks passing
- ✅ API documentation complete
- ✅ Test coverage > 80%

### Non-Functional
- ✅ Response time < 100ms (p95)
- ✅ Uptime > 99.9%
- ✅ Handle 1000 requests/second
- ✅ Zero data loss

### Operational
- ✅ Automated deployment working
- ✅ Monitoring dashboards setup
- ✅ Alerts configured
- ✅ Documentation complete

---

## Timeline Summary

| Phase | Duration | Effort |
|---|---|---|
| Phase 1: Foundation & Setup | Week 1 | 44 hours |
| Phase 2: Enhancement & Testing | Week 2 | 22 hours |
| Phase 3: Production Readiness | Week 3 | 23 hours |
| **Total** | **3 weeks** | **89 hours** |

**Team Size:** 2 engineers + 1 QA + 1 DevOps

---

## Next Steps

1. **Get approval** on architecture and implementation plan
2. **Assign tasks** to software engineers via their `instructions.md`
3. **Setup infrastructure** (database, CI/CD)
4. **Begin Phase 1** implementation
5. **Daily standups** to track progress
6. **Weekly reviews** to adjust course

---

## Questions for Engineering Manager

Before proceeding, please clarify:

1. **Timeline flexibility:** Is 3-week timeline acceptable or need faster MVP?
2. **Team allocation:** How many engineers available? Full-time or part-time?
3. **Infrastructure:** Do we have PostgreSQL available or need to provision?
4. **Authentication:** Required for MVP or can defer to Phase 2?
5. **Deployment target:** Kubernetes, Docker Compose, or other?
6. **Monitoring:** Prometheus/Grafana already available or need setup?

---

## Conclusion

This implementation guide provides a clear, step-by-step roadmap for building the Todo API from scratch to production deployment. Each step is scoped, sequenced logically, and includes concrete deliverables.

The architecture leverages wildwest's existing patterns (clean architecture, Gin, PostgreSQL, pgx) ensuring consistency and maintainability.

Ready to proceed with implementation upon approval.
