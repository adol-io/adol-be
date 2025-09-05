package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name        string
		tenantName  string
		domain      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid tenant creation",
			tenantName:  "Test Company",
			domain:      "test-company.com",
			expectError: false,
		},
		{
			name:        "Valid tenant without domain",
			tenantName:  "Test Company",
			domain:      "",
			expectError: false,
		},
		{
			name:        "Empty name",
			tenantName:  "",
			domain:      "test.com",
			expectError: true,
			errorMsg:    "tenant name is required",
		},
		{
			name:        "Name too short",
			tenantName:  "A",
			domain:      "test.com",
			expectError: true,
			errorMsg:    "tenant name too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := NewTenant(tt.tenantName, tt.domain, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, tenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tenant)
				assert.Equal(t, tt.tenantName, tenant.Name)
				assert.Equal(t, TenantStatusTrial, tenant.Status)
				assert.NotEqual(t, uuid.Nil, tenant.ID)
				assert.False(t, tenant.CreatedAt.IsZero())
				assert.False(t, tenant.UpdatedAt.IsZero())
				assert.NotNil(t, tenant.TrialStart)
				assert.NotNil(t, tenant.TrialEnd)
				
				if tt.domain != "" {
					assert.Equal(t, &tt.domain, tenant.Domain)
				} else {
					assert.Nil(t, tenant.Domain)
				}
			}
		})
	}
}

func TestTenant_IsInTrial(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Test Company",
		Status: TenantStatusTrial,
	}

	// Test without trial dates
	assert.False(t, tenant.IsInTrial())

	// Test with active trial
	now := time.Now()
	trialStart := now.Add(-10 * 24 * time.Hour)
	trialEnd := now.Add(10 * 24 * time.Hour)
	tenant.TrialStart = &trialStart
	tenant.TrialEnd = &trialEnd
	assert.True(t, tenant.IsInTrial())

	// Test with expired trial
	expiredEnd := now.Add(-5 * 24 * time.Hour)
	tenant.TrialEnd = &expiredEnd
	assert.False(t, tenant.IsInTrial())

	// Test with non-trial status
	tenant.Status = TenantStatusActive
	trialEnd = now.Add(10 * 24 * time.Hour)
	tenant.TrialEnd = &trialEnd
	assert.False(t, tenant.IsInTrial())
}

func TestTenant_GetTrialDaysRemaining(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Test Company",
		Status: TenantStatusTrial,
	}

	// Test without trial
	assert.Equal(t, 0, tenant.GetTrialDaysRemaining())

	// Test with active trial
	now := time.Now()
	trialStart := now.Add(-10 * 24 * time.Hour)
	trialEnd := now.Add(5 * 24 * time.Hour)
	tenant.TrialStart = &trialStart
	tenant.TrialEnd = &trialEnd
	remaining := tenant.GetTrialDaysRemaining()
	assert.True(t, remaining >= 4 && remaining <= 5) // Account for timing variations

	// Test with expired trial
	expiredEnd := now.Add(-5 * 24 * time.Hour)
	tenant.TrialEnd = &expiredEnd
	assert.Equal(t, 0, tenant.GetTrialDaysRemaining())
}

func TestTenant_UpdateTenant(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Old Name",
		Status: TenantStatusActive,
	}
	originalUpdatedAt := tenant.UpdatedAt

	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	err := tenant.UpdateTenant("New Name", "new-domain.com")
	assert.NoError(t, err)
	assert.Equal(t, "New Name", tenant.Name)
	assert.Equal(t, "new-domain.com", *tenant.Domain)
	assert.True(t, tenant.UpdatedAt.After(originalUpdatedAt))
}

func TestTenant_UpdateTenantEmpty(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Old Name",
		Status: TenantStatusActive,
	}

	err := tenant.UpdateTenant("", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant name is required")
	assert.Equal(t, "Old Name", tenant.Name) // Name should not change
}

func TestTenant_ChangeStatus(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Test Company",
		Status: TenantStatusActive,
	}

	err := tenant.ChangeStatus(TenantStatusSuspended)
	assert.NoError(t, err)
	assert.Equal(t, TenantStatusSuspended, tenant.Status)

	// Test invalid status
	err = tenant.ChangeStatus("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant status")
}

func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   TenantStatus
		expected bool
	}{
		{"Active status", TenantStatusActive, true},
		{"Inactive status", TenantStatusInactive, false},
		{"Suspended status", TenantStatusSuspended, false},
		{"Trial status", TenantStatusTrial, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				ID:     uuid.New(),
				Name:   "Test Company",
				Status: tt.status,
			}

			result := tenant.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenant_HasFeature(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Test Company",
		Status: TenantStatusActive,
		Configuration: TenantConfiguration{
			FeatureFlags: map[string]bool{
				"advanced_reporting": true,
				"basic_feature":      false,
			},
		},
	}

	assert.True(t, tenant.HasFeature("advanced_reporting"))
	assert.False(t, tenant.HasFeature("basic_feature"))
	assert.False(t, tenant.HasFeature("non_existent_feature"))
}

func TestTenant_EnableDisableFeature(t *testing.T) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Test Company",
		Status: TenantStatusActive,
		Configuration: TenantConfiguration{
			FeatureFlags: make(map[string]bool),
		},
	}

	tenant.EnableFeature("advanced_reporting")
	assert.True(t, tenant.HasFeature("advanced_reporting"))

	tenant.DisableFeature("advanced_reporting")
	assert.False(t, tenant.HasFeature("advanced_reporting"))
}

func TestGenerateSlugFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple name", "Test Company", "test-company"},
		{"Name with special characters", "Test & Co., Inc!", "test-co-inc"},
		{"Name with numbers", "Company 123", "company-123"},
		{"Name with multiple spaces", "Test   Company   Name", "test-company-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlugFromName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTenantStatus(t *testing.T) {
	validStatuses := []TenantStatus{
		TenantStatusActive,
		TenantStatusInactive,
		TenantStatusSuspended,
		TenantStatusTrial,
	}

	for _, status := range validStatuses {
		err := ValidateTenantStatus(status)
		assert.NoError(t, err)
	}

	// Test invalid status
	err := ValidateTenantStatus("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant status")
}