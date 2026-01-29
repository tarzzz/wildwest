# User Management API - System Architecture

## Executive Summary

This document describes the system architecture for the User Management REST API. The API follows clean architecture principles with clear separation between layers, dependency injection for testability, and stateless design for horizontal scalability.

## Architecture Overview

### Architecture Style
**Clean Architecture (Hexagonal/Ports & Adapters)**

The system is organized in concentric layers where:
- Inner layers contain business logic and are framework-agnostic
- Outer layers handle infrastructure concerns (HTTP, database, etc.)
- Dependencies point inward (outer layers depend on inner layers, never the reverse)

```
┌─────────────────────────────────────────────────────────────┐
│                    External Clients                          │
│              (Web Browser, Mobile App, CLI)                  │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTPS
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                     Handler Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Auth       │  │    User      │  │   Health     │      │
│  │   Handlers   │  │   Handlers   │  │   Handlers   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         HTTP Request/Response Mapping                        │
└─────────────────┬───────────────────────────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌───────────┐ ┌───────────┐ ┌───────────┐
│  Auth     │ │   Rate    │ │  Logger   │  Middleware Layer
│  JWT      │ │  Limiter  │ │  Request  │
└───────────┘ └───────────┘ └───────────┘
    │             │             │
    └─────────────┼─────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Auth       │  │    User      │  │   Token      │      │
│  │   Service    │  │   Service    │  │   Service    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         Business Logic & Use Cases                           │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │            User Repository Interface                  │   │
│  │  • Create(user)    • GetByEmail(email)              │   │
│  │  • GetByID(id)     • Update(user)                   │   │
│  │  • Delete(id)      • List(filters)                  │   │
│  └──────────────────────────────────────────────────────┘   │
│                          ▲                                   │
│                          │ implements                        │
│                          │                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │       PostgreSQL User Repository                     │   │
│  │         (Concrete Implementation)                    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Database Layer                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              PostgreSQL Database                     │   │
│  │         (pgx with connection pooling)                │   │
│  │                                                      │   │
│  │  Tables: users, sessions, audit_logs                │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Handler Layer (Outer)
**Location**: `internal/handler/`
**Purpose**: HTTP interface and request/response transformation

**Responsibilities**:
- Parse and validate HTTP requests
- Route requests to appropriate service methods
- Transform service results into HTTP responses
- Handle HTTP-specific concerns (status codes, headers, cookies)
- Input validation using struct tags

**Dependencies**: Service layer interfaces

**Example Flow**:
```
POST /api/v1/auth/register
  ↓
AuthHandler.Register()
  ↓ parse JSON body
  ↓ validate input
  ↓ call AuthService.Register()
  ↓ format response
  ↓ return 201 Created
```

### 2. Middleware Layer (Cross-cutting)
**Location**: `internal/middleware/`
**Purpose**: Cross-cutting concerns applied to all/specific routes

**Components**:
- **Auth Middleware**: JWT token validation, user context injection
- **Rate Limiter**: Request throttling per IP/user
- **Logger**: Request/response logging with correlation IDs
- **Error Handler**: Standardized error response formatting
- **CORS**: Cross-origin resource sharing configuration
- **Recovery**: Panic recovery and graceful error handling

**Middleware Chain Example**:
```
Request → Logger → CORS → Rate Limiter → Auth → Handler → Response
```

### 3. Service Layer (Core Business Logic)
**Location**: `internal/service/`
**Purpose**: Business logic orchestration and use case implementation

**Services**:

**AuthService**:
- User registration with validation
- Login with credential verification
- Token generation and refresh
- Password reset flow
- Session management

**UserService**:
- User CRUD operations
- Profile management
- Role assignment
- User listing with pagination
- Soft delete implementation

**TokenService**:
- JWT token creation
- Token validation and parsing
- Refresh token management
- Token revocation

**Responsibilities**:
- Implement business rules and validation
- Coordinate between multiple repositories
- Transaction management
- Error handling and business exceptions
- Audit logging

**Dependencies**: Repository interfaces, domain entities

### 4. Repository Layer (Data Access)
**Location**: `internal/repository/`
**Purpose**: Data persistence abstraction

**Pattern**: Repository pattern with interfaces

**UserRepository Interface**:
```go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filters ListFilters) ([]*domain.User, int, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

**Implementations**:
- `PostgresUserRepository`: Production PostgreSQL implementation
- `MemoryUserRepository`: In-memory implementation for testing

**Responsibilities**:
- Execute database queries (CRUD operations)
- Map between database rows and domain entities
- Handle database errors and connection issues
- Implement pagination and filtering
- Transaction support

### 5. Domain Layer (Core)
**Location**: `internal/domain/`
**Purpose**: Core business entities and rules

**Entities**:
- `User`: Core user entity with validation
- `Role`: User role enumeration
- `Session`: User session data
- Domain-specific errors

**Characteristics**:
- Pure Go code (no external dependencies)
- Business validation methods
- Immutable where appropriate
- Rich domain models

**Example**:
```go
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
    // Business validation rules
}

func (u *User) IsDeleted() bool {
    return u.DeletedAt != nil
}
```

## Component Interactions

### Registration Flow
```
1. Client sends POST /api/v1/auth/register
2. Handler receives request → validates input
3. Handler calls AuthService.Register()
4. AuthService validates business rules:
   - Email format and uniqueness
   - Password strength requirements
   - Required fields present
5. AuthService hashes password (bcrypt)
6. AuthService calls UserRepository.Create()
7. Repository inserts user into PostgreSQL
8. AuthService generates JWT token
9. Handler returns 201 Created with token
```

### Authentication Flow
```
1. Client sends POST /api/v1/auth/login with credentials
2. Handler validates input
3. Handler calls AuthService.Login()
4. AuthService calls UserRepository.GetByEmail()
5. AuthService verifies password with bcrypt.Compare()
6. AuthService calls TokenService.GenerateToken()
7. TokenService creates JWT with user claims
8. Handler returns 200 OK with token
```

### Protected Endpoint Flow
```
1. Client sends GET /api/v1/users/me with Authorization header
2. Auth middleware extracts JWT token
3. Middleware validates token signature and expiry
4. Middleware extracts user ID from claims
5. Middleware injects user context
6. Handler calls UserService.GetByID()
7. Service calls UserRepository.GetByID()
8. Repository queries database
9. Handler returns 200 OK with user data
```

## Security Architecture

### Authentication
- **JWT Tokens**: Stateless authentication with signed tokens
- **Token Structure**:
  ```json
  {
    "sub": "user-uuid",
    "email": "user@example.com",
    "role": "user",
    "exp": 1234567890,
    "iat": 1234567890
  }
  ```
- **Token Lifetime**: 15 minutes for access tokens, 7 days for refresh tokens
- **Signing Algorithm**: HS256 (HMAC with SHA-256)

### Authorization
- **Role-Based Access Control (RBAC)**
- **Roles**: Admin, User, Guest
- **Middleware Enforcement**: `RequireRole()` middleware
- **Permissions**:
  - Admin: Full access to all endpoints
  - User: Read/update own profile, read public data
  - Guest: Read public data only

### Password Security
- **Hashing**: bcrypt with cost factor 12
- **Salt**: Automatic per password (built into bcrypt)
- **Storage**: Only hashed passwords stored, never plaintext
- **Validation**: Minimum 8 characters, complexity requirements

### Input Validation
- **Request Validation**: Struct tags with go-playground/validator
- **SQL Injection Prevention**: Parameterized queries with pgx
- **XSS Prevention**: JSON encoding escapes special characters
- **CSRF Protection**: Stateless API (no cookies for auth)

### Rate Limiting
- **Strategy**: Token bucket algorithm
- **Limits**: 100 requests per minute per IP
- **Auth Endpoints**: Stricter limits (10 requests per minute)
- **Implementation**: In-memory store with sliding window

## Data Flow Architecture

### Request Processing Pipeline
```
HTTP Request
  ↓
[CORS Middleware]
  ↓
[Logger Middleware] ← Log request
  ↓
[Recovery Middleware] ← Catch panics
  ↓
[Rate Limit Middleware] ← Check request quota
  ↓
[Auth Middleware] ← Validate JWT (if protected route)
  ↓
[Handler] ← Route to appropriate handler
  ↓
[Service] ← Execute business logic
  ↓
[Repository] ← Data access
  ↓
[Database] ← Query execution
  ↓
[Repository] ← Map results
  ↓
[Service] ← Transform data
  ↓
[Handler] ← Format response
  ↓
HTTP Response
```

### Error Handling Flow
```
Error occurs at any layer
  ↓
Layer wraps error with context
  ↓
Error bubbles up through layers
  ↓
Handler catches error
  ↓
Error middleware formats error
  ↓
Standardized JSON error response:
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID xyz not found",
    "timestamp": "2024-01-20T10:00:00Z"
  }
}
```

## Scalability Architecture

### Horizontal Scaling
- **Stateless Design**: No session state stored in application
- **JWT Tokens**: Self-contained, no server-side session storage
- **Load Balancer**: Multiple API instances behind load balancer
- **Shared Database**: All instances connect to same PostgreSQL cluster

### Database Optimization
- **Connection Pooling**: pgxpool with 25 max connections
- **Indexes**: On frequently queried columns (email, id)
- **Prepared Statements**: Query plan caching
- **Read Replicas**: Future support for read scaling

### Caching Strategy (Future)
- **User Data**: Redis cache for frequently accessed profiles
- **Token Blacklist**: Redis for revoked tokens
- **Rate Limits**: Redis for distributed rate limiting
- **TTL**: Short-lived cache (5-15 minutes)

### Performance Targets
- **Response Time**: < 200ms for 95th percentile
- **Throughput**: 1000+ requests per second per instance
- **Concurrent Users**: 1000+ simultaneous connections
- **Database Queries**: < 50ms for simple queries

## Deployment Architecture

### Container Structure
```
┌────────────────────────────────────────┐
│      Docker Container (API)            │
│  ┌──────────────────────────────────┐  │
│  │  Go Binary (user-api)            │  │
│  │  Port: 8080                      │  │
│  │  Environment Variables           │  │
│  └──────────────────────────────────┘  │
│                                        │
│  Base: alpine:latest                   │
│  Size: ~20-30 MB                       │
└────────────────────────────────────────┘
         │
         │ Network: bridge
         ▼
┌────────────────────────────────────────┐
│   Docker Container (PostgreSQL)        │
│  ┌──────────────────────────────────┐  │
│  │  PostgreSQL 15                   │  │
│  │  Port: 5432                      │  │
│  │  Volume: /var/lib/postgresql/data│  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

### Environment Configuration
- **Development**: Local Docker Compose
- **Staging**: Kubernetes cluster
- **Production**: Kubernetes cluster with auto-scaling

### Health Monitoring
- **Health Check Endpoint**: GET /health
  - Returns: Database connectivity, version, uptime
- **Metrics Endpoint**: GET /metrics
  - Prometheus-compatible metrics
  - Request rates, latency, error rates

## Technology Stack

| Layer | Technology | Justification |
|-------|-----------|---------------|
| Web Framework | Gin | High performance, rich middleware |
| Database | PostgreSQL 15 | ACID compliance, JSON support |
| Database Driver | pgx | Native PostgreSQL driver |
| Authentication | JWT (golang-jwt/jwt v5) | Stateless, standard |
| Password Hashing | bcrypt | Industry standard |
| Configuration | viper | Flexible, 12-factor |
| Logging | zerolog | High performance, structured |
| Validation | validator | Declarative, comprehensive |
| Testing | testify, dockertest | Assertions, integration tests |
| Documentation | swaggo/swag | OpenAPI generation |
| Container | Docker | Multi-stage builds |

## Design Patterns

### Dependency Injection
- Constructor injection for services and repositories
- Interface-based dependencies for testability
- Composition over inheritance

### Repository Pattern
- Abstract data access behind interfaces
- Enable easy mocking for tests
- Support multiple implementations (Postgres, Memory)

### Factory Pattern
- Configuration loading and validation
- Database connection creation
- Logger initialization

### Middleware Chain
- Request processing pipeline
- Composable cross-cutting concerns
- Clean separation of concerns

## Testing Strategy

### Unit Tests
- Test each layer independently
- Mock dependencies using interfaces
- Coverage target: 80%+
- Location: Next to implementation files

### Integration Tests
- Test full request/response cycle
- Use real PostgreSQL (dockertest)
- Test authentication flows
- Test error scenarios
- Location: `tests/integration/`

### Test Pyramid
```
      /\
     /  \    E2E Tests (few)
    /────\
   /      \  Integration Tests (some)
  /────────\
 /          \ Unit Tests (many)
/────────────\
```

## Error Handling Strategy

### Error Types
- **Domain Errors**: Business rule violations (user already exists)
- **Repository Errors**: Database failures (connection lost)
- **Validation Errors**: Input validation failures (invalid email)
- **Authentication Errors**: Auth failures (invalid token)

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": {
      "field": "email",
      "reason": "already_exists"
    },
    "timestamp": "2024-01-20T10:00:00Z",
    "request_id": "uuid"
  }
}
```

### HTTP Status Codes
- 200: Success
- 201: Created
- 400: Bad Request (validation)
- 401: Unauthorized (auth required)
- 403: Forbidden (insufficient permissions)
- 404: Not Found
- 429: Too Many Requests (rate limit)
- 500: Internal Server Error

## Configuration Management

### Environment Variables
```
APP_ENV=production
APP_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_NAME=userapi
DB_USER=apiuser
DB_PASSWORD=<secret>
JWT_SECRET=<secret>
JWT_EXPIRY=15m
RATE_LIMIT_REQUESTS=100
LOG_LEVEL=info
```

### Configuration Layers
1. Default values (code)
2. Configuration file (.env)
3. Environment variables (override)
4. Command-line flags (highest priority)

## Future Enhancements

### Phase 2 Considerations
- Redis caching layer
- Email service integration
- OAuth2 social login
- Two-factor authentication
- User activity logging
- Admin dashboard API
- Webhook support
- API versioning strategy

### Scalability Improvements
- Read replicas for database
- CDN for static assets
- Message queue for async tasks
- Distributed tracing (OpenTelemetry)
- Service mesh (Istio)

## Appendix: Key Design Decisions

### Why Clean Architecture?
- **Testability**: Easy to mock and test layers independently
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations (database, framework)
- **Scalability**: Layers can be optimized independently

### Why Gin Framework?
- Proven performance in production
- Rich middleware ecosystem
- Large community support
- Built-in validation

### Why PostgreSQL?
- ACID compliance for data integrity
- Rich type system
- Excellent performance
- Mature and battle-tested
- JSON support for flexibility

### Why JWT for Auth?
- Stateless (no server-side session storage)
- Scales horizontally easily
- Self-contained (includes user info)
- Industry standard

---

**Document Version**: 1.0
**Last Updated**: 2026-01-27
**Author**: Solutions Architect (berners-lee)
**Status**: Phase 1 Deliverable
