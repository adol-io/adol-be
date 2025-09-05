package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nicklaros/adol/internal/infrastructure/config"
	"github.com/nicklaros/adol/internal/infrastructure/database"
	httpInfra "github.com/nicklaros/adol/internal/infrastructure/http"
	"github.com/nicklaros/adol/pkg/logger"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	logger.Info("Starting ADOL POS Backend System")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgreSQL(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize HTTP server
	server := httpInfra.NewServer(cfg, db, logger)

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %s", cfg.Server.Port))
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
