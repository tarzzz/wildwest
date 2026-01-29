# User Management API - Database Schema

## Overview

This document defines the complete database schema for the User Management API, including table structures, relationships, indexes, constraints, and migration strategies.

**Database**: PostgreSQL 15+
**Driver**: pgx with connection pooling
**Migration Tool**: golang-migrate
**Naming Convention**: snake_case for all database objects

## Entity Relationship Diagram (ERD)

```
┌─────────────────────────────────────────────────────┐
│                      users                          │
├─────────────────────────────────────────────────────┤
│ PK  id                  UUID                        │
│ UQ  email               VARCHAR(255)                │
│     password_hash       VARCHAR(255)                │
│     name                VARCHAR(255)                │
│     bio                 TEXT                        │
│     avatar_url          VARCHAR(500)                │
│     role                VARCHAR(20)                 │
│     is_active           BOOLEAN                     │
│     email_verified      BOOLEAN                     │
│     email_verified_at   TIMESTAMP                   │
│     last_login          TIMESTAMP                   │
│     failed_login_count  INTEGER                     │
│     locked_until        TIMESTAMP                   │
│     created_at          TIMESTAMP                   │
│     updated_at          TIMESTAMP                   │
│     deleted_at          TIMESTAMP                   │
└──────────────────┬──────────────────────────────────┘
                   │
                   │ 1:N
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│                refresh_tokens                       │
├─────────────────────────────────────────────────────┤
│ PK  id                  UUID                        │
│ FK  user_id             UUID                        │
│     token_hash          VARCHAR(255)                │
│     expires_at          TIMESTAMP                   │
│     revoked             BOOLEAN                     │
│     revoked_at          TIMESTAMP                   │
│     created_at          TIMESTAMP                   │
│ IDX (user_id)                                       │
│ IDX (token_hash)                                    │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│                  audit_logs                         │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGSERIAL                   │
│ FK  user_id             UUID                        │
│     action              VARCHAR(100)                │
│     resource_type       VARCHAR(100)                │
│     resource_id         UUID                        │
│     ip_address          VARCHAR(45)                 │
│     user_agent          TEXT                        │
│     details             JSONB                       │
│     created_at          TIMESTAMP                   │
│ IDX (user_id)                                       │
│ IDX (action)                                        │
│ IDX (created_at)                                    │
│ IDX (resource_type, resource_id)                    │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│             password_reset_tokens                   │
├─────────────────────────────────────────────────────┤
│ PK  id                  UUID                        │
│ FK  user_id             UUID                        │
│     token_hash          VARCHAR(255)                │
│     expires_at          TIMESTAMP                   │
│     used                BOOLEAN                     │
│     used_at             TIMESTAMP                   │
│     created_at          TIMESTAMP                   │
│ IDX (user_id)                                       │
│ IDX (token_hash)                                    │
└─────────────────────────────────────────────────────┘
```

## Table Definitions

### users

Core user account table with authentication credentials and profile information.

```sql
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255) NOT NULL UNIQUE,
    password_hash       VARCHAR(255) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    bio                 TEXT,
    avatar_url          VARCHAR(500),
    role                VARCHAR(20) NOT NULL DEFAULT 'user',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    email_verified      BOOLEAN NOT NULL DEFAULT false,
    email_verified_at   TIMESTAMP WITH TIME ZONE,
    last_login          TIMESTAMP WITH TIME ZONE,
    failed_login_count  INTEGER NOT NULL DEFAULT 0,
    locked_until        TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMP WITH TIME ZONE,

    CONSTRAINT check_role CHECK (role IN ('admin', 'user', 'guest')),
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
    CONSTRAINT check_failed_login_count CHECK (failed_login_count >= 0)
);

-- Indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_is_active ON users(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Updated timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Column Descriptions**:
- `id`: Unique user identifier (UUID v4)
- `email`: User email address (unique, used for login)
- `password_hash`: bcrypt hashed password (never store plaintext)
- `name`: User's display name
- `bio`: Optional user biography/description
- `avatar_url`: Optional URL to user's profile picture
- `role`: User role for RBAC (admin, user, guest)
- `is_active`: Account active status (false = deactivated)
- `email_verified`: Whether email has been verified
- `email_verified_at`: Timestamp of email verification
- `last_login`: Timestamp of last successful login
- `failed_login_count`: Count of consecutive failed login attempts
- `locked_until`: Account locked until this timestamp (for brute force protection)
- `created_at`: Account creation timestamp
- `updated_at`: Last update timestamp (auto-updated by trigger)
- `deleted_at`: Soft delete timestamp (NULL = not deleted)

**Constraints**:
- Email must be unique and match email format
- Role must be one of: admin, user, guest
- Failed login count must be non-negative

### refresh_tokens

Stores refresh tokens for JWT authentication, allowing token renewal without re-login.

```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT false,
    revoked_at      TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT check_revoked_at CHECK (
        (revoked = true AND revoked_at IS NOT NULL) OR
        (revoked = false AND revoked_at IS NULL)
    )
);

-- Indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked) WHERE revoked = false;
```

**Column Descriptions**:
- `id`: Unique token identifier
- `user_id`: Foreign key to users table
- `token_hash`: SHA-256 hash of the refresh token (never store plaintext tokens)
- `expires_at`: Token expiration timestamp (typically 7 days)
- `revoked`: Whether token has been revoked (logout, security)
- `revoked_at`: Timestamp when token was revoked
- `created_at`: Token creation timestamp

**Business Rules**:
- One user can have multiple active refresh tokens (multiple devices)
- Tokens automatically expire after configured duration
- Revoked tokens cannot be used even if not expired
- Cascade delete when user is deleted

### audit_logs

Comprehensive audit trail for security and compliance.

```sql
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(100) NOT NULL,
    resource_type   VARCHAR(100) NOT NULL,
    resource_id     UUID,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    details         JSONB,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_details ON audit_logs USING GIN(details);

-- Partition by month for better performance (optional)
-- CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

**Column Descriptions**:
- `id`: Auto-incrementing log entry ID
- `user_id`: User who performed the action (NULL for anonymous actions)
- `action`: Action performed (e.g., "user.login", "user.update", "user.delete")
- `resource_type`: Type of resource affected (e.g., "user", "profile")
- `resource_id`: ID of the affected resource
- `ip_address`: Client IP address (IPv4 or IPv6)
- `user_agent`: Client user agent string
- `details`: JSON object with additional context (old values, new values, etc.)
- `created_at`: Timestamp of the action

**Audit Actions**:
- `user.register`: New user registration
- `user.login`: Successful login
- `user.login_failed`: Failed login attempt
- `user.logout`: User logout
- `user.update`: User profile update
- `user.delete`: User account deletion
- `user.password_change`: Password change
- `user.password_reset`: Password reset
- `user.email_verify`: Email verification
- `user.role_change`: Role modification (admin only)

### password_reset_tokens

Temporary tokens for password reset flow.

```sql
CREATE TABLE password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    used        BOOLEAN NOT NULL DEFAULT false,
    used_at     TIMESTAMP WITH TIME ZONE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT check_used_at CHECK (
        (used = true AND used_at IS NOT NULL) OR
        (used = false AND used_at IS NULL)
    )
);

-- Indexes
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
```

**Column Descriptions**:
- `id`: Unique token identifier
- `user_id`: Foreign key to users table
- `token_hash`: SHA-256 hash of the reset token (sent via email)
- `expires_at`: Token expiration (typically 1 hour)
- `used`: Whether token has been used
- `used_at`: Timestamp when token was used
- `created_at`: Token creation timestamp

**Business Rules**:
- One user can have multiple active reset tokens (multiple requests)
- Tokens expire after 1 hour
- Used tokens cannot be reused
- Only most recent token should be used (security best practice)

## Data Types and Conventions

### UUID
- Primary keys use UUID v4 for distributed ID generation
- Better for security (no sequential guessing)
- Better for distributed systems (no coordination needed)

### Timestamps
- All timestamps use `TIMESTAMP WITH TIME ZONE`
- Ensures timezone-aware storage
- Default to NOW() for creation timestamps
- Automatic update trigger for `updated_at`

### VARCHAR vs TEXT
- `VARCHAR(n)`: Fixed maximum length fields (email, name)
- `TEXT`: Unbounded length fields (bio, user_agent)

### Boolean Defaults
- Default to `false` for safety (opt-in behavior)
- Explicit NOT NULL for data integrity

### Soft Deletes
- `deleted_at TIMESTAMP`: NULL = active, non-NULL = deleted
- Indexes exclude deleted rows for performance
- Cascade deletes for related tables (tokens)

## Indexes Strategy

### Primary Indexes
- All primary keys (UUID) have automatic indexes

### Lookup Indexes
- `users.email`: Most common lookup (login)
- `refresh_tokens.token_hash`: Token validation
- `password_reset_tokens.token_hash`: Reset validation

### Filter Indexes
- Partial indexes on `deleted_at IS NULL` (exclude soft-deleted)
- Partial indexes on `revoked = false` (exclude revoked tokens)

### Performance Indexes
- `audit_logs.created_at DESC`: Recent logs query
- `users.created_at`: User registration trends

### Composite Indexes
- `(resource_type, resource_id)`: Audit log resource queries

### JSONB GIN Index
- `audit_logs.details`: Enable fast JSONB queries

## Constraints

### Foreign Key Constraints
- **ON DELETE CASCADE**: Tokens deleted when user deleted
- **ON DELETE SET NULL**: Audit logs preserved but user_id nullified

### Check Constraints
- Email format validation
- Role enumeration enforcement
- Non-negative counts
- Conditional field requirements (revoked_at, used_at)

### Unique Constraints
- Email uniqueness
- Token hash uniqueness

## Migration Strategy

### Migration Files

**Migration 1: Initial Schema**
```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_refresh_tokens_table.up.sql
├── 000002_create_refresh_tokens_table.down.sql
├── 000003_create_audit_logs_table.up.sql
├── 000003_create_audit_logs_table.down.sql
├── 000004_create_password_reset_tokens_table.up.sql
├── 000004_create_password_reset_tokens_table.down.sql
└── 000005_create_indexes.up.sql
└── 000005_create_indexes.down.sql
```

### Migration Commands
```bash
# Apply all migrations
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback one migration
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" down 1

# Check migration version
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" version
```

### Migration Best Practices
1. **Always write down migrations**: Every up has a corresponding down
2. **Test migrations on copy**: Never run on production first
3. **Idempotent migrations**: Use `IF NOT EXISTS` where possible
4. **Small migrations**: One logical change per migration
5. **Data migrations separate**: Schema and data in different migrations

## Seed Data

### Development Seed Data

```sql
-- Admin user (password: admin123)
INSERT INTO users (id, email, password_hash, name, role, is_active, email_verified, email_verified_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJLOKMfim', -- bcrypt hash
    'Admin User',
    'admin',
    true,
    true,
    NOW()
);

-- Regular user (password: user123)
INSERT INTO users (id, email, password_hash, name, role, is_active, email_verified, email_verified_at)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'user@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJLOKMfim',
    'Test User',
    'user',
    true,
    true,
    NOW()
);

-- Guest user (password: guest123)
INSERT INTO users (id, email, password_hash, name, role, is_active, email_verified, email_verified_at)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'guest@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJLOKMfim',
    'Guest User',
    'guest',
    true,
    true,
    NOW()
);
```

## Performance Considerations

### Connection Pooling
```go
// Recommended pgxpool configuration
config, _ := pgxpool.ParseConfig(databaseURL)
config.MaxConns = 25                   // Maximum connections
config.MinConns = 5                    // Minimum connections
config.MaxConnLifetime = time.Hour     // Recycle connections
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = time.Minute
```

### Query Optimization
- Use prepared statements for repeated queries
- Limit result sets with pagination
- Use covering indexes where possible
- Avoid N+1 queries (use joins or batching)

### Monitoring Queries
```sql
-- Check table sizes
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan ASC;

-- Find slow queries
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

## Security Considerations

### Password Storage
- NEVER store plaintext passwords
- Use bcrypt with cost factor 12
- Hash passwords in application layer, not database

### Token Storage
- Store only hashed tokens in database
- Use SHA-256 for token hashing
- Send plaintext token only once (via email or response)

### SQL Injection Prevention
- Always use parameterized queries
- Never concatenate user input into SQL
- Use pgx's parameter binding: `$1, $2, $3`

### Audit Logging
- Log all authentication events
- Log all data modifications
- Store IP address and user agent
- Retain logs for compliance period

### Data Privacy
- Hash sensitive tokens
- Soft delete preserves audit trail
- Consider GDPR right to erasure

## Backup and Recovery

### Backup Strategy
```bash
# Full database backup
pg_dump -h localhost -U postgres -d userapi -F c -f backup.dump

# Restore from backup
pg_restore -h localhost -U postgres -d userapi backup.dump

# Incremental backup (WAL archiving)
# Configure in postgresql.conf:
# wal_level = replica
# archive_mode = on
# archive_command = 'cp %p /path/to/archive/%f'
```

### Recovery Procedures
1. Regular backups (daily full + continuous WAL)
2. Test restore procedures monthly
3. Off-site backup storage
4. Point-in-time recovery capability

## Future Schema Enhancements

### Phase 2 Additions
- `user_sessions`: Track active sessions per device
- `user_preferences`: Store user settings as JSONB
- `oauth_providers`: Social login integration
- `two_factor_auth`: TOTP secrets and backup codes
- `api_keys`: Programmatic access tokens
- `rate_limit_buckets`: Distributed rate limiting state

### Partitioning
- Partition `audit_logs` by month for better query performance
- Partition `refresh_tokens` by expiration date

### Read Replicas
- Configure streaming replication for read scaling
- Route read queries to replicas
- Route write queries to primary

---

**Document Version**: 1.0
**Last Updated**: 2026-01-27
**Author**: Solutions Architect (berners-lee)
**Status**: Phase 1 Deliverable
