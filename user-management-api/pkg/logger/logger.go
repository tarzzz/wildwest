package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, pretty
	OutputPath string // stdout, stderr, or file path
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Set log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zerolog.SetGlobalLevel(level)

	// Set output writer
	var output io.Writer
	switch strings.ToLower(cfg.OutputPath) {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// File output
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Set format
	if strings.ToLower(cfg.Format) == "pretty" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create logger with common fields
	logger := zerolog.New(output).With().
		Timestamp().
		Str("service", "user-management-api").
		Logger()

	return &Logger{logger: logger}, nil
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info", "":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// WithRequestID returns a logger with request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithError returns a logger with error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// GetZerolog returns the underlying zerolog.Logger
// Useful for integration with libraries that accept zerolog
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

// SetGlobalLogger sets the global logger for the log package
func (l *Logger) SetGlobalLogger() {
	log.Logger = l.logger
}

// Default creates a logger with default settings (info level, JSON format, stdout)
func Default() *Logger {
	logger, _ := New(Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	})
	return logger
}

// HTTPAccessLog logs HTTP access in a structured format
func (l *Logger) HTTPAccessLog(method, path string, status int, duration time.Duration, clientIP, userAgent string) {
	l.logger.Info().
		Str("type", "http_access").
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration_ms", duration).
		Str("client_ip", clientIP).
		Str("user_agent", userAgent).
		Msg("HTTP request")
}

// DatabaseLog logs database operations
func (l *Logger) DatabaseLog(operation, query string, duration time.Duration, err error) {
	event := l.logger.Info()
	if err != nil {
		event = l.logger.Error().Err(err)
	}

	event.
		Str("type", "database").
		Str("operation", operation).
		Str("query", query).
		Dur("duration_ms", duration).
		Msg("Database operation")
}

// AuthLog logs authentication events
func (l *Logger) AuthLog(event, userID, email string, success bool) {
	level := l.logger.Info()
	if !success {
		level = l.logger.Warn()
	}

	level.
		Str("type", "auth").
		Str("event", event).
		Str("user_id", userID).
		Str("email", email).
		Bool("success", success).
		Msg("Authentication event")
}
