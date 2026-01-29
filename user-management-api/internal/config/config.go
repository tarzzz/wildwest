package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	RateLimit RateLimitConfig
	Log      LogConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string
	Environment string // development, staging, production
	Version     string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
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

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret                string
	AccessTokenDuration   time.Duration
	RefreshTokenDuration  time.Duration
	Issuer                string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	AuthRequestsPerMinute int
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json, pretty
	OutputPath string // stdout, stderr, or file path
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Load from .env file if it exists
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("..")
	v.AddConfigPath("../..")

	// Read config file (optional, won't error if not found)
	_ = v.ReadInConfig()

	// Override with environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into Config struct
	var cfg Config

	cfg.App = AppConfig{
		Name:        v.GetString("app.name"),
		Environment: v.GetString("app.environment"),
		Version:     v.GetString("app.version"),
	}

	cfg.Server = ServerConfig{
		Port:            v.GetInt("server.port"),
		ReadTimeout:     v.GetDuration("server.read_timeout"),
		WriteTimeout:    v.GetDuration("server.write_timeout"),
		ShutdownTimeout: v.GetDuration("server.shutdown_timeout"),
		TrustedProxies:  v.GetStringSlice("server.trusted_proxies"),
	}

	cfg.Database = DatabaseConfig{
		Host:            v.GetString("database.host"),
		Port:            v.GetInt("database.port"),
		User:            v.GetString("database.user"),
		Password:        v.GetString("database.password"),
		Database:        v.GetString("database.database"),
		SSLMode:         v.GetString("database.sslmode"),
		MaxConnections:  v.GetInt("database.max_connections"),
		MinConnections:  v.GetInt("database.min_connections"),
		MaxConnLifetime: v.GetDuration("database.max_conn_lifetime"),
		MaxConnIdleTime: v.GetDuration("database.max_conn_idle_time"),
	}

	cfg.JWT = JWTConfig{
		Secret:               v.GetString("jwt.secret"),
		AccessTokenDuration:  v.GetDuration("jwt.access_token_duration"),
		RefreshTokenDuration: v.GetDuration("jwt.refresh_token_duration"),
		Issuer:               v.GetString("jwt.issuer"),
	}

	cfg.RateLimit = RateLimitConfig{
		RequestsPerMinute:     v.GetInt("ratelimit.requests_per_minute"),
		AuthRequestsPerMinute: v.GetInt("ratelimit.auth_requests_per_minute"),
	}

	cfg.Log = LogConfig{
		Level:      v.GetString("log.level"),
		Format:     v.GetString("log.format"),
		OutputPath: v.GetString("log.output_path"),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "user-management-api")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.version", "1.0.0")

	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 30*time.Second)
	v.SetDefault("server.trusted_proxies", []string{})

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "userapi")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.min_connections", 5)
	v.SetDefault("database.max_conn_lifetime", 1*time.Hour)
	v.SetDefault("database.max_conn_idle_time", 10*time.Minute)

	// JWT defaults
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_token_duration", 15*time.Minute)
	v.SetDefault("jwt.refresh_token_duration", 7*24*time.Hour)
	v.SetDefault("jwt.issuer", "user-management-api")

	// Rate limit defaults
	v.SetDefault("ratelimit.requests_per_minute", 100)
	v.SetDefault("ratelimit.auth_requests_per_minute", 10)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_path", "stdout")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate app config
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if c.Database.MaxConnections < 1 {
		return fmt.Errorf("database.max_connections must be at least 1")
	}
	if c.Database.MinConnections < 0 {
		return fmt.Errorf("database.min_connections must be non-negative")
	}
	if c.Database.MinConnections > c.Database.MaxConnections {
		return fmt.Errorf("database.min_connections cannot exceed max_connections")
	}

	// Validate JWT config
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if c.JWT.Secret == "change-me-in-production" && c.App.Environment == "production" {
		return fmt.Errorf("jwt.secret must be changed in production")
	}
	if c.JWT.AccessTokenDuration <= 0 {
		return fmt.Errorf("jwt.access_token_duration must be positive")
	}
	if c.JWT.RefreshTokenDuration <= 0 {
		return fmt.Errorf("jwt.refresh_token_duration must be positive")
	}

	// Validate log config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("log.level must be one of: debug, info, warn, error")
	}
	validFormats := map[string]bool{"json": true, "pretty": true}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("log.format must be one of: json, pretty")
	}

	return nil
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// IsDevelopment returns true if the app is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the app is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
