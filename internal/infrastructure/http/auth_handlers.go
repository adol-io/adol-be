package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/usecases"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/errors"
)

// AuthHandlers contains authentication-related HTTP handlers
type AuthHandlers struct {
	authUseCase *usecases.AuthUseCase
	userUseCase *usecases.UserUseCase
}

// login handles user login
func (s *Server) login(c *gin.Context) {
	var req usecases.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// Set IP address and user agent from request
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	// TODO: Use actual auth use case
	// For now, return mock response
	response := &usecases.LoginResponse{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		User: &entities.User{
			ID:        uuid.New(),
			Username:  req.Username,
			Email:     "user@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      entities.RoleAdmin,
			Status:    entities.UserStatusActive,
		},
	}

	s.logger.WithFields(map[string]interface{}{
		"username":   req.Username,
		"ip_address": req.IPAddress,
	}).Info("User login attempt")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// refreshToken handles token refresh
func (s *Server) refreshToken(c *gin.Context) {
	var req usecases.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// TODO: Use actual auth use case
	// For now, return mock response
	response := &usecases.LoginResponse{
		AccessToken:  "new-mock-access-token",
		RefreshToken: "new-mock-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		User: &entities.User{
			ID:        uuid.New(),
			Username:  "user",
			Email:     "user@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      entities.RoleAdmin,
			Status:    entities.UserStatusActive,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// logout handles user logout
func (s *Server) logout(c *gin.Context) {
	userID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	token, exists := c.Get("token")
	if !exists {
		s.respondWithError(c, errors.NewUnauthorizedError("token not found"))
		return
	}

	tokenStr, ok := token.(string)
	if !ok {
		s.respondWithError(c, errors.NewInternalError("invalid token format", nil))
		return
	}

	// TODO: Use actual auth use case to logout
	_ = userID
	_ = tokenStr

	s.logger.WithField("user_id", userID).Info("User logged out")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// changePassword handles password change
func (s *Server) changePassword(c *gin.Context) {
	userID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	var req usecases.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// TODO: Use actual auth use case
	// For now, return success
	_ = userID
	_ = req

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// resetPassword handles password reset (admin only)
func (s *Server) resetPassword(c *gin.Context) {
	adminID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	// Check admin permission
	if err := s.checkPermission(c, "users", "update"); err != nil {
		s.respondWithError(c, err)
		return
	}

	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid user ID", "user ID must be a valid UUID"))
		return
	}

	var req usecases.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	req.UserID = userID

	// TODO: Use actual auth use case
	_ = adminID
	_ = req

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}