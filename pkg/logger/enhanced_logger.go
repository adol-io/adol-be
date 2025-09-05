package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// ContextKey represents a key for context values
type ContextKey string

const (
	// ContextKeyRequestID is the key for request ID in context
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeyUserID is the key for user ID in context
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyOperation is the key for operation name in context
	ContextKeyOperation ContextKey = "operation"
	// ContextKeyTenantID is the key for tenant ID in context
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyTenantSlug is the key for tenant slug in context
	ContextKeyTenantSlug ContextKey = "tenant_slug"
)

// LogLevel represents different log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogFormat represents different log formats
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// EnhancedLogger extends the basic Logger interface with additional capabilities
type EnhancedLogger interface {
	Logger

	// Context-aware logging
	WithContext(ctx context.Context) Logger

	// Structured logging with tags
	WithTag(tag string) Logger
	WithTags(tags []string) Logger

	// Performance logging
	LogDuration(operation string, duration time.Duration)
	StartTimer(operation string) func()

	// Audit logging
	LogAudit(action, resource string, userID, details interface{})

	// Security logging
	LogSecurity(event, description string, userID interface{}, metadata map[string]interface{})

	// Business event logging
	LogBusinessEvent(event string, data interface{})

	// Set log level dynamically
	SetLevel(level LogLevel)

	// Tenant-aware logging
	LogTenantEvent(tenantID, event string, data interface{})
	LogTenantAudit(tenantID, action, resource string, userID, details interface{})
	LogTenantSecurity(tenantID, event, description string, userID interface{}, metadata map[string]interface{})
	LogTenantUsage(tenantID string, resource string, usage, limit int64)
	LogTenantSubscriptionEvent(tenantID, event string, subscriptionData interface{})

	// Health check logging
	LogHealthCheck(service string, status bool, responseTime time.Duration, details string)
}

// enhancedLogrusLogger implements EnhancedLogger interface
type enhancedLogrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
	tags   []string
}

// NewEnhancedLogger creates a new enhanced logger instance
func NewEnhancedLogger(level LogLevel, format LogFormat) EnhancedLogger {
	logger := logrus.New()

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Set formatter based on format
	switch format {
	case JSONFormat:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
	case TextFormat:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// Set log level
	switch level {
	case DebugLevel:
		logger.SetLevel(logrus.DebugLevel)
	case WarnLevel:
		logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Add default fields
	entry := logger.WithFields(logrus.Fields{
		"service":    "adol-pos",
		"version":    "1.0.0",
		"hostname":   getHostname(),
		"process_id": os.Getpid(),
	})

	return &enhancedLogrusLogger{
		logger: logger,
		entry:  entry,
		tags:   make([]string, 0),
	}
}

// Basic logging methods (implementing Logger interface)

func (l *enhancedLogrusLogger) Debug(args ...interface{}) {
	l.addCallerInfo().Debug(args...)
}

func (l *enhancedLogrusLogger) Info(args ...interface{}) {
	l.addCallerInfo().Info(args...)
}

func (l *enhancedLogrusLogger) Warn(args ...interface{}) {
	l.addCallerInfo().Warn(args...)
}

func (l *enhancedLogrusLogger) Error(args ...interface{}) {
	l.addCallerInfo().Error(args...)
}

func (l *enhancedLogrusLogger) Fatal(args ...interface{}) {
	l.addCallerInfo().Fatal(args...)
}

func (l *enhancedLogrusLogger) WithField(key string, value interface{}) Logger {
	return &enhancedLogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
		tags:   l.tags,
	}
}

func (l *enhancedLogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &enhancedLogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
		tags:   l.tags,
	}
}

// Enhanced logging methods

func (l *enhancedLogrusLogger) WithContext(ctx context.Context) Logger {
	fields := logrus.Fields{}

	if requestID := ctx.Value(ContextKeyRequestID); requestID != nil {
		fields["request_id"] = requestID
	}

	if userID := ctx.Value(ContextKeyUserID); userID != nil {
		fields["user_id"] = userID
	}

	if tenantID := ctx.Value(ContextKeyTenantID); tenantID != nil {
		fields["tenant_id"] = tenantID
	}

	if tenantSlug := ctx.Value(ContextKeyTenantSlug); tenantSlug != nil {
		fields["tenant_slug"] = tenantSlug
	}

	if operation := ctx.Value(ContextKeyOperation); operation != nil {
		fields["operation"] = operation
	}

	return &enhancedLogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
		tags:   l.tags,
	}
}

func (l *enhancedLogrusLogger) WithTag(tag string) Logger {
	newTags := append(l.tags, tag)
	return &enhancedLogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField("tags", newTags),
		tags:   newTags,
	}
}

func (l *enhancedLogrusLogger) WithTags(tags []string) Logger {
	newTags := append(l.tags, tags...)
	return &enhancedLogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField("tags", newTags),
		tags:   newTags,
	}
}

func (l *enhancedLogrusLogger) LogDuration(operation string, duration time.Duration) {
	l.entry.WithFields(logrus.Fields{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
		"duration_ns": duration.Nanoseconds(),
		"log_type":    "performance",
	}).Info(fmt.Sprintf("Operation '%s' completed in %v", operation, duration))
}

func (l *enhancedLogrusLogger) StartTimer(operation string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		l.LogDuration(operation, duration)
	}
}

func (l *enhancedLogrusLogger) LogAudit(action, resource string, userID, details interface{}) {
	l.entry.WithFields(logrus.Fields{
		"log_type":  "audit",
		"action":    action,
		"resource":  resource,
		"user_id":   userID,
		"details":   details,
		"timestamp": time.Now().UTC(),
	}).Info(fmt.Sprintf("Audit: %s performed on %s", action, resource))
}

func (l *enhancedLogrusLogger) LogSecurity(event, description string, userID interface{}, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"log_type":    "security",
		"event":       event,
		"description": description,
		"user_id":     userID,
		"timestamp":   time.Now().UTC(),
	}

	// Add metadata fields
	for key, value := range metadata {
		fields[key] = value
	}

	l.entry.WithFields(fields).Warn(fmt.Sprintf("Security Event: %s - %s", event, description))
}

func (l *enhancedLogrusLogger) LogBusinessEvent(event string, data interface{}) {
	l.entry.WithFields(logrus.Fields{
		"log_type":  "business_event",
		"event":     event,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}).Info(fmt.Sprintf("Business Event: %s", event))
}

func (l *enhancedLogrusLogger) SetLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		l.logger.SetLevel(logrus.DebugLevel)
	case WarnLevel:
		l.logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		l.logger.SetLevel(logrus.FatalLevel)
	default:
		l.logger.SetLevel(logrus.InfoLevel)
	}
}

func (l *enhancedLogrusLogger) LogHealthCheck(service string, status bool, responseTime time.Duration, details string) {
	statusStr := "healthy"
	logLevel := logrus.InfoLevel

	if !status {
		statusStr = "unhealthy"
		logLevel = logrus.ErrorLevel
	}

	l.entry.WithFields(logrus.Fields{
		"log_type":      "health_check",
		"service":       service,
		"status":        statusStr,
		"response_time": responseTime.Milliseconds(),
		"details":       details,
		"timestamp":     time.Now().UTC(),
	}).Log(logLevel, fmt.Sprintf("Health Check: %s is %s", service, statusStr))
}

// Tenant-specific logging methods

func (l *enhancedLogrusLogger) LogTenantEvent(tenantID, event string, data interface{}) {
	l.entry.WithFields(logrus.Fields{
		"log_type":  "tenant_event",
		"tenant_id": tenantID,
		"event":     event,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}).Info(fmt.Sprintf("Tenant Event [%s]: %s", tenantID, event))
}

func (l *enhancedLogrusLogger) LogTenantAudit(tenantID, action, resource string, userID, details interface{}) {
	l.entry.WithFields(logrus.Fields{
		"log_type":  "tenant_audit",
		"tenant_id": tenantID,
		"action":    action,
		"resource":  resource,
		"user_id":   userID,
		"details":   details,
		"timestamp": time.Now().UTC(),
	}).Info(fmt.Sprintf("Tenant Audit [%s]: %s performed on %s", tenantID, action, resource))
}

func (l *enhancedLogrusLogger) LogTenantSecurity(tenantID, event, description string, userID interface{}, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"log_type":    "tenant_security",
		"tenant_id":   tenantID,
		"event":       event,
		"description": description,
		"user_id":     userID,
		"timestamp":   time.Now().UTC(),
	}

	// Add metadata fields
	for key, value := range metadata {
		fields[key] = value
	}

	l.entry.WithFields(fields).Warn(fmt.Sprintf("Tenant Security [%s]: %s - %s", tenantID, event, description))
}

func (l *enhancedLogrusLogger) LogTenantUsage(tenantID string, resource string, usage, limit int64) {
	usagePercent := float64(usage) / float64(limit) * 100
	logLevel := logrus.InfoLevel

	// Log warning if usage is over 80%
	if usagePercent >= 80 {
		logLevel = logrus.WarnLevel
	}
	// Log error if usage is at or over limit
	if usage >= limit {
		logLevel = logrus.ErrorLevel
	}

	l.entry.WithFields(logrus.Fields{
		"log_type":       "tenant_usage",
		"tenant_id":      tenantID,
		"resource":       resource,
		"current_usage":  usage,
		"usage_limit":    limit,
		"usage_percent":  fmt.Sprintf("%.2f%%", usagePercent),
		"timestamp":      time.Now().UTC(),
	}).Log(logLevel, fmt.Sprintf("Tenant Usage [%s]: %s usage at %.2f%% (%d/%d)", tenantID, resource, usagePercent, usage, limit))
}

func (l *enhancedLogrusLogger) LogTenantSubscriptionEvent(tenantID, event string, subscriptionData interface{}) {
	l.entry.WithFields(logrus.Fields{
		"log_type":          "tenant_subscription",
		"tenant_id":         tenantID,
		"event":             event,
		"subscription_data": subscriptionData,
		"timestamp":         time.Now().UTC(),
	}).Info(fmt.Sprintf("Tenant Subscription [%s]: %s", tenantID, event))
}

// Helper methods

func (l *enhancedLogrusLogger) addCallerInfo() *logrus.Entry {
	// Get caller information
	pc, file, line, ok := runtime.Caller(2) // Skip 2 frames to get the actual caller
	if !ok {
		return l.entry
	}

	funcName := runtime.FuncForPC(pc).Name()

	return l.entry.WithFields(logrus.Fields{
		"caller": fmt.Sprintf("%s:%d", file, line),
		"func":   funcName,
	})
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// LoggerFromContext creates a logger with context information
func LoggerFromContext(ctx context.Context, baseLogger EnhancedLogger) Logger {
	return baseLogger.WithContext(ctx)
}

// AddRequestIDToContext adds request ID to context
func AddRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// AddUserIDToContext adds user ID to context
func AddUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// AddTenantToContext adds tenant information to context
func AddTenantToContext(ctx context.Context, tenantID, tenantSlug string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
	return context.WithValue(ctx, ContextKeyTenantSlug, tenantSlug)
}

// GetTenantFromContext retrieves tenant information from context
func GetTenantFromContext(ctx context.Context) (tenantID, tenantSlug string) {
	if id := ctx.Value(ContextKeyTenantID); id != nil {
		if idStr, ok := id.(string); ok {
			tenantID = idStr
		}
	}
	if slug := ctx.Value(ContextKeyTenantSlug); slug != nil {
		if slugStr, ok := slug.(string); ok {
			tenantSlug = slugStr
		}
	}
	return tenantID, tenantSlug
}

// AddOperationToContext adds operation name to context
func AddOperationToContext(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, ContextKeyOperation, operation)
}
