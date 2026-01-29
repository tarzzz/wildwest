package handler

import (
	"net/http"
	"time"

	"github.com/example/user-management-api/pkg/database"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db        *database.Database
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.Database, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Health handles GET /health - Basic health check
// @Summary Basic health check
// @Description Returns basic health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "user-management-api",
		Version:   h.version,
		Uptime:    uptime.String(),
	}

	c.JSON(http.StatusOK, response)
}

// Ready handles GET /health/ready - Readiness check
// @Summary Readiness check
// @Description Returns readiness status including database connectivity
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	uptime := time.Since(h.startTime)
	details := make(map[string]interface{})

	// Check database connectivity
	dbHealthy := false
	if h.db != nil {
		if err := h.db.Ping(c.Request.Context()); err == nil {
			dbHealthy = true
			details["database"] = map[string]interface{}{
				"status":      "healthy",
				"connections": h.db.GetConnectionInfo(),
			}
		} else {
			details["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		}
	} else {
		details["database"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// Determine overall status
	status := "ready"
	httpStatus := http.StatusOK

	if !dbHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "user-management-api",
		Version:   h.version,
		Uptime:    uptime.String(),
		Details:   details,
	}

	c.JSON(httpStatus, response)
}

// MetricsResponse represents the metrics response
type MetricsResponse struct {
	Timestamp  string                 `json:"timestamp"`
	Service    string                 `json:"service"`
	Version    string                 `json:"version"`
	Uptime     string                 `json:"uptime"`
	UptimeMS   int64                  `json:"uptime_ms"`
	Metrics    map[string]interface{} `json:"metrics"`
}

// Metrics handles GET /metrics - Basic metrics
// @Summary Application metrics
// @Description Returns basic application metrics
// @Tags health
// @Produce json
// @Success 200 {object} MetricsResponse
// @Router /metrics [get]
func (h *HealthHandler) Metrics(c *gin.Context) {
	uptime := time.Since(h.startTime)
	metrics := make(map[string]interface{})

	// Database metrics
	if h.db != nil {
		stats := h.db.Stats()
		metrics["database"] = map[string]interface{}{
			"total_conns":                  stats.TotalConns(),
			"acquired_conns":               stats.AcquiredConns(),
			"idle_conns":                   stats.IdleConns(),
			"max_conns":                    stats.MaxConns(),
			"constructing_conns":           stats.ConstructingConns(),
			"acquire_count":                stats.AcquireCount(),
			"empty_acquire_count":          stats.EmptyAcquireCount(),
			"canceled_acquire_count":       stats.CanceledAcquireCount(),
			"max_lifetime_destroy_count":   stats.MaxLifetimeDestroyCount(),
			"max_idle_destroy_count":       stats.MaxIdleDestroyCount(),
		}
	}

	// System metrics (basic)
	metrics["system"] = map[string]interface{}{
		"uptime_seconds": int64(uptime.Seconds()),
	}

	response := MetricsResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "user-management-api",
		Version:   h.version,
		Uptime:    uptime.String(),
		UptimeMS:  uptime.Milliseconds(),
		Metrics:   metrics,
	}

	c.JSON(http.StatusOK, response)
}
