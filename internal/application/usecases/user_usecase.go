package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/ports"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
	"github.com/nicklaros/adol/pkg/utils"
)

// UserUseCase handles user management operations
type UserUseCase struct {
	userRepo repositories.UserRepository
	audit    ports.AuditPort
	logger   logger.Logger
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(
	userRepo repositories.UserRepository,
	audit ports.AuditPort,
	logger logger.Logger,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		audit:    audit,
		logger:   logger,
	}
}

// CreateUserRequest represents create user request
type CreateUserRequest struct {
	Username  string              `json:"username" validate:"required,min=3"`
	Email     string              `json:"email" validate:"required,email"`
	FirstName string              `json:"first_name" validate:"required"`
	LastName  string              `json:"last_name" validate:"required"`
	Password  string              `json:"password" validate:"required,min=8"`
	Role      entities.UserRole   `json:"role" validate:"required"`
	Status    entities.UserStatus `json:"status,omitempty"`
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	FirstName string               `json:"first_name,omitempty"`
	LastName  string               `json:"last_name,omitempty"`
	Email     string               `json:"email,omitempty"`
	Role      *entities.UserRole   `json:"role,omitempty"`
	Status    *entities.UserStatus `json:"status,omitempty"`
}

// UserResponse represents user response
type UserResponse struct {
	ID          uuid.UUID           `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	Role        entities.UserRole   `json:"role"`
	Status      entities.UserStatus `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	LastLoginAt *time.Time          `json:"last_login_at,omitempty"`
}

// UserListResponse represents user list response
type UserListResponse struct {
	Users      []*UserResponse      `json:"users"`
	Pagination utils.PaginationInfo `json:"pagination"`
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, adminID uuid.UUID, req CreateUserRequest) (*UserResponse, error) {
	// Check if username already exists
	exists, err := uc.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to check username existence")
		return nil, errors.NewInternalError("failed to check username", err)
	}
	if exists {
		return nil, errors.NewConflictError("username already exists")
	}

	// Check if email already exists
	exists, err = uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to check email existence")
		return nil, errors.NewInternalError("failed to check email", err)
	}
	if exists {
		return nil, errors.NewConflictError("email already exists")
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = entities.UserStatusActive
	}

	// Create user entity
	user, err := entities.NewUser(
		req.Username,
		req.Email,
		req.FirstName,
		req.LastName,
		req.Password,
		req.Role,
	)
	if err != nil {
		return nil, err
	}

	// Set status
	if err := user.ChangeStatus(status); err != nil {
		return nil, err
	}

	// Save user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
			"error":    err.Error(),
		}).Error("Failed to create user")
		return nil, errors.NewInternalError("failed to create user", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     adminID,
		Action:     "create",
		Resource:   "user",
		ResourceID: user.ID.String(),
		NewValue: map[string]interface{}{
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"status":     user.Status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"admin_id": adminID,
	}).Info("User created successfully")

	return uc.toUserResponse(user), nil
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return uc.toUserResponse(user), nil
}

// GetUserByUsername retrieves a user by username
func (uc *UserUseCase) GetUserByUsername(ctx context.Context, username string) (*UserResponse, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return uc.toUserResponse(user), nil
}

// UpdateUser updates an existing user
func (uc *UserUseCase) UpdateUser(ctx context.Context, adminID, userID uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Store old values for audit log
	oldValue := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
		"status":     user.Status,
	}

	// Update profile if provided
	if req.FirstName != "" || req.LastName != "" || req.Email != "" {
		firstName := user.FirstName
		lastName := user.LastName
		email := user.Email

		if req.FirstName != "" {
			firstName = req.FirstName
		}
		if req.LastName != "" {
			lastName = req.LastName
		}
		if req.Email != "" {
			// Check if new email already exists (for different user)
			if req.Email != user.Email {
				exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
				if err != nil {
					uc.logger.WithField("error", err.Error()).Error("Failed to check email existence")
					return nil, errors.NewInternalError("failed to check email", err)
				}
				if exists {
					return nil, errors.NewConflictError("email already exists")
				}
			}
			email = req.Email
		}

		if err := user.UpdateProfile(firstName, lastName, email); err != nil {
			return nil, err
		}
	}

	// Update role if provided
	if req.Role != nil {
		if err := user.ChangeRole(*req.Role); err != nil {
			return nil, err
		}
	}

	// Update status if provided
	if req.Status != nil {
		if err := user.ChangeStatus(*req.Status); err != nil {
			return nil, err
		}
	}

	// Save user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to update user")
		return nil, errors.NewInternalError("failed to update user", err)
	}

	// New values for audit log
	newValue := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
		"status":     user.Status,
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     adminID,
		Action:     "update",
		Resource:   "user",
		ResourceID: userID.String(),
		OldValue:   oldValue,
		NewValue:   newValue,
		Timestamp:  time.Now(),
		Success:    true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"admin_id": adminID,
	}).Info("User updated successfully")

	return uc.toUserResponse(user), nil
}

// DeleteUser deletes a user (soft delete)
func (uc *UserUseCase) DeleteUser(ctx context.Context, adminID, userID uuid.UUID) error {
	// Get user to ensure it exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Prevent deleting yourself
	if adminID == userID {
		return errors.NewValidationError("cannot delete yourself", "admin cannot delete their own account")
	}

	// Soft delete user
	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to delete user")
		return errors.NewInternalError("failed to delete user", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     adminID,
		Action:     "delete",
		Resource:   "user",
		ResourceID: userID.String(),
		OldValue: map[string]interface{}{
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"status":     user.Status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"admin_id": adminID,
	}).Info("User deleted successfully")

	return nil
}

// ListUsers retrieves users with pagination and filtering
func (uc *UserUseCase) ListUsers(ctx context.Context, filter repositories.UserFilter, pagination utils.PaginationInfo) (*UserListResponse, error) {
	users, paginationResult, err := uc.userRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list users")
		return nil, errors.NewInternalError("failed to list users", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = uc.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Pagination: paginationResult,
	}, nil
}

// ActivateUser activates a user account
func (uc *UserUseCase) ActivateUser(ctx context.Context, adminID, userID uuid.UUID) error {
	return uc.changeUserStatus(ctx, adminID, userID, entities.UserStatusActive, "activate")
}

// DeactivateUser deactivates a user account
func (uc *UserUseCase) DeactivateUser(ctx context.Context, adminID, userID uuid.UUID) error {
	return uc.changeUserStatus(ctx, adminID, userID, entities.UserStatusInactive, "deactivate")
}

// SuspendUser suspends a user account
func (uc *UserUseCase) SuspendUser(ctx context.Context, adminID, userID uuid.UUID) error {
	return uc.changeUserStatus(ctx, adminID, userID, entities.UserStatusSuspended, "suspend")
}

// changeUserStatus is a helper method to change user status
func (uc *UserUseCase) changeUserStatus(ctx context.Context, adminID, userID uuid.UUID, status entities.UserStatus, action string) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	oldStatus := user.Status

	// Change status
	if err := user.ChangeStatus(status); err != nil {
		return err
	}

	// Save user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to change user status")
		return errors.NewInternalError("failed to change user status", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     adminID,
		Action:     action,
		Resource:   "user",
		ResourceID: userID.String(),
		OldValue: map[string]interface{}{
			"status": oldStatus,
		},
		NewValue: map[string]interface{}{
			"status": status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"admin_id": adminID,
		"action":   action,
		"status":   status,
	}).Info("User status changed successfully")

	return nil
}

// toUserResponse converts user entity to response
func (uc *UserUseCase) toUserResponse(user *entities.User) *UserResponse {
	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Role:        user.Role,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
