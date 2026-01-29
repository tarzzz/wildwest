# User Management API - Development Guide

This guide will help you set up and run the User Management API locally.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
- [Development Workflows](#development-workflows)
- [Environment Variables](#environment-variables)
- [Database](#database)
- [Testing](#testing)
- [Docker Development](#docker-development)
- [Common Issues](#common-issues)
- [Code Quality](#code-quality)

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download here](https://golang.org/dl/)
- **Docker & Docker Compose**: [Download here](https://www.docker.com/products/docker-desktop)
- **PostgreSQL 15+** (for local development without Docker): [Download here](https://www.postgresql.org/download/)
- **Make**: Usually pre-installed on macOS/Linux, for Windows use [Chocolatey](https://chocolatey.org/)

### Optional Tools

- **Air**: Hot reload for Go (installed automatically via `make dev`)
- **golangci-lint**: Go linter (installed automatically via `make lint`)
- **migrate**: Database migration tool (installed automatically)

## Project Structure

```
user-management-api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── domain/                  # Domain models
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # HTTP middleware
│   ├── repository/              # Data access layer
│   └── service/                 # Business logic
├── pkg/
│   ├── database/                # Database utilities
│   ├── logger/                  # Logging utilities
│   └── validator/               # Validation utilities
├── migrations/                  # Database migrations
│   ├── 001_create_users_table.up.sql
│   ├── 001_create_users_table.down.sql
│   └── ...
├── .air.toml                    # Air configuration
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Docker services orchestration
├── Makefile                     # Development commands
├── go.mod                       # Go module definition
└── DEVELOPMENT.md              # This file
```

## Quick Start

### Option 1: Docker (Recommended for Quick Start)

1. **Clone the repository and navigate to the project**:
   ```bash
   cd user-management-api
   ```

2. **Start services with Docker Compose**:
   ```bash
   make docker-up
   ```

3. **Run migrations**:
   ```bash
   # Wait a few seconds for PostgreSQL to be ready, then:
   make migrate-up
   ```

4. **Test the API**:
   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:
   ```json
   {
     "status": "healthy",
     "version": "1.0.0",
     "uptime_seconds": 10,
     "timestamp": "2024-01-20T10:30:00Z"
   }
   ```

### Option 2: Local Development

1. **Install development tools**:
   ```bash
   make install-tools
   ```

2. **Start PostgreSQL** (if not using Docker):
   ```bash
   # macOS with Homebrew
   brew services start postgresql@15

   # Or use Docker for just PostgreSQL
   docker run -d \
     --name userapi-postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=userdb \
     -p 5432:5432 \
     postgres:15-alpine
   ```

3. **Create environment file**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run migrations**:
   ```bash
   make migrate-up
   ```

5. **Start the API with hot reload**:
   ```bash
   make dev
   ```

## Development Workflows

### Running the Application

**With hot reload (recommended for development)**:
```bash
make dev
```

**Without hot reload**:
```bash
make run
```

**Build and run binary**:
```bash
make build
./bin/api
```

### Database Migrations

**Apply all pending migrations**:
```bash
make migrate-up
```

**Rollback last migration**:
```bash
make migrate-down
```

**Create new migration**:
```bash
make migrate-create name=add_email_verification
```

This creates two files:
- `migrations/00X_add_email_verification.up.sql`
- `migrations/00X_add_email_verification.down.sql`

**Check current migration version**:
```bash
make migrate-version
```

**Force migration to specific version** (use with caution):
```bash
make migrate-force version=3
```

**Reset database** (drops and recreates):
```bash
make db-reset
make migrate-up
```

### Testing

**Run all tests**:
```bash
make test
```

**Run unit tests only**:
```bash
make test-unit
```

**Run integration tests**:
```bash
make test-integration
```

**Generate coverage report**:
```bash
make test-coverage
# Opens coverage.html in browser
```

### Code Quality

**Format code**:
```bash
make fmt
```

**Run linter**:
```bash
make lint
```

**Run go vet**:
```bash
make vet
```

**Run all quality checks**:
```bash
make fmt && make vet && make lint
```

## Environment Variables

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=userdb
DB_SSLMODE=disable
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Environment
APP_ENV=development
```

### Configuration Precedence

1. Environment variables (highest priority)
2. `.env` file
3. Default values in code (lowest priority)

## Database

### Connection String Format

```
postgres://[user]:[password]@[host]:[port]/[database]?sslmode=[mode]
```

### Local PostgreSQL Setup

**macOS with Homebrew**:
```bash
brew install postgresql@15
brew services start postgresql@15
createdb userdb
```

**Linux (Ubuntu/Debian)**:
```bash
sudo apt-get install postgresql-15
sudo systemctl start postgresql
sudo -u postgres createdb userdb
```

### Database Schema

The API uses the following tables:
- `users`: Core user accounts
- `refresh_tokens`: JWT refresh tokens
- `audit_logs`: Audit trail
- `password_reset_tokens`: Password reset tokens

See `DATABASE_SCHEMA.md` for complete schema documentation.

### Seed Data

To insert test data:
```bash
psql -h localhost -U postgres -d userdb -f migrations/seed.sql
```

## Testing

### Test Structure

```
internal/
├── handler/
│   ├── auth_handler.go
│   └── auth_handler_test.go
├── service/
│   ├── user_service.go
│   └── user_service_test.go
└── repository/
    ├── user_repository.go
    └── user_repository_test.go
```

### Writing Tests

**Unit Test Example**:
```go
func TestUserService_Create(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)

    // Act
    user, err := service.Create(ctx, createReq)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

**Integration Test Example**:
```go
func TestIntegration_UserAPI(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Test API endpoints
    resp := makeRequest(t, "POST", "/api/v1/auth/register", body)
    assert.Equal(t, 201, resp.StatusCode)
}
```

### Test Coverage Goals

- **Overall**: > 80%
- **Handler**: > 70%
- **Service**: > 90%
- **Repository**: > 80%

## Docker Development

### Docker Commands

**Build image**:
```bash
make docker-build
```

**Start all services**:
```bash
make docker-up
```

**Stop all services**:
```bash
make docker-down
```

**View logs**:
```bash
make docker-logs
```

**Restart services**:
```bash
make docker-restart
```

**Clean up (removes volumes and images)**:
```bash
make docker-clean
```

### Docker Compose Services

The `docker-compose.yml` defines:

1. **postgres**: PostgreSQL 15 database
   - Port: 5432
   - Data persistence via volume

2. **api**: API server
   - Port: 8080
   - Auto-restarts on failure
   - Health checks enabled

### Accessing Services

**API**:
```bash
curl http://localhost:8080/health
```

**PostgreSQL**:
```bash
psql -h localhost -U postgres -d userdb
# Password: postgres
```

**Docker logs**:
```bash
docker-compose logs -f api
docker-compose logs -f postgres
```

## Common Issues

### Port Already in Use

**Problem**: `Error starting userCan't bind on port 8080`

**Solution**:
```bash
# Find process using port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8081
```

### Database Connection Failed

**Problem**: `connection refused` or `authentication failed`

**Solution**:
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check database exists
psql -h localhost -U postgres -l

# Reset database
make db-reset
make migrate-up
```

### Migration Failed

**Problem**: `Dirty database version X`

**Solution**:
```bash
# Force migration to previous version
make migrate-force version=<X-1>

# Then re-run migrations
make migrate-up
```

### Hot Reload Not Working

**Problem**: Changes not triggering reload

**Solution**:
```bash
# Clean tmp directory
rm -rf tmp/

# Reinstall Air
go install github.com/cosmtrek/air@latest

# Restart dev server
make dev
```

### Docker Build Fails

**Problem**: `docker build` fails with permission errors

**Solution**:
```bash
# Clean Docker cache
docker builder prune -a

# Rebuild
make docker-build
```

## Code Quality

### Pre-commit Checklist

Before committing code:

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run tests
make test

# 4. Check coverage
make test-coverage
```

### Linter Configuration

The project uses `golangci-lint` with configuration in `.golangci.yml`.

Enabled linters:
- `errcheck`: Check for unchecked errors
- `gosimple`: Simplify code
- `govet`: Examine Go source code
- `ineffassign`: Detect ineffectual assignments
- `staticcheck`: Static analysis
- `typecheck`: Type checking
- `unused`: Check for unused code

### Code Style Guidelines

1. **Error Handling**: Always check and handle errors
   ```go
   if err != nil {
       return fmt.Errorf("failed to create user: %w", err)
   }
   ```

2. **Logging**: Use structured logging
   ```go
   log.Info("user created", "user_id", user.ID, "email", user.Email)
   ```

3. **Context**: Pass context through function calls
   ```go
   func GetUser(ctx context.Context, id string) (*User, error)
   ```

4. **Naming**: Use clear, descriptive names
   ```go
   // Good
   func CreateUser(ctx context.Context, req CreateUserRequest) error

   // Bad
   func CU(c context.Context, r CUR) error
   ```

## API Documentation

### Swagger UI

When running locally, access API documentation:
```
http://localhost:8080/swagger/index.html
```

### Generating Swagger Docs

```bash
# Install swaggo
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs/swagger
```

### Manual API Testing

**Register user**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "name": "Test User"
  }'
```

**Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }'
```

**Get current user**:
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

## Getting Help

### Documentation

- `API_SPEC.md`: Complete API specification
- `DATABASE_SCHEMA.md`: Database schema documentation
- `ARCHITECTURE.md`: System architecture
- `PROJECT_REQUIREMENTS.md`: Project requirements

### Makefile Help

```bash
make help
```

### Troubleshooting

1. Check logs: `docker-compose logs -f`
2. Verify environment: `env | grep DB_`
3. Test database: `psql -h localhost -U postgres -d userdb`
4. Check migrations: `make migrate-version`

## Contributing

1. Create a feature branch
2. Make your changes
3. Run quality checks: `make fmt && make lint && make test`
4. Submit a pull request

---

**Happy coding! 🚀**
