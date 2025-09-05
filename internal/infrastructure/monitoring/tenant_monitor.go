package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/logger"
)

// TenantMonitor handles tenant-specific monitoring and metrics
type TenantMonitor interface {
	// Usage tracking
	TrackUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int64) error
	GetUsage(ctx context.Context, tenantID uuid.UUID, resource string) (*UsageMetrics, error)
	
	// Performance monitoring
	TrackResponse(ctx context.Context, tenantID uuid.UUID, operation string, duration time.Duration, success bool)
	GetPerformanceMetrics(ctx context.Context, tenantID uuid.UUID) (*PerformanceMetrics, error)
	
	// Subscription monitoring
	TrackSubscriptionEvent(ctx context.Context, tenantID uuid.UUID, event string, data interface{})
	CheckSubscriptionLimits(ctx context.Context, tenantID uuid.UUID) (*LimitStatus, error)
	
	// Health monitoring
	RecordHealthCheck(ctx context.Context, tenantID uuid.UUID, service string, status bool, responseTime time.Duration)
	GetTenantHealth(ctx context.Context, tenantID uuid.UUID) (*TenantHealth, error)
	
	// Alert management
	CheckAlerts(ctx context.Context, tenantID uuid.UUID) ([]*Alert, error)
	CreateAlert(ctx context.Context, alert *Alert) error
}

// UsageMetrics represents usage statistics for a tenant
type UsageMetrics struct {
	TenantID     uuid.UUID         `json:"tenant_id"`
	Resource     string            `json:"resource"`
	CurrentUsage int64             `json:"current_usage"`
	Limit        int64             `json:"limit"`
	ResetDate    time.Time         `json:"reset_date"`
	History      []UsageDataPoint  `json:"history"`
}

// UsageDataPoint represents a single usage measurement
type UsageDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     int64     `json:"usage"`
}

// PerformanceMetrics represents performance statistics for a tenant
type PerformanceMetrics struct {
	TenantID         uuid.UUID                    `json:"tenant_id"`
	AverageResponse  time.Duration                `json:"average_response"`
	SuccessRate      float64                      `json:"success_rate"`
	RequestCount     int64                        `json:"request_count"`
	ErrorCount       int64                        `json:"error_count"`
	OperationMetrics map[string]*OperationMetrics `json:"operation_metrics"`
	LastUpdated      time.Time                    `json:"last_updated"`
}

// OperationMetrics represents metrics for a specific operation
type OperationMetrics struct {
	Operation       string        `json:"operation"`
	AverageResponse time.Duration `json:"average_response"`
	SuccessRate     float64       `json:"success_rate"`
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	P50Response     time.Duration `json:"p50_response"`
	P95Response     time.Duration `json:"p95_response"`
	P99Response     time.Duration `json:"p99_response"`
}

// LimitStatus represents the current status against subscription limits
type LimitStatus struct {
	TenantID   uuid.UUID                 `json:"tenant_id"`
	PlanType   entities.SubscriptionPlanType `json:"plan_type"`
	Limits     map[string]*ResourceLimit `json:"limits"`
	IsOverLimit bool                     `json:"is_over_limit"`
	Warnings   []string                  `json:"warnings"`
}

// ResourceLimit represents limit status for a specific resource
type ResourceLimit struct {
	Resource     string    `json:"resource"`
	CurrentUsage int64     `json:"current_usage"`
	Limit        int64     `json:"limit"`
	Percentage   float64   `json:"percentage"`
	IsExceeded   bool      `json:"is_exceeded"`
	IsWarning    bool      `json:"is_warning"` // Above 80%
	LastUpdated  time.Time `json:"last_updated"`
}

// TenantHealth represents the overall health status of a tenant
type TenantHealth struct {
	TenantID       uuid.UUID                  `json:"tenant_id"`
	OverallStatus  HealthStatus               `json:"overall_status"`
	Services       map[string]*ServiceHealth  `json:"services"`
	LastChecked    time.Time                  `json:"last_checked"`
	Issues         []string                   `json:"issues"`
}

// ServiceHealth represents health status of a specific service for a tenant
type ServiceHealth struct {
	Service       string        `json:"service"`
	Status        HealthStatus  `json:"status"`
	ResponseTime  time.Duration `json:"response_time"`
	LastCheck     time.Time     `json:"last_check"`
	ErrorCount    int64         `json:"error_count"`
	Message       string        `json:"message"`
}

// HealthStatus represents different health statuses
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusWarning   HealthStatus = "warning"
	HealthStatusCritical  HealthStatus = "critical"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Alert represents a monitoring alert
type Alert struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Type        AlertType    `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time    `json:"created_at"`
	ResolvedAt  *time.Time   `json:"resolved_at,omitempty"`
	Status      AlertStatus  `json:"status"`
}

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeUsage         AlertType = "usage"
	AlertTypePerformance   AlertType = "performance"
	AlertTypeSubscription  AlertType = "subscription"
	AlertTypeHealth        AlertType = "health"
	AlertTypeSecurity      AlertType = "security"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// tenantMonitor implements TenantMonitor interface
type tenantMonitor struct {
	logger          logger.EnhancedLogger
	usageStore      map[uuid.UUID]map[string]*UsageMetrics
	performanceStore map[uuid.UUID]*PerformanceMetrics
	healthStore     map[uuid.UUID]*TenantHealth
	alertStore      map[uuid.UUID][]*Alert
	mu              sync.RWMutex
}

// NewTenantMonitor creates a new tenant monitor instance
func NewTenantMonitor(logger logger.EnhancedLogger) TenantMonitor {
	return &tenantMonitor{
		logger:          logger,
		usageStore:      make(map[uuid.UUID]map[string]*UsageMetrics),
		performanceStore: make(map[uuid.UUID]*PerformanceMetrics),
		healthStore:     make(map[uuid.UUID]*TenantHealth),
		alertStore:      make(map[uuid.UUID][]*Alert),
	}
}

// TrackUsage tracks resource usage for a tenant
func (tm *tenantMonitor) TrackUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Initialize tenant usage map if not exists
	if tm.usageStore[tenantID] == nil {
		tm.usageStore[tenantID] = make(map[string]*UsageMetrics)
	}

	// Get or create usage metrics for resource
	usage, exists := tm.usageStore[tenantID][resource]
	if !exists {
		usage = &UsageMetrics{
			TenantID:     tenantID,
			Resource:     resource,
			CurrentUsage: 0,
			Limit:        getDefaultLimit(resource),
			ResetDate:    getNextResetDate(),
			History:      make([]UsageDataPoint, 0),
		}
		tm.usageStore[tenantID][resource] = usage
	}

	// Update usage
	usage.CurrentUsage += amount
	usage.History = append(usage.History, UsageDataPoint{
		Timestamp: time.Now(),
		Usage:     amount,
	})

	// Keep only last 100 data points
	if len(usage.History) > 100 {
		usage.History = usage.History[len(usage.History)-100:]
	}

	// Log usage tracking
	tm.logger.LogTenantUsage(tenantID.String(), resource, usage.CurrentUsage, usage.Limit)

	// Check for alerts
	tm.checkUsageAlerts(ctx, tenantID, usage)

	return nil
}

// GetUsage retrieves usage metrics for a tenant resource
func (tm *tenantMonitor) GetUsage(ctx context.Context, tenantID uuid.UUID, resource string) (*UsageMetrics, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tenantUsage, exists := tm.usageStore[tenantID]; exists {
		if usage, exists := tenantUsage[resource]; exists {
			return usage, nil
		}
	}

	// Return zero usage if not tracked yet
	return &UsageMetrics{
		TenantID:     tenantID,
		Resource:     resource,
		CurrentUsage: 0,
		Limit:        getDefaultLimit(resource),
		ResetDate:    getNextResetDate(),
		History:      make([]UsageDataPoint, 0),
	}, nil
}

// Helper functions

func getDefaultLimit(resource string) int64 {
	limits := map[string]int64{
		"api_requests":  10000,  // requests per month
		"storage":      1000000, // bytes
		"users":        10,      // user count
		"products":     1000,    // product count
		"sales":        1000,    // sales per month
	}
	
	if limit, exists := limits[resource]; exists {
		return limit
	}
	return 1000 // Default limit
}

func getNextResetDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
}

func (tm *tenantMonitor) checkUsageAlerts(ctx context.Context, tenantID uuid.UUID, usage *UsageMetrics) {
	percentage := float64(usage.CurrentUsage) / float64(usage.Limit) * 100

	// Create warning alert at 80%
	if percentage >= 80 && percentage < 100 {
		alert := &Alert{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Type:        AlertTypeUsage,
			Severity:    AlertSeverityWarning,
			Title:       "Usage Warning",
			Description: fmt.Sprintf("Resource '%s' usage at %.1f%%", usage.Resource, percentage),
			Metadata: map[string]interface{}{
				"resource":    usage.Resource,
				"usage":       usage.CurrentUsage,
				"limit":       usage.Limit,
				"percentage":  percentage,
			},
			CreatedAt: time.Now(),
			Status:    AlertStatusActive,
		}
		tm.CreateAlert(ctx, alert)
	}

	// Create critical alert at 100%
	if percentage >= 100 {
		alert := &Alert{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Type:        AlertTypeUsage,
			Severity:    AlertSeverityCritical,
			Title:       "Usage Limit Exceeded",
			Description: fmt.Sprintf("Resource '%s' usage exceeded limit: %.1f%%", usage.Resource, percentage),
			Metadata: map[string]interface{}{
				"resource":    usage.Resource,
				"usage":       usage.CurrentUsage,
				"limit":       usage.Limit,
				"percentage":  percentage,
			},
			CreatedAt: time.Now(),
			Status:    AlertStatusActive,
		}
		tm.CreateAlert(ctx, alert)
	}
}

// Additional implementation methods would continue here...
// (Implementing the remaining interface methods)

// TrackResponse tracks API response performance for a tenant
func (tm *tenantMonitor) TrackResponse(ctx context.Context, tenantID uuid.UUID, operation string, duration time.Duration, success bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get or create performance metrics
	metrics, exists := tm.performanceStore[tenantID]
	if !exists {
		metrics = &PerformanceMetrics{
			TenantID:         tenantID,
			OperationMetrics: make(map[string]*OperationMetrics),
			LastUpdated:      time.Now(),
		}
		tm.performanceStore[tenantID] = metrics
	}

	// Update overall metrics
	metrics.RequestCount++
	if !success {
		metrics.ErrorCount++
	}
	metrics.SuccessRate = float64(metrics.RequestCount-metrics.ErrorCount) / float64(metrics.RequestCount) * 100
	metrics.LastUpdated = time.Now()

	// Update operation-specific metrics
	opMetrics, exists := metrics.OperationMetrics[operation]
	if !exists {
		opMetrics = &OperationMetrics{
			Operation: operation,
		}
		metrics.OperationMetrics[operation] = opMetrics
	}

	opMetrics.RequestCount++
	if !success {
		opMetrics.ErrorCount++
	}
	opMetrics.SuccessRate = float64(opMetrics.RequestCount-opMetrics.ErrorCount) / float64(opMetrics.RequestCount) * 100

	// Log performance tracking
	tm.logger.WithFields(map[string]interface{}{
		"tenant_id":  tenantID.String(),
		"operation":  operation,
		"duration":   duration.Milliseconds(),
		"success":    success,
	}).Debug("Performance tracked")
}

// GetPerformanceMetrics retrieves performance metrics for a tenant
func (tm *tenantMonitor) GetPerformanceMetrics(ctx context.Context, tenantID uuid.UUID) (*PerformanceMetrics, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if metrics, exists := tm.performanceStore[tenantID]; exists {
		return metrics, nil
	}

	// Return empty metrics if not tracked yet
	return &PerformanceMetrics{
		TenantID:         tenantID,
		OperationMetrics: make(map[string]*OperationMetrics),
		LastUpdated:      time.Now(),
	}, nil
}

// TrackSubscriptionEvent tracks subscription-related events
func (tm *tenantMonitor) TrackSubscriptionEvent(ctx context.Context, tenantID uuid.UUID, event string, data interface{}) {
	tm.logger.LogTenantSubscriptionEvent(tenantID.String(), event, data)
}

// CheckSubscriptionLimits checks current usage against subscription limits
func (tm *tenantMonitor) CheckSubscriptionLimits(ctx context.Context, tenantID uuid.UUID) (*LimitStatus, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	limitStatus := &LimitStatus{
		TenantID:    tenantID,
		PlanType:    entities.PlanStarter, // Default, should be fetched from subscription
		Limits:      make(map[string]*ResourceLimit),
		IsOverLimit: false,
		Warnings:    make([]string, 0),
	}

	// Check usage against limits for each resource
	if tenantUsage, exists := tm.usageStore[tenantID]; exists {
		for resource, usage := range tenantUsage {
			percentage := float64(usage.CurrentUsage) / float64(usage.Limit) * 100
			isExceeded := usage.CurrentUsage >= usage.Limit
			isWarning := percentage >= 80

			limitStatus.Limits[resource] = &ResourceLimit{
				Resource:     resource,
				CurrentUsage: usage.CurrentUsage,
				Limit:        usage.Limit,
				Percentage:   percentage,
				IsExceeded:   isExceeded,
				IsWarning:    isWarning,
				LastUpdated:  time.Now(),
			}

			if isExceeded {
				limitStatus.IsOverLimit = true
			}

			if isWarning {
				limitStatus.Warnings = append(limitStatus.Warnings, 
					fmt.Sprintf("%s usage at %.1f%%", resource, percentage))
			}
		}
	}

	return limitStatus, nil
}

// RecordHealthCheck records health check results for a tenant service
func (tm *tenantMonitor) RecordHealthCheck(ctx context.Context, tenantID uuid.UUID, service string, status bool, responseTime time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get or create tenant health
	health, exists := tm.healthStore[tenantID]
	if !exists {
		health = &TenantHealth{
			TenantID:      tenantID,
			OverallStatus: HealthStatusHealthy,
			Services:      make(map[string]*ServiceHealth),
			LastChecked:   time.Now(),
			Issues:        make([]string, 0),
		}
		tm.healthStore[tenantID] = health
	}

	// Update service health
	serviceHealth := &ServiceHealth{
		Service:      service,
		Status:       HealthStatusHealthy,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
		Message:      "Service is healthy",
	}

	if !status {
		serviceHealth.Status = HealthStatusCritical
		serviceHealth.Message = "Service is unhealthy"
		serviceHealth.ErrorCount++
	} else if responseTime > 5*time.Second {
		serviceHealth.Status = HealthStatusWarning
		serviceHealth.Message = "Service responding slowly"
	}

	health.Services[service] = serviceHealth
	health.LastChecked = time.Now()

	// Update overall health status
	tm.updateOverallHealth(health)

	// Log health check
	tm.logger.LogHealthCheck(fmt.Sprintf("%s[%s]", service, tenantID.String()), status, responseTime, serviceHealth.Message)
}

// GetTenantHealth retrieves health status for a tenant
func (tm *tenantMonitor) GetTenantHealth(ctx context.Context, tenantID uuid.UUID) (*TenantHealth, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if health, exists := tm.healthStore[tenantID]; exists {
		return health, nil
	}

	// Return default healthy status if not tracked yet
	return &TenantHealth{
		TenantID:      tenantID,
		OverallStatus: HealthStatusUnknown,
		Services:      make(map[string]*ServiceHealth),
		LastChecked:   time.Now(),
		Issues:        []string{"No health data available"},
	}, nil
}

// CheckAlerts retrieves active alerts for a tenant
func (tm *tenantMonitor) CheckAlerts(ctx context.Context, tenantID uuid.UUID) ([]*Alert, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if alerts, exists := tm.alertStore[tenantID]; exists {
		// Filter active alerts
		activeAlerts := make([]*Alert, 0)
		for _, alert := range alerts {
			if alert.Status == AlertStatusActive {
				activeAlerts = append(activeAlerts, alert)
			}
		}
		return activeAlerts, nil
	}

	return make([]*Alert, 0), nil
}

// CreateAlert creates a new alert for a tenant
func (tm *tenantMonitor) CreateAlert(ctx context.Context, alert *Alert) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Initialize tenant alerts if not exists
	if tm.alertStore[alert.TenantID] == nil {
		tm.alertStore[alert.TenantID] = make([]*Alert, 0)
	}

	// Add alert
	tm.alertStore[alert.TenantID] = append(tm.alertStore[alert.TenantID], alert)

	// Log alert creation
	tm.logger.WithFields(map[string]interface{}{
		"tenant_id": alert.TenantID.String(),
		"alert_id":  alert.ID.String(),
		"type":      alert.Type,
		"severity":  alert.Severity,
		"title":     alert.Title,
	}).Warn("Alert created")

	return nil
}

// Helper method to update overall health status
func (tm *tenantMonitor) updateOverallHealth(health *TenantHealth) {
	if len(health.Services) == 0 {
		health.OverallStatus = HealthStatusUnknown
		return
	}

	criticalCount := 0
	warningCount := 0
	healthyCount := 0
	health.Issues = make([]string, 0)

	for _, service := range health.Services {
		switch service.Status {
		case HealthStatusCritical:
			criticalCount++
			health.Issues = append(health.Issues, fmt.Sprintf("%s: %s", service.Service, service.Message))
		case HealthStatusWarning:
			warningCount++
			health.Issues = append(health.Issues, fmt.Sprintf("%s: %s", service.Service, service.Message))
		case HealthStatusHealthy:
			healthyCount++
		}
	}

	if criticalCount > 0 {
		health.OverallStatus = HealthStatusCritical
	} else if warningCount > 0 {
		health.OverallStatus = HealthStatusWarning
	} else {
		health.OverallStatus = HealthStatusHealthy
	}
}