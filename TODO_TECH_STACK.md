# Todo API Technology Stack

## Overview
Technology choices and architectural decisions for the Todo List API, aligned with the wildwest project's existing patterns.

---

## Core Technology Stack

### 1. Programming Language: **Go 1.24**

**Rationale:**
- ✅ Already used in wildwest project
- ✅ Excellent performance and concurrency
- ✅ Strong standard library
- ✅ Fast compilation and deployment
- ✅ Great for building RESTful APIs
- ✅ Strong typing and compile-time safety
- ✅ Excellent tooling and ecosystem

**Version:** 1.24.0 (aligned with wildwest's go.mod)

---

### 2. HTTP Framework: **Gin**

**Package:** `github.com/gin-gonic/gin` v1.11.0

**Rationale:**
- ✅ Already in wildwest's go.mod
- ✅ High performance (fastest Go web framework)
- ✅ Clean routing and middleware support
- ✅ Built-in validation with struct tags
- ✅ JSON serialization/deserialization
- ✅ Extensive middleware ecosystem
- ✅ Excellent documentation
- ✅ Used in user-management-api (consistency)

**Alternatives Considered:**
- **Echo**: Similar performance, less adoption
- **Fiber**: Fastest but Express.js-like API (not idiomatic Go)
- **Chi**: Lightweight but more manual setup required
- **Standard net/http**: Too low-level for rapid development

**Winner:** Gin (already used, proven in wildwest)

---

### 3. Database: **PostgreSQL 15+**

**Driver:** `github.com/jackc/pgx/v5` v5.8.0

**Rationale:**
- ✅ Already used in user-management-api
- ✅ ACID compliance
- ✅ Advanced features (JSON, full-text search, triggers)
- ✅ Excellent performance
- ✅ Strong community support
- ✅ pgx/v5 is fastest Go PostgreSQL driver
- ✅ Built-in connection pooling
- ✅ Production-ready and battle-tested

**Why pgx over database/sql?**
- Better performance (native protocol)
- Type safety with PostgreSQL types
- Connection pooling built-in
- Better error messages
- More PostgreSQL-specific features

**Alternatives Considered:**
- **MySQL**: Less advanced features, weaker JSON support
- **SQLite**: Not suitable for production, no concurrent writes
- **MongoDB**: Overkill for simple CRUD, schema flexibility not needed

**Winner:** PostgreSQL with pgx/v5 (consistency with wildwest)

---

### 4. Database Migrations: **golang-migrate**

**Package:** `github.com/golang-migrate/migrate/v4`

**Rationale:**
- ✅ Industry standard for Go
- ✅ CLI tool + Go library
- ✅ Version control for schema
- ✅ Up/down migrations
- ✅ Multiple database support
- ✅ Simple SQL-based migrations

**Migration Strategy:**
- Store migrations in `migrations/` directory
- Naming: `000001_description.up.sql` / `000001_description.down.sql`
- Run migrations on startup (development)
- Manual migrations in production (safety)

---

### 5. Configuration Management: **Viper**

**Package:** `github.com/spf13/viper` v1.18.2

**Rationale:**
- ✅ Already used in wildwest
- ✅ Multiple config sources (files, env vars, flags)
- ✅ Live config reloading
- ✅ Multiple format support (YAML, JSON, TOML)
- ✅ Environment variable overrides
- ✅ Integration with Cobra CLI

**Configuration Priority:**
1. Command-line flags
2. Environment variables
3. Config file (`.todo-api.yaml`)
4. Defaults

---

### 6. Logging: **Zerolog**

**Package:** `github.com/rs/zerolog` v1.34.0

**Rationale:**
- ✅ Already used in wildwest (user-management-api)
- ✅ Zero allocation (high performance)
- ✅ Structured logging (JSON output)
- ✅ Contextual logging
- ✅ Multiple log levels
- ✅ Pretty console output for development

**Log Levels:**
- `debug`: Development debugging
- `info`: Standard operational messages
- `warn`: Warning conditions
- `error`: Error conditions
- `fatal`: Critical errors (exit)

---

### 7. Validation: **Gin Binding + validator**

**Package:** `github.com/go-playground/validator/v10`

**Rationale:**
- ✅ Integrated with Gin
- ✅ Struct tag-based validation
- ✅ Custom validators support
- ✅ Comprehensive validation rules
- ✅ i18n support for error messages

**Example:**
```go
type CreateTodoRequest struct {
    Title string `json:"title" binding:"required,min=1,max=200"`
    Priority string `json:"priority" binding:"omitempty,oneof=low medium high"`
}
```

---

### 8. UUID Generation: **Google UUID**

**Package:** `github.com/google/uuid`

**Rationale:**
- ✅ Industry standard
- ✅ Clean API
- ✅ UUID v4 support
- ✅ PostgreSQL uuid type compatible

---

### 9. Testing Framework: **Testify**

**Package:** `github.com/stretchr/testify` v1.11.1

**Rationale:**
- ✅ Already in wildwest's go.mod
- ✅ Rich assertion library
- ✅ Mocking support
- ✅ Test suites
- ✅ Table-driven test helpers

**Testing Strategy:**
- Unit tests for business logic
- Integration tests for database operations
- API tests for endpoints
- Mock database for unit tests
- Real database for integration tests

---

### 10. API Documentation: **Swaggo**

**Package:** `github.com/swaggo/gin-swagger`

**Rationale:**
- ✅ Generates OpenAPI/Swagger docs from code comments
- ✅ Interactive Swagger UI
- ✅ Keeps docs in sync with code
- ✅ Industry standard format

**Usage:**
```go
// @Summary Create a new todo
// @Description Create a new todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body CreateTodoRequest true "Todo to create"
// @Success 201 {object} TodoResponse
// @Router /todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) { ... }
```

---

## Project Structure

### Clean Architecture Pattern

Following user-management-api's structure:

```
todo-api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handler/                 # HTTP handlers (controllers)
│   │   ├── todo_handler.go
│   │   ├── health_handler.go
│   │   └── handler.go
│   ├── service/                 # Business logic
│   │   ├── todo_service.go
│   │   └── service.go
│   ├── repository/              # Data access layer
│   │   ├── todo_repository.go
│   │   ├── postgres/
│   │   │   └── todo_repository_impl.go
│   │   └── repository.go
│   ├── middleware/              # HTTP middleware
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   └── ratelimit.go
│   ├── domain/                  # Domain models
│   │   ├── todo.go
│   │   ├── stats.go
│   │   └── errors.go
│   └── config/                  # Configuration
│       └── config.go
├── pkg/
│   ├── database/                # Database connection
│   │   └── postgres.go
│   └── logger/                  # Logger setup
│       └── logger.go
├── migrations/                  # Database migrations
│   ├── 000001_create_todos.up.sql
│   └── 000001_create_todos.down.sql
├── docs/                        # Generated API docs (Swagger)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── scripts/                     # Utility scripts
│   ├── migrate.sh
│   └── seed.sh
├── .env.example                 # Example environment variables
├── .todo-api.yaml               # Default configuration
├── Dockerfile                   # Container image
├── Makefile                     # Build automation
├── go.mod
├── go.sum
└── README.md
```

**Architecture Layers:**

1. **Handler Layer** (HTTP):
   - Request parsing and validation
   - Response formatting
   - HTTP status codes
   - Calls service layer

2. **Service Layer** (Business Logic):
   - Business rules and validation
   - Orchestrates repository calls
   - Transaction management
   - No HTTP or database knowledge

3. **Repository Layer** (Data Access):
   - CRUD operations
   - Query building
   - Database-specific logic
   - Returns domain models

4. **Domain Layer** (Models):
   - Data structures
   - Business entities
   - No dependencies

**Benefits:**
- Clear separation of concerns
- Testable (mock each layer)
- Maintainable
- Scalable
- Follows SOLID principles

---

## Development Tools

### 1. Build Tool: **Make**

**Makefile targets:**
```makefile
build:          Build the application
run:            Run the application
test:           Run all tests
test-unit:      Run unit tests
test-integration: Run integration tests
migrate-up:     Apply database migrations
migrate-down:   Rollback database migrations
seed:           Seed database with sample data
lint:           Run linter (golangci-lint)
fmt:            Format code (gofmt)
swagger:        Generate Swagger docs
clean:          Clean build artifacts
docker-build:   Build Docker image
docker-run:     Run Docker container
```

---

### 2. Linter: **golangci-lint**

**Configuration:** `.golangci.yml`

**Enabled Linters:**
- `gofmt`: Code formatting
- `govet`: Go vet
- `errcheck`: Unchecked errors
- `staticcheck`: Static analysis
- `unused`: Unused code
- `gosimple`: Simplifications
- `structcheck`: Unused struct fields
- `ineffassign`: Ineffective assignments

---

### 3. Local Development: **Air** (Hot Reload)

**Package:** `github.com/cosmtrek/air`

**Rationale:**
- Auto-reload on file changes
- Fast development iteration
- No manual restarts

**Configuration:** `.air.toml`

---

### 4. Database GUI: **pgAdmin** or **DBeaver**

For local database management and queries.

---

## Deployment Stack

### 1. Containerization: **Docker**

**Dockerfile:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o todo-api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/todo-api .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./todo-api"]
```

---

### 2. Orchestration: **Kubernetes** (Production)

**Resources:**
- Deployment: 3 replicas
- Service: LoadBalancer
- ConfigMap: Configuration
- Secret: Database credentials
- HorizontalPodAutoscaler: Auto-scaling

---

### 3. Database: **PostgreSQL StatefulSet**

Or managed PostgreSQL (AWS RDS, Google Cloud SQL, Azure Database).

---

### 4. Monitoring: **Prometheus + Grafana**

**Metrics Exposed:**
- HTTP request duration
- HTTP request count
- Database connection pool stats
- Error rates
- Response status codes

**Package:** `github.com/prometheus/client_golang`

---

### 5. Logging: **ELK Stack** or **Loki**

- Centralized log aggregation
- Log search and analysis
- Dashboard visualization

---

### 6. Tracing: **OpenTelemetry** (Future)

For distributed tracing in microservices architecture.

---

## Security Stack

### 1. Authentication: **JWT** (Phase 2)

**Package:** `github.com/golang-jwt/jwt/v5`

**Strategy:**
- Bearer token in Authorization header
- Token validation middleware
- Token expiration (1 hour)
- Refresh token mechanism (7 days)

---

### 2. Rate Limiting: **go-rate**

**Package:** `golang.org/x/time/rate`

**Strategy:**
- Per-IP rate limiting
- 100 requests per minute
- Sliding window algorithm
- Redis-backed (multi-instance)

---

### 3. CORS: **Gin CORS Middleware**

**Package:** `github.com/gin-contrib/cors`

**Configuration:**
- Whitelist origins in production
- Allow credentials
- Max age: 12 hours

---

### 4. Input Sanitization: **bluemonday**

**Package:** `github.com/microcosm-cc/bluemonday`

**Usage:**
- Sanitize HTML in user inputs
- Prevent XSS attacks
- Strip dangerous tags

---

### 5. SQL Injection Prevention

- Use prepared statements (pgx handles this)
- Never concatenate SQL strings
- Validate input types

---

## Performance Optimization

### 1. Database Connection Pooling

- Min: 2 connections
- Max: 10 connections
- Connection lifetime: 1 hour
- Idle timeout: 10 minutes

---

### 2. Caching: **Redis** (Phase 2)

**Package:** `github.com/redis/go-redis/v9`

**Cache Strategy:**
- Cache frequently accessed todos
- TTL: 5 minutes
- Cache invalidation on updates
- Cache statistics endpoint

---

### 3. Indexes

- Primary key (UUID)
- Status index
- Priority index
- Due date index
- Full-text search index (GIN)
- Composite indexes for common queries

---

### 4. Pagination

- Default page size: 20
- Max page size: 100
- Offset-based pagination (simple)
- Cursor-based pagination (Phase 2, better performance)

---

### 5. JSON Serialization: **sonic**

**Package:** `github.com/bytedance/sonic` (already in go.mod)

**Rationale:**
- Fastest JSON library for Go
- Drop-in replacement for encoding/json
- Used automatically by Gin when available

---

## Environment-Specific Configuration

### Development
```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  sslmode: disable

log:
  level: debug
  format: console
```

### Production
```yaml
server:
  port: 8080
  mode: release

database:
  host: postgres.internal
  port: 5432
  sslmode: require

log:
  level: info
  format: json
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

**Stages:**
1. **Lint**: golangci-lint
2. **Test**: Run all tests
3. **Build**: Compile binary
4. **Docker Build**: Build container image
5. **Push**: Push to container registry
6. **Deploy**: Deploy to Kubernetes (production)

**Workflow File:** `.github/workflows/ci.yml`

---

## Dependencies Summary

### Direct Dependencies
```go
require (
    github.com/gin-gonic/gin v1.11.0
    github.com/jackc/pgx/v5 v5.8.0
    github.com/rs/zerolog v1.34.0
    github.com/spf13/viper v1.18.2
    github.com/google/uuid v1.6.0
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/swaggo/gin-swagger v1.6.0
    github.com/swaggo/swag v1.16.3
    github.com/stretchr/testify v1.11.1
    github.com/gin-contrib/cors v1.7.2
    golang.org/x/time v0.9.0
)
```

### Development Dependencies
```go
require (
    github.com/cosmtrek/air v1.52.0
    github.com/golangci/golangci-lint v1.55.2
)
```

---

## Versioning Strategy

### Semantic Versioning

**Format:** `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking API changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

**API Versioning:**
- URL path: `/api/v1/`
- Response header: `API-Version: 1.0.0`

---

## Scalability Considerations

### Horizontal Scaling
- Stateless API servers
- Multiple replicas behind load balancer
- Shared PostgreSQL database
- Redis for distributed caching (Phase 2)

### Vertical Scaling
- Increase database resources
- Optimize queries and indexes
- Connection pool tuning

### Future Enhancements
- Read replicas for PostgreSQL
- CQRS pattern (separate read/write)
- Event-driven architecture with message queue
- GraphQL API for flexible queries

---

## Technology Decision Matrix

| Category | Technology | Status | Rationale |
|---|---|---|---|
| Language | Go 1.24 | ✅ Selected | Consistency with wildwest |
| HTTP Framework | Gin | ✅ Selected | Already used, high performance |
| Database | PostgreSQL | ✅ Selected | Already used, feature-rich |
| DB Driver | pgx/v5 | ✅ Selected | Fastest driver, already used |
| Configuration | Viper | ✅ Selected | Already used, flexible |
| Logging | Zerolog | ✅ Selected | Already used, performant |
| Validation | validator/v10 | ✅ Selected | Integrated with Gin |
| Testing | Testify | ✅ Selected | Already used |
| Migrations | golang-migrate | ✅ Selected | Standard tool |
| API Docs | Swaggo | ✅ Selected | Code-first documentation |
| UUID | google/uuid | ✅ Selected | Standard library |
| Auth (Phase 2) | JWT | 📋 Planned | Industry standard |
| Caching (Phase 2) | Redis | 📋 Planned | Fast, distributed |
| Rate Limiting | go-rate | ✅ Selected | Standard library |
| Monitoring | Prometheus | 📋 Planned | Industry standard |

---

## Final Recommendations

### Phase 1 (MVP)
- ✅ Go + Gin + PostgreSQL
- ✅ Clean architecture
- ✅ Full CRUD operations
- ✅ Health checks
- ✅ API documentation
- ✅ Unit + integration tests
- ✅ Docker deployment
- ❌ No authentication (public API)
- ❌ No caching

### Phase 2 (Production)
- ✅ JWT authentication
- ✅ Redis caching
- ✅ Rate limiting
- ✅ Monitoring with Prometheus
- ✅ Kubernetes deployment
- ✅ CI/CD pipeline
- ✅ Load testing

---

## Conclusion

The technology stack is carefully chosen to:
1. **Leverage existing wildwest patterns** (consistency)
2. **Use proven, production-ready technologies**
3. **Prioritize performance and scalability**
4. **Maintain code quality and testability**
5. **Enable rapid development and deployment**

All choices align with modern Go best practices and the existing wildwest architecture.
