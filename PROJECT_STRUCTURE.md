# Go Project Structure

## Directory Layout

```
user-management-api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/                       # Private application code
│   ├── config/
│   │   └── config.go              # Configuration loading (viper)
│   ├── domain/
│   │   ├── user.go                # User entity and business rules
│   │   ├── role.go                # Role entity
│   │   └── errors.go              # Domain-specific errors
│   ├── repository/
│   │   ├── user_repository.go     # User repository interface
│   │   └── postgres/
│   │       ├── user_postgres.go   # PostgreSQL user implementation
│   │       └── migrations/        # SQL migration files
│   ├── service/
│   │   ├── auth_service.go        # Authentication business logic
│   │   ├── user_service.go        # User management business logic
│   │   └── token_service.go       # JWT token operations
│   ├── handler/
│   │   ├── auth_handler.go        # Auth endpoints (register, login, etc.)
│   │   ├── user_handler.go        # User CRUD endpoints
│   │   ├── health_handler.go      # Health check endpoints
│   │   └── response.go            # Standard response formats
│   ├── middleware/
│   │   ├── auth.go                # JWT authentication middleware
│   │   ├── rate_limit.go          # Rate limiting middleware
│   │   ├── logger.go              # Request logging middleware
│   │   └── error.go               # Error handling middleware
│   └── validator/
│       └── custom.go              # Custom validation rules
├── pkg/                            # Public, reusable packages
│   ├── logger/
│   │   └── logger.go              # Logging setup (zerolog)
│   ├── database/
│   │   └── postgres.go            # PostgreSQL connection setup
│   └── utils/
│       ├── jwt.go                 # JWT utility functions
│       └── password.go            # Password hashing utilities
├── migrations/                     # Database migrations (golang-migrate)
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_indexes.up.sql
│   └── 000002_add_indexes.down.sql
├── api/                            # API documentation
│   └── openapi.yaml               # OpenAPI/Swagger spec
├── tests/
│   ├── integration/               # Integration tests
│   │   ├── auth_test.go
│   │   └── user_test.go
│   └── fixtures/                  # Test data
│       └── users.json
├── scripts/
│   ├── migrate.sh                 # Database migration script
│   └── seed.sh                    # Database seeding script
├── docker/
│   ├── Dockerfile                 # Multi-stage Docker build
│   └── docker-compose.yml         # Local development stack
├── .air.toml                      # Hot reload configuration
├── .env.example                   # Example environment variables
├── .gitignore
├── go.mod                         # Go module dependencies
├── go.sum
├── Makefile                       # Build and development tasks
└── README.md                      # Project documentation
```

---

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
**Purpose**: Core business entities and rules

**Responsibilities:**
- Define entity structs (User, Role, etc.)
- Business validation rules
- Domain-specific errors
- Pure Go code, no external dependencies

**Example:**
```go
// internal/domain/user.go
type User struct {
    ID            uuid.UUID
    Email         string
    PasswordHash  string
    Name          string
    Bio           *string
    AvatarURL     *string
    Role          Role
    IsActive      bool
    EmailVerified bool
    LastLogin     *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

func (u *User) Validate() error {
    // Business validation logic
}
```

### 2. Repository Layer (`internal/repository/`)
**Purpose**: Data access abstraction

**Responsibilities:**
- Define repository interfaces
- Implement database operations
- Handle database connections
- Map between database and domain models
- Transaction management

**Pattern:**
```go
// internal/repository/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filters ListFilters) ([]*domain.User, int, error)
}
```

### 3. Service Layer (`internal/service/`)
**Purpose**: Business logic orchestration

**Responsibilities:**
- Implement use cases
- Coordinate between repositories
- Handle business logic
- Transaction coordination
- Call external services if needed

**Example:**
```go
// internal/service/user_service.go
type UserService struct {
    userRepo repository.UserRepository
    logger   logger.Logger
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*domain.User, error) {
    // Validate request
    // Hash password
    // Create user via repository
    // Return result
}
```

### 4. Handler Layer (`internal/handler/`)
**Purpose**: HTTP request/response handling

**Responsibilities:**
- Parse HTTP requests
- Validate input
- Call service layer
- Format HTTP responses
- Handle HTTP-specific concerns (status codes, headers)

**Example:**
```go
// internal/handler/user_handler.go
type UserHandler struct {
    userService *service.UserService
}

func (h *UserHandler) Create(c *gin.Context) {
    // Parse request
    // Call service
    // Return response
}
```

### 5. Middleware Layer (`internal/middleware/`)
**Purpose**: Cross-cutting concerns

**Responsibilities:**
- Authentication/authorization
- Logging
- Rate limiting
- Error handling
- CORS
- Request ID tracking

---

## Dependency Flow

```
HTTP Request
     ↓
Handler (HTTP layer)
     ↓
Service (Business logic)
     ↓
Repository (Data access)
     ↓
Database
```

**Key Principle**: Inner layers do not depend on outer layers
- Domain: No dependencies
- Repository: Depends on Domain
- Service: Depends on Domain and Repository interfaces
- Handler: Depends on Service

---

## File Naming Conventions

1. **Entities**: Singular nouns (`user.go`, `role.go`)
2. **Interfaces**: Descriptive names with "_test" for tests (`user_repository.go`, `user_repository_test.go`)
3. **Implementations**: Interface name + implementation (`user_postgres.go`, `user_memory.go`)
4. **Tests**: Same name as file + `_test.go` suffix
5. **Handlers**: Resource + "_handler.go" (`user_handler.go`)
6. **Services**: Purpose + "_service.go" (`auth_service.go`)

---

## Package Organization Principles

### Internal vs Pkg

**internal/**: Code that is private to this application
- Cannot be imported by other projects
- Application-specific logic
- Domain models, handlers, services

**pkg/**: Code that could be reused
- Can be imported by other projects
- Generic utilities
- Database connection helpers
- Logger setup

### Why This Structure?

1. **Clean Architecture**: Clear separation of concerns
2. **Testability**: Easy to mock interfaces between layers
3. **Dependency Injection**: Services depend on interfaces, not implementations
4. **Scalability**: Easy to add new features without breaking existing code
5. **Go Conventions**: Follows standard Go project layout
6. **Team Collaboration**: Clear boundaries for different developers

---

## Configuration Management

### Environment Variables (`.env`)
```bash
# Server
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=userapi
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Security
BCRYPT_COST=12
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

### Loading Configuration
```go
// internal/config/config.go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Security SecurityConfig
    Log      LogConfig
}

func Load() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    // Load and return config
}
```

---

## Makefile Commands

```makefile
.PHONY: help run build test lint migrate-up migrate-down docker-build

help:           ## Show this help
run:            ## Run the application
build:          ## Build the binary
test:           ## Run tests
test-coverage:  ## Run tests with coverage
lint:           ## Run linters
migrate-up:     ## Run database migrations
migrate-down:   ## Rollback migrations
docker-build:   ## Build Docker image
docker-up:      ## Start Docker compose
seed:           ## Seed database with test data
```

---

## Development Workflow

1. **Start Development Environment**:
   ```bash
   docker-compose up -d postgres  # Start PostgreSQL
   make migrate-up                # Run migrations
   make run                       # Start API server (with hot reload)
   ```

2. **Add New Feature**:
   - Define domain entity in `internal/domain/`
   - Create repository interface in `internal/repository/`
   - Implement repository in `internal/repository/postgres/`
   - Create service in `internal/service/`
   - Create handler in `internal/handler/`
   - Add routes in `cmd/api/main.go`
   - Write tests

3. **Run Tests**:
   ```bash
   make test              # Unit tests
   make test-coverage     # With coverage report
   ```

4. **Build and Deploy**:
   ```bash
   make build             # Build binary
   make docker-build      # Build Docker image
   ```

---

## Testing Strategy

### Unit Tests
- Test each layer independently
- Mock dependencies using interfaces
- Located next to the code they test (`user_service_test.go`)

### Integration Tests
- Test full request/response cycle
- Use real PostgreSQL (via dockertest)
- Located in `tests/integration/`

### Test Structure
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)

    // Act
    result, err := service.CreateUser(context.Background(), validRequest)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

---

## Benefits of This Structure

1. **Maintainability**: Clear organization makes code easy to find and modify
2. **Testability**: Interface-based design enables comprehensive testing
3. **Scalability**: Structure supports growth from small to large codebases
4. **Team Collaboration**: Multiple developers can work on different layers simultaneously
5. **Go Idiomatic**: Follows Go community standards and best practices
6. **Migration Ready**: Easy to swap implementations (e.g., change database)
