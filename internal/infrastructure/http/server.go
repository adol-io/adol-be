package http

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/nicklaros/adol/internal/infrastructure/config"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
	"github.com/nicklaros/adol/pkg/monitoring"
)

// Server represents the HTTP server
type Server struct {
	config  *config.Config
	db      *sql.DB
	logger  logger.EnhancedLogger
	router  *gin.Engine
	server  *http.Server
	metrics *monitoring.MetricsCollector
	health  *monitoring.HealthChecker
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *sql.DB, baseLogger logger.Logger) *Server {
	// Convert to enhanced logger
	enhancedLogger := logger.NewEnhancedLogger(
		logger.LogLevel(cfg.Logger.Level),
		logger.LogFormat(cfg.Logger.Format),
	)

	// Set Gin mode based on environment
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create metrics collector and health checker
	metricsCollector := monitoring.NewMetricsCollector(enhancedLogger)
	healthChecker := monitoring.NewHealthChecker(enhancedLogger)

	server := &Server{
		config:  cfg,
		db:      db,
		logger:  enhancedLogger,
		router:  router,
		metrics: metricsCollector,
		health:  healthChecker,
	}

	// Add enhanced middleware
	router.Use(gin.Recovery())
	router.Use(server.ErrorHandlingMiddleware())
	router.Use(server.RequestTrackingMiddleware())
	router.Use(server.SecurityHeadersMiddleware())
	router.Use(corsMiddleware())
	router.Use(server.RateLimitingMiddleware())

	// Register health checks
	server.registerHealthChecks()

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.server = &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start metrics collection
	go server.startMetricsCollection()

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server on port " + s.config.Server.Port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// setupRoutes sets up all the routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/health/detailed", s.detailedHealthCheck)
	s.router.GET("/metrics", s.metricsEndpoint)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", s.logout)
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(s.authMiddleware())
		{
			// User management routes
			users := protected.Group("/users")
			{
				users.GET("", s.listUsers)
				users.POST("", s.createUser)
				users.GET("/:id", s.getUser)
				users.PUT("/:id", s.updateUser)
				users.DELETE("/:id", s.deleteUser)
				users.PUT("/:id/activate", s.activateUser)
				users.PUT("/:id/deactivate", s.deactivateUser)
				users.PUT("/:id/suspend", s.suspendUser)
				users.PUT("/change-password", s.changePassword)
				users.PUT("/:id/reset-password", s.resetPassword)
			}

			// Product management routes
			products := protected.Group("/products")
			{
				products.GET("", s.listProducts)
				products.POST("", s.createProduct)
				products.GET("/:id", s.getProduct)
				products.PUT("/:id", s.updateProduct)
				products.DELETE("/:id", s.deleteProduct)
				products.GET("/categories", s.getCategories)
				products.GET("/low-stock", s.getLowStockProducts)
				products.GET("/sku/:sku", s.getProductBySKU)
			}

			// Stock management routes
			stock := protected.Group("/stock")
			{
				stock.GET("", s.listStock)
				stock.GET("/:productId", s.getStock)
				stock.POST("/adjust", s.adjustStock)
				stock.POST("/reserve", s.reserveStock)
				stock.POST("/release", s.releaseReservedStock)
				stock.GET("/low-stock", s.getLowStockItems)
				stock.GET("/movements", s.getStockMovements)
				stock.GET("/movements/:productId", s.getProductStockMovements)
			}

			// Sales management routes
			sales := protected.Group("/sales")
			{
				sales.GET("", s.listSales)
				sales.POST("", s.createSale)
				sales.GET("/:id", s.getSale)
				sales.PUT("/:id/cancel", s.cancelSale)
				sales.POST("/:id/items", s.addSaleItem)
				sales.PUT("/:id/items", s.updateSaleItem)
				sales.DELETE("/:id/items/:productId", s.removeSaleItem)
				sales.POST("/:id/complete", s.completeSale)
				sales.GET("/number/:saleNumber", s.getSaleBySaleNumber)
			}

			// Invoice management routes
			invoices := protected.Group("/invoices")
			{
				invoices.GET("", s.listInvoices)
				invoices.POST("", s.createInvoice)
				invoices.GET("/:id", s.getInvoice)
				invoices.PUT("/:id/paid", s.markInvoiceAsPaid)
				invoices.PUT("/:id/cancel", s.cancelInvoice)
				invoices.GET("/:id/pdf", s.generateInvoicePDF)
				invoices.GET("/:id/preview", s.getInvoicePreview)
				invoices.POST("/:id/email", s.sendInvoiceEmail)
				invoices.POST("/:id/print", s.printInvoice)
				invoices.GET("/number/:invoiceNumber", s.getInvoiceByNumber)
				invoices.GET("/overdue", s.getOverdueInvoices)
				invoices.GET("/templates", s.getInvoiceTemplates)
				invoices.GET("/paper-sizes", s.getPaperSizes)
				invoices.GET("/printers", s.getAvailablePrinters)
			}

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.GET("/sales", s.getSalesReport)
				reports.GET("/sales/daily", s.getDailySalesReport)
				reports.GET("/invoices", s.getInvoiceReport)
				reports.GET("/products/top-selling", s.getTopSellingProducts)
			}
		}
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	// Simple health check - just verify database connection
	if err := s.db.Ping(); err != nil {
		s.RespondWithError(c, errors.NewInternalError("Database connection failed", err))
		return
	}

	s.RespondWithSuccess(c, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "adol-pos-api",
		"version":   "1.0.0",
	}, "Service is healthy")
}

// detailedHealthCheck provides detailed health information
func (s *Server) detailedHealthCheck(c *gin.Context) {
	health := s.health.RunChecks()
	overallStatus := s.health.GetOverallHealth()

	response := gin.H{
		"status":         string(overallStatus),
		"timestamp":      time.Now().UTC(),
		"service":        "adol-pos-api",
		"version":        "1.0.0",
		"checks":         health,
		"overall_status": overallStatus,
	}

	statusCode := http.StatusOK
	if overallStatus == monitoring.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == monitoring.HealthStatusDegraded {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, response)
}

// metricsEndpoint provides application metrics
func (s *Server) metricsEndpoint(c *gin.Context) {
	metrics := s.metrics.GetAllMetrics()
	s.RespondWithSuccess(c, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now().UTC(),
	}, "Metrics retrieved successfully")
}

// registerHealthChecks registers various health checks
func (s *Server) registerHealthChecks() {
	// Database health check
	s.health.RegisterCheck("database", func() monitoring.HealthCheck {
		err := s.db.Ping()
		if err != nil {
			return monitoring.HealthCheck{
				Name:    "database",
				Status:  monitoring.HealthStatusUnhealthy,
				Message: "Database connection failed: " + err.Error(),
			}
		}
		return monitoring.HealthCheck{
			Name:    "database",
			Status:  monitoring.HealthStatusHealthy,
			Message: "Database connection is healthy",
		}
	})

	// Memory health check
	s.health.RegisterCheck("memory", func() monitoring.HealthCheck {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Consider memory unhealthy if using more than 1GB
		memoryUsageMB := float64(m.Alloc) / 1024 / 1024
		if memoryUsageMB > 1024 {
			return monitoring.HealthCheck{
				Name:    "memory",
				Status:  monitoring.HealthStatusDegraded,
				Message: fmt.Sprintf("High memory usage: %.2f MB", memoryUsageMB),
				Details: map[string]interface{}{
					"memory_usage_mb": memoryUsageMB,
					"goroutines":      runtime.NumGoroutine(),
				},
			}
		}

		return monitoring.HealthCheck{
			Name:    "memory",
			Status:  monitoring.HealthStatusHealthy,
			Message: fmt.Sprintf("Memory usage is normal: %.2f MB", memoryUsageMB),
			Details: map[string]interface{}{
				"memory_usage_mb": memoryUsageMB,
				"goroutines":      runtime.NumGoroutine(),
			},
		}
	})
}

// startMetricsCollection starts background metrics collection
func (s *Server) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			s.metrics.RecordSystemMetrics()
		}
	}()
}

// corsMiddleware sets up CORS middleware
func corsMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.ExposeHeaders = []string{"X-Request-ID"}
	return cors.New(config)
}