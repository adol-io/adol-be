package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/usecases"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/utils"
)

// listUsers handles listing users with pagination and filtering
func (s *Server) listUsers(c *gin.Context) {
	// Check permission
	if err := s.checkPermission(c, "users", "read"); err != nil {
		s.respondWithError(c, err)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	pagination := utils.PaginationInfo{
		Page:  page,
		Limit: limit,
	}

	// Parse filter parameters
	filter := repositories.UserFilter{
		Search:   c.Query("search"),
		OrderBy:  c.DefaultQuery("order_by", "created_at"),
		OrderDir: c.DefaultQuery("order_dir", "DESC"),
	}

	if role := c.Query("role"); role != "" {
		userRole := entities.UserRole(role)
		filter.Role = &userRole
	}

	if status := c.Query("status"); status != "" {
		userStatus := entities.UserStatus(status)
		filter.Status = &userStatus
	}

	// TODO: Use actual user use case
	// For now, return mock response
	_ = filter
	_ = pagination
	response := &usecases.UserListResponse{
		Users: []*usecases.UserResponse{
			{
				ID:        uuid.New(),
				Username:  "admin",
				Email:     "admin@example.com",
				FirstName: "Admin",
				LastName:  "User",
				Role:      entities.RoleAdmin,
				Status:    entities.UserStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		Pagination: utils.PaginationInfo{
			Page:       page,
			Limit:      limit,
			TotalCount: 1,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// createUser handles creating a new user
func (s *Server) createUser(c *gin.Context) {
	// Check permission
	if err := s.checkPermission(c, "users", "create"); err != nil {
		s.respondWithError(c, err)
		return
	}

	adminID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	var req usecases.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// TODO: Use actual user use case
	_ = adminID
	_ = req

	// Mock response
	response := &usecases.UserResponse{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		Status:    entities.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// getUser handles retrieving a user by ID
func (s *Server) getUser(c *gin.Context) {
	// Check permission
	if err := s.checkPermission(c, "users", "read"); err != nil {
		s.respondWithError(c, err)
		return
	}

	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid user ID", "user ID must be a valid UUID"))
		return
	}

	// TODO: Use actual user use case
	_ = userID

	// Mock response
	response := &usecases.UserResponse{
		ID:        userID,
		Username:  "user",
		Email:     "user@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      entities.RoleEmployee,
		Status:    entities.UserStatusActive,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// updateUser handles updating a user
func (s *Server) updateUser(c *gin.Context) {
	// Check permission
	if err := s.checkPermission(c, "users", "update"); err != nil {
		s.respondWithError(c, err)
		return
	}

	adminID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid user ID", "user ID must be a valid UUID"))
		return
	}

	var req usecases.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// TODO: Use actual user use case
	_ = adminID
	_ = userID
	_ = req

	// Mock response
	response := &usecases.UserResponse{
		ID:        userID,
		Username:  "user",
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      entities.RoleEmployee,
		Status:    entities.UserStatusActive,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// deleteUser handles deleting a user
func (s *Server) deleteUser(c *gin.Context) {
	// Check permission
	if err := s.checkPermission(c, "users", "delete"); err != nil {
		s.respondWithError(c, err)
		return
	}

	adminID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid user ID", "user ID must be a valid UUID"))
		return
	}

	// TODO: Use actual user use case
	_ = adminID
	_ = userID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

// activateUser handles activating a user
func (s *Server) activateUser(c *gin.Context) {
	s.changeUserStatus(c, "activate")
}

// deactivateUser handles deactivating a user
func (s *Server) deactivateUser(c *gin.Context) {
	s.changeUserStatus(c, "deactivate")
}

// suspendUser handles suspending a user
func (s *Server) suspendUser(c *gin.Context) {
	s.changeUserStatus(c, "suspend")
}

// changeUserStatus is a helper function to change user status
func (s *Server) changeUserStatus(c *gin.Context, action string) {
	// Check permission
	if err := s.checkPermission(c, "users", "update"); err != nil {
		s.respondWithError(c, err)
		return
	}

	adminID, err := s.getCurrentUser(c)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid user ID", "user ID must be a valid UUID"))
		return
	}

	// TODO: Use actual user use case
	_ = adminID
	_ = userID
	_ = action

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User " + action + "d successfully",
	})
}
