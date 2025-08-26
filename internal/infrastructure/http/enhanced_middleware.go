package http

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/pkg/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorInfo `json:"error"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
	Meta      interface{} `json:"meta,omitempty"`
}

// RequestTrackingMiddleware adds request tracking and correlation ID
func (s *Server) RequestTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logLevel := "info"
		if statusCode >= 500 {
			logLevel = "error"
		} else if statusCode >= 400 {
			logLevel = "warn"
		}

		logFields := map[string]interface{}{
			"request_id":  requestID,
			"method":      c.Request.Method,
			"url":         c.Request.URL.String(),
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"ip":          c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		}

		logMessage := fmt.Sprintf("%s %s - %d (%dms)", c.Request.Method, c.Request.URL.String(), statusCode, duration.Milliseconds())

		switch logLevel {
		case "error":
			s.logger.WithFields(logFields).Error(logMessage)
		case "warn":
			s.logger.WithFields(logFields).Warn(logMessage)
		default:
			s.logger.WithFields(logFields).Info(logMessage)
		}
	}
}

// ErrorHandlingMiddleware provides comprehensive error handling
func (s *Server) ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log panic with stack trace
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)

				s.logger.WithFields(map[string]interface{}{
					"panic":       r,
					"stack_trace": string(stack[:length]),
					"request_id":  s.getRequestID(c),
					"method":      c.Request.Method,
					"url":         c.Request.URL.String(),
				}).Error("Panic recovered")

				// Respond with internal server error
				s.RespondWithError(c, errors.NewInternalError("Internal server error", fmt.Errorf("%v", r)))
			}
		}()

		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			s.RespondWithError(c, lastError.Err)
		}
	}
}

// SecurityHeadersMiddleware adds security headers
func (s *Server) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// RateLimitingMiddleware provides basic rate limiting
func (s *Server) RateLimitingMiddleware() gin.HandlerFunc {
	requests := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old requests (older than 1 minute)
		if reqs, exists := requests[clientIP]; exists {
			filtered := make([]time.Time, 0)
			for _, reqTime := range reqs {
				if now.Sub(reqTime) < time.Minute {
					filtered = append(filtered, reqTime)
				}
			}
			requests[clientIP] = filtered
		}

		// Check rate limit (100 requests per minute per IP)
		if len(requests[clientIP]) >= 100 {
			s.RespondWithError(c, errors.NewAppError(errors.ErrorTypeRateLimit, "Rate limit exceeded", nil))
			c.Abort()
			return
		}

		// Add current request
		requests[clientIP] = append(requests[clientIP], now)

		c.Next()
	}
}

// Enhanced response methods

// RespondWithError sends standardized error response
func (s *Server) RespondWithError(c *gin.Context, err error) {
	requestID := s.getRequestID(c)

	// Log the error with context
	s.logError(c, err, requestID)

	var errorResponse ErrorResponse

	if appErr, ok := errors.IsAppError(err); ok {
		errorResponse = ErrorResponse{
			Success: false,
			Error: ErrorInfo{
				Type:    string(appErr.Type),
				Message: appErr.Message,
				Code:    fmt.Sprintf("%d", appErr.Code),
			},
			RequestID: requestID,
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		}

		c.JSON(appErr.Code, errorResponse)
	} else {
		// Handle unknown errors
		errorResponse = ErrorResponse{
			Success: false,
			Error: ErrorInfo{
				Type:    string(errors.ErrorTypeInternal),
				Message: "Internal server error",
				Code:    "500",
			},
			RequestID: requestID,
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		}

		c.JSON(http.StatusInternalServerError, errorResponse)
	}
}

// RespondWithSuccess sends standardized success response
func (s *Server) RespondWithSuccess(c *gin.Context, data interface{}, message string) {
	requestID := s.getRequestID(c)

	response := SuccessResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// RespondWithSuccessAndMeta sends success response with metadata
func (s *Server) RespondWithSuccessAndMeta(c *gin.Context, data interface{}, meta interface{}, message string) {
	requestID := s.getRequestID(c)

	response := SuccessResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// getRequestID extracts request ID from context
func (s *Server) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// logError logs error with comprehensive context
func (s *Server) logError(c *gin.Context, err error, requestID string) {
	logFields := map[string]interface{}{
		"request_id": requestID,
		"method":     c.Request.Method,
		"url":        c.Request.URL.String(),
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"error":      err.Error(),
	}

	// Add user ID if available
	if userID, exists := c.Get("user_id"); exists {
		logFields["user_id"] = userID
	}

	// Add error type if it's an AppError
	if appErr, ok := errors.IsAppError(err); ok {
		logFields["error_type"] = appErr.Type
		logFields["error_code"] = appErr.Code
		if appErr.Internal != nil {
			logFields["internal_error"] = appErr.Internal.Error()
		}
	}

	s.logger.WithFields(logFields).Error("Error occurred during request processing")
}