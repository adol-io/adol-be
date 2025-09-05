package entities

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/pkg/errors"
)

// TenantStatus represents tenant status
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusTrial     TenantStatus = "trial"
)

// TenantConfiguration represents tenant-specific configuration
type TenantConfiguration struct {
	BusinessInfo BusinessInfo            `json:"business_info"`
	POSSettings  POSSettings             `json:"pos_settings"`
	FeatureFlags map[string]bool         `json:"feature_flags"`
	CustomFields map[string]interface{}  `json:"custom_fields,omitempty"`
}

// BusinessInfo represents tenant's business information
type BusinessInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	TaxID    string `json:"tax_id,omitempty"`
	Currency string `json:"currency"`
	TaxRate  float64 `json:"tax_rate"`
}

// POSSettings represents POS-specific settings
type POSSettings struct {
	DefaultCurrency    string `json:"default_currency"`
	TaxRate           float64 `json:"tax_rate"`
	ReceiptTemplate   string `json:"receipt_template"`
	AutoPrintReceipts bool   `json:"auto_print_receipts"`
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID            uuid.UUID           `json:"id"`
	Name          string              `json:"name"`
	Slug          string              `json:"slug"`
	Domain        *string             `json:"domain,omitempty"`
	Status        TenantStatus        `json:"status"`
	Configuration TenantConfiguration `json:"configuration"`
	TrialStart    *time.Time          `json:"trial_start,omitempty"`
	TrialEnd      *time.Time          `json:"trial_end,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	CreatedBy     *uuid.UUID          `json:"created_by,omitempty"`
}

// NewTenant creates a new tenant
func NewTenant(name, domain string, createdBy *uuid.UUID) (*Tenant, error) {
	if err := validateTenantInput(name, domain); err != nil {
		return nil, err
	}

	slug := generateSlugFromName(name)
	if err := validateSlug(slug); err != nil {
		return nil, err
	}

	now := time.Now()
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   name,
		Slug:   slug,
		Status: TenantStatusTrial,
		Configuration: TenantConfiguration{
			BusinessInfo: BusinessInfo{
				Name:     name,
				Currency: "USD",
				TaxRate:  0.0,
			},
			POSSettings: POSSettings{
				DefaultCurrency:    "USD",
				TaxRate:           0.0,
				ReceiptTemplate:   "standard",
				AutoPrintReceipts: true,
			},
			FeatureFlags: map[string]bool{
				"pos":                true,
				"inventory":          true,
				"reporting":          true,
				"advanced_reporting": false,
				"multi_location":     false,
				"api_access":         false,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
	}

	if domain != "" {
		tenant.Domain = &domain
	}

	// Set trial period for new tenants
	tenant.TrialStart = &now
	trialEnd := now.AddDate(0, 0, 30) // 30-day trial
	tenant.TrialEnd = &trialEnd

	return tenant, nil
}

// UpdateTenant updates tenant information
func (t *Tenant) UpdateTenant(name, domain string) error {
	if err := validateTenantUpdateInput(name, domain); err != nil {
		return err
	}

	t.Name = name
	if domain != "" {
		t.Domain = &domain
	} else {
		t.Domain = nil
	}
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateConfiguration updates tenant configuration
func (t *Tenant) UpdateConfiguration(config TenantConfiguration) error {
	if err := validateTenantConfiguration(config); err != nil {
		return err
	}

	t.Configuration = config
	t.UpdatedAt = time.Now()
	return nil
}

// UpdateBusinessInfo updates business information
func (t *Tenant) UpdateBusinessInfo(businessInfo BusinessInfo) error {
	if businessInfo.Name == "" {
		return errors.NewValidationError("business name is required", "business_name cannot be empty")
	}
	if businessInfo.Currency == "" {
		businessInfo.Currency = "USD"
	}

	t.Configuration.BusinessInfo = businessInfo
	t.UpdatedAt = time.Now()
	return nil
}

// ChangeStatus changes the tenant's status
func (t *Tenant) ChangeStatus(status TenantStatus) error {
	if err := ValidateTenantStatus(status); err != nil {
		return err
	}

	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

// ActivateTenant activates the tenant and ends trial if applicable
func (t *Tenant) ActivateTenant() error {
	t.Status = TenantStatusActive
	
	// End trial period if still in trial
	if t.Status == TenantStatusTrial && t.TrialEnd != nil {
		now := time.Now()
		if t.TrialEnd.After(now) {
			t.TrialEnd = &now
		}
	}
	
	t.UpdatedAt = time.Now()
	return nil
}

// SuspendTenant suspends the tenant
func (t *Tenant) SuspendTenant() error {
	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now()
	return nil
}

// IsActive checks if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsInTrial checks if the tenant is in trial period
func (t *Tenant) IsInTrial() bool {
	if t.Status != TenantStatusTrial {
		return false
	}
	
	if t.TrialEnd == nil {
		return false
	}
	
	return t.TrialEnd.After(time.Now())
}

// GetTrialDaysRemaining returns the number of trial days remaining
func (t *Tenant) GetTrialDaysRemaining() int {
	if !t.IsInTrial() {
		return 0
	}
	
	days := int(t.TrialEnd.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		return 0
	}
	
	return days
}

// HasFeature checks if the tenant has a specific feature enabled
func (t *Tenant) HasFeature(feature string) bool {
	if t.Configuration.FeatureFlags == nil {
		return false
	}
	
	enabled, exists := t.Configuration.FeatureFlags[feature]
	return exists && enabled
}

// EnableFeature enables a feature for the tenant
func (t *Tenant) EnableFeature(feature string) {
	if t.Configuration.FeatureFlags == nil {
		t.Configuration.FeatureFlags = make(map[string]bool)
	}
	
	t.Configuration.FeatureFlags[feature] = true
	t.UpdatedAt = time.Now()
}

// DisableFeature disables a feature for the tenant
func (t *Tenant) DisableFeature(feature string) {
	if t.Configuration.FeatureFlags == nil {
		t.Configuration.FeatureFlags = make(map[string]bool)
	}
	
	t.Configuration.FeatureFlags[feature] = false
	t.UpdatedAt = time.Now()
}

// GetCurrency returns the tenant's currency
func (t *Tenant) GetCurrency() string {
	if t.Configuration.BusinessInfo.Currency != "" {
		return t.Configuration.BusinessInfo.Currency
	}
	return "USD"
}

// GetTaxRate returns the tenant's tax rate
func (t *Tenant) GetTaxRate() decimal.Decimal {
	return decimal.NewFromFloat(t.Configuration.BusinessInfo.TaxRate)
}

// ValidateTenantStatus validates if the status is valid
func ValidateTenantStatus(status TenantStatus) error {
	switch status {
	case TenantStatusActive, TenantStatusInactive, TenantStatusSuspended, TenantStatusTrial:
		return nil
	default:
		return errors.NewValidationError("invalid tenant status", "status must be one of: active, inactive, suspended, trial")
	}
}

// Helper functions

func validateTenantInput(name, domain string) error {
	if name == "" {
		return errors.NewValidationError("tenant name is required", "name cannot be empty")
	}
	if len(name) < 2 {
		return errors.NewValidationError("tenant name too short", "name must be at least 2 characters long")
	}
	if len(name) > 255 {
		return errors.NewValidationError("tenant name too long", "name cannot exceed 255 characters")
	}
	
	if domain != "" {
		if err := validateDomain(domain); err != nil {
			return err
		}
	}
	
	return nil
}

func validateTenantUpdateInput(name, domain string) error {
	return validateTenantInput(name, domain)
}

func validateTenantConfiguration(config TenantConfiguration) error {
	if config.BusinessInfo.Name == "" {
		return errors.NewValidationError("business name is required", "business_info.name cannot be empty")
	}
	
	if config.BusinessInfo.Currency == "" {
		config.BusinessInfo.Currency = "USD"
	}
	
	return nil
}

// GenerateSlugFromName generates a slug from a tenant name
func GenerateSlugFromName(name string) string {
	return generateSlugFromName(name)
}

func generateSlugFromName(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	// Remove special characters, keep only alphanumeric and hyphens
	var cleanSlug strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanSlug.WriteRune(r)
		}
	}
	
	// Remove multiple consecutive hyphens and trim
	result := strings.TrimSpace(cleanSlug.String())
	result = strings.Trim(result, "-")
	
	// Handle multiple consecutive hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	
	return result
}

func validateSlug(slug string) error {
	if slug == "" {
		return errors.NewValidationError("invalid slug", "slug cannot be empty")
	}
	if len(slug) < 2 {
		return errors.NewValidationError("slug too short", "slug must be at least 2 characters long")
	}
	if len(slug) > 100 {
		return errors.NewValidationError("slug too long", "slug cannot exceed 100 characters")
	}
	
	// Validate slug format (alphanumeric and hyphens only)
	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return errors.NewValidationError("invalid slug format", "slug can only contain lowercase letters, numbers, and hyphens")
		}
	}
	
	// Slug cannot start or end with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return errors.NewValidationError("invalid slug format", "slug cannot start or end with hyphen")
	}
	
	return nil
}

func validateDomain(domain string) error {
	if len(domain) < 3 {
		return errors.NewValidationError("domain too short", "domain must be at least 3 characters long")
	}
	if len(domain) > 255 {
		return errors.NewValidationError("domain too long", "domain cannot exceed 255 characters")
	}
	
	// Basic domain validation - in production, you might want more robust validation
	if !strings.Contains(domain, ".") {
		return errors.NewValidationError("invalid domain format", "domain must contain at least one dot")
	}
	
	return nil
}