package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/utils"
)

// DatabasePort defines the interface for database operations
type DatabasePort interface {
	// Transaction management
	BeginTransaction(ctx context.Context) (TransactionPort, error)
	Health(ctx context.Context) error
}

// TransactionPort defines the interface for database transactions
type TransactionPort interface {
	Commit() error
	Rollback() error
	GetUserRepository() repositories.UserRepository
	GetProductRepository() repositories.ProductRepository
	GetStockRepository() repositories.StockRepository
	GetStockMovementRepository() repositories.StockMovementRepository
	GetSaleRepository() repositories.SaleRepository
	GetSaleItemRepository() repositories.SaleItemRepository
	GetInvoiceRepository() repositories.InvoiceRepository
	GetInvoiceItemRepository() repositories.InvoiceItemRepository
}

// CachePort defines the interface for caching operations
type CachePort interface {
	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Advanced operations
	SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags []string) error
	InvalidateByTags(ctx context.Context, tags []string) error
	
	// User session management
	SetUserSession(ctx context.Context, userID uuid.UUID, sessionData interface{}, expiration time.Duration) error
	GetUserSession(ctx context.Context, userID uuid.UUID, dest interface{}) error
	DeleteUserSession(ctx context.Context, userID uuid.UUID) error
	
	// Token blacklist management
	BlacklistToken(ctx context.Context, token string, expiration time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// EventBusPort defines the interface for event publishing and handling
type EventBusPort interface {
	// Publish publishes an event
	Publish(ctx context.Context, event DomainEvent) error
	
	// Subscribe subscribes to events of a specific type
	Subscribe(eventType string, handler EventHandler) error
	
	// Unsubscribe unsubscribes from events
	Unsubscribe(eventType string, handler EventHandler) error
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
}

// DomainEvent represents a domain event
type DomainEvent interface {
	GetEventType() string
	GetEventID() uuid.UUID
	GetAggregateID() uuid.UUID
	GetTimestamp() time.Time
	GetPayload() interface{}
}

// FileStoragePort defines the interface for file storage operations
type FileStoragePort interface {
	// Store stores a file and returns the file path
	Store(ctx context.Context, filename string, data []byte) (string, error)
	
	// Retrieve retrieves a file by path
	Retrieve(ctx context.Context, filepath string) ([]byte, error)
	
	// Delete deletes a file by path
	Delete(ctx context.Context, filepath string) error
	
	// Exists checks if a file exists
	Exists(ctx context.Context, filepath string) (bool, error)
	
	// GetURL returns a public URL for a file
	GetURL(ctx context.Context, filepath string) (string, error)
	
	// GetSignedURL returns a signed URL for private file access
	GetSignedURL(ctx context.Context, filepath string, expiration time.Duration) (string, error)
}

// NotificationPort defines the interface for notifications
type NotificationPort interface {
	// Send email notification
	SendEmail(ctx context.Context, notification EmailNotification) error
	
	// Send SMS notification
	SendSMS(ctx context.Context, notification SMSNotification) error
	
	// Send push notification
	SendPushNotification(ctx context.Context, notification PushNotification) error
	
	// Send webhook notification
	SendWebhook(ctx context.Context, notification WebhookNotification) error
}

// EmailNotification represents email notification
type EmailNotification struct {
	To          []string `json:"to"`
	CC          []string `json:"cc,omitempty"`
	BCC         []string `json:"bcc,omitempty"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	IsHTML      bool     `json:"is_html"`
	Attachments []FileAttachment `json:"attachments,omitempty"`
	Priority    string   `json:"priority,omitempty"` // low, normal, high
}

// SMSNotification represents SMS notification
type SMSNotification struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// PushNotification represents push notification
type PushNotification struct {
	UserID  uuid.UUID `json:"user_id"`
	Title   string    `json:"title"`
	Message string    `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// WebhookNotification represents webhook notification
type WebhookNotification struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers,omitempty"`
	Payload map[string]interface{} `json:"payload"`
}

// FileAttachment represents file attachment
type FileAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// AuditPort defines the interface for audit logging
type AuditPort interface {
	// Log logs an audit event
	Log(ctx context.Context, event AuditEvent) error
	
	// Query queries audit events
	Query(ctx context.Context, filter AuditFilter, pagination utils.PaginationInfo) ([]AuditEvent, utils.PaginationInfo, error)
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	OldValue    map[string]interface{} `json:"old_value,omitempty"`
	NewValue    map[string]interface{} `json:"new_value,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Success     bool                   `json:"success"`
	ErrorMessage string                `json:"error_message,omitempty"`
}

// AuditFilter represents audit event filter
type AuditFilter struct {
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Action     string     `json:"action,omitempty"`
	Resource   string     `json:"resource,omitempty"`
	ResourceID string     `json:"resource_id,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
	Success    *bool      `json:"success,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	OrderBy    string     `json:"order_by,omitempty"`
	OrderDir   string     `json:"order_dir,omitempty"`
}

// MetricsPort defines the interface for metrics collection
type MetricsPort interface {
	// Counter operations
	IncrementCounter(name string, tags map[string]string)
	IncrementCounterBy(name string, value float64, tags map[string]string)
	
	// Gauge operations
	SetGauge(name string, value float64, tags map[string]string)
	
	// Histogram operations
	RecordHistogram(name string, value float64, tags map[string]string)
	
	// Timing operations
	RecordTiming(name string, duration time.Duration, tags map[string]string)
	StartTimer(name string, tags map[string]string) TimerPort
}

// TimerPort defines the interface for timing operations
type TimerPort interface {
	Stop()
}