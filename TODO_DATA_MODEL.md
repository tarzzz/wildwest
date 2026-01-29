# Todo List Data Model

## Overview
Data structures, database schemas, and validation rules for the Todo API.

---

## Domain Model

### Todo Entity

**Go Struct Definition:**
```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// Todo represents a todo item
type Todo struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    Title       string     `json:"title" db:"title" binding:"required,min=1,max=200"`
    Description string     `json:"description" db:"description" binding:"max=2000"`
    Priority    Priority   `json:"priority" db:"priority" binding:"oneof=low medium high"`
    Status      Status     `json:"status" db:"status" binding:"oneof=pending in_progress completed"`
    DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
    Tags        []string   `json:"tags" db:"-"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
    CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
    DeletedAt   *time.Time `json:"-" db:"deleted_at"` // Soft delete
}

// Priority levels
type Priority string

const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
)

// Status values
type Status string

const (
    StatusPending    Status = "pending"
    StatusInProgress Status = "in_progress"
    StatusCompleted  Status = "completed"
)

// TableName returns the table name for Todo
func (Todo) TableName() string {
    return "todos"
}
```

---

## Database Schema

### PostgreSQL Tables

#### 1. `todos` Table

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL CHECK (length(trim(title)) > 0),
    description TEXT,
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_title CHECK (length(trim(title)) BETWEEN 1 AND 200),
    CONSTRAINT valid_description CHECK (description IS NULL OR length(description) <= 2000),
    CONSTRAINT completed_at_consistency CHECK (
        (status = 'completed' AND completed_at IS NOT NULL) OR
        (status != 'completed' AND completed_at IS NULL)
    )
);

-- Indexes for performance
CREATE INDEX idx_todos_status ON todos(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_priority ON todos(priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE deleted_at IS NULL AND due_date IS NOT NULL;
CREATE INDEX idx_todos_created_at ON todos(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_updated_at ON todos(updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_deleted_at ON todos(deleted_at) WHERE deleted_at IS NOT NULL;

-- Full-text search index for title and description
CREATE INDEX idx_todos_search ON todos USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')))
WHERE deleted_at IS NULL;

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_todos_updated_at BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to auto-set completed_at
CREATE OR REPLACE FUNCTION set_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        NEW.completed_at = NOW();
    ELSIF NEW.status != 'completed' AND OLD.status = 'completed' THEN
        NEW.completed_at = NULL;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER set_todos_completed_at BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION set_completed_at();
```

---

#### 2. `todo_tags` Table

Many-to-many relationship for flexible tag management.

```sql
CREATE TABLE todo_tags (
    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL CHECK (length(trim(tag)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (todo_id, tag)
);

-- Indexes
CREATE INDEX idx_todo_tags_tag ON todo_tags(tag);
CREATE INDEX idx_todo_tags_todo_id ON todo_tags(todo_id);

-- Constraint: max 10 tags per todo
CREATE OR REPLACE FUNCTION check_max_tags()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM todo_tags WHERE todo_id = NEW.todo_id) >= 10 THEN
        RAISE EXCEPTION 'Maximum 10 tags allowed per todo';
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER enforce_max_tags BEFORE INSERT ON todo_tags
    FOR EACH ROW EXECUTE FUNCTION check_max_tags();
```

---

## Database Migrations

### Migration Files Structure
```
migrations/
├── 000001_create_todos_table.up.sql
├── 000001_create_todos_table.down.sql
├── 000002_create_todo_tags_table.up.sql
├── 000002_create_todo_tags_table.down.sql
├── 000003_add_indexes.up.sql
└── 000003_add_indexes.down.sql
```

### Migration Tool
Use **golang-migrate** for database migrations:
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/todos?sslmode=disable" up
```

---

## Repository Layer Interface

**Go Interface Definition:**
```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "wildwest/todo-api/internal/domain"
)

// TodoRepository defines data access methods
type TodoRepository interface {
    // Create creates a new todo
    Create(ctx context.Context, todo *domain.Todo) error

    // GetByID retrieves a todo by ID
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error)

    // List retrieves todos with filters and pagination
    List(ctx context.Context, filter ListFilter) ([]*domain.Todo, *Pagination, error)

    // Update updates an existing todo
    Update(ctx context.Context, todo *domain.Todo) error

    // Delete soft-deletes a todo
    Delete(ctx context.Context, id uuid.UUID) error

    // BulkUpdate updates multiple todos
    BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) (int, error)

    // GetStats retrieves statistics
    GetStats(ctx context.Context) (*domain.TodoStats, error)
}

// ListFilter defines filtering options
type ListFilter struct {
    Page      int
    PageSize  int
    Status    string
    Priority  string
    Tags      []string
    Search    string
    SortBy    string
    SortOrder string
}

// Pagination defines pagination metadata
type Pagination struct {
    Page        int  `json:"page"`
    PageSize    int  `json:"page_size"`
    TotalItems  int  `json:"total_items"`
    TotalPages  int  `json:"total_pages"`
    HasNext     bool `json:"has_next"`
    HasPrev     bool `json:"has_prev"`
}
```

---

## Statistics Model

```go
package domain

// TodoStats represents todo statistics
type TodoStats struct {
    Total       int                `json:"total"`
    ByStatus    map[string]int     `json:"by_status"`
    ByPriority  map[string]int     `json:"by_priority"`
    Overdue     int                `json:"overdue"`
    DueToday    int                `json:"due_today"`
    DueThisWeek int                `json:"due_this_week"`
}
```

**SQL Query for Statistics:**
```sql
SELECT
    COUNT(*) FILTER (WHERE deleted_at IS NULL) as total,
    COUNT(*) FILTER (WHERE status = 'pending' AND deleted_at IS NULL) as pending,
    COUNT(*) FILTER (WHERE status = 'in_progress' AND deleted_at IS NULL) as in_progress,
    COUNT(*) FILTER (WHERE status = 'completed' AND deleted_at IS NULL) as completed,
    COUNT(*) FILTER (WHERE priority = 'low' AND deleted_at IS NULL) as priority_low,
    COUNT(*) FILTER (WHERE priority = 'medium' AND deleted_at IS NULL) as priority_medium,
    COUNT(*) FILTER (WHERE priority = 'high' AND deleted_at IS NULL) as priority_high,
    COUNT(*) FILTER (WHERE due_date < NOW() AND status != 'completed' AND deleted_at IS NULL) as overdue,
    COUNT(*) FILTER (WHERE DATE(due_date) = CURRENT_DATE AND deleted_at IS NULL) as due_today,
    COUNT(*) FILTER (WHERE due_date BETWEEN NOW() AND NOW() + INTERVAL '7 days' AND deleted_at IS NULL) as due_this_week
FROM todos;
```

---

## Validation Rules

### Field Validations

**Title:**
- Required: ✅
- Type: string
- Min length: 1 (trimmed)
- Max length: 200
- No whitespace-only: ✅

**Description:**
- Required: ❌
- Type: string
- Max length: 2000

**Priority:**
- Required: ❌ (default: "medium")
- Type: enum
- Values: "low", "medium", "high"

**Status:**
- Required: ✅ (auto-set on create)
- Type: enum
- Values: "pending", "in_progress", "completed"

**Due Date:**
- Required: ❌
- Type: timestamp with timezone
- Format: ISO8601/RFC3339
- Validation: Must be valid date

**Tags:**
- Required: ❌
- Type: array of strings
- Max items: 10
- Max length per tag: 50
- Normalization: lowercase, trimmed
- Deduplication: automatic

---

## Data Transfer Objects (DTOs)

### Create Todo Request
```go
type CreateTodoRequest struct {
    Title       string    `json:"title" binding:"required,min=1,max=200"`
    Description string    `json:"description" binding:"max=2000"`
    Priority    string    `json:"priority" binding:"omitempty,oneof=low medium high"`
    DueDate     *string   `json:"due_date" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    Tags        []string  `json:"tags" binding:"omitempty,max=10,dive,max=50"`
}
```

### Update Todo Request
```go
type UpdateTodoRequest struct {
    Title       *string   `json:"title" binding:"omitempty,min=1,max=200"`
    Description *string   `json:"description" binding:"omitempty,max=2000"`
    Priority    *string   `json:"priority" binding:"omitempty,oneof=low medium high"`
    Status      *string   `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
    DueDate     *string   `json:"due_date" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    Tags        *[]string `json:"tags" binding:"omitempty,max=10,dive,max=50"`
}
```

### Bulk Update Request
```go
type BulkUpdateRequest struct {
    IDs     []string               `json:"ids" binding:"required,min=1,max=100,dive,uuid"`
    Updates map[string]interface{} `json:"updates" binding:"required"`
}
```

### Todo Response
```go
type TodoResponse struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Priority    string    `json:"priority"`
    Status      string    `json:"status"`
    DueDate     *string   `json:"due_date,omitempty"`
    Tags        []string  `json:"tags"`
    CreatedAt   string    `json:"created_at"`
    UpdatedAt   string    `json:"updated_at"`
    CompletedAt *string   `json:"completed_at,omitempty"`
}
```

---

## Database Connection Configuration

```go
type DatabaseConfig struct {
    Host            string `mapstructure:"host" validate:"required"`
    Port            int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    User            string `mapstructure:"user" validate:"required"`
    Password        string `mapstructure:"password" validate:"required"`
    DBName          string `mapstructure:"dbname" validate:"required"`
    SSLMode         string `mapstructure:"sslmode" validate:"oneof=disable require verify-ca verify-full"`
    MaxConns        int    `mapstructure:"max_conns" validate:"min=1,max=100"`
    MinConns        int    `mapstructure:"min_conns" validate:"min=0,max=100"`
    MaxConnLifetime int    `mapstructure:"max_conn_lifetime" validate:"min=0"` // seconds
    MaxConnIdleTime int    `mapstructure:"max_conn_idle_time" validate:"min=0"` // seconds
}
```

**Default Values:**
- Host: `localhost`
- Port: `5432`
- SSLMode: `disable` (dev), `require` (prod)
- MaxConns: `10`
- MinConns: `2`
- MaxConnLifetime: `3600` (1 hour)
- MaxConnIdleTime: `600` (10 minutes)

---

## Query Patterns

### List with Filters
```sql
SELECT
    t.id, t.title, t.description, t.priority, t.status,
    t.due_date, t.created_at, t.updated_at, t.completed_at,
    COALESCE(array_agg(tt.tag) FILTER (WHERE tt.tag IS NOT NULL), '{}') as tags
FROM todos t
LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE t.deleted_at IS NULL
    AND ($1::text IS NULL OR t.status = $1)
    AND ($2::text IS NULL OR t.priority = $2)
    AND ($3::text IS NULL OR to_tsvector('english', t.title || ' ' || COALESCE(t.description, '')) @@ plainto_tsquery('english', $3))
    AND ($4::text[] IS NULL OR tt.tag = ANY($4))
GROUP BY t.id
ORDER BY
    CASE WHEN $5 = 'created_at' THEN t.created_at END DESC,
    CASE WHEN $5 = 'updated_at' THEN t.updated_at END DESC,
    CASE WHEN $5 = 'due_date' THEN t.due_date END ASC,
    CASE WHEN $5 = 'priority' THEN
        CASE t.priority
            WHEN 'high' THEN 1
            WHEN 'medium' THEN 2
            WHEN 'low' THEN 3
        END
    END
LIMIT $6 OFFSET $7;
```

### Get by ID with Tags
```sql
SELECT
    t.id, t.title, t.description, t.priority, t.status,
    t.due_date, t.created_at, t.updated_at, t.completed_at,
    COALESCE(array_agg(tt.tag) FILTER (WHERE tt.tag IS NOT NULL), '{}') as tags
FROM todos t
LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE t.id = $1 AND t.deleted_at IS NULL
GROUP BY t.id;
```

---

## Transaction Management

Use database transactions for operations that modify multiple tables:

```go
func (r *todoRepository) Create(ctx context.Context, todo *domain.Todo) error {
    return r.db.RunInTransaction(ctx, func(ctx context.Context) error {
        // Insert todo
        err := r.insertTodo(ctx, todo)
        if err != nil {
            return err
        }

        // Insert tags
        if len(todo.Tags) > 0 {
            err = r.insertTags(ctx, todo.ID, todo.Tags)
            if err != nil {
                return err
            }
        }

        return nil
    })
}
```

---

## Data Consistency Rules

1. **Status-CompletedAt Consistency**:
   - When `status = 'completed'`, `completed_at` must be set
   - When `status != 'completed'`, `completed_at` must be NULL
   - Enforced by database trigger

2. **Soft Delete**:
   - Never DELETE records, always set `deleted_at`
   - Queries always filter `WHERE deleted_at IS NULL`

3. **Tag Normalization**:
   - Tags stored in lowercase
   - Whitespace trimmed
   - Duplicates removed before insert

4. **Timestamp Consistency**:
   - `created_at`: Set once on creation, never modified
   - `updated_at`: Auto-updated on every modification via trigger
   - `completed_at`: Auto-managed by status trigger

---

## Performance Considerations

### Indexes
- Composite index on `(status, deleted_at)` for filtered queries
- Index on `due_date` for date-based queries
- GIN index for full-text search
- Partial indexes to exclude soft-deleted records

### Connection Pooling
- Min 2, Max 10 connections by default
- Connection lifetime: 1 hour
- Idle timeout: 10 minutes

### Query Optimization
- Use prepared statements
- Fetch only required columns
- JOIN tags in single query (avoid N+1)
- Paginate all list queries

---

## Backup & Recovery

### Backup Strategy
- Daily full backups at 2 AM
- Point-in-time recovery with WAL archiving
- Retention: 30 days

### Backup Command
```bash
pg_dump -h localhost -U postgres -d todos -F c -f todos_backup_$(date +%Y%m%d).dump
```

### Restore Command
```bash
pg_restore -h localhost -U postgres -d todos -c todos_backup_20260129.dump
```

---

## Seed Data (Development)

```sql
-- Sample todos for development
INSERT INTO todos (title, description, priority, status, due_date) VALUES
('Setup development environment', 'Install Go, PostgreSQL, and dependencies', 'high', 'completed', NULL),
('Write API documentation', 'Complete OpenAPI specification', 'high', 'in_progress', '2026-02-01T17:00:00Z'),
('Implement authentication', 'Add JWT-based auth middleware', 'medium', 'pending', '2026-02-15T17:00:00Z'),
('Write unit tests', 'Achieve 80% code coverage', 'medium', 'pending', '2026-02-20T17:00:00Z'),
('Deploy to staging', 'Deploy API to staging environment', 'low', 'pending', '2026-03-01T17:00:00Z');

-- Sample tags
INSERT INTO todo_tags (todo_id, tag)
SELECT id, 'setup' FROM todos WHERE title = 'Setup development environment';
INSERT INTO todo_tags (todo_id, tag)
SELECT id, 'documentation' FROM todos WHERE title = 'Write API documentation';
INSERT INTO todo_tags (todo_id, tag)
SELECT id, 'security' FROM todos WHERE title = 'Implement authentication';
```
