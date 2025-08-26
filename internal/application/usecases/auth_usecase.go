package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/ports"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/internal/domain/services"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
)

// AuthUseCase handles authentication-related operations
type AuthUseCase struct {
	userRepo    repositories.UserRepository
	authService services.AuthService
	jwtService  services.JWTService
	cache       ports.CachePort
	audit       ports.AuditPort
	logger      logger.Logger
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(
	userRepo repositories.UserRepository,
	authService services.AuthService,
	jwtService services.JWTService,
	cache ports.CachePort,
	audit ports.AuditPort,
	logger logger.Logger,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		authService: authService,
		jwtService:  jwtService,
		cache:       cache,
		audit:       audit,
		logger:      logger,
	}
}

// LoginRequest represents login request
type LoginRequest struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User         *entities.User `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at"`
	TokenType    string         `json:"token_type"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	NewPassword string    `json:"new_password" validate:"required,min=8"`
}

// Login authenticates a user and returns JWT tokens
func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Audit log for login attempt
	defer func() {
		auditEvent := ports.AuditEvent{
			ID:        uuid.New(),
			Action:    "login",
			Resource:  "user",
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
			Timestamp: time.Now(),
		}
		
		// Get user for audit log (if exists)
		if user, err := uc.userRepo.GetByUsername(ctx, req.Username); err == nil {
			auditEvent.UserID = user.ID
			auditEvent.ResourceID = user.ID.String()
		}
		
		uc.audit.Log(ctx, auditEvent)
	}()

	// Get user by username
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		uc.logger.WithField("username", req.Username).Warn("Login attempt with invalid username")
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive() {
		uc.logger.WithField("user_id", user.ID).Warn("Login attempt with inactive user")
		return nil, errors.NewForbiddenError("user account is not active")
	}

	// Validate password
	if !user.ValidatePassword(req.Password) {
		uc.logger.WithField("user_id", user.ID).Warn("Login attempt with invalid password")
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(user)
	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate JWT tokens")
		return nil, errors.NewInternalError("failed to generate tokens", err)
	}

	// Update last login time
	user.UpdateLastLogin()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Warn("Failed to update last login time")
		// Don't fail the login for this
	}

	// Store user session in cache
	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"login_time": time.Now(),
		"ip_address": req.IPAddress,
		"user_agent": req.UserAgent,
	}
	
	if err := uc.cache.SetUserSession(ctx, user.ID, sessionData, 24*time.Hour); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Warn("Failed to store user session")
		// Don't fail the login for this
	}

	uc.logger.WithField("user_id", user.ID).Info("User logged in successfully")

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken refreshes an expired access token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := uc.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Warn("Invalid refresh token")
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		uc.logger.WithField("user_id", claims.UserID).Error("User not found for refresh token")
		return nil, errors.NewUnauthorizedError("user not found")
	}

	// Check if user is still active
	if !user.IsActive() {
		uc.logger.WithField("user_id", user.ID).Warn("Refresh token attempt with inactive user")
		return nil, errors.NewForbiddenError("user account is not active")
	}

	// Generate new token pair
	tokenPair, err := uc.jwtService.GenerateTokenPair(user)
	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate new JWT tokens")
		return nil, errors.NewInternalError("failed to generate tokens", err)
	}

	// Revoke old refresh token
	if err := uc.jwtService.RevokeToken(req.RefreshToken); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Warn("Failed to revoke old refresh token")
		// Don't fail the refresh for this
	}

	uc.logger.WithField("user_id", user.ID).Info("Token refreshed successfully")

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// Logout logs out a user and revokes tokens
func (uc *AuthUseCase) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	// Revoke access token
	if err := uc.jwtService.RevokeToken(accessToken); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Warn("Failed to revoke access token")
	}

	// Remove user session from cache
	if err := uc.cache.DeleteUserSession(ctx, userID); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Warn("Failed to remove user session")
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "logout",
		Resource:  "user",
		ResourceID: userID.String(),
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithField("user_id", userID).Info("User logged out successfully")
	return nil
}

// ValidateToken validates a JWT token and returns user information
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*entities.User, error) {
	// Check if token is revoked
	if revoked := uc.jwtService.IsTokenRevoked(token); revoked {
		return nil, errors.NewUnauthorizedError("token has been revoked")
	}

	// Validate token
	claims, err := uc.jwtService.ValidateAccessToken(token)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Warn("Invalid access token")
		return nil, errors.NewUnauthorizedError("invalid token")
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		uc.logger.WithField("user_id", claims.UserID).Error("User not found for token")
		return nil, errors.NewUnauthorizedError("user not found")
	}

	// Check if user is still active
	if !user.IsActive() {
		uc.logger.WithField("user_id", user.ID).Warn("Token validation for inactive user")
		return nil, errors.NewForbiddenError("user account is not active")
	}

	return user, nil
}

// ChangePassword changes a user's password
func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Validate old password
	if !user.ValidatePassword(req.OldPassword) {
		uc.logger.WithField("user_id", userID).Warn("Change password attempt with invalid old password")
		return errors.NewUnauthorizedError("invalid old password")
	}

	// Update password
	if err := user.UpdatePassword(req.NewPassword); err != nil {
		return err
	}

	// Save user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to update user password")
		return errors.NewInternalError("failed to update password", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "change_password",
		Resource:   "user",
		ResourceID: userID.String(),
		Timestamp:  time.Now(),
		Success:    true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithField("user_id", userID).Info("Password changed successfully")
	return nil
}

// ResetPassword resets a user's password (admin only)
func (uc *AuthUseCase) ResetPassword(ctx context.Context, adminID uuid.UUID, req ResetPasswordRequest) error {
	// Get admin user to check permissions
	admin, err := uc.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return errors.NewNotFoundError("admin user")
	}

	// Check if admin has permission to reset passwords
	if !admin.CanManageUsers() {
		return errors.NewForbiddenError("insufficient permissions to reset password")
	}

	// Get target user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Update password
	if err := user.UpdatePassword(req.NewPassword); err != nil {
		return err
	}

	// Save user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"user_id":  req.UserID,
			"admin_id": adminID,
			"error":    err.Error(),
		}).Error("Failed to reset user password")
		return errors.NewInternalError("failed to reset password", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     adminID,
		Action:     "reset_password",
		Resource:   "user",
		ResourceID: req.UserID.String(),
		Timestamp:  time.Now(),
		Success:    true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"user_id":  req.UserID,
		"admin_id": adminID,
	}).Info("Password reset successfully")

	return nil
}

// CheckPermission checks if a user has permission to perform an action
func (uc *AuthUseCase) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	// Get user from cache first
	var sessionData map[string]interface{}
	if err := uc.cache.GetUserSession(ctx, userID, &sessionData); err == nil {
		if roleStr, ok := sessionData["role"].(string); ok {
			role := entities.UserRole(roleStr)
			return services.HasPermission(role, resource, action), nil
		}
	}

	// Fallback to database
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, errors.NewNotFoundError("user")
	}

	if !user.IsActive() {
		return false, errors.NewForbiddenError("user account is not active")
	}

	return services.HasPermission(user.Role, resource, action), nil
}