package entities

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/nicklaros/adol/pkg/errors"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleCashier  UserRole = "cashier"
	RoleEmployee UserRole = "employee"
)

// UserStatus represents user status
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Role        UserRole   `json:"role"`
	Status      UserStatus `json:"status"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// NewUser creates a new user
func NewUser(username, email, firstName, lastName, password string, role UserRole) (*User, error) {
	if err := validateUserInput(username, email, firstName, lastName, password); err != nil {
		return nil, err
	}

	if err := ValidateUserRole(role); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, errors.NewInternalError("failed to hash password", err)
	}

	now := time.Now()
	user := &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		Status:       UserStatusActive,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return user, nil
}

// ValidatePassword checks if the provided password matches the user's password
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(newPassword string) error {
	if len(newPassword) < 8 {
		return errors.NewValidationError("password too short", "password must be at least 8 characters long")
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return errors.NewInternalError("failed to hash password", err)
	}

	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName, email string) error {
	if firstName == "" {
		return errors.NewValidationError("first name is required", "first_name cannot be empty")
	}
	if lastName == "" {
		return errors.NewValidationError("last name is required", "last_name cannot be empty")
	}
	if !isValidEmail(email) {
		return errors.NewValidationError("invalid email format", "email must be a valid email address")
	}

	u.FirstName = firstName
	u.LastName = lastName
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// ChangeStatus changes the user's status
func (u *User) ChangeStatus(status UserStatus) error {
	if err := ValidateUserStatus(status); err != nil {
		return err
	}

	u.Status = status
	u.UpdatedAt = time.Now()
	return nil
}

// ChangeRole changes the user's role
func (u *User) ChangeRole(role UserRole) error {
	if err := ValidateUserRole(role); err != nil {
		return err
	}

	u.Role = role
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanManageUsers checks if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanManageProducts checks if the user can manage products
func (u *User) CanManageProducts() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanProcessSales checks if the user can process sales
func (u *User) CanProcessSales() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleCashier
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// ValidateUserRole validates if the role is valid
func ValidateUserRole(role UserRole) error {
	switch role {
	case RoleAdmin, RoleManager, RoleCashier, RoleEmployee:
		return nil
	default:
		return errors.NewValidationError("invalid user role", "role must be one of: admin, manager, cashier, employee")
	}
}

// ValidateUserStatus validates if the status is valid
func ValidateUserStatus(status UserStatus) error {
	switch status {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return nil
	default:
		return errors.NewValidationError("invalid user status", "status must be one of: active, inactive, suspended")
	}
}

// Helper functions

func validateUserInput(username, email, firstName, lastName, password string) error {
	if username == "" {
		return errors.NewValidationError("username is required", "username cannot be empty")
	}
	if len(username) < 3 {
		return errors.NewValidationError("username too short", "username must be at least 3 characters long")
	}
	if email == "" {
		return errors.NewValidationError("email is required", "email cannot be empty")
	}
	if !isValidEmail(email) {
		return errors.NewValidationError("invalid email format", "email must be a valid email address")
	}
	if firstName == "" {
		return errors.NewValidationError("first name is required", "first_name cannot be empty")
	}
	if lastName == "" {
		return errors.NewValidationError("last name is required", "last_name cannot be empty")
	}
	if len(password) < 8 {
		return errors.NewValidationError("password too short", "password must be at least 8 characters long")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func isValidEmail(email string) bool {
	// Simple email validation - in production, you might want to use a more robust validation
	return len(email) > 0 && 
		   len(email) <= 254 && 
		   email[0] != '@' && 
		   email[len(email)-1] != '@' &&
		   countChar(email, '@') == 1
}

func countChar(s string, char byte) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == char {
			count++
		}
	}
	return count
}