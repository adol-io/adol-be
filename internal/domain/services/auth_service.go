package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
)

// AuthService defines the interface for authentication and authorization
type AuthService interface {
	// Login authenticates a user and returns a JWT token
	Login(ctx context.Context, username, password string) (*AuthResponse, error)
	
	// RefreshToken refreshes an expired JWT token
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	
	// Logout invalidates a JWT token
	Logout(ctx context.Context, token string) error
	
	// ValidateToken validates a JWT token and returns user information
	ValidateToken(ctx context.Context, token string) (*entities.User, error)
	
	// GenerateToken generates a JWT token for a user
	GenerateToken(ctx context.Context, user *entities.User) (*TokenPair, error)
	
	// HashPassword hashes a password using bcrypt
	HashPassword(password string) (string, error)
	
	// ValidatePassword validates a password against a hash
	ValidatePassword(password, hash string) bool
	
	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	
	// ResetPassword resets a user's password (admin only)
	ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string, adminID uuid.UUID) error
	
	// CheckPermission checks if a user has permission to perform an action
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// JWTService defines the interface for JWT token operations
type JWTService interface {
	// GenerateTokenPair generates access and refresh tokens
	GenerateTokenPair(user *entities.User) (*TokenPair, error)
	
	// ValidateAccessToken validates an access token and returns claims
	ValidateAccessToken(tokenString string) (*JWTClaims, error)
	
	// ValidateRefreshToken validates a refresh token and returns claims
	ValidateRefreshToken(tokenString string) (*JWTClaims, error)
	
	// ExtractUserIDFromToken extracts user ID from a token
	ExtractUserIDFromToken(tokenString string) (uuid.UUID, error)
	
	// IsTokenExpired checks if a token is expired
	IsTokenExpired(tokenString string) bool
	
	// RevokeToken revokes a token (adds to blacklist)
	RevokeToken(tokenString string) error
	
	// IsTokenRevoked checks if a token is revoked
	IsTokenRevoked(tokenString string) bool
}

// AuthResponse represents the response from login/refresh operations
type AuthResponse struct {
	User         *entities.User `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at"`
	TokenType    string         `json:"token_type"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	AccessExpiry  time.Time `json:"access_expiry"`
	RefreshExpiry time.Time `json:"refresh_expiry"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    uuid.UUID         `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Role      entities.UserRole `json:"role"`
	TokenType string            `json:"token_type"` // "access" or "refresh"
	IssuedAt  time.Time         `json:"issued_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Issuer    string            `json:"issuer"`
}

// Permission represents a permission in the system
type Permission struct {
	Resource string `json:"resource"` // e.g., "users", "products", "sales"
	Action   string `json:"action"`   // e.g., "create", "read", "update", "delete"
}

// Role-based permissions
var (
	// Admin permissions - full access
	AdminPermissions = []Permission{
		// User management
		{"users", "create"},
		{"users", "read"},
		{"users", "update"},
		{"users", "delete"},
		
		// Product management
		{"products", "create"},
		{"products", "read"},
		{"products", "update"},
		{"products", "delete"},
		
		// Stock management
		{"stock", "create"},
		{"stock", "read"},
		{"stock", "update"},
		{"stock", "delete"},
		
		// Sales management
		{"sales", "create"},
		{"sales", "read"},
		{"sales", "update"},
		{"sales", "delete"},
		
		// Invoice management
		{"invoices", "create"},
		{"invoices", "read"},
		{"invoices", "update"},
		{"invoices", "delete"},
		
		// Reports
		{"reports", "read"},
		
		// System settings
		{"settings", "read"},
		{"settings", "update"},
	}
	
	// Manager permissions - most operations except user management
	ManagerPermissions = []Permission{
		// Product management
		{"products", "create"},
		{"products", "read"},
		{"products", "update"},
		{"products", "delete"},
		
		// Stock management
		{"stock", "create"},
		{"stock", "read"},
		{"stock", "update"},
		{"stock", "delete"},
		
		// Sales management
		{"sales", "create"},
		{"sales", "read"},
		{"sales", "update"},
		{"sales", "delete"},
		
		// Invoice management
		{"invoices", "create"},
		{"invoices", "read"},
		{"invoices", "update"},
		{"invoices", "delete"},
		
		// Reports
		{"reports", "read"},
		
		// Limited user operations
		{"users", "read"},
	}
	
	// Cashier permissions - sales and basic operations
	CashierPermissions = []Permission{
		// Product read access
		{"products", "read"},
		
		// Stock read access
		{"stock", "read"},
		
		// Sales management
		{"sales", "create"},
		{"sales", "read"},
		{"sales", "update"},
		
		// Invoice management
		{"invoices", "create"},
		{"invoices", "read"},
		
		// Limited reports
		{"reports", "read"},
	}
	
	// Employee permissions - read-only access
	EmployeePermissions = []Permission{
		// Product read access
		{"products", "read"},
		
		// Stock read access
		{"stock", "read"},
		
		// Sales read access
		{"sales", "read"},
		
		// Invoice read access
		{"invoices", "read"},
	}
)

// GetPermissionsByRole returns permissions for a specific role
func GetPermissionsByRole(role entities.UserRole) []Permission {
	switch role {
	case entities.RoleAdmin:
		return AdminPermissions
	case entities.RoleManager:
		return ManagerPermissions
	case entities.RoleCashier:
		return CashierPermissions
	case entities.RoleEmployee:
		return EmployeePermissions
	default:
		return []Permission{}
	}
}

// HasPermission checks if a role has a specific permission
func HasPermission(role entities.UserRole, resource, action string) bool {
	permissions := GetPermissionsByRole(role)
	for _, permission := range permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}
	return false
}