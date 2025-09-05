package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/infrastructure/monitoring"
	"github.com/nicklaros/adol/pkg/logger"
)

// TenantLoggingMiddleware handles tenant-aware logging
type TenantLoggingMiddleware struct {
	logger  logger.EnhancedLogger
	monitor monitoring.TenantMonitor
}

// NewTenantLoggingMiddleware creates a new tenant logging middleware
func NewTenantLoggingMiddleware(logger logger.EnhancedLogger, monitor monitoring.TenantMonitor) *TenantLoggingMiddleware {
	return &TenantLoggingMiddleware{
		logger:  logger,
		monitor: monitor,
	}
}

// TenantAwareLogger returns a middleware that adds tenant context to all logs
func (tlm *TenantLoggingMiddleware) TenantAwareLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// Get tenant context if available
		tenantContext := GetTenantContext(c)
		var tenantID uuid.UUID
		var tenantSlug string
		
		if tenantContext != nil {
			tenantID = tenantContext.TenantID
			tenantSlug = tenantContext.TenantSlug
		}

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Create enriched context
		ctx := c.Request.Context()
		ctx = logger.AddRequestIDToContext(ctx, requestID)
		if tenantContext != nil {
			ctx = logger.AddTenantToContext(ctx, tenantID.String(), tenantSlug)
		}
		c.Request = c.Request.WithContext(ctx)

		// Create tenant-aware logger
		tenantLogger := tlm.logger.WithContext(ctx).WithFields(map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
			"ip":         c.ClientIP(),
		})

		// Store logger in context for use by handlers
		c.Set("logger", tenantLogger)

		// Log request start
		if tenantContext != nil {
			tenantLogger.Info("Request started")
		} else {
			tenantLogger.Debug("Request started (no tenant context)")
		}

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(startTime)
		
		// Log request completion
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 400
		
		logFields := map[string]interface{}{
			"status_code":   statusCode,
			"response_time": duration.Milliseconds(),
			"success":       success,
		}

		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.String()
		}

		requestLogger := tenantLogger.WithFields(logFields)
		
		if success {
			requestLogger.Info("Request completed successfully")
		} else {
			requestLogger.Warn("Request completed with error")
		}

		// Track performance metrics if tenant context is available
		if tenantContext != nil {
			operation := c.Request.Method + " " + c.FullPath()
			tlm.monitor.TrackResponse(ctx, tenantID, operation, duration, success)
			
			// Track API usage
			if err := tlm.monitor.TrackUsage(ctx, tenantID, "api_requests", 1); err != nil {
				tlm.logger.WithFields(map[string]interface{}{
					"tenant_id": tenantID.String(),
					"error":     err.Error(),
				}).Error("Failed to track API usage")
			}
		}
	}
}

// TenantUsageTracker returns a middleware that tracks specific resource usage
func (tlm *TenantLoggingMiddleware) TenantUsageTracker(resource string, amount int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant context
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.Next()
			return
		}

		// Track usage before processing request
		ctx := c.Request.Context()
		if err := tlm.monitor.TrackUsage(ctx, tenantContext.TenantID, resource, amount); err != nil {
			tlm.logger.WithFields(map[string]interface{}{
				"tenant_id": tenantContext.TenantID.String(),
				"resource":  resource,
				"amount":    amount,
				"error":     err.Error(),
			}).Error("Failed to track resource usage")
		}

		c.Next()
	}
}

// AuditLogger returns a middleware that logs audit events
func (tlm *TenantLoggingMiddleware) AuditLogger(action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Log audit event after processing
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			return
		}

		// Get user ID from context (assuming it's set by auth middleware)
		userID := getUserIDFromContext(c)
		
		// Prepare audit details
		details := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"ip":          c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		}

		// Add request/response details if relevant
		if c.Request.Method != "GET" {
			if bodySize := c.Request.ContentLength; bodySize > 0 {
				details["request_size"] = bodySize
			}
		}

		// Log audit event
		tlm.logger.LogTenantAudit(
			tenantContext.TenantID.String(),
			action,
			resource,
			userID,
			details,
		)
	}
}

// SecurityLogger returns a middleware that logs security events
func (tlm *TenantLoggingMiddleware) SecurityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for security-relevant events
		statusCode := c.Writer.Status()
		tenantContext := GetTenantContext(c)
		
		// Log authentication failures
		if statusCode == 401 {
			metadata := map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}

			if tenantContext != nil {
				tlm.logger.LogTenantSecurity(
					tenantContext.TenantID.String(),
					"authentication_failure",
					"Unauthorized access attempt",
					nil,
					metadata,
				)
			} else {
				tlm.logger.LogSecurity(
					"authentication_failure",
					"Unauthorized access attempt without tenant context",
					nil,
					metadata,
				)
			}
		}

		// Log authorization failures
		if statusCode == 403 {
			userID := getUserIDFromContext(c)
			metadata := map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}

			if tenantContext != nil {
				tlm.logger.LogTenantSecurity(
					tenantContext.TenantID.String(),
					"authorization_failure",
					"Access denied to protected resource",
					userID,
					metadata,
				)
			}
		}

		// Log suspicious activity (too many requests from same IP)
		if shouldLogSuspiciousActivity(c) {
			if tenantContext != nil {
				tlm.logger.LogTenantSecurity(
					tenantContext.TenantID.String(),
					"suspicious_activity",
					"High request rate detected",
					nil,
					map[string]interface{}{
						"ip":         c.ClientIP(),
						"user_agent": c.Request.UserAgent(),
						"path":       c.Request.URL.Path,
					},
				)
			}
		}
	}
}

// HealthCheckLogger returns a middleware for health check logging
func (tlm *TenantLoggingMiddleware) HealthCheckLogger(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		c.Next()

		duration := time.Since(startTime)
		success := c.Writer.Status() >= 200 && c.Writer.Status() < 400

		// Log health check for each tenant if this is a tenant-specific endpoint
		tenantContext := GetTenantContext(c)
		if tenantContext != nil {
			tlm.monitor.RecordHealthCheck(
				c.Request.Context(),
				tenantContext.TenantID,
				service,
				success,
				duration,
			)
		}

		// Always log general health check
		tlm.logger.LogHealthCheck(service, success, duration, getHealthCheckMessage(c))
	}
}

// Helper functions

func getUserIDFromContext(c *gin.Context) interface{} {
	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*entities.User); ok {
			return user.ID.String()
		}
	}
	return nil
}

func shouldLogSuspiciousActivity(c *gin.Context) bool {
	// This is a simple implementation. In production, you might want to use
	// a more sophisticated rate limiting or anomaly detection system
	// For now, we'll just return false to avoid false positives
	return false
}

func getHealthCheckMessage(c *gin.Context) string {
	if len(c.Errors) > 0 {
		return c.Errors.String()
	}
	
	statusCode := c.Writer.Status()
	if statusCode >= 200 && statusCode < 300 {
		return "Service is healthy"
	} else if statusCode >= 300 && statusCode < 400 {
		return "Service responded with redirect"
	} else if statusCode >= 400 && statusCode < 500 {
		return "Service responded with client error"
	} else {
		return "Service responded with server error"
	}
}

// GetLoggerFromContext retrieves the tenant-aware logger from gin context
func GetLoggerFromContext(c *gin.Context) logger.Logger {
	if loggerInterface, exists := c.Get("logger"); exists {
		if logger, ok := loggerInterface.(logger.Logger); ok {
			return logger
		}
	}
	// Fallback to basic logger
	return logger.NewLogger()
}