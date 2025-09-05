package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/pkg/errors"
)

// SubscriptionPlanType represents subscription plan types
type SubscriptionPlanType string

const (
	PlanStarter      SubscriptionPlanType = "starter"
	PlanProfessional SubscriptionPlanType = "professional"
	PlanEnterprise   SubscriptionPlanType = "enterprise"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusInactive  SubscriptionStatus = "inactive"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusTrial     SubscriptionStatus = "trial"
)

// SubscriptionFeatures represents features included in a subscription
type SubscriptionFeatures struct {
	POS               bool `json:"pos"`
	Inventory         bool `json:"inventory"`
	Reporting         bool `json:"reporting"`
	AdvancedReporting bool `json:"advanced_reporting"`
	MultiLocation     bool `json:"multi_location"`
	APIAccess         bool `json:"api_access"`
	CustomIntegration bool `json:"custom_integration"`
}

// SubscriptionLimits represents usage limits for a subscription
type SubscriptionLimits struct {
	Users            int `json:"users"`             // -1 means unlimited
	Products         int `json:"products"`          // -1 means unlimited
	SalesPerMonth    int `json:"sales_per_month"`   // -1 means unlimited
	APICallsPerMonth int `json:"api_calls_per_month"` // -1 means unlimited
}

// SubscriptionUsage represents current usage statistics
type SubscriptionUsage struct {
	Users         int `json:"users"`
	Products      int `json:"products"`
	SalesThisMonth int `json:"sales_this_month"`
	APICallsThisMonth int `json:"api_calls_this_month"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TenantSubscription represents a tenant's subscription
type TenantSubscription struct {
	ID           uuid.UUID            `json:"id"`
	TenantID     uuid.UUID            `json:"tenant_id"`
	PlanType     SubscriptionPlanType `json:"plan_type"`
	Status       SubscriptionStatus   `json:"status"`
	BillingStart *time.Time           `json:"billing_start,omitempty"`
	BillingEnd   *time.Time           `json:"billing_end,omitempty"`
	MonthlyFee   decimal.Decimal      `json:"monthly_fee"`
	Features     SubscriptionFeatures `json:"features"`
	UsageLimits  SubscriptionLimits   `json:"usage_limits"`
	CurrentUsage SubscriptionUsage    `json:"current_usage"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// NewTenantSubscription creates a new tenant subscription
func NewTenantSubscription(tenantID uuid.UUID, planType SubscriptionPlanType) (*TenantSubscription, error) {
	if err := ValidateSubscriptionPlanType(planType); err != nil {
		return nil, err
	}

	now := time.Now()
	subscription := &TenantSubscription{
		ID:       uuid.New(),
		TenantID: tenantID,
		PlanType: planType,
		Status:   SubscriptionStatusTrial,
		CurrentUsage: SubscriptionUsage{
			LastUpdated: now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set plan-specific configurations
	if err := subscription.ApplyPlanConfiguration(planType); err != nil {
		return nil, err
	}

	return subscription, nil
}

// ApplyPlanConfiguration applies the configuration for the specified plan
func (s *TenantSubscription) ApplyPlanConfiguration(planType SubscriptionPlanType) error {
	switch planType {
	case PlanStarter:
		s.MonthlyFee = decimal.Zero // 0 + 1% transaction fee
		s.Features = SubscriptionFeatures{
			POS:               true,
			Inventory:         true,
			Reporting:         true,
			AdvancedReporting: false,
			MultiLocation:     false,
			APIAccess:         false,
			CustomIntegration: false,
		}
		s.UsageLimits = SubscriptionLimits{
			Users:            2,
			Products:         -1, // unlimited
			SalesPerMonth:    -1, // unlimited
			APICallsPerMonth: 0,  // no API access
		}

	case PlanProfessional:
		s.MonthlyFee = decimal.NewFromFloat(300000) // Rp300,000
		s.Features = SubscriptionFeatures{
			POS:               true,
			Inventory:         true,
			Reporting:         true,
			AdvancedReporting: true,
			MultiLocation:     true,
			APIAccess:         false,
			CustomIntegration: false,
		}
		s.UsageLimits = SubscriptionLimits{
			Users:            10,
			Products:         -1, // unlimited
			SalesPerMonth:    -1, // unlimited
			APICallsPerMonth: 0,  // no API access
		}

	case PlanEnterprise:
		s.MonthlyFee = decimal.NewFromFloat(1500000) // Rp1,500,000
		s.Features = SubscriptionFeatures{
			POS:               true,
			Inventory:         true,
			Reporting:         true,
			AdvancedReporting: true,
			MultiLocation:     true,
			APIAccess:         true,
			CustomIntegration: true,
		}
		s.UsageLimits = SubscriptionLimits{
			Users:            -1, // unlimited
			Products:         -1, // unlimited
			SalesPerMonth:    -1, // unlimited
			APICallsPerMonth: 10000,
		}

	default:
		return errors.NewValidationError("invalid plan type", "plan_type must be one of: starter, professional, enterprise")
	}

	s.PlanType = planType
	s.UpdatedAt = time.Now()
	return nil
}

// UpgradePlan upgrades the subscription to a higher plan
func (s *TenantSubscription) UpgradePlan(newPlanType SubscriptionPlanType) error {
	if err := ValidateSubscriptionPlanType(newPlanType); err != nil {
		return err
	}

	// Validate upgrade path (cannot downgrade using this method)
	if err := s.validateUpgradePath(s.PlanType, newPlanType); err != nil {
		return err
	}

	if err := s.ApplyPlanConfiguration(newPlanType); err != nil {
		return err
	}

	s.Status = SubscriptionStatusActive
	now := time.Now()
	s.BillingStart = &now
	
	// Set billing end to one month from now
	billingEnd := now.AddDate(0, 1, 0)
	s.BillingEnd = &billingEnd

	s.UpdatedAt = now
	return nil
}

// DowngradePlan downgrades the subscription to a lower plan
func (s *TenantSubscription) DowngradePlan(newPlanType SubscriptionPlanType) error {
	if err := ValidateSubscriptionPlanType(newPlanType); err != nil {
		return err
	}

	// Check if current usage fits within new plan limits
	if err := s.validateUsageAgainstPlan(newPlanType); err != nil {
		return err
	}

	if err := s.ApplyPlanConfiguration(newPlanType); err != nil {
		return err
	}

	s.UpdatedAt = time.Now()
	return nil
}

// ChangeStatus changes the subscription status
func (s *TenantSubscription) ChangeStatus(status SubscriptionStatus) error {
	if err := ValidateSubscriptionStatus(status); err != nil {
		return err
	}

	s.Status = status
	s.UpdatedAt = time.Now()
	return nil
}

// ActivateSubscription activates the subscription
func (s *TenantSubscription) ActivateSubscription() error {
	s.Status = SubscriptionStatusActive
	now := time.Now()
	s.BillingStart = &now
	
	// Set billing end to one month from now
	billingEnd := now.AddDate(0, 1, 0)
	s.BillingEnd = &billingEnd
	
	s.UpdatedAt = now
	return nil
}

// SuspendSubscription suspends the subscription
func (s *TenantSubscription) SuspendSubscription() error {
	s.Status = SubscriptionStatusSuspended
	s.UpdatedAt = time.Now()
	return nil
}

// CancelSubscription cancels the subscription
func (s *TenantSubscription) CancelSubscription() error {
	s.Status = SubscriptionStatusCancelled
	s.UpdatedAt = time.Now()
	return nil
}

// IsActive checks if the subscription is active
func (s *TenantSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive
}

// IsInTrial checks if the subscription is in trial
func (s *TenantSubscription) IsInTrial() bool {
	return s.Status == SubscriptionStatusTrial
}

// HasFeature checks if the subscription includes a specific feature
func (s *TenantSubscription) HasFeature(feature string) bool {
	switch feature {
	case "pos":
		return s.Features.POS
	case "inventory":
		return s.Features.Inventory
	case "reporting":
		return s.Features.Reporting
	case "advanced_reporting":
		return s.Features.AdvancedReporting
	case "multi_location":
		return s.Features.MultiLocation
	case "api_access":
		return s.Features.APIAccess
	case "custom_integration":
		return s.Features.CustomIntegration
	default:
		return false
	}
}

// UpdateUsage updates the current usage statistics
func (s *TenantSubscription) UpdateUsage(usage SubscriptionUsage) {
	s.CurrentUsage = usage
	s.CurrentUsage.LastUpdated = time.Now()
	s.UpdatedAt = time.Now()
}

// CanAddUser checks if the subscription allows adding another user
func (s *TenantSubscription) CanAddUser() bool {
	if s.UsageLimits.Users == -1 {
		return true // unlimited
	}
	return s.CurrentUsage.Users < s.UsageLimits.Users
}

// CanAddProduct checks if the subscription allows adding another product
func (s *TenantSubscription) CanAddProduct() bool {
	if s.UsageLimits.Products == -1 {
		return true // unlimited
	}
	return s.CurrentUsage.Products < s.UsageLimits.Products
}

// CanProcessSale checks if the subscription allows processing another sale this month
func (s *TenantSubscription) CanProcessSale() bool {
	if s.UsageLimits.SalesPerMonth == -1 {
		return true // unlimited
	}
	return s.CurrentUsage.SalesThisMonth < s.UsageLimits.SalesPerMonth
}

// CanMakeAPICall checks if the subscription allows making another API call this month
func (s *TenantSubscription) CanMakeAPICall() bool {
	if s.UsageLimits.APICallsPerMonth == -1 {
		return true // unlimited
	}
	return s.CurrentUsage.APICallsThisMonth < s.UsageLimits.APICallsPerMonth
}

// GetUsagePercentage returns the usage percentage for a specific limit type
func (s *TenantSubscription) GetUsagePercentage(limitType string) float64 {
	switch limitType {
	case "users":
		if s.UsageLimits.Users == -1 {
			return 0 // unlimited
		}
		if s.UsageLimits.Users == 0 {
			return 100
		}
		return float64(s.CurrentUsage.Users) / float64(s.UsageLimits.Users) * 100

	case "products":
		if s.UsageLimits.Products == -1 {
			return 0 // unlimited
		}
		if s.UsageLimits.Products == 0 {
			return 100
		}
		return float64(s.CurrentUsage.Products) / float64(s.UsageLimits.Products) * 100

	case "sales":
		if s.UsageLimits.SalesPerMonth == -1 {
			return 0 // unlimited
		}
		if s.UsageLimits.SalesPerMonth == 0 {
			return 100
		}
		return float64(s.CurrentUsage.SalesThisMonth) / float64(s.UsageLimits.SalesPerMonth) * 100

	case "api_calls":
		if s.UsageLimits.APICallsPerMonth == -1 {
			return 0 // unlimited
		}
		if s.UsageLimits.APICallsPerMonth == 0 {
			return 100
		}
		return float64(s.CurrentUsage.APICallsThisMonth) / float64(s.UsageLimits.APICallsPerMonth) * 100

	default:
		return 0
	}
}

// ValidateSubscriptionPlanType validates if the plan type is valid
func ValidateSubscriptionPlanType(planType SubscriptionPlanType) error {
	switch planType {
	case PlanStarter, PlanProfessional, PlanEnterprise:
		return nil
	default:
		return errors.NewValidationError("invalid subscription plan type", "plan_type must be one of: starter, professional, enterprise")
	}
}

// ValidateSubscriptionStatus validates if the status is valid
func ValidateSubscriptionStatus(status SubscriptionStatus) error {
	switch status {
	case SubscriptionStatusActive, SubscriptionStatusInactive, SubscriptionStatusSuspended, SubscriptionStatusCancelled, SubscriptionStatusTrial:
		return nil
	default:
		return errors.NewValidationError("invalid subscription status", "status must be one of: active, inactive, suspended, cancelled, trial")
	}
}

// Helper functions

func (s *TenantSubscription) validateUpgradePath(currentPlan, newPlan SubscriptionPlanType) error {
	planHierarchy := map[SubscriptionPlanType]int{
		PlanStarter:      1,
		PlanProfessional: 2,
		PlanEnterprise:   3,
	}

	currentLevel := planHierarchy[currentPlan]
	newLevel := planHierarchy[newPlan]

	if newLevel <= currentLevel {
		return errors.NewValidationError("invalid upgrade path", "new plan must be higher than current plan")
	}

	return nil
}

func (s *TenantSubscription) validateUsageAgainstPlan(newPlanType SubscriptionPlanType) error {
	// Create a temporary subscription with the new plan to check limits
	tempSub := &TenantSubscription{PlanType: newPlanType}
	if err := tempSub.ApplyPlanConfiguration(newPlanType); err != nil {
		return err
	}

	// Check if current usage exceeds new plan limits
	if tempSub.UsageLimits.Users != -1 && s.CurrentUsage.Users > tempSub.UsageLimits.Users {
		return errors.NewValidationError("usage exceeds plan limit", "current user count exceeds new plan limit")
	}

	if tempSub.UsageLimits.Products != -1 && s.CurrentUsage.Products > tempSub.UsageLimits.Products {
		return errors.NewValidationError("usage exceeds plan limit", "current product count exceeds new plan limit")
	}

	if tempSub.UsageLimits.SalesPerMonth != -1 && s.CurrentUsage.SalesThisMonth > tempSub.UsageLimits.SalesPerMonth {
		return errors.NewValidationError("usage exceeds plan limit", "current monthly sales exceeds new plan limit")
	}

	if tempSub.UsageLimits.APICallsPerMonth != -1 && s.CurrentUsage.APICallsThisMonth > tempSub.UsageLimits.APICallsPerMonth {
		return errors.NewValidationError("usage exceeds plan limit", "current monthly API calls exceeds new plan limit")
	}

	return nil
}