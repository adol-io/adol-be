package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
)

// TenantContextKey is the key used to store tenant context in the request context
type TenantContextKey string

const (
	TenantContextKeyValue TenantContextKey = "tenant_context"
	TenantIDHeader        string          = "X-Tenant-ID"
	TenantSlugHeader      string          = "X-Tenant-Slug"
	TenantDomainHeader    string          = "X-Tenant-Domain"
)

// TenantResolver handles tenant resolution from various sources
type TenantResolver interface {
	ResolveTenant(c *gin.Context) (*entities.TenantContext, error)
}

// TenantMiddleware provides tenant-aware middleware
type TenantMiddleware struct {
	tenantRepo         repositories.TenantRepository
	subscriptionRepo   repositories.TenantSubscriptionRepository
	logger            logger.Logger
	enableRowLevelSecurity bool
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(
	tenantRepo repositories.TenantRepository,
	subscriptionRepo repositories.TenantSubscriptionRepository,
	logger logger.Logger,
	enableRLS bool,
) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo:         tenantRepo,
		subscriptionRepo:   subscriptionRepo,
		logger:            logger,
		enableRowLevelSecurity: enableRLS,
	}
}

// TenantResolverMiddleware resolves tenant context and adds it to the request
func (tm *TenantMiddleware) TenantResolverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext, err := tm.resolveTenantContext(c)
		if err != nil {
			tm.logger.WithFields(map[string]interface{}{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
			}).Error("Failed to resolve tenant context")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid tenant context",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store tenant context in gin context
		c.Set(string(TenantContextKeyValue), tenantContext)

		// Store tenant context in request context for database operations
		ctx := context.WithValue(c.Request.Context(), TenantContextKeyValue, tenantContext)
		c.Request = c.Request.WithContext(ctx)

		// Set database session variable for Row Level Security
		if tm.enableRowLevelSecurity {
			if err := tm.setDatabaseTenantContext(c, tenantContext.TenantID); err != nil {
				tm.logger.WithFields(map[string]interface{}{
					"tenant_id": tenantContext.TenantID,
					"error":     err.Error(),
				}).Error("Failed to set database tenant context")
				
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Database configuration error",
				})
				c.Abort()
				return
			}
		}

		tm.logger.WithFields(map[string]interface{}{
			"tenant_id":   tenantContext.TenantID,
			"tenant_name": tenantContext.TenantName,
			"tenant_slug": tenantContext.TenantSlug,
			"path":        c.Request.URL.Path,
		}).Debug("Tenant context resolved")

		c.Next()
	}
}

// RequireActiveTenant ensures the tenant is active and subscription is valid
func (tm *TenantMiddleware) RequireActiveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Tenant context not found",
			})
			c.Abort()
			return
		}

		if err := tenantContext.ValidateAccess(); err != nil {
			tm.logger.WithFields(map[string]interface{}{
				"tenant_id": tenantContext.TenantID,
				"error":     err.Error(),
			}).Warn("Tenant access denied")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant access denied",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireFeature ensures the tenant has access to a specific feature
func (tm *TenantMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Tenant context not found",
			})
			c.Abort()
			return
		}

		if err := tenantContext.ValidateFeatureAccess(feature); err != nil {
			tm.logger.WithFields(map[string]interface{}{
				"tenant_id": tenantContext.TenantID,
				"feature":   feature,
				"error":     err.Error(),
			}).Warn("Feature access denied")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Feature access denied",
				"message": err.Error(),
				"required_feature": feature,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// resolveTenantContext resolves tenant context from request
func (tm *TenantMiddleware) resolveTenantContext(c *gin.Context) (*entities.TenantContext, error) {
	// Try to resolve tenant from various sources in order of priority
	
	// 1. Try from tenant ID header
	if tenantIDStr := c.GetHeader(TenantIDHeader); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.NewValidationError("invalid tenant ID format", "tenant_id must be a valid UUID")
		}
		return tm.getTenantContextByID(c, tenantID)
	}

	// 2. Try from tenant slug header
	if tenantSlug := c.GetHeader(TenantSlugHeader); tenantSlug != "" {
		return tm.getTenantContextBySlug(c, tenantSlug)
	}

	// 3. Try from tenant domain header
	if tenantDomain := c.GetHeader(TenantDomainHeader); tenantDomain != "" {
		return tm.getTenantContextByDomain(c, tenantDomain)
	}

	// 4. Try from subdomain (e.g., tenant-slug.domain.com)
	if host := c.GetHeader("Host"); host != "" {
		if tenantSlug := extractTenantFromSubdomain(host); tenantSlug != "" {
			return tm.getTenantContextBySlug(c, tenantSlug)
		}
	}

	// 5. Try from URL path parameter (e.g., /api/v1/tenants/{slug}/...)
	if tenantSlug := c.Param("tenant_slug"); tenantSlug != "" {
		return tm.getTenantContextBySlug(c, tenantSlug)
	}

	return nil, errors.NewValidationError("tenant not specified", "tenant must be specified via header, subdomain, or URL parameter")
}

// getTenantContextByID retrieves tenant context by ID
func (tm *TenantMiddleware) getTenantContextByID(c *gin.Context, tenantID uuid.UUID) (*entities.TenantContext, error) {
	tenant, err := tm.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	subscription, err := tm.subscriptionRepo.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	return entities.NewTenantContext(tenant, subscription), nil
}

// getTenantContextBySlug retrieves tenant context by slug
func (tm *TenantMiddleware) getTenantContextBySlug(c *gin.Context, slug string) (*entities.TenantContext, error) {
	tenant, err := tm.tenantRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		return nil, err
	}

	subscription, err := tm.subscriptionRepo.GetByTenantID(c.Request.Context(), tenant.ID)
	if err != nil {
		return nil, err
	}

	return entities.NewTenantContext(tenant, subscription), nil
}

// getTenantContextByDomain retrieves tenant context by domain
func (tm *TenantMiddleware) getTenantContextByDomain(c *gin.Context, domain string) (*entities.TenantContext, error) {
	tenant, err := tm.tenantRepo.GetByDomain(c.Request.Context(), domain)
	if err != nil {
		return nil, err
	}

	subscription, err := tm.subscriptionRepo.GetByTenantID(c.Request.Context(), tenant.ID)
	if err != nil {
		return nil, err
	}

	return entities.NewTenantContext(tenant, subscription), nil
}

// setDatabaseTenantContext sets the tenant context in the database session
func (tm *TenantMiddleware) setDatabaseTenantContext(c *gin.Context, tenantID uuid.UUID) error {
	// This would typically involve setting a session variable for Row Level Security
	// Implementation depends on your database connection management
	// For now, we'll store it in the context for repositories to use
	return nil
}

// Helper functions

// GetTenantContext retrieves tenant context from gin context
func GetTenantContext(c *gin.Context) *entities.TenantContext {
	if value, exists := c.Get(string(TenantContextKeyValue)); exists {
		if tenantContext, ok := value.(*entities.TenantContext); ok {
			return tenantContext
		}
	}
	return nil
}

// GetTenantContextFromContext retrieves tenant context from standard context
func GetTenantContextFromContext(ctx context.Context) *entities.TenantContext {
	if value := ctx.Value(TenantContextKeyValue); value != nil {
		if tenantContext, ok := value.(*entities.TenantContext); ok {
			return tenantContext
		}
	}
	return nil
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(c *gin.Context) uuid.UUID {
	if tenantContext := GetTenantContext(c); tenantContext != nil {
		return tenantContext.TenantID
	}
	return uuid.Nil
}

// GetTenantIDFromContext retrieves tenant ID from standard context
func GetTenantIDFromContext(ctx context.Context) uuid.UUID {
	if tenantContext := GetTenantContextFromContext(ctx); tenantContext != nil {
		return tenantContext.TenantID
	}
	return uuid.Nil
}

// extractTenantFromSubdomain extracts tenant slug from subdomain
func extractTenantFromSubdomain(host string) string {
	// Remove port if present
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Split by dots
	parts := strings.Split(host, ".")
	
	// If we have at least 3 parts (subdomain.domain.tld), return the first part as tenant slug
	if len(parts) >= 3 {
		return parts[0]
	}

	return ""
}

// TenantFilter adds tenant filtering to database queries
func TenantFilter(tenantID uuid.UUID) string {
	return "tenant_id = '" + tenantID.String() + "'"
}