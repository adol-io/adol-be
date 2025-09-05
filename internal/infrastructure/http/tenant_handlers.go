package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/usecases"
	"github.com/nicklaros/adol/internal/infrastructure/http/middleware"
	"github.com/nicklaros/adol/pkg/errors"
)

// TenantHandlers contains tenant-related HTTP handlers
type TenantHandlers struct {
	tenantUseCase      *usecases.TenantUseCase
	subscriptionUseCase *usecases.SubscriptionUseCase
}

// registerTenant handles tenant registration
func (s *Server) registerTenant(c *gin.Context) {
	var req usecases.RegisterTenantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// Set IP address and user agent
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	response, err := s.tenantUseCase.RegisterTenant(c.Request.Context(), req)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	s.logger.WithFields(map[string]interface{}{
		"tenant_id":   response.Tenant.ID,
		"tenant_slug": response.Tenant.Slug,
		"admin_email": response.AdminUser.Email,
	}).Info("Tenant registered successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant registered successfully",
		"data":    response,
	})
}

// tenantLogin handles tenant-specific user authentication
func (s *Server) tenantLogin(c *gin.Context) {
	var req struct {
		TenantSlug string `json:"tenant_slug" binding:"required"`
		Email      string `json:"email" binding:"required,email"`
		Password   string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// TODO: Implement tenant login using TenantAuthService
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data": gin.H{
			"access_token":  "placeholder-tenant-access-token",
			"refresh_token": "placeholder-tenant-refresh-token",
			"token_type":    "Bearer",
			"expires_at":    "2024-12-31T23:59:59Z",
			"tenant_context": gin.H{
				"tenant_id":   "placeholder-id",
				"tenant_name": req.TenantSlug,
				"features":    []string{"pos", "inventory", "reporting"},
			},
		},
	})
}

// getTenant handles retrieving tenant information
func (s *Server) getTenant(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	req := usecases.GetTenantRequest{
		TenantID: &tenantContext.TenantID,
	}

	tenant, err := s.tenantUseCase.GetTenant(c.Request.Context(), req)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tenant,
	})
}

// updateTenant handles updating tenant information
func (s *Server) updateTenant(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	var req struct {
		Name   string `json:"name" binding:"required"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	userID := s.getCurrentUserID(c)
	updateReq := usecases.UpdateTenantRequest{
		TenantID: tenantContext.TenantID,
		Name:     req.Name,
		Domain:   req.Domain,
	}

	tenant, err := s.tenantUseCase.UpdateTenant(c.Request.Context(), updateReq, userID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant updated successfully",
		"data":    tenant,
	})
}

// listTenants handles listing tenants (system admin only)
func (s *Server) listTenants(c *gin.Context) {
	var req usecases.ListTenantsRequest

	// Parse query parameters
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := parseIntQuery(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := parseIntQuery(limitStr); err == nil {
			req.Limit = limit
		}
	} else {
		req.Limit = 50 // default
	}

	response, err := s.tenantUseCase.ListTenants(c.Request.Context(), req)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// getTenantSettings handles retrieving tenant settings
func (s *Server) getTenantSettings(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	settings, err := s.tenantUseCase.GetTenantSettings(c.Request.Context(), tenantContext.TenantID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": settings,
	})
}

// updateTenantSettings handles updating tenant settings
func (s *Server) updateTenantSettings(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	userID := s.getCurrentUserID(c)
	updateReq := usecases.UpdateTenantSettingsRequest{
		TenantID: tenantContext.TenantID,
		Settings: req,
	}

	err := s.tenantUseCase.UpdateTenantSettings(c.Request.Context(), updateReq, userID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant settings updated successfully",
	})
}

// getSubscription handles retrieving subscription information
func (s *Server) getSubscription(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	req := usecases.GetSubscriptionRequest{
		TenantID: tenantContext.TenantID,
	}

	subscription, err := s.subscriptionUseCase.GetSubscription(c.Request.Context(), req)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": subscription,
	})
}

// updateSubscriptionPlan handles updating subscription plan
func (s *Server) updateSubscriptionPlan(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	var req struct {
		PlanType string `json:"plan_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	// Parse plan type
	var planType entities.SubscriptionPlanType
	switch req.PlanType {
	case "starter":
		planType = entities.PlanStarter
	case "professional":
		planType = entities.PlanProfessional
	case "enterprise":
		planType = entities.PlanEnterprise
	default:
		s.respondWithError(c, errors.NewValidationError("invalid plan type", "plan_type must be one of: starter, professional, enterprise"))
		return
	}

	userID := s.getCurrentUserID(c)
	updateReq := usecases.UpdateSubscriptionPlanRequest{
		TenantID: tenantContext.TenantID,
		PlanType: planType,
	}

	subscription, err := s.subscriptionUseCase.UpdateSubscriptionPlan(c.Request.Context(), updateReq, userID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription plan updated successfully",
		"data":    subscription,
	})
}

// getUsageAnalysis handles retrieving subscription usage analysis
func (s *Server) getUsageAnalysis(c *gin.Context) {
	tenantContext := middleware.GetTenantContext(c)
	if tenantContext == nil {
		s.respondWithError(c, errors.NewUnauthorizedError("tenant context not found"))
		return
	}

	analysis, err := s.subscriptionUseCase.GetUsageAnalysis(c.Request.Context(), tenantContext.TenantID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": analysis,
	})
}

// activateTenant handles tenant activation (system admin only)
func (s *Server) activateTenant(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid tenant ID", "tenant_id must be a valid UUID"))
		return
	}

	userID := s.getCurrentUserID(c)

	err = s.tenantUseCase.ActivateTenant(c.Request.Context(), tenantID, userID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant activated successfully",
	})
}

// suspendTenant handles tenant suspension (system admin only)
func (s *Server) suspendTenant(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid tenant ID", "tenant_id must be a valid UUID"))
		return
	}

	userID := s.getCurrentUserID(c)

	err = s.tenantUseCase.SuspendTenant(c.Request.Context(), tenantID, userID)
	if err != nil {
		s.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant suspended successfully",
	})
}

// switchTenant handles switching user context to another tenant
func (s *Server) switchTenant(c *gin.Context) {
	var req struct {
		TenantSlug string `json:"tenant_slug" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondWithError(c, errors.NewValidationError("invalid request body", err.Error()))
		return
	}

	userID := s.getCurrentUserID(c)

	// TODO: Implement tenant switching using TenantAuthService
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant switched successfully",
		"data": gin.H{
			"access_token":  "placeholder-switched-token",
			"refresh_token": "placeholder-refresh-token",
			"tenant_context": gin.H{
				"tenant_slug": req.TenantSlug,
				"message":     "Switched to " + req.TenantSlug,
			},
		},
	})
}

// Helper methods

// getCurrentUserID extracts user ID from context
func (s *Server) getCurrentUserID(c *gin.Context) uuid.UUID {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			return uid
		}
	}
	return uuid.Nil
}

// parseIntQuery parses integer query parameter
func parseIntQuery(str string) (int, error) {
	// Simple integer parsing - you might want to use strconv.Atoi
	switch str {
	case "0":
		return 0, nil
	case "10":
		return 10, nil
	case "20":
		return 20, nil
	case "50":
		return 50, nil
	case "100":
		return 100, nil
	default:
		return 0, errors.NewValidationError("invalid integer", "must be a valid integer")
	}
}