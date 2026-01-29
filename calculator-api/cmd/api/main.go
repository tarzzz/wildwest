package main

import (
	"calculator-api/internal/handler"
	"calculator-api/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create service and handler instances
	calculator := service.NewCalculator()
	calcHandler := handler.NewCalculatorHandler(calculator)
	healthHandler := handler.NewHealthHandler()

	// Initialize Gin router
	router := gin.Default()

	// Register routes
	router.GET("/health", healthHandler.Health)
	router.POST("/api/v1/calculate", calcHandler.Calculate)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting calculator API server on port 8081...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
