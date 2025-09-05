package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTenantSubscription(t *testing.T) {
	tenantID := uuid.New()
	
	subscription, err := NewTenantSubscription(tenantID, PlanStarter)
	
	assert.NoError(t, err)
	assert.NotNil(t, subscription)
	assert.Equal(t, tenantID, subscription.TenantID)
	assert.Equal(t, PlanStarter, subscription.PlanType)
	assert.Equal(t, SubscriptionStatusTrial, subscription.Status)
	assert.NotEqual(t, uuid.Nil, subscription.ID)
	assert.False(t, subscription.CreatedAt.IsZero())
	assert.False(t, subscription.UpdatedAt.IsZero())
}

func TestTenantSubscription_UpgradePlan(t *testing.T) {
	tenantID := uuid.New()
	subscription, _ := NewTenantSubscription(tenantID, PlanStarter)

	err := subscription.UpgradePlan(PlanProfessional)
	assert.NoError(t, err)
	assert.Equal(t, PlanProfessional, subscription.PlanType)
	assert.Equal(t, SubscriptionStatusActive, subscription.Status)
	assert.NotNil(t, subscription.BillingStart)
	assert.NotNil(t, subscription.BillingEnd)
}

func TestTenantSubscription_ChangeStatus(t *testing.T) {
	tenantID := uuid.New()
	subscription, _ := NewTenantSubscription(tenantID, PlanStarter)

	err := subscription.ChangeStatus(SubscriptionStatusCancelled)
	assert.NoError(t, err)
	assert.Equal(t, SubscriptionStatusCancelled, subscription.Status)

	// Test invalid status
	err = subscription.ChangeStatus("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subscription status")
}

func TestTenantSubscription_IsActive(t *testing.T) {
	tenantID := uuid.New()
	
	tests := []struct {
		name     string
		status   SubscriptionStatus
		expected bool
	}{
		{"Active subscription", SubscriptionStatusActive, true},
		{"Cancelled subscription", SubscriptionStatusCancelled, false},
		{"Suspended subscription", SubscriptionStatusSuspended, false},
		{"Trial subscription", SubscriptionStatusTrial, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription, _ := NewTenantSubscription(tenantID, PlanStarter)
			subscription.Status = tt.status

			result := subscription.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantSubscription_HasFeature(t *testing.T) {
	tests := []struct {
		name     string
		planType SubscriptionPlanType
		feature  string
		expected bool
	}{
		{"Starter has POS", PlanStarter, "pos", true},
		{"Starter no advanced reporting", PlanStarter, "advanced_reporting", false},
		{"Professional has advanced reporting", PlanProfessional, "advanced_reporting", true},
		{"Professional no API access", PlanProfessional, "api_access", false},
		{"Enterprise has API access", PlanEnterprise, "api_access", true},
		{"Enterprise has multi-location", PlanEnterprise, "multi_location", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID := uuid.New()
			subscription, _ := NewTenantSubscription(tenantID, tt.planType)

			result := subscription.HasFeature(tt.feature)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantSubscription_UpdateUsage(t *testing.T) {
	tenantID := uuid.New()
	subscription, _ := NewTenantSubscription(tenantID, PlanStarter)

	newUsage := SubscriptionUsage{
		Users:    3,
		Products: 150,
	}

	subscription.UpdateUsage(newUsage)
	assert.Equal(t, 3, subscription.CurrentUsage.Users)
	assert.Equal(t, 150, subscription.CurrentUsage.Products)
}

func TestValidateSubscriptionPlanType(t *testing.T) {
	validPlans := []SubscriptionPlanType{
		PlanStarter,
		PlanProfessional,
		PlanEnterprise,
	}

	for _, plan := range validPlans {
		err := ValidateSubscriptionPlanType(plan)
		assert.NoError(t, err)
	}

	// Test invalid plan
	err := ValidateSubscriptionPlanType("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subscription plan type")
}

func TestValidateSubscriptionStatus(t *testing.T) {
	validStatuses := []SubscriptionStatus{
		SubscriptionStatusActive,
		SubscriptionStatusInactive,
		SubscriptionStatusSuspended,
		SubscriptionStatusCancelled,
		SubscriptionStatusTrial,
	}

	for _, status := range validStatuses {
		err := ValidateSubscriptionStatus(status)
		assert.NoError(t, err)
	}

	// Test invalid status
	err := ValidateSubscriptionStatus("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subscription status")
}