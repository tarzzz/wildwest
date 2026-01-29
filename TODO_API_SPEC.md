# Todo List API - Complete Specification

## Overview

This document provides the complete REST API specification for the Todo List API, including endpoint definitions, request/response formats, error handling, and usage examples.

**Base URL**: `http://localhost:8082` (development)
**API Version**: v1
**API Prefix**: `/api/v1`
**Content Type**: `application/json`
**Authentication**: None (Phase 1 - Public API)

---

## API Conventions

### HTTP Methods
- `GET`: Retrieve resources (idempotent, cacheable)
- `POST`: Create new resources
- `PUT`: Update entire resource (idempotent)
- `DELETE`: Delete resource (soft delete)

### Response Codes
- `200 OK`: Successful GET, PUT
- `201 Created`: Successful POST (resource created)
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Invalid request format or validation error
- `404 Not Found`: Resource not found
- `422 Unprocessable Entity`: Validation error with details
- `500 Internal Server Error`: Server error

### Pagination Format
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

### Timestamps
- All timestamps in ISO 8601 format with timezone
- Example: `2026-01-29T14:30:00Z`

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": [
      {
        "field": "field_name",
        "message": "Specific field error"
      }
    ],
    "timestamp": "2026-01-29T14:30:00Z",
    "path": "/api/v1/todos"
  }
}
```

---

## Endpoints

### 1. Create Todo

**POST** `/api/v1/todos`

Create a new todo item.

**Request Body**:
```json
{
  "title": "Complete project documentation",
  "description": "Write comprehensive docs for the API",
  "priority": "high",
  "due_date": "2026-02-15T17:00:00Z",
  "tags": ["documentation", "urgent"]
}
```

**Request Schema**:
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| title | string | Yes | 1-200 characters, not empty/whitespace |
| description | string | No | Max 2000 characters |
| priority | string | No | Enum: "low", "medium", "high" (default: "medium") |
| due_date | string | No | ISO 8601 format, valid datetime |
| tags | array[string] | No | Max 10 tags, each max 50 chars, case-insensitive |

**Success Response** (201 Created):
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

**Error Responses**:
- `400 Bad Request`: Malformed JSON
- `422 Unprocessable Entity`: Validation errors

**Example Error**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed for one or more fields",
    "details": [
      {
        "field": "title",
        "message": "title is required and must be between 1 and 200 characters"
      }
    ],
    "timestamp": "2026-01-29T14:30:00Z",
    "path": "/api/v1/todos"
  }
}
```

---

### 2. List Todos

**GET** `/api/v1/todos`

Retrieve a paginated list of todos with optional filtering and sorting.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | integer | No | 1 | Page number (min: 1) |
| page_size | integer | No | 20 | Items per page (min: 1, max: 100) |
| status | string | No | all | Filter: "pending", "in_progress", "completed", "all" |
| priority | string | No | all | Filter: "low", "medium", "high", "all" |
| tags | string | No | - | Comma-separated tags (OR logic) |
| sort_by | string | No | created_at | Sort field: "created_at", "updated_at", "due_date", "priority", "title" |
| sort_order | string | No | desc | Sort order: "asc", "desc" |
| search | string | No | - | Search in title and description (case-insensitive) |

**Example Request**:
```
GET /api/v1/todos?status=pending&priority=high&sort_by=due_date&sort_order=asc&page=1&page_size=20
```

**Success Response** (200 OK):
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

**Error Responses**:
- `400 Bad Request`: Invalid query parameters

---

### 3. Get Todo by ID

**GET** `/api/v1/todos/:id`

Retrieve a specific todo by its UUID.

**Path Parameters**:
- `id`: UUID of the todo

**Success Response** (200 OK):
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

**Error Responses**:
- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Todo not found

---

### 4. Update Todo

**PUT** `/api/v1/todos/:id`

Update an existing todo. All fields are optional - only provided fields will be updated.

**Path Parameters**:
- `id`: UUID of the todo

**Request Body**:
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

**Request Schema**:
All fields are optional. Same validation rules as create endpoint.

**Success Response** (200 OK):
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

**Business Rules**:
- When status changes to "completed", `completed_at` is automatically set to current timestamp
- When status changes from "completed" to other status, `completed_at` is set to null
- `updated_at` is automatically updated on every modification

**Error Responses**:
- `400 Bad Request`: Invalid UUID or malformed request
- `404 Not Found`: Todo not found
- `422 Unprocessable Entity`: Validation errors

---

### 5. Delete Todo

**DELETE** `/api/v1/todos/:id`

Soft delete a todo item. The todo is marked as deleted but not removed from the database.

**Path Parameters**:
- `id`: UUID of the todo

**Success Response** (204 No Content)

**Error Responses**:
- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Todo not found

**Note**: This is a soft delete. The record remains in the database with `deleted_at` timestamp set.

---

### 6. Get Todo Statistics

**GET** `/api/v1/todos/stats`

Retrieve statistics about todos.

**Success Response** (200 OK):
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

---

### 7. Bulk Update Todos

**PATCH** `/api/v1/todos/bulk`

Update multiple todos at once.

**Request Body**:
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

**Request Schema**:
- `ids`: Array of UUIDs (max 100)
- `updates`: Object with fields to update (same validation as single update)

**Success Response** (200 OK):
```json
{
  "updated": 2,
  "failed": 0,
  "errors": []
}
```

**Partial Success Response** (200 OK):
```json
{
  "updated": 1,
  "failed": 1,
  "errors": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "error": "Todo not found"
    }
  ]
}
```

---

## Health Check Endpoints

### GET /health

Basic liveness check.

**Success Response** (200 OK):
```json
{
  "status": "ok",
  "timestamp": "2026-01-29T14:30:00Z",
  "service": "todo-api",
  "version": "1.0.0"
}
```

---

### GET /health/ready

Readiness check including database connectivity.

**Success Response** (200 OK):
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

**Error Response** (503 Service Unavailable):
```json
{
  "status": "not_ready",
  "timestamp": "2026-01-29T14:30:00Z",
  "service": "todo-api",
  "version": "1.0.0",
  "details": {
    "database": "error: database connection failed"
  }
}
```

---

## Error Codes

### Validation Errors
- `VALIDATION_ERROR`: Request validation failed
- `REQUIRED_FIELD_MISSING`: Required field not provided
- `INVALID_FORMAT`: Invalid field format
- `INVALID_VALUE`: Value not in allowed set

### Resource Errors
- `NOT_FOUND`: Todo not found
- `ALREADY_EXISTS`: Resource already exists

### Server Errors
- `INTERNAL_ERROR`: Unexpected server error
- `DATABASE_ERROR`: Database operation failed

---

## Field Validation Rules

### Title
- **Required**: Yes
- **Type**: String
- **Min Length**: 1
- **Max Length**: 200
- **Validation**: Cannot be only whitespace

### Description
- **Required**: No
- **Type**: String
- **Max Length**: 2000

### Priority
- **Required**: No
- **Type**: Enum
- **Default**: "medium"
- **Allowed Values**: "low", "medium", "high"

### Status
- **Required**: No (auto-set to "pending" on creation)
- **Type**: Enum
- **Allowed Values**: "pending", "in_progress", "completed"

### Due Date
- **Required**: No
- **Type**: Timestamp
- **Format**: ISO 8601 (RFC3339)
- **Validation**: Must be valid datetime (warning if past date)

### Tags
- **Required**: No
- **Type**: Array of strings
- **Max Tags**: 10
- **Max Length per Tag**: 50
- **Normalization**: Stored as lowercase
- **Deduplication**: Duplicates removed automatically

---

## cURL Examples

### Create a Todo
```bash
curl -X POST http://localhost:8082/api/v1/todos \
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
curl -X GET "http://localhost:8082/api/v1/todos?status=pending&priority=high&sort_by=due_date&sort_order=asc"
```

### Get Todo by ID
```bash
curl -X GET http://localhost:8082/api/v1/todos/550e8400-e29b-41d4-a716-446655440000
```

### Update Todo
```bash
curl -X PUT http://localhost:8082/api/v1/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

### Delete Todo
```bash
curl -X DELETE http://localhost:8082/api/v1/todos/550e8400-e29b-41d4-a716-446655440000
```

### Get Statistics
```bash
curl -X GET http://localhost:8082/api/v1/todos/stats
```

### Bulk Update
```bash
curl -X PATCH http://localhost:8082/api/v1/todos/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["550e8400-e29b-41d4-a716-446655440000"],
    "updates": {"status": "completed"}
  }'
```

---

## Future Enhancements (Phase 2)

### Authentication
- JWT-based authentication
- Bearer token in Authorization header
- Per-user todo isolation

### Additional Features
- Todo categories/projects
- Subtasks
- Attachments
- Comments
- Recurring todos
- Reminders

---

**Document Version**: 1.0
**Last Updated**: 2026-01-29
**Author**: Solutions Architect (ford)
**Status**: Design Phase Complete
