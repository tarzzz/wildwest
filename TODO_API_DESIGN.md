# Todo List API Design

## Overview
RESTful API for managing todo items with CRUD operations. Follows the clean architecture pattern established in wildwest's user-management-api.

---

## Base URL
```
/api/v1
```

---

## Endpoints

### 1. Create Todo
**POST** `/todos`

Create a new todo item.

**Request Body:**
```json
{
  "title": "Complete project documentation",
  "description": "Write comprehensive docs for the API",
  "priority": "high",
  "due_date": "2026-02-15T17:00:00Z",
  "tags": ["documentation", "urgent"]
}
```

**Request Schema:**
| Field | Type | Required | Description |
|---|---|---|---|
| title | string | Yes | Todo title (1-200 chars) |
| description | string | No | Detailed description (max 2000 chars) |
| priority | string | No | Priority level: "low", "medium", "high" (default: "medium") |
| due_date | string (ISO8601) | No | Due date/time in ISO8601 format |
| tags | array[string] | No | Array of tags (max 10 tags, each max 50 chars) |

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete project documentation",
  "description": "Write comprehensive docs for the API",
  "priority": "high",
  "status": "pending",
  "due_date": "2026-02-15T17:00:00Z",
  "tags": ["documentation", "urgent"],
  "created_at": "2026-01-29T14:30:00Z",
  "updated_at": "2026-01-29T14:30:00Z",
  "completed_at": null
}
```

**Error Responses:**
- `400 Bad Request`: Invalid input data
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Server error

---

### 2. List Todos
**GET** `/todos`

Retrieve a paginated list of todos with optional filtering and sorting.

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| page | integer | No | 1 | Page number (min: 1) |
| page_size | integer | No | 20 | Items per page (min: 1, max: 100) |
| status | string | No | all | Filter by status: "pending", "in_progress", "completed", "all" |
| priority | string | No | all | Filter by priority: "low", "medium", "high", "all" |
| tags | string | No | - | Comma-separated tags (returns todos matching ANY tag) |
| sort_by | string | No | created_at | Sort field: "created_at", "updated_at", "due_date", "priority", "title" |
| sort_order | string | No | desc | Sort order: "asc", "desc" |
| search | string | No | - | Search in title and description (case-insensitive) |

**Example Request:**
```
GET /api/v1/todos?status=pending&priority=high&sort_by=due_date&sort_order=asc&page=1&page_size=20
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Complete project documentation",
      "description": "Write comprehensive docs for the API",
      "priority": "high",
      "status": "pending",
      "due_date": "2026-02-15T17:00:00Z",
      "tags": ["documentation", "urgent"],
      "created_at": "2026-01-29T14:30:00Z",
      "updated_at": "2026-01-29T14:30:00Z",
      "completed_at": null
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 45,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid query parameters
- `500 Internal Server Error`: Server error

---

### 3. Get Todo by ID
**GET** `/todos/:id`

Retrieve a specific todo by its ID.

**Path Parameters:**
| Parameter | Type | Description |
|---|---|---|
| id | UUID | Todo unique identifier |

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete project documentation",
  "description": "Write comprehensive docs for the API",
  "priority": "high",
  "status": "pending",
  "due_date": "2026-02-15T17:00:00Z",
  "tags": ["documentation", "urgent"],
  "created_at": "2026-01-29T14:30:00Z",
  "updated_at": "2026-01-29T14:30:00Z",
  "completed_at": null
}
```

**Error Responses:**
- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Todo not found
- `500 Internal Server Error`: Server error

---

### 4. Update Todo
**PUT** `/todos/:id`

Update an existing todo. All fields are optional - only provided fields will be updated.

**Path Parameters:**
| Parameter | Type | Description |
|---|---|---|
| id | UUID | Todo unique identifier |

**Request Body:**
```json
{
  "title": "Complete project documentation (updated)",
  "description": "Write comprehensive docs including examples",
  "priority": "high",
  "status": "in_progress",
  "due_date": "2026-02-20T17:00:00Z",
  "tags": ["documentation", "in-progress"]
}
```

**Request Schema:**
| Field | Type | Required | Description |
|---|---|---|---|
| title | string | No | Todo title (1-200 chars) |
| description | string | No | Detailed description (max 2000 chars) |
| priority | string | No | Priority level: "low", "medium", "high" |
| status | string | No | Status: "pending", "in_progress", "completed" |
| due_date | string (ISO8601) | No | Due date/time in ISO8601 format |
| tags | array[string] | No | Array of tags (max 10 tags, each max 50 chars) |

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete project documentation (updated)",
  "description": "Write comprehensive docs including examples",
  "priority": "high",
  "status": "in_progress",
  "due_date": "2026-02-20T17:00:00Z",
  "tags": ["documentation", "in-progress"],
  "created_at": "2026-01-29T14:30:00Z",
  "updated_at": "2026-01-29T15:45:00Z",
  "completed_at": null
}
```

**Notes:**
- When status is changed to "completed", `completed_at` is automatically set to current timestamp
- When status is changed from "completed" to other statuses, `completed_at` is set to null
- `updated_at` is automatically updated on every modification

**Error Responses:**
- `400 Bad Request`: Invalid UUID format or input data
- `404 Not Found`: Todo not found
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Server error

---

### 5. Delete Todo
**DELETE** `/todos/:id`

Soft delete a todo item (marks as deleted, doesn't remove from database).

**Path Parameters:**
| Parameter | Type | Description |
|---|---|---|
| id | UUID | Todo unique identifier |

**Response:** `204 No Content`

**Error Responses:**
- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Todo not found
- `500 Internal Server Error`: Server error

---

### 6. Get Todo Statistics
**GET** `/todos/stats`

Retrieve statistics about todos.

**Response:** `200 OK`
```json
{
  "total": 45,
  "by_status": {
    "pending": 20,
    "in_progress": 15,
    "completed": 10
  },
  "by_priority": {
    "low": 10,
    "medium": 25,
    "high": 10
  },
  "overdue": 5,
  "due_today": 3,
  "due_this_week": 12
}
```

**Error Responses:**
- `500 Internal Server Error`: Server error

---

### 7. Bulk Update Todos
**PATCH** `/todos/bulk`

Update multiple todos at once (e.g., bulk status change, bulk priority update).

**Request Body:**
```json
{
  "ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ],
  "updates": {
    "status": "completed",
    "priority": "low"
  }
}
```

**Request Schema:**
| Field | Type | Required | Description |
|---|---|---|---|
| ids | array[UUID] | Yes | Array of todo IDs to update (max 100) |
| updates | object | Yes | Fields to update (same as single update) |

**Response:** `200 OK`
```json
{
  "updated": 2,
  "failed": 0,
  "errors": []
}
```

**Error Responses:**
- `400 Bad Request`: Invalid input data
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Server error

---

## Health Check Endpoints

### 1. Liveness Check
**GET** `/health`

Basic liveness check to verify service is running.

**Response:** `200 OK`
```json
{
  "status": "ok",
  "timestamp": "2026-01-29T14:30:00Z",
  "service": "todo-api",
  "version": "1.0.0"
}
```

---

### 2. Readiness Check
**GET** `/health/ready`

Readiness check including database connectivity.

**Response:** `200 OK`
```json
{
  "status": "ready",
  "timestamp": "2026-01-29T14:30:00Z",
  "service": "todo-api",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "details": {
    "database": "ok"
  }
}
```

**Error Response:** `503 Service Unavailable` (if database is down)

---

### 3. Metrics
**GET** `/metrics`

Service metrics including database connection pool stats.

**Response:** `200 OK`
```json
{
  "database": {
    "acquire_count": 1250,
    "acquire_duration_ms": 2.5,
    "acquired_conns": 3,
    "canceled_acquire_count": 0,
    "constructing_conns": 0,
    "empty_acquire_count": 100,
    "idle_conns": 2,
    "max_conns": 10,
    "total_conns": 5
  }
}
```

---

## Error Response Format

All error responses follow a consistent structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "title",
        "message": "title is required and must be between 1 and 200 characters"
      },
      {
        "field": "priority",
        "message": "priority must be one of: low, medium, high"
      }
    ],
    "timestamp": "2026-01-29T14:30:00Z",
    "path": "/api/v1/todos"
  }
}
```

**Error Codes:**
- `VALIDATION_ERROR`: Input validation failure
- `NOT_FOUND`: Resource not found
- `INTERNAL_ERROR`: Server error
- `BAD_REQUEST`: Malformed request

---

## Authentication & Authorization

**Phase 1 (MVP)**: No authentication - public API for prototyping

**Phase 2 (Production)**: JWT-based authentication
- Use Bearer token in Authorization header: `Authorization: Bearer <token>`
- Token validation middleware
- Per-user todo isolation
- Rate limiting per user

---

## Rate Limiting

**Default Limits:**
- 100 requests per minute per IP
- 429 Too Many Requests response when exceeded

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1643472000
```

---

## CORS Policy

**Development:**
- Allow all origins: `*`

**Production:**
- Whitelist specific origins
- Allow methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
- Allow headers: Content-Type, Authorization
- Max age: 3600 seconds

---

## API Versioning

- URL path versioning: `/api/v1/`
- Version in response headers: `API-Version: 1.0.0`
- Backward compatibility maintained within major versions
- Deprecation warnings in headers for deprecated endpoints

---

## Content Types

**Supported:**
- `application/json` (default)

**Request:**
- Content-Type: `application/json`

**Response:**
- Content-Type: `application/json; charset=utf-8`

---

## HTTP Status Codes

| Code | Description | Usage |
|---|---|---|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Malformed request |
| 404 | Not Found | Resource doesn't exist |
| 422 | Unprocessable Entity | Validation error |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service down (e.g., database unavailable) |

---

## Pagination

All list endpoints support pagination:

**Request:**
```
GET /api/v1/todos?page=2&page_size=50
```

**Response includes:**
```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "page_size": 50,
    "total_items": 245,
    "total_pages": 5,
    "has_next": true,
    "has_prev": true
  }
}
```

---

## Field Validation Rules

### Title
- Required
- Min length: 1
- Max length: 200
- Type: string
- Cannot be only whitespace

### Description
- Optional
- Max length: 2000
- Type: string

### Priority
- Optional (default: "medium")
- Allowed values: "low", "medium", "high"
- Type: string enum

### Status
- Auto-set to "pending" on creation
- Allowed values: "pending", "in_progress", "completed"
- Type: string enum

### Due Date
- Optional
- Format: ISO8601 (RFC3339)
- Must be valid future date (warning if past)
- Type: timestamp with timezone

### Tags
- Optional
- Max tags: 10
- Max length per tag: 50
- Type: array of strings
- Tags are case-insensitive (stored lowercase)
- Duplicate tags removed automatically

---

## Example Usage

### Create a Todo
```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Review pull request",
    "description": "Review PR #123 for authentication feature",
    "priority": "high",
    "due_date": "2026-01-30T17:00:00Z",
    "tags": ["code-review", "urgent"]
  }'
```

### List Todos with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/todos?status=pending&priority=high&sort_by=due_date&sort_order=asc"
```

### Update Todo Status
```bash
curl -X PUT http://localhost:8080/api/v1/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

### Delete Todo
```bash
curl -X DELETE http://localhost:8080/api/v1/todos/550e8400-e29b-41d4-a716-446655440000
```

---

## OpenAPI/Swagger Documentation

Interactive API documentation will be available at:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

Generated using `swaggo/gin-swagger`.
