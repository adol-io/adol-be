package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/pkg/errors"
)

// TenantSetting represents a tenant-specific setting
type TenantSetting struct {
	ID           uuid.UUID   `json:"id"`
	TenantID     uuid.UUID   `json:"tenant_id"`
	SettingKey   string      `json:"setting_key"`
	SettingValue interface{} `json:"setting_value"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// TenantContext represents the context for a tenant in the current request
type TenantContext struct {
	TenantID           uuid.UUID           `json:"tenant_id"`
	TenantName         string              `json:"tenant_name"`
	TenantSlug         string              `json:"tenant_slug"`
	TenantStatus       TenantStatus        `json:"tenant_status"`
	Configuration      TenantConfiguration `json:"configuration"`
	SubscriptionStatus SubscriptionStatus  `json:"subscription_status"`
	Features           SubscriptionFeatures `json:"features"`
	UsageLimits        SubscriptionLimits  `json:"usage_limits"`
}

// NewTenantSetting creates a new tenant setting
func NewTenantSetting(tenantID uuid.UUID, key string, value interface{}) (*TenantSetting, error) {
	if err := validateTenantSettingInput(key, value); err != nil {
		return nil, err
	}

	now := time.Now()
	setting := &TenantSetting{
		ID:           uuid.New(),
		TenantID:     tenantID,
		SettingKey:   key,
		SettingValue: value,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return setting, nil
}

// UpdateValue updates the setting value
func (ts *TenantSetting) UpdateValue(value interface{}) error {
	if err := validateSettingValue(value); err != nil {
		return err
	}

	ts.SettingValue = value
	ts.UpdatedAt = time.Now()
	return nil
}

// NewTenantContext creates a new tenant context
func NewTenantContext(tenant *Tenant, subscription *TenantSubscription) *TenantContext {
	return &TenantContext{
		TenantID:           tenant.ID,
		TenantName:         tenant.Name,
		TenantSlug:         tenant.Slug,
		TenantStatus:       tenant.Status,
		Configuration:      tenant.Configuration,
		SubscriptionStatus: subscription.Status,
		Features:           subscription.Features,
		UsageLimits:        subscription.UsageLimits,
	}
}

// IsActive checks if the tenant context is active
func (tc *TenantContext) IsActive() bool {
	return tc.TenantStatus == TenantStatusActive && 
		   (tc.SubscriptionStatus == SubscriptionStatusActive || tc.SubscriptionStatus == SubscriptionStatusTrial)
}

// HasFeature checks if a feature is enabled
func (tc *TenantContext) HasFeature(feature string) bool {
	switch feature {
	case "pos":
		return tc.Features.POS
	case "inventory":
		return tc.Features.Inventory
	case "reporting":
		return tc.Features.Reporting
	case "advanced_reporting":
		return tc.Features.AdvancedReporting
	case "multi_location":
		return tc.Features.MultiLocation
	case "api_access":
		return tc.Features.APIAccess
	case "custom_integration":
		return tc.Features.CustomIntegration
	default:
		return false
	}
}

// GetUserLimit returns the user limit for the tenant
func (tc *TenantContext) GetUserLimit() int {
	return tc.UsageLimits.Users
}

// GetProductLimit returns the product limit for the tenant
func (tc *TenantContext) GetProductLimit() int {
	return tc.UsageLimits.Products
}

// GetSalesLimit returns the monthly sales limit for the tenant
func (tc *TenantContext) GetSalesLimit() int {
	return tc.UsageLimits.SalesPerMonth
}

// GetAPILimit returns the monthly API call limit for the tenant
func (tc *TenantContext) GetAPILimit() int {
	return tc.UsageLimits.APICallsPerMonth
}

// GetCurrency returns the tenant's default currency
func (tc *TenantContext) GetCurrency() string {
	if tc.Configuration.BusinessInfo.Currency != "" {
		return tc.Configuration.BusinessInfo.Currency
	}
	return "USD"
}

// GetTaxRate returns the tenant's tax rate
func (tc *TenantContext) GetTaxRate() float64 {
	return tc.Configuration.BusinessInfo.TaxRate
}

// GetBusinessInfo returns the tenant's business information
func (tc *TenantContext) GetBusinessInfo() BusinessInfo {
	return tc.Configuration.BusinessInfo
}

// GetPOSSettings returns the tenant's POS settings
func (tc *TenantContext) GetPOSSettings() POSSettings {
	return tc.Configuration.POSSettings
}

// ValidateAccess validates if the tenant has access to the system
func (tc *TenantContext) ValidateAccess() error {
	if !tc.IsActive() {
		return errors.NewForbiddenError("tenant access denied", "tenant is not active or subscription is suspended")
	}
	
	return nil
}

// ValidateFeatureAccess validates if the tenant has access to a specific feature
func (tc *TenantContext) ValidateFeatureAccess(feature string) error {
	if err := tc.ValidateAccess(); err != nil {
		return err
	}
	
	if !tc.HasFeature(feature) {
		return errors.NewForbiddenError("feature access denied", "tenant does not have access to this feature")
	}
	
	return nil
}

// Helper functions

func validateTenantSettingInput(key string, value interface{}) error {
	if key == "" {
		return errors.NewValidationError("setting key is required", "setting_key cannot be empty")
	}
	
	if len(key) > 255 {
		return errors.NewValidationError("setting key too long", "setting_key cannot exceed 255 characters")
	}
	
	return validateSettingValue(value)
}

func validateSettingValue(value interface{}) error {
	if value == nil {
		return errors.NewValidationError("setting value is required", "setting_value cannot be nil")
	}
	
	// Additional validation can be added here based on setting type
	return nil
}