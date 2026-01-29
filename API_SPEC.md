# User Management API - API Specification

## Overview

This document provides the complete REST API specification for the User Management API, including endpoint definitions, request/response formats, authentication requirements, and error handling.

**Base URL**: `http://localhost:8080` (development)
**API Version**: v1
**API Prefix**: `/api/v1`
**Content Type**: `application/json`
**Authentication**: JWT Bearer Token

## API Conventions

### HTTP Methods
- `GET`: Retrieve resources (idempotent, cacheable)
- `POST`: Create new resources or execute actions
- `PUT`: Update entire resource (idempotent)
- `PATCH`: Partial update of resource (idempotent)
- `DELETE`: Delete resource (idempotent)

### Response Codes
- `200 OK`: Successful GET, PUT, PATCH, DELETE
- `201 Created`: Successful POST (resource created)
- `204 No Content`: Successful DELETE with no response body
- `400 Bad Request`: Invalid request format or validation error
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Valid token but insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., email already exists)
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Pagination
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 100,
    "total_pages": 5
  }
}
```

### Timestamps
- All timestamps in ISO 8601 format with timezone
- Example: `2024-01-20T10:30:00Z`

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "reason": "already_exists"
    },
    "timestamp": "2024-01-20T10:30:00Z",
    "request_id": "uuid"
  }
}
```

## Authentication Endpoints

### POST /api/v1/auth/register

Register a new user account.

**Authentication**: None (public endpoint)

**Rate Limit**: 10 requests/hour per IP

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "John Doe"
}
```

**Request Validation**:
- `email`: Required, valid email format, max 255 chars, unique
- `password`: Required, min 8 chars, max 72 chars, must contain uppercase, lowercase, number
- `name`: Required, min 2 chars, max 255 chars

**Success Response** (201 Created):
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "is_active": true,
    "email_verified": false,
    "created_at": "2024-01-20T10:30:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Error Responses**:
- `400`: Validation error (invalid email format, weak password)
- `409`: Email already exists
- `429`: Too many registration attempts

**Example Error**:
```json
{
  "error": {
    "code": "EMAIL_ALREADY_EXISTS",
    "message": "An account with this email already exists",
    "details": {
      "field": "email",
      "value": "user@example.com"
    },
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

---

### POST /api/v1/auth/login

Authenticate user and obtain access tokens.

**Authentication**: None (public endpoint)

**Rate Limit**: 10 requests/15min per IP

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "is_active": true,
    "email_verified": true,
    "last_login": "2024-01-20T10:30:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Error Responses**:
- `400`: Missing required fields
- `401`: Invalid credentials
- `403`: Account locked (too many failed attempts)
- `429`: Too many login attempts

**Account Lockout**:
- After 5 failed login attempts, account locked for 15 minutes
- Lock response includes `locked_until` timestamp

---

### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Authentication**: Refresh token in request body

**Rate Limit**: 100 requests/hour per user

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Error Responses**:
- `400`: Missing refresh token
- `401`: Invalid or expired refresh token
- `403`: Token revoked

---

### POST /api/v1/auth/logout

Revoke refresh token and logout user.

**Authentication**: Bearer token required

**Rate Limit**: 100 requests/hour per user

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

**Error Responses**:
- `400`: Missing refresh token
- `401`: Unauthorized

---

### POST /api/v1/auth/forgot-password

Request password reset token.

**Authentication**: None (public endpoint)

**Rate Limit**: 3 requests/hour per IP

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Success Response** (200 OK):
```json
{
  "message": "If an account exists with this email, a password reset link has been sent."
}
```

**Note**: Always returns success to prevent email enumeration attacks.

---

### POST /api/v1/auth/reset-password

Reset password using token from email.

**Authentication**: None (token in request body)

**Rate Limit**: 10 requests/hour per IP

**Request Body**:
```json
{
  "token": "reset-token-from-email",
  "new_password": "NewSecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password reset successfully"
}
```

**Error Responses**:
- `400`: Invalid token or weak password
- `410`: Token expired

---

### POST /api/v1/auth/verify-email

Verify email address using token from email.

**Authentication**: None (token in request body)

**Rate Limit**: 10 requests/hour per IP

**Request Body**:
```json
{
  "token": "verification-token-from-email"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Email verified successfully"
}
```

---

## User Management Endpoints

### GET /api/v1/users

List all users with pagination and filtering.

**Authentication**: Bearer token required

**Authorization**: Admin role only

**Rate Limit**: 100 requests/minute per user

**Query Parameters**:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `sort`: Sort field (default: created_at)
- `order`: Sort order (asc/desc, default: desc)
- `role`: Filter by role (admin/user/guest)
- `is_active`: Filter by active status (true/false)
- `search`: Search in name and email

**Example Request**:
```
GET /api/v1/users?page=1&page_size=20&role=user&sort=created_at&order=desc
```

**Success Response** (200 OK):
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "name": "John Doe",
      "bio": "Software developer",
      "avatar_url": "https://example.com/avatar.jpg",
      "role": "user",
      "is_active": true,
      "email_verified": true,
      "last_login": "2024-01-20T10:30:00Z",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-20T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 100,
    "total_pages": 5
  }
}
```

**Error Responses**:
- `401`: Unauthorized (missing/invalid token)
- `403`: Forbidden (not admin)

---

### GET /api/v1/users/me

Get current authenticated user's profile.

**Authentication**: Bearer token required

**Authorization**: Any authenticated user

**Rate Limit**: 100 requests/minute per user

**Success Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "bio": "Software developer",
  "avatar_url": "https://example.com/avatar.jpg",
  "role": "user",
  "is_active": true,
  "email_verified": true,
  "last_login": "2024-01-20T10:30:00Z",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-20T10:30:00Z"
}
```

---

### GET /api/v1/users/:id

Get user by ID.

**Authentication**: Bearer token required

**Authorization**:
- Admin: Can view any user
- User: Can only view own profile
- Guest: Cannot access

**Rate Limit**: 100 requests/minute per user

**Path Parameters**:
- `id`: User UUID

**Success Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "bio": "Software developer",
  "avatar_url": "https://example.com/avatar.jpg",
  "role": "user",
  "is_active": true,
  "email_verified": true,
  "last_login": "2024-01-20T10:30:00Z",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-20T10:30:00Z"
}
```

**Error Responses**:
- `401`: Unauthorized
- `403`: Forbidden (insufficient permissions)
- `404`: User not found

---

### POST /api/v1/users

Create a new user (admin only).

**Authentication**: Bearer token required

**Authorization**: Admin role only

**Rate Limit**: 100 requests/hour per user

**Request Body**:
```json
{
  "email": "newuser@example.com",
  "password": "SecurePassword123!",
  "name": "Jane Smith",
  "role": "user",
  "is_active": true
}
```

**Success Response** (201 Created):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "newuser@example.com",
  "name": "Jane Smith",
  "role": "user",
  "is_active": true,
  "email_verified": false,
  "created_at": "2024-01-20T10:30:00Z",
  "updated_at": "2024-01-20T10:30:00Z"
}
```

**Error Responses**:
- `400`: Validation error
- `403`: Forbidden (not admin)
- `409`: Email already exists

---

### PUT /api/v1/users/:id

Update user (full update).

**Authentication**: Bearer token required

**Authorization**:
- Admin: Can update any user
- User: Can only update own profile (excluding role, is_active)
- Guest: Cannot update

**Rate Limit**: 100 requests/hour per user

**Path Parameters**:
- `id`: User UUID

**Request Body**:
```json
{
  "name": "John Doe Updated",
  "bio": "Senior software developer",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

**Admin-only fields** (regular users cannot modify):
- `role`
- `is_active`
- `email_verified`

**Success Response** (200 OK):
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe Updated",
  "bio": "Senior software developer",
  "avatar_url": "https://example.com/new-avatar.jpg",
  "role": "user",
  "is_active": true,
  "email_verified": true,
  "updated_at": "2024-01-20T11:00:00Z"
}
```

**Error Responses**:
- `400`: Validation error
- `403`: Forbidden
- `404`: User not found

---

### PATCH /api/v1/users/:id

Partial update of user.

**Authentication**: Bearer token required

**Authorization**: Same as PUT endpoint

**Rate Limit**: 100 requests/hour per user

**Request Body** (only include fields to update):
```json
{
  "bio": "Updated bio only"
}
```

**Success Response** (200 OK):
Same as PUT endpoint

---

### DELETE /api/v1/users/:id

Soft delete user account.

**Authentication**: Bearer token required

**Authorization**:
- Admin: Can delete any user
- User: Can only delete own account
- Guest: Cannot delete

**Rate Limit**: 10 requests/hour per user

**Path Parameters**:
- `id`: User UUID

**Success Response** (200 OK):
```json
{
  "message": "User deleted successfully"
}
```

**Error Responses**:
- `403`: Forbidden
- `404`: User not found

**Note**: Soft delete sets `deleted_at` timestamp. User can be restored by admin.

---

### PATCH /api/v1/users/:id/change-password

Change user password.

**Authentication**: Bearer token required

**Authorization**: User can only change own password

**Rate Limit**: 10 requests/hour per user

**Request Body**:
```json
{
  "current_password": "CurrentPassword123!",
  "new_password": "NewSecurePassword123!"
}
```

**Success Response** (200 OK):
```json
{
  "message": "Password changed successfully"
}
```

**Error Responses**:
- `400`: Invalid current password or weak new password
- `401`: Unauthorized
- `403`: Forbidden

---

## Health & Monitoring Endpoints

### GET /health

Basic health check endpoint.

**Authentication**: None (public endpoint)

**Rate Limit**: No limit

**Success Response** (200 OK):
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "timestamp": "2024-01-20T10:30:00Z"
}
```

---

### GET /health/ready

Readiness check (includes database connectivity).

**Authentication**: None (public endpoint)

**Rate Limit**: No limit

**Success Response** (200 OK):
```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "migrations": "ok"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

**Error Response** (503 Service Unavailable):
```json
{
  "status": "not_ready",
  "checks": {
    "database": "error: connection refused",
    "migrations": "ok"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

---

### GET /metrics

Prometheus-compatible metrics endpoint.

**Authentication**: None (should be restricted via network policy)

**Rate Limit**: No limit

**Response Format**: Prometheus text format

**Example Response**:
```
# HELP api_requests_total Total number of API requests
# TYPE api_requests_total counter
api_requests_total{method="GET",endpoint="/api/v1/users",status="200"} 1234

# HELP api_request_duration_seconds API request duration
# TYPE api_request_duration_seconds histogram
api_request_duration_seconds_bucket{le="0.1"} 1000
api_request_duration_seconds_bucket{le="0.5"} 1200
```

---

## Authentication & Authorization

### JWT Token Structure

**Access Token Claims**:
```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "role": "user",
  "type": "access",
  "exp": 1705751400,
  "iat": 1705750500
}
```

**Refresh Token Claims**:
```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "type": "refresh",
  "exp": 1706355300,
  "iat": 1705750500
}
```

### Authorization Header

All protected endpoints require:
```
Authorization: Bearer <access_token>
```

### Role Permissions Matrix

| Endpoint | Admin | User | Guest | Anonymous |
|----------|-------|------|-------|-----------|
| POST /auth/register | ✓ | ✓ | ✓ | ✓ |
| POST /auth/login | ✓ | ✓ | ✓ | ✓ |
| POST /auth/logout | ✓ | ✓ | ✓ | ✗ |
| GET /users | ✓ | ✗ | ✗ | ✗ |
| GET /users/me | ✓ | ✓ | ✓ | ✗ |
| GET /users/:id | ✓ | Own only | ✗ | ✗ |
| POST /users | ✓ | ✗ | ✗ | ✗ |
| PUT /users/:id | ✓ | Own only* | ✗ | ✗ |
| DELETE /users/:id | ✓ | Own only | ✗ | ✗ |

*Users cannot modify `role`, `is_active`, `email_verified` fields

---

## Error Codes

### Authentication Errors
- `AUTH_TOKEN_MISSING`: Authorization header missing
- `AUTH_TOKEN_INVALID`: Invalid token format or signature
- `AUTH_TOKEN_EXPIRED`: Token has expired
- `AUTH_TOKEN_REVOKED`: Token has been revoked
- `AUTH_INVALID_CREDENTIALS`: Wrong email or password
- `AUTH_ACCOUNT_LOCKED`: Account locked due to failed attempts

### Validation Errors
- `VALIDATION_FAILED`: Request validation failed
- `EMAIL_INVALID`: Invalid email format
- `EMAIL_ALREADY_EXISTS`: Email already registered
- `PASSWORD_TOO_WEAK`: Password doesn't meet requirements
- `REQUIRED_FIELD_MISSING`: Required field not provided

### Resource Errors
- `USER_NOT_FOUND`: User does not exist
- `RESOURCE_NOT_FOUND`: Generic resource not found

### Permission Errors
- `FORBIDDEN`: Insufficient permissions
- `UNAUTHORIZED`: Authentication required

### Rate Limit Errors
- `RATE_LIMIT_EXCEEDED`: Too many requests

### Server Errors
- `INTERNAL_SERVER_ERROR`: Unexpected server error
- `DATABASE_ERROR`: Database operation failed

---

## Rate Limiting

### Rate Limit Headers

Included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705751400
```

### Rate Limit Exceeded Response

**Status**: 429 Too Many Requests
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 60 seconds.",
    "details": {
      "limit": 100,
      "window": "1 minute",
      "retry_after": 60
    },
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

**Headers**:
```
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705751460
```

---

## Request Examples

### cURL Examples

**Register**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "name": "John Doe"
  }'
```

**Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

**Get Current User**:
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

**Update Profile**:
```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Updated bio"
  }'
```

---

## Testing with Postman

### Environment Variables
```json
{
  "base_url": "http://localhost:8080",
  "access_token": "",
  "refresh_token": "",
  "user_id": ""
}
```

### Collection Structure
```
User Management API/
├── Auth/
│   ├── Register
│   ├── Login
│   ├── Refresh Token
│   ├── Logout
│   ├── Forgot Password
│   └── Reset Password
├── Users/
│   ├── List Users
│   ├── Get Current User
│   ├── Get User by ID
│   ├── Create User
│   ├── Update User
│   ├── Delete User
│   └── Change Password
└── Health/
    ├── Health Check
    ├── Readiness Check
    └── Metrics
```

---

## Swagger/OpenAPI Annotations for Go

### Handler Annotation Examples

Below are examples of how to annotate Go handlers for automatic OpenAPI/Swagger documentation generation using `swaggo/swag`.

#### Register Endpoint Example

```go
// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201  {object}  RegisterResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      429  {object}  ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    // Implementation
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email" example:"user@example.com"`
    Password string `json:"password" binding:"required,min=8" example:"SecurePassword123!"`
    Name     string `json:"name" binding:"required,min=2" example:"John Doe"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
    User   UserResponse  `json:"user"`
    Tokens TokenResponse `json:"tokens"`
}
```

#### Login Endpoint Example

```go
// Login godoc
// @Summary      Authenticate user
// @Description  Login with email and password to obtain JWT tokens
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      429  {object}  ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    // Implementation
}

// LoginRequest represents the login request body
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email" example:"user@example.com"`
    Password string `json:"password" binding:"required" example:"SecurePassword123!"`
}
```

#### Get Current User Example

```go
// GetCurrentUser godoc
// @Summary      Get current user profile
// @Description  Retrieve the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    // Implementation
}
```

#### List Users Example

```go
// ListUsers godoc
// @Summary      List all users
// @Description  Get paginated list of users with filtering (Admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false  "Page number"               default(1)
// @Param        page_size  query  int     false  "Items per page"            default(20)
// @Param        role       query  string  false  "Filter by role"            Enums(admin, user, guest)
// @Param        is_active  query  bool    false  "Filter by active status"
// @Param        search     query  string  false  "Search in name and email"
// @Param        sort       query  string  false  "Sort field"                default(created_at)
// @Param        order      query  string  false  "Sort order"                Enums(asc, desc) default(desc)
// @Success      200  {object}  ListUsersResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
    // Implementation
}
```

#### Update User Example

```go
// UpdateUser godoc
// @Summary      Update user
// @Description  Update user profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    string          true  "User ID"
// @Param        request body    UpdateUserRequest true  "User update data"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    // Implementation
}
```

### Common Response Models

```go
// UserResponse represents a user in API responses
type UserResponse struct {
    ID            string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    Email         string    `json:"email" example:"user@example.com"`
    Name          string    `json:"name" example:"John Doe"`
    Bio           *string   `json:"bio,omitempty" example:"Software developer"`
    AvatarURL     *string   `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
    Role          string    `json:"role" example:"user" enums:"admin,user,guest"`
    IsActive      bool      `json:"is_active" example:"true"`
    EmailVerified bool      `json:"email_verified" example:"true"`
    LastLogin     *string   `json:"last_login,omitempty" example:"2024-01-20T10:30:00Z"`
    CreatedAt     string    `json:"created_at" example:"2024-01-15T09:00:00Z"`
    UpdatedAt     string    `json:"updated_at" example:"2024-01-20T10:30:00Z"`
}

// TokenResponse represents JWT tokens in API responses
type TokenResponse struct {
    AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    TokenType    string `json:"token_type" example:"Bearer"`
    ExpiresIn    int    `json:"expires_in" example:"900"`
}

// ErrorResponse represents an error in API responses
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
    Code      string                 `json:"code" example:"VALIDATION_FAILED"`
    Message   string                 `json:"message" example:"Validation error occurred"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp string                 `json:"timestamp" example:"2024-01-20T10:30:00Z"`
    RequestID string                 `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
    Page       int `json:"page" example:"1"`
    PageSize   int `json:"page_size" example:"20"`
    TotalItems int `json:"total_items" example:"100"`
    TotalPages int `json:"total_pages" example:"5"`
}

// ListUsersResponse represents the list users response
type ListUsersResponse struct {
    Data       []UserResponse     `json:"data"`
    Pagination PaginationResponse `json:"pagination"`
}
```

### Main Application Annotation

```go
// @title           User Management API
// @version         1.0
// @description     A REST API for user management with authentication and RBAC
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description Authentication endpoints for login, registration, and token management

// @tag.name users
// @tag.description User management endpoints for CRUD operations
func main() {
    // Application initialization
}
```

### Generating Swagger Documentation

```bash
# Install swag CLI tool
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init -g cmd/server/main.go -o docs/swagger

# Swagger files will be generated in docs/swagger/:
# - docs.go
# - swagger.json
# - swagger.yaml
```

### Integrating Swagger UI with Gin

```go
import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "your-module/docs/swagger" // Import generated docs
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Swagger endpoint
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API routes
    // ...

    return r
}
```

---

## Authentication Flow Diagrams

### Registration Flow

```
┌─────────┐                ┌─────────┐                ┌──────────┐
│ Client  │                │   API   │                │ Database │
└────┬────┘                └────┬────┘                └─────┬────┘
     │                          │                           │
     │  POST /auth/register     │                           │
     ├─────────────────────────>│                           │
     │  {email, password, name} │                           │
     │                          │                           │
     │                          │  Validate input           │
     │                          │  (email format, password  │
     │                          │   strength, etc.)         │
     │                          │                           │
     │                          │  Check email uniqueness   │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  Email available          │
     │                          │<──────────────────────────┤
     │                          │                           │
     │                          │  Hash password (bcrypt)   │
     │                          │                           │
     │                          │  Create user record       │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  User created             │
     │                          │<──────────────────────────┤
     │                          │                           │
     │                          │  Generate JWT tokens      │
     │                          │  (access + refresh)       │
     │                          │                           │
     │                          │  Store refresh token      │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  Token stored             │
     │                          │<──────────────────────────┤
     │                          │                           │
     │  201 Created             │                           │
     │  {user, tokens}          │                           │
     │<─────────────────────────┤                           │
     │                          │                           │
```

### Login Flow

```
┌─────────┐                ┌─────────┐                ┌──────────┐
│ Client  │                │   API   │                │ Database │
└────┬────┘                └────┬────┘                └─────┬────┘
     │                          │                           │
     │  POST /auth/login        │                           │
     ├─────────────────────────>│                           │
     │  {email, password}       │                           │
     │                          │                           │
     │                          │  Validate input           │
     │                          │                           │
     │                          │  Get user by email        │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  User record              │
     │                          │<──────────────────────────┤
     │                          │                           │
     │                          │  Check account status     │
     │                          │  (active, not locked)     │
     │                          │                           │
     │                          │  Verify password          │
     │                          │  (bcrypt compare)         │
     │                          │                           │
     │                          ├─ Success ────────────────>│
     │                          │                           │
     │                          │  Update last_login        │
     │                          │  Reset failed_login_count │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  Generate JWT tokens      │
     │                          │                           │
     │                          │  Store refresh token      │
     │                          ├──────────────────────────>│
     │                          │                           │
     │  200 OK                  │                           │
     │  {user, tokens}          │                           │
     │<─────────────────────────┤                           │
     │                          │                           │
     │                          ├─ Failure ────────────────>│
     │                          │                           │
     │                          │  Increment failed_login_  │
     │                          │  count, check if locked   │
     │                          ├──────────────────────────>│
     │                          │                           │
     │  401 Unauthorized        │                           │
     │  or 403 Account Locked   │                           │
     │<─────────────────────────┤                           │
     │                          │                           │
```

### Authenticated Request Flow

```
┌─────────┐                ┌─────────┐                ┌──────────┐
│ Client  │                │   API   │                │ Database │
└────┬────┘                └────┬────┘                └─────┬────┘
     │                          │                           │
     │  GET /users/me           │                           │
     │  Authorization: Bearer   │                           │
     │  <access_token>          │                           │
     ├─────────────────────────>│                           │
     │                          │                           │
     │                          │  Auth Middleware:         │
     │                          │  - Extract token          │
     │                          │  - Verify signature       │
     │                          │  - Check expiration       │
     │                          │  - Extract claims         │
     │                          │    (user_id, role)        │
     │                          │                           │
     │                          │  Authorization Check:     │
     │                          │  - Verify permissions     │
     │                          │                           │
     │                          │  Get user by ID           │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  User record              │
     │                          │<──────────────────────────┤
     │                          │                           │
     │  200 OK                  │                           │
     │  {user data}             │                           │
     │<─────────────────────────┤                           │
     │                          │                           │
```

### Token Refresh Flow

```
┌─────────┐                ┌─────────┐                ┌──────────┐
│ Client  │                │   API   │                │ Database │
└────┬────┘                └────┬────┘                └─────┬────┘
     │                          │                           │
     │  POST /auth/refresh      │                           │
     ├─────────────────────────>│                           │
     │  {refresh_token}         │                           │
     │                          │                           │
     │                          │  Validate refresh token   │
     │                          │  - Verify signature       │
     │                          │  - Check expiration       │
     │                          │  - Extract user_id        │
     │                          │                           │
     │                          │  Verify token in DB       │
     │                          │  (check not revoked)      │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  Token valid              │
     │                          │<──────────────────────────┤
     │                          │                           │
     │                          │  Get user by ID           │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  User record              │
     │                          │<──────────────────────────┤
     │                          │                           │
     │                          │  Generate new tokens      │
     │                          │                           │
     │                          │  Store new refresh token  │
     │                          ├──────────────────────────>│
     │                          │                           │
     │                          │  Revoke old refresh token │
     │                          ├──────────────────────────>│
     │                          │                           │
     │  200 OK                  │                           │
     │  {new tokens}            │                           │
     │<─────────────────────────┤                           │
     │                          │                           │
```

### Password Reset Flow

```
┌─────────┐                ┌─────────┐                ┌──────────┐         ┌───────┐
│ Client  │                │   API   │                │ Database │         │ Email │
└────┬────┘                └────┬────┘                └─────┬────┘         └───┬───┘
     │                          │                           │                   │
     │  POST /auth/             │                           │                   │
     │  forgot-password         │                           │                   │
     ├─────────────────────────>│                           │                   │
     │  {email}                 │                           │                   │
     │                          │                           │                   │
     │                          │  Get user by email        │                   │
     │                          ├──────────────────────────>│                   │
     │                          │                           │                   │
     │                          │  User record (or null)    │                   │
     │                          │<──────────────────────────┤                   │
     │                          │                           │                   │
     │                          │  If user exists:          │                   │
     │                          │  - Generate reset token   │                   │
     │                          │  - Hash token (SHA-256)   │                   │
     │                          │                           │                   │
     │                          │  Store reset token        │                   │
     │                          │  (expires in 1 hour)      │                   │
     │                          ├──────────────────────────>│                   │
     │                          │                           │                   │
     │                          │  Send email with token    │                   │
     │                          ├───────────────────────────────────────────────>│
     │                          │                           │                   │
     │  200 OK                  │                           │                   │
     │  {generic message}       │                           │                   │
     │<─────────────────────────┤                           │                   │
     │                          │                           │                   │
     │  (User receives email    │                           │                   │
     │   with reset link)       │                           │                   │
     │<─────────────────────────────────────────────────────────────────────────┤
     │                          │                           │                   │
     │  POST /auth/             │                           │                   │
     │  reset-password          │                           │                   │
     ├─────────────────────────>│                           │                   │
     │  {token, new_password}   │                           │                   │
     │                          │                           │                   │
     │                          │  Validate token           │                   │
     │                          │  - Hash token             │                   │
     │                          │  - Check expiration       │                   │
     │                          │  - Check not used         │                   │
     │                          ├──────────────────────────>│                   │
     │                          │                           │                   │
     │                          │  Token valid              │                   │
     │                          │<──────────────────────────┤                   │
     │                          │                           │                   │
     │                          │  Hash new password        │                   │
     │                          │                           │                   │
     │                          │  Update user password     │                   │
     │                          │  Mark token as used       │                   │
     │                          │  Revoke all refresh tokens│                   │
     │                          ├──────────────────────────>│                   │
     │                          │                           │                   │
     │  200 OK                  │                           │                   │
     │  {success message}       │                           │                   │
     │<─────────────────────────┤                           │                   │
     │                          │                           │                   │
```

### Authorization Decision Flow

```
┌────────────────┐
│ Request arrives│
└───────┬────────┘
        │
        ▼
┌───────────────────┐
│ Extract JWT token │
│ from Authorization│
│      header       │
└───────┬───────────┘
        │
        ▼
┌───────────────────┐           ┌──────────────┐
│ Token present?    │───No─────>│ 401 Unauth   │
└───────┬───────────┘           └──────────────┘
        │ Yes
        ▼
┌───────────────────┐           ┌──────────────┐
│ Valid signature   │───No─────>│ 401 Invalid  │
│ & not expired?    │           │    Token     │
└───────┬───────────┘           └──────────────┘
        │ Yes
        ▼
┌───────────────────┐
│ Extract claims:   │
│ - user_id         │
│ - role            │
│ - email           │
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ Check endpoint    │
│ permissions       │
└───────┬───────────┘
        │
        ▼
┌───────────────────┐           ┌──────────────┐
│ Required role?    │           │              │
│ Admin only?       │───No─────>│ Continue to  │
│ Owner only?       │           │   handler    │
└───────┬───────────┘           └──────────────┘
        │ Yes
        ▼
┌───────────────────┐           ┌──────────────┐
│ User has required │───No─────>│ 403 Forbidden│
│   permissions?    │           └──────────────┘
└───────┬───────────┘
        │ Yes
        ▼
┌───────────────────┐
│ Continue to       │
│    handler        │
└───────────────────┘
```

---

## Versioning Strategy

### URL Versioning
- Current: `/api/v1/...`
- Future: `/api/v2/...`

### Breaking Changes
Changes requiring new version:
- Removing endpoints
- Changing response structure
- Changing authentication method
- Changing required parameters

### Non-Breaking Changes
Can be added to existing version:
- Adding optional parameters
- Adding new endpoints
- Adding response fields
- Adding error codes

---

## OpenAPI/Swagger Integration

### Swagger UI
- **URL**: `http://localhost:8080/swagger/index.html`
- Interactive API documentation
- Try endpoints directly from browser

### OpenAPI Spec
- **URL**: `http://localhost:8080/swagger/doc.json`
- OpenAPI 3.0 specification
- Can be imported into Postman, Insomnia, etc.

---

**Document Version**: 1.0
**Last Updated**: 2026-01-27
**Author**: Solutions Architect (berners-lee)
**Status**: Phase 1 Deliverable
