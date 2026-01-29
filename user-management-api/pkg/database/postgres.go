package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Database wraps the PostgreSQL connection pool
type Database struct {
	Pool   *pgxpool.Pool
	logger zerolog.Logger
}

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConnections  int
	MinConnections  int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// New creates a new database connection pool
func New(ctx context.Context, cfg Config, logger zerolog.Logger) (*Database, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		cfg.MaxConnections,
		cfg.MinConnections,
	)

	// Parse connection config
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Set connection timeout
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	// Create connection pool with retries
	var pool *pgxpool.Pool
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			// Test connection
			if err = pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}

		if i < maxRetries-1 {
			logger.Warn().
				Err(err).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Msg("Failed to connect to database, retrying...")
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to connect to database after %d attempts: %w", maxRetries, err)
	}

	logger.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Database).
		Int("max_conns", cfg.MaxConnections).
		Int("min_conns", cfg.MinConnections).
		Msg("Database connection pool created successfully")

	return &Database{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.logger.Info().Msg("Closing database connection pool")
		db.Pool.Close()
	}
}

// Ping checks if the database is accessible
func (db *Database) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// HealthCheck performs a comprehensive health check
func (db *Database) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check connection
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Execute a simple query
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

// Stats returns connection pool statistics
func (db *Database) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// GetConnectionInfo returns human-readable connection pool info
func (db *Database) GetConnectionInfo() map[string]interface{} {
	stats := db.Pool.Stat()
	return map[string]interface{}{
		"total_conns":       stats.TotalConns(),
		"acquired_conns":    stats.AcquiredConns(),
		"idle_conns":        stats.IdleConns(),
		"max_conns":         stats.MaxConns(),
		"constructing_conns": stats.ConstructingConns(),
		"acquire_count":     stats.AcquireCount(),
		"empty_acquire_count": stats.EmptyAcquireCount(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
		"max_lifetime_destroy_count": stats.MaxLifetimeDestroyCount(),
		"max_idle_destroy_count": stats.MaxIdleDestroyCount(),
	}
}

// WaitForDatabase waits for the database to become available
func WaitForDatabase(ctx context.Context, db *Database, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database: %w", ctx.Err())
		case <-ticker.C:
			if err := db.Ping(ctx); err == nil {
				return nil
			}
		}
	}
}

// RunInTransaction executes a function within a database transaction
func (db *Database) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error().
				Err(rbErr).
				Msg("Failed to rollback transaction")
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IsHealthy returns true if the database is healthy
func (db *Database) IsHealthy(ctx context.Context) bool {
	return db.HealthCheck(ctx) == nil
}
