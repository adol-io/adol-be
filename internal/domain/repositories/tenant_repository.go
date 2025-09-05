package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
)

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *entities.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error)

	// GetBySlug retrieves a tenant by slug
	GetBySlug(ctx context.Context, slug string) (*entities.Tenant, error)

	// GetByDomain retrieves a tenant by domain
	GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error)

	// Update updates a tenant
	Update(ctx context.Context, tenant *entities.Tenant) error

	// Delete soft deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves tenants with pagination
	List(ctx context.Context, offset, limit int) ([]*entities.Tenant, int, error)

	// ExistsBySlug checks if a tenant exists with the given slug
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// ExistsByDomain checks if a tenant exists with the given domain
	ExistsByDomain(ctx context.Context, domain string) (bool, error)

	// GetActiveCount returns the count of active tenants
	GetActiveCount(ctx context.Context) (int, error)

	// GetTrialTenants retrieves tenants that are in trial period
	GetTrialTenants(ctx context.Context) ([]*entities.Tenant, error)

	// GetExpiredTrialTenants retrieves tenants with expired trials
	GetExpiredTrialTenants(ctx context.Context) ([]*entities.Tenant, error)
}

// TenantSubscriptionRepository defines the interface for tenant subscription operations
type TenantSubscriptionRepository interface {
	// Create creates a new tenant subscription
	Create(ctx context.Context, subscription *entities.TenantSubscription) error

	// GetByTenantID retrieves a subscription by tenant ID
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entities.TenantSubscription, error)

	// Update updates a tenant subscription
	Update(ctx context.Context, subscription *entities.TenantSubscription) error

	// Delete deletes a tenant subscription
	Delete(ctx context.Context, tenantID uuid.UUID) error

	// GetActiveSubscriptions retrieves all active subscriptions
	GetActiveSubscriptions(ctx context.Context) ([]*entities.TenantSubscription, error)

	// GetSubscriptionsByPlan retrieves subscriptions by plan type
	GetSubscriptionsByPlan(ctx context.Context, planType entities.SubscriptionPlanType) ([]*entities.TenantSubscription, error)

	// GetExpiredSubscriptions retrieves subscriptions that have expired
	GetExpiredSubscriptions(ctx context.Context) ([]*entities.TenantSubscription, error)

	// UpdateUsage updates the usage statistics for a subscription
	UpdateUsage(ctx context.Context, tenantID uuid.UUID, usage entities.SubscriptionUsage) error

	// GetUsageByTenant retrieves usage statistics for a tenant
	GetUsageByTenant(ctx context.Context, tenantID uuid.UUID) (*entities.SubscriptionUsage, error)
}

// TenantSettingRepository defines the interface for tenant settings operations
type TenantSettingRepository interface {
	// Create creates a new tenant setting
	Create(ctx context.Context, setting *entities.TenantSetting) error

	// GetByTenantAndKey retrieves a setting by tenant ID and key
	GetByTenantAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*entities.TenantSetting, error)

	// GetByTenant retrieves all settings for a tenant
	GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entities.TenantSetting, error)

	// Update updates a tenant setting
	Update(ctx context.Context, setting *entities.TenantSetting) error

	// Delete deletes a tenant setting
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error

	// Upsert creates or updates a tenant setting
	Upsert(ctx context.Context, tenantID uuid.UUID, key string, value interface{}) error

	// GetSettings retrieves settings as a key-value map
	GetSettings(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error)
}