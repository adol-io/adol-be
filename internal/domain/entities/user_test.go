package entities

import (
	"testing"
	"time"

	"github.com/nicklaros/adol/pkg/errors"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		firstName   string
		lastName    string
		password    string
		role        UserRole
		expectError bool
	}{
		{
			name:        "Valid user creation",
			username:    "testuser",
			email:       "test@example.com",
			firstName:   "John",
			lastName:    "Doe",
			password:    "password123",
			role:        RoleEmployee,
			expectError: false,
		},
		{
			name:        "Empty username",
			username:    "",
			email:       "test@example.com",
			firstName:   "John",
			lastName:    "Doe",
			password:    "password123",
			role:        RoleEmployee,
			expectError: true,
		},
		{
			name:        "Short username",
			username:    "ab",
			email:       "test@example.com",
			firstName:   "John",
			lastName:    "Doe",
			password:    "password123",
			role:        RoleEmployee,
			expectError: true,
		},
		{
			name:        "Invalid email",
			username:    "testuser",
			email:       "invalid-email",
			firstName:   "John",
			lastName:    "Doe",
			password:    "password123",
			role:        RoleEmployee,
			expectError: true,
		},
		{
			name:        "Short password",
			username:    "testuser",
			email:       "test@example.com",
			firstName:   "John",
			lastName:    "Doe",
			password:    "123",
			role:        RoleEmployee,
			expectError: true,
		},
		{
			name:        "Invalid role",
			username:    "testuser",
			email:       "test@example.com",
			firstName:   "John",
			lastName:    "Doe",
			password:    "password123",
			role:        UserRole("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.firstName, tt.lastName, tt.password, tt.role)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if user == nil {
					t.Errorf("Expected user to be created")
				}
				if user != nil {
					if user.Username != tt.username {
						t.Errorf("Expected username %s, got %s", tt.username, user.Username)
					}
					if user.Email != tt.email {
						t.Errorf("Expected email %s, got %s", tt.email, user.Email)
					}
					if user.Role != tt.role {
						t.Errorf("Expected role %s, got %s", tt.role, user.Role)
					}
					if user.Status != UserStatusActive {
						t.Errorf("Expected status to be active, got %s", user.Status)
					}
				}
			}
		})
	}
}

func TestUserValidatePassword(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "John", "Doe", "password123", RoleEmployee)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Correct password",
			password: "password123",
			expected: true,
		},
		{
			name:     "Incorrect password",
			password: "wrongpassword",
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.ValidatePassword(tt.password)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUserUpdatePassword(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "John", "Doe", "password123", RoleEmployee)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name        string
		newPassword string
		expectError bool
	}{
		{
			name:        "Valid new password",
			newPassword: "newpassword123",
			expectError: false,
		},
		{
			name:        "Short password",
			newPassword: "123",
			expectError: true,
		},
		{
			name:        "Empty password",
			newPassword: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.UpdatePassword(tt.newPassword)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify the password was updated
				if !user.ValidatePassword(tt.newPassword) {
					t.Errorf("Password was not updated correctly")
				}
			}
		})
	}
}

func TestUserPermissions(t *testing.T) {
	tests := []struct {
		name              string
		role              UserRole
		canManageUsers    bool
		canManageProducts bool
		canProcessSales   bool
	}{
		{
			name:              "Admin permissions",
			role:              RoleAdmin,
			canManageUsers:    true,
			canManageProducts: true,
			canProcessSales:   true,
		},
		{
			name:              "Manager permissions",
			role:              RoleManager,
			canManageUsers:    true,
			canManageProducts: true,
			canProcessSales:   true,
		},
		{
			name:              "Cashier permissions",
			role:              RoleCashier,
			canManageUsers:    false,
			canManageProducts: false,
			canProcessSales:   true,
		},
		{
			name:              "Employee permissions",
			role:              RoleEmployee,
			canManageUsers:    false,
			canManageProducts: false,
			canProcessSales:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser("testuser", "test@example.com", "John", "Doe", "password123", tt.role)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			if user.CanManageUsers() != tt.canManageUsers {
				t.Errorf("Expected CanManageUsers %v, got %v", tt.canManageUsers, user.CanManageUsers())
			}
			if user.CanManageProducts() != tt.canManageProducts {
				t.Errorf("Expected CanManageProducts %v, got %v", tt.canManageProducts, user.CanManageProducts())
			}
			if user.CanProcessSales() != tt.canProcessSales {
				t.Errorf("Expected CanProcessSales %v, got %v", tt.canProcessSales, user.CanProcessSales())
			}
		})
	}
}

func TestUserStatusChanges(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "John", "Doe", "password123", RoleEmployee)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test activating user (should already be active)
	err = user.ChangeStatus(UserStatusActive)
	if err != nil {
		t.Errorf("Unexpected error changing to active: %v", err)
	}
	if user.Status != UserStatusActive {
		t.Errorf("Expected status to be active")
	}

	// Test deactivating user
	err = user.ChangeStatus(UserStatusInactive)
	if err != nil {
		t.Errorf("Unexpected error changing to inactive: %v", err)
	}
	if user.Status != UserStatusInactive {
		t.Errorf("Expected status to be inactive")
	}
	if user.IsActive() {
		t.Errorf("User should not be active")
	}

	// Test suspending user
	err = user.ChangeStatus(UserStatusSuspended)
	if err != nil {
		t.Errorf("Unexpected error changing to suspended: %v", err)
	}
	if user.Status != UserStatusSuspended {
		t.Errorf("Expected status to be suspended")
	}

	// Test invalid status
	err = user.ChangeStatus(UserStatus("invalid"))
	if err == nil {
		t.Errorf("Expected error for invalid status")
	}
}

func TestUserLastLogin(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "John", "Doe", "password123", RoleEmployee)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Initially should be nil
	if user.LastLoginAt != nil {
		t.Errorf("Expected LastLoginAt to be nil initially")
	}

	// Update last login
	before := time.Now()
	user.UpdateLastLogin()
	after := time.Now()

	if user.LastLoginAt == nil {
		t.Errorf("Expected LastLoginAt to be set")
	}
	if user.LastLoginAt.Before(before) || user.LastLoginAt.After(after) {
		t.Errorf("LastLoginAt time is not within expected range")
	}
}

func TestValidateUserRole(t *testing.T) {
	tests := []struct {
		name        string
		role        UserRole
		expectError bool
	}{
		{name: "Valid admin role", role: RoleAdmin, expectError: false},
		{name: "Valid manager role", role: RoleManager, expectError: false},
		{name: "Valid cashier role", role: RoleCashier, expectError: false},
		{name: "Valid employee role", role: RoleEmployee, expectError: false},
		{name: "Invalid role", role: UserRole("invalid"), expectError: true},
		{name: "Empty role", role: UserRole(""), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserRole(tt.role)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if appErr, ok := errors.IsAppError(err); ok {
					if appErr.Type != errors.ErrorTypeValidation {
						t.Errorf("Expected validation error, got %s", appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateUserStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      UserStatus
		expectError bool
	}{
		{name: "Valid active status", status: UserStatusActive, expectError: false},
		{name: "Valid inactive status", status: UserStatusInactive, expectError: false},
		{name: "Valid suspended status", status: UserStatusSuspended, expectError: false},
		{name: "Invalid status", status: UserStatus("invalid"), expectError: true},
		{name: "Empty status", status: UserStatus(""), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserStatus(tt.status)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
