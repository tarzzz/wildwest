package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/user-management-api/internal/config"
	"github.com/example/user-management-api/internal/handler"
	"github.com/example/user-management-api/pkg/database"
	"github.com/example/user-management-api/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Create root context
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		OutputPath: cfg.Log.OutputPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.SetGlobalLogger()

	log.Infof("Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize database connection
	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxConnections:  cfg.Database.MaxConnections,
		MinConnections:  cfg.Database.MinConnections,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxConnIdleTime: cfg.Database.MaxConnIdleTime,
	}

	db, err := database.New(ctx, dbCfg, log.GetZerolog())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Info("Database connection pool initialized successfully")

	// Initialize Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(requestIDMiddleware(log))
	router.Use(loggerMiddleware(log))
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg))

	// Health check endpoints (no authentication required)
	healthHandler := handler.NewHealthHandler(db, cfg.App.Version)
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Ready)
	router.GET("/metrics", healthHandler.Metrics)

	// API v1 routes
	v1 := router.Group("/api/v1")
	_ = v1 // Will be used when auth and user handlers are implemented
	{
		// Auth endpoints (public)
		// auth := v1.Group("/auth")
		// {
		//     auth.POST("/register", authHandler.Register)
		//     auth.POST("/login", authHandler.Login)
		//     auth.POST("/refresh", authHandler.RefreshToken)
		//     auth.POST("/logout", authHandler.Logout)
		//     auth.POST("/forgot-password", authHandler.ForgotPassword)
		//     auth.POST("/reset-password", authHandler.ResetPassword)
		//     auth.POST("/verify-email", authHandler.VerifyEmail)
		// }

		// User endpoints (protected)
		// users := v1.Group("/users")
		// users.Use(authMiddleware())
		// {
		//     users.GET("", userHandler.ListUsers)          // Admin only
		//     users.GET("/me", userHandler.GetCurrentUser)  // All authenticated users
		//     users.GET("/:id", userHandler.GetUser)        // Owner or admin
		//     users.POST("", userHandler.CreateUser)        // Admin only
		//     users.PUT("/:id", userHandler.UpdateUser)     // Owner or admin
		//     users.PATCH("/:id", userHandler.PatchUser)    // Owner or admin
		//     users.DELETE("/:id", userHandler.DeleteUser)  // Owner or admin
		//     users.PATCH("/:id/change-password", userHandler.ChangePassword) // Owner only
		// }
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":      "NOT_FOUND",
				"message":   "The requested endpoint does not exist",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Starting HTTP server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server shutdown complete")
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID exists in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context
		c.Set("request_id", requestID)

		// Set request ID in response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// loggerMiddleware logs HTTP requests
func loggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Build path with query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		duration := time.Since(start)
		requestID, _ := c.Get("request_id")
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		requestLog := log.WithFields(map[string]interface{}{
			"request_id":  requestID,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"status":      statusCode,
			"duration_ms": duration.Milliseconds(),
			"user_agent":  userAgent,
		})

		// Log at different levels based on status code
		if statusCode >= 500 {
			requestLog.Error("HTTP request")
		} else if statusCode >= 400 {
			requestLog.Warn("HTTP request")
		} else {
			requestLog.Info("HTTP request")
		}
	}
}

// corsMiddleware configures CORS
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	if cfg.App.Environment == "development" {
		// Allow all origins in development
		corsConfig.AllowAllOrigins = true
	} else {
		// Configure specific origins in production
		corsConfig.AllowOrigins = []string{
			// Add your production domains here
			// "https://example.com",
		}
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Request-ID",
	}
	corsConfig.ExposeHeaders = []string{
		"Content-Length",
		"X-Request-ID",
	}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

	return cors.New(corsConfig)
}
