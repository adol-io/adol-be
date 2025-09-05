package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
)

// TenantAuthService defines the interface for multi-tenant authentication and authorization
type TenantAuthService interface {
	// TenantLogin authenticates a user within a tenant context
	TenantLogin(ctx context.Context, tenantSlug, email, password string) (*TenantAuthResponse, error)
	
	// RegisterTenant creates a new tenant with an admin user
	RegisterTenant(ctx context.Context, req *TenantRegistrationRequest) (*TenantRegistrationResponse, error)
	
	// RefreshTenantToken refreshes a tenant-scoped JWT token
	RefreshTenantToken(ctx context.Context, refreshToken string) (*TenantAuthResponse, error)
	
	// ValidateTenantToken validates a tenant-scoped JWT token
	ValidateTenantToken(ctx context.Context, token string) (*TenantTokenInfo, error)
	
	// LogoutFromTenant invalidates a tenant-scoped JWT token
	LogoutFromTenant(ctx context.Context, token string) error
	
	// GenerateTenantToken generates a tenant-scoped JWT token
	GenerateTenantToken(ctx context.Context, user *entities.User, tenantContext *entities.TenantContext) (*TenantTokenPair, error)
	
	// ValidateTenantAccess validates if a user has access to a tenant
	ValidateTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) (*entities.TenantContext, error)
	
	// CheckTenantPermission checks if a user has permission within a tenant context
	CheckTenantPermission(ctx context.Context, userID, tenantID uuid.UUID, resource, action string) (bool, error)
	
	// SwitchTenant switches user context to another tenant (if they have access)
	SwitchTenant(ctx context.Context, userID uuid.UUID, newTenantSlug string) (*TenantAuthResponse, error)
}

// TenantJWTService defines the interface for tenant-aware JWT operations
type TenantJWTService interface {
	// GenerateTenantTokenPair generates access and refresh tokens with tenant context
	GenerateTenantTokenPair(user *entities.User, tenantContext *entities.TenantContext) (*TenantTokenPair, error)
	
	// ValidateTenantAccessToken validates a tenant-aware access token
	ValidateTenantAccessToken(tokenString string) (*TenantJWTClaims, error)
	
	// ValidateTenantRefreshToken validates a tenant-aware refresh token
	ValidateTenantRefreshToken(tokenString string) (*TenantJWTClaims, error)
	
	// ExtractTenantInfoFromToken extracts user and tenant info from token
	ExtractTenantInfoFromToken(tokenString string) (*TenantTokenInfo, error)
	
	// RevokeTenantToken revokes a tenant-scoped token
	RevokeTenantToken(tokenString string) error
	
	// IsTenantTokenRevoked checks if a tenant token is revoked
	IsTenantTokenRevoked(tokenString string) bool
}

// TenantAuthResponse represents the response from tenant authentication
type TenantAuthResponse struct {
	User           *entities.User         `json:"user"`
	TenantContext  *entities.TenantContext `json:"tenant_context"`
	AccessToken    string                 `json:"access_token"`
	RefreshToken   string                 `json:"refresh_token"`
	ExpiresAt      time.Time              `json:"expires_at"`
	TokenType      string                 `json:"token_type"`
}

// TenantTokenPair represents tenant-scoped access and refresh tokens
type TenantTokenPair struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	AccessExpiry  time.Time `json:"access_expiry"`
	RefreshExpiry time.Time `json:"refresh_expiry"`
	TenantID      uuid.UUID `json:"tenant_id"`
	TenantSlug    string    `json:"tenant_slug"`
}

// TenantJWTClaims represents JWT claims with tenant context
type TenantJWTClaims struct {
	UserID       uuid.UUID         `json:"user_id"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	Role         entities.UserRole `json:"role"`
	TenantID     uuid.UUID         `json:"tenant_id"`
	TenantSlug   string            `json:"tenant_slug"`
	TenantName   string            `json:"tenant_name"`
	TokenType    string            `json:"token_type"` // "access" or "refresh"
	Permissions  []string          `json:"permissions"`
	Features     []string          `json:"features"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Issuer       string            `json:"issuer"`
}

// TenantTokenInfo represents information extracted from a validated token
type TenantTokenInfo struct {
	UserID       uuid.UUID         `json:"user_id"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	Role         entities.UserRole `json:"role"`
	TenantID     uuid.UUID         `json:"tenant_id"`
	TenantSlug   string            `json:"tenant_slug"`
	TenantName   string            `json:"tenant_name"`
	Permissions  []string          `json:"permissions"`
	Features     []string          `json:"features"`
	IsValid      bool              `json:"is_valid"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

// TenantRegistrationRequest represents a tenant registration request
type TenantRegistrationRequest struct {
	TenantName       string `json:"tenant_name" binding:"required"`
	Domain           string `json:"domain,omitempty"`
	AdminEmail       string `json:"admin_email" binding:"required,email"`
	AdminPassword    string `json:"admin_password" binding:"required,min=8"`
	AdminFirstName   string `json:"admin_first_name" binding:"required"`
	AdminLastName    string `json:"admin_last_name" binding:"required"`
	SubscriptionPlan string `json:"subscription_plan,omitempty"` // defaults to "starter"
}

// TenantRegistrationResponse represents a tenant registration response
type TenantRegistrationResponse struct {
	Tenant        *entities.Tenant        `json:"tenant"`
	AdminUser     *entities.User          `json:"admin_user"`
	Subscription  *entities.TenantSubscription `json:"subscription"`
	AuthResponse  *TenantAuthResponse     `json:"auth_response"`
}

// TenantSwitchRequest represents a tenant switch request
type TenantSwitchRequest struct {
	TenantSlug string `json:"tenant_slug" binding:"required"`
}

// TenantPermission represents a tenant-specific permission
type TenantPermission struct {
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Restricted bool   `json:"restricted"` // Whether this permission is restricted by subscription
}

// Enhanced permission checking with tenant and subscription awareness
var (
	// Tenant Admin permissions - includes tenant management
	TenantAdminPermissions = append(AdminPermissions, []Permission{
		{"tenant", "read"},
		{"tenant", "update"},
		{"tenant_settings", "read"},
		{"tenant_settings", "update"},
		{"tenant_users", "create"},
		{"tenant_users", "read"},
		{"tenant_users", "update"},
		{"tenant_users", "delete"},
		{"subscription", "read"},
		{"subscription", "update"},
	}...)
	
	// System Admin permissions - cross-tenant access (for platform management)
	SystemAdminPermissions = append(TenantAdminPermissions, []Permission{
		{"tenants", "create"},
		{"tenants", "read"},
		{"tenants", "update"},
		{"tenants", "delete"},
		{"system_settings", "read"},
		{"system_settings", "update"},
		{"system_reports", "read"},
	}...)
)

// GetTenantPermissionsByRole returns permissions for a role within a tenant context
func GetTenantPermissionsByRole(role entities.UserRole, tenantContext *entities.TenantContext) []Permission {
	basePermissions := GetPermissionsByRole(role)
	
	// Add tenant-specific permissions for admin role
	if role == entities.RoleAdmin {
		basePermissions = append(basePermissions, []Permission{
			{"tenant", "read"},
			{"tenant", "update"},
			{"tenant_settings", "read"},
			{"tenant_settings", "update"},
			{"subscription", "read"},
		}...)
	}
	
	// Filter permissions based on subscription features
	if tenantContext != nil {
		return filterPermissionsBySubscription(basePermissions, tenantContext)
	}
	
	return basePermissions
}

// HasTenantPermission checks if a role has a specific permission within a tenant context
func HasTenantPermission(role entities.UserRole, tenantContext *entities.TenantContext, resource, action string) bool {
	permissions := GetTenantPermissionsByRole(role, tenantContext)
	for _, permission := range permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}
	return false
}

// filterPermissionsBySubscription filters permissions based on subscription features
func filterPermissionsBySubscription(permissions []Permission, tenantContext *entities.TenantContext) []Permission {
	if tenantContext == nil {
		return permissions
	}
	
	var filteredPermissions []Permission
	
	for _, permission := range permissions {
		// Check if permission requires specific features
		switch permission.Resource {
		case "reports":
			if permission.Action == "advanced" && !tenantContext.HasFeature("advanced_reporting") {
				continue // Skip this permission
			}
		case "api":
			if !tenantContext.HasFeature("api_access") {
				continue // Skip API permissions
			}
		case "integrations":
			if !tenantContext.HasFeature("custom_integration") {
				continue // Skip integration permissions
			}
		case "locations":
			if !tenantContext.HasFeature("multi_location") {
				continue // Skip multi-location permissions
			}
		}
		
		filteredPermissions = append(filteredPermissions, permission)
	}
	
	return filteredPermissions
}

// GetEnabledFeatures returns a list of enabled features for a tenant context
func GetEnabledFeatures(tenantContext *entities.TenantContext) []string {
	if tenantContext == nil {
		return []string{}
	}
	
	var features []string
	
	featureMap := map[string]bool{
		"pos":                tenantContext.Features.POS,
		"inventory":          tenantContext.Features.Inventory,
		"reporting":          tenantContext.Features.Reporting,
		"advanced_reporting": tenantContext.Features.AdvancedReporting,
		"multi_location":     tenantContext.Features.MultiLocation,
		"api_access":         tenantContext.Features.APIAccess,
		"custom_integration": tenantContext.Features.CustomIntegration,
	}
	
	for feature, enabled := range featureMap {
		if enabled {
			features = append(features, feature)
		}
	}
	
	return features
}

// GetPermissionStrings converts Permission slice to string slice
func GetPermissionStrings(permissions []Permission) []string {
	var permissionStrings []string
	for _, permission := range permissions {
		permissionStrings = append(permissionStrings, permission.Resource+":"+permission.Action)
	}
	return permissionStrings
}