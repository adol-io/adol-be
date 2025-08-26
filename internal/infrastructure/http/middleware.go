package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/errors"
)

// AuthMiddleware provides authentication middleware
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			s.respondWithError(c, errors.NewUnauthorizedError("authorization header required"))
			c.Abort()
			return
		}

		// Check Bearer token format
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			s.respondWithError(c, errors.NewUnauthorizedError("invalid authorization header format"))
			c.Abort()
			return
		}

		token := bearerToken[1]

		// TODO: Validate token using auth service
		// For now, we'll create a mock validation
		userID, err := s.validateToken(token)
		if err != nil {
			s.respondWithError(c, err)
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", userID)
		c.Set("token", token)

		c.Next()
	}
}

// PermissionMiddleware checks if user has required permission
func (s *Server) permissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			s.respondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
			c.Abort()
			return
		}

		// TODO: Check permission using auth service
		// For now, we'll allow all authenticated users
		_ = userID

		c.Next()
	}
}

// AdminOnlyMiddleware ensures only admin users can access the endpoint
func (s *Server) adminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			s.respondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
			c.Abort()
			return
		}

		// TODO: Check if user is admin using auth service
		// For now, we'll allow all authenticated users
		_ = userID

		c.Next()
	}
}

// validateToken validates JWT token and returns user ID
// TODO: This is a mock implementation, replace with actual JWT validation
func (s *Server) validateToken(token string) (uuid.UUID, error) {
	// Mock implementation - in real scenario, validate JWT token
	if token == "mock-token" {
		return uuid.New(), nil
	}
	return uuid.Nil, errors.NewUnauthorizedError("invalid token")
}

// getCurrentUser gets current user from context
func (s *Server) getCurrentUser(c *gin.Context) (uuid.UUID, error) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.NewUnauthorizedError("user not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.NewInternalError("invalid user ID format", nil)
	}

	return userID, nil
}

// getCurrentUserRole gets current user role from context
// TODO: This is a mock implementation
func (s *Server) getCurrentUserRole(c *gin.Context) (entities.UserRole, error) {
	// Mock implementation - in real scenario, get user role from token or database
	return entities.RoleAdmin, nil
}

// checkPermission checks if current user has required permission
func (s *Server) checkPermission(c *gin.Context, resource, action string) error {
	userRole, err := s.getCurrentUserRole(c)
	if err != nil {
		return err
	}

	// TODO: Use actual permission checking logic
	// For now, allow admin and manager for all operations
	if userRole == entities.RoleAdmin || userRole == entities.RoleManager {
		return nil
	}

	// Cashier can only process sales and view products
	if userRole == entities.RoleCashier {
		if (resource == "sales" && (action == "create" || action == "read" || action == "update")) ||
			(resource == "products" && action == "read") ||
			(resource == "stock" && action == "read") ||
			(resource == "invoices" && (action == "create" || action == "read")) {
			return nil
		}
	}

	// Employee has read-only access
	if userRole == entities.RoleEmployee {
		if action == "read" {
			return nil
		}
	}

	return errors.NewForbiddenError("insufficient permissions")
}

// respondWithError sends error response
func (s *Server) respondWithError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		c.JSON(appErr.Code, gin.H{
			"error": gin.H{
				"type":    appErr.Type,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
	}
}
