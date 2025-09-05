package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/infrastructure/config"
	"github.com/nicklaros/adol/pkg/logger"
)

// MultiTenantIntegrationTestSuite provides integration testing for multi-tenant functionality
type MultiTenantIntegrationTestSuite struct {
	suite.Suite
	cfg    *config.Config
	logger logger.Logger
	
	// Test data
	testTenant1 entities.Tenant
	testTenant2 entities.Tenant
	testUser1   entities.User
	testUser2   entities.User
}

// SetupSuite initializes the test environment
func (suite *MultiTenantIntegrationTestSuite) SetupSuite() {
	// Load test configuration
	cfg, err := config.Load()
	suite.Require().NoError(err)
	
	// Override database name for testing
	cfg.Database.DBName = "adol_pos_test"
	suite.cfg = cfg

	// Initialize logger
	suite.logger = logger.NewLogger()

	// Setup test data
	suite.setupTestData()
}

// setupTestData creates test tenants and users
func (suite *MultiTenantIntegrationTestSuite) setupTestData() {
	// Create test tenants
	tenant1, err := entities.NewTenant("Test Company 1", "test1.example.com", nil)
	suite.Require().NoError(err)
	suite.testTenant1 = *tenant1

	tenant2, err := entities.NewTenant("Test Company 2", "test2.example.com", nil)
	suite.Require().NoError(err)
	suite.testTenant2 = *tenant2

	// Create test users
	user1, err := entities.NewUser(
		suite.testTenant1.ID,
		"admin1",
		"admin1@test1.example.com",
		"Admin",
		"One",
		"password123",
		entities.RoleAdmin,
	)
	suite.Require().NoError(err)
	suite.testUser1 = *user1

	user2, err := entities.NewUser(
		suite.testTenant2.ID,
		"admin2",
		"admin2@test2.example.com",
		"Admin",
		"Two",
		"password123",
		entities.RoleAdmin,
	)
	suite.Require().NoError(err)
	suite.testUser2 = *user2
}

// TestTenantIsolationLogic tests the logic for tenant isolation
func (suite *MultiTenantIntegrationTestSuite) TestTenantIsolationLogic() {
	// Create mock products for each tenant
	product1, err := entities.NewProduct(
		suite.testTenant1.ID,
		"PROD001",
		"Test Product 1", 
		"A test product",
		"Electronics",
		"pcs",
		decimal.NewFromFloat(99.99),
		decimal.NewFromFloat(50.00),
		10,
		suite.testUser1.ID,
	)
	suite.NoError(err)

	product2, err := entities.NewProduct(
		suite.testTenant2.ID,
		"PROD001", // Same SKU but different tenant
		"Test Product 2",
		"Another test product", 
		"Electronics",
		"pcs",
		decimal.NewFromFloat(149.99),
		decimal.NewFromFloat(75.00),
		15,
		suite.testUser2.ID,
	)
	suite.NoError(err)

	// Verify that products belong to different tenants
	suite.Equal(suite.testTenant1.ID, product1.TenantID)
	suite.Equal(suite.testTenant2.ID, product2.TenantID)
	
	// Even with the same SKU, they should be isolated by tenant
	suite.Equal("PROD001", product1.SKU)
	suite.Equal("PROD001", product2.SKU)
	suite.NotEqual(product1.TenantID, product2.TenantID)
}

// TestSubscriptionFeatureAccess tests subscription feature access logic
func (suite *MultiTenantIntegrationTestSuite) TestSubscriptionFeatureAccess() {
	// Create subscriptions for different plans
	starterSub, err := entities.NewTenantSubscription(suite.testTenant1.ID, entities.PlanStarter)
	suite.NoError(err)

	enterpriseSub, err := entities.NewTenantSubscription(suite.testTenant2.ID, entities.PlanEnterprise)
	suite.NoError(err)

	// Test feature access
	testFeatures := []string{"pos", "inventory", "reporting", "advanced_reporting", "api_access"}

	for _, feature := range testFeatures {
		starterHasFeature := starterSub.HasFeature(feature)
		enterpriseHasFeature := enterpriseSub.HasFeature(feature)

		switch feature {
		case "pos", "inventory", "reporting":
			// Basic features should be available to all plans
			suite.True(starterHasFeature, "Starter should have %s", feature)
			suite.True(enterpriseHasFeature, "Enterprise should have %s", feature)
		case "advanced_reporting":
			// Advanced features only for higher plans
			suite.False(starterHasFeature, "Starter should not have %s", feature)
			suite.True(enterpriseHasFeature, "Enterprise should have %s", feature)
		case "api_access":
			// Premium features only for enterprise
			suite.False(starterHasFeature, "Starter should not have %s", feature)
			suite.True(enterpriseHasFeature, "Enterprise should have %s", feature)
		}
	}
}

// TestUsageLimitValidation tests usage limit validation logic
func (suite *MultiTenantIntegrationTestSuite) TestUsageLimitValidation() {
	// Create a starter subscription (has user limits)
	subscription, err := entities.NewTenantSubscription(suite.testTenant1.ID, entities.PlanStarter)
	suite.NoError(err)

	// Test user creation up to limit
	for i := 0; i < 5; i++ {
		newUsage := entities.SubscriptionUsage{
			Users:    i + 1,
			Products: 10,
		}
		
		subscription.UpdateUsage(newUsage)
		
		// Check if limit is exceeded (starter plan allows 2 users)
		exceeded := subscription.IsUsageLimitExceeded("users")
		
		if i < 2 { // First 2 users should be allowed
			suite.False(exceeded, "Usage should not be exceeded for %d users", i+1)
		} else { // 3+ users should exceed limit
			suite.True(exceeded, "Usage should be exceeded for %d users", i+1)
		}
	}
}

// TestTenantConfigurationIsolation tests tenant configuration isolation
func (suite *MultiTenantIntegrationTestSuite) TestTenantConfigurationIsolation() {
	// Test that tenant configurations are isolated
	tenant1Config := entities.TenantConfiguration{
		BusinessInfo: entities.BusinessInfo{
			Name:     "Company 1",
			Currency: "USD",
			TaxRate:  0.08,
		},
		FeatureFlags: map[string]bool{
			"custom_feature": true,
		},
	}

	tenant2Config := entities.TenantConfiguration{
		BusinessInfo: entities.BusinessInfo{
			Name:     "Company 2", 
			Currency: "EUR",
			TaxRate:  0.20,
		},
		FeatureFlags: map[string]bool{
			"custom_feature": false,
		},
	}

	// Update tenant configurations
	err := suite.testTenant1.UpdateConfiguration(tenant1Config)
	suite.NoError(err)

	err = suite.testTenant2.UpdateConfiguration(tenant2Config)
	suite.NoError(err)

	// Verify configurations are different and isolated
	suite.Equal("USD", suite.testTenant1.Configuration.BusinessInfo.Currency)
	suite.Equal("EUR", suite.testTenant2.Configuration.BusinessInfo.Currency)
	suite.Equal(0.08, suite.testTenant1.Configuration.BusinessInfo.TaxRate)
	suite.Equal(0.20, suite.testTenant2.Configuration.BusinessInfo.TaxRate)

	suite.True(suite.testTenant1.HasFeature("custom_feature"))
	suite.False(suite.testTenant2.HasFeature("custom_feature"))
}

// TestMultiTenantIntegrationTestSuite runs the integration test suite
func TestMultiTenantIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MultiTenantIntegrationTestSuite))
}

// TestBasicMultiTenantFlow tests a basic multi-tenant flow
func TestBasicMultiTenantFlow(t *testing.T) {
	// Test tenant creation
	tenant, err := entities.NewTenant("Integration Test Company", "test.example.com", nil)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "Integration Test Company", tenant.Name)

	// Test subscription creation
	subscription, err := entities.NewTenantSubscription(tenant.ID, entities.PlanProfessional)
	assert.NoError(t, err)
	assert.NotNil(t, subscription)
	assert.Equal(t, tenant.ID, subscription.TenantID)
	assert.Equal(t, entities.PlanProfessional, subscription.PlanType)

	// Test user creation
	user, err := entities.NewUser(
		tenant.ID,
		"testuser",
		"test@example.com", 
		"Test",
		"User",
		"password123",
		entities.RoleManager,
	)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, tenant.ID, user.TenantID)

	// Test product creation
	product, err := entities.NewProduct(
		tenant.ID,
		"INT-001",
		"Integration Test Product",
		"A product for integration testing",
		"Test Category",
		"pcs",
		decimal.NewFromFloat(99.99),
		decimal.NewFromFloat(49.99),
		10,
		user.ID,
	)
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, tenant.ID, product.TenantID)
	assert.Equal(t, user.ID, product.CreatedBy)

	// Verify relationships
	assert.Equal(t, tenant.ID, user.TenantID)
	assert.Equal(t, tenant.ID, product.TenantID) 
	assert.Equal(t, tenant.ID, subscription.TenantID)
}

// TestTenantContextValidation tests tenant context validation
func TestTenantContextValidation(t *testing.T) {
	// Create test tenant
	tenant, err := entities.NewTenant("Context Test Company", "", nil)
	assert.NoError(t, err)

	// Create subscription 
	subscription, err := entities.NewTenantSubscription(tenant.ID, entities.PlanProfessional)
	assert.NoError(t, err)

	// Create tenant context
	tenantContext := entities.NewTenantContext(tenant, subscription)
	assert.NotNil(t, tenantContext)
	assert.Equal(t, tenant.ID, tenantContext.TenantID)
	assert.Equal(t, tenant.Name, tenantContext.TenantName)
	assert.Equal(t, tenant.Slug, tenantContext.TenantSlug)

	// Test access validation
	err = tenantContext.ValidateAccess()
	assert.NoError(t, err, "Active tenant with valid subscription should have access")

	// Test feature access validation
	err = tenantContext.ValidateFeatureAccess("advanced_reporting")
	assert.NoError(t, err, "Professional plan should have advanced reporting")

	err = tenantContext.ValidateFeatureAccess("api_access")
	assert.Error(t, err, "Professional plan should not have API access")
}