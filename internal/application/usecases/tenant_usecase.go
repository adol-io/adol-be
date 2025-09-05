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

// TenantUseCase handles tenant-related operations
type TenantUseCase struct {
	tenantRepo       repositories.TenantRepository
	subscriptionRepo repositories.TenantSubscriptionRepository
	userRepo         repositories.UserRepository
	settingRepo      repositories.TenantSettingRepository
	tenantAuthService services.TenantAuthService
	audit           ports.AuditPort
	logger          logger.Logger
}

// NewTenantUseCase creates a new tenant use case
func NewTenantUseCase(
	tenantRepo repositories.TenantRepository,
	subscriptionRepo repositories.TenantSubscriptionRepository,
	userRepo repositories.UserRepository,
	settingRepo repositories.TenantSettingRepository,
	tenantAuthService services.TenantAuthService,
	audit ports.AuditPort,
	logger logger.Logger,
) *TenantUseCase {
	return &TenantUseCase{
		tenantRepo:       tenantRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		settingRepo:      settingRepo,
		tenantAuthService: tenantAuthService,
		audit:           audit,
		logger:          logger,
	}
}

// RegisterTenantRequest represents a tenant registration request
type RegisterTenantRequest struct {
	TenantName       string `json:"tenant_name" validate:"required"`
	Domain           string `json:"domain,omitempty"`
	AdminEmail       string `json:"admin_email" validate:"required,email"`
	AdminPassword    string `json:"admin_password" validate:"required,min=8"`
	AdminFirstName   string `json:"admin_first_name" validate:"required"`
	AdminLastName    string `json:"admin_last_name" validate:"required"`
	SubscriptionPlan string `json:"subscription_plan,omitempty"` // defaults to "starter"
	IPAddress        string `json:"ip_address,omitempty"`
	UserAgent        string `json:"user_agent,omitempty"`
}

// RegisterTenantResponse represents a tenant registration response
type RegisterTenantResponse struct {
	Tenant       *entities.Tenant             `json:"tenant"`
	AdminUser    *entities.User               `json:"admin_user"`
	Subscription *entities.TenantSubscription `json:"subscription"`
	AccessToken  string                       `json:"access_token"`
	RefreshToken string                       `json:"refresh_token"`
	ExpiresAt    time.Time                    `json:"expires_at"`
}

// GetTenantRequest represents a get tenant request
type GetTenantRequest struct {
	TenantID   *uuid.UUID `json:"tenant_id,omitempty"`
	TenantSlug *string    `json:"tenant_slug,omitempty"`
	Domain     *string    `json:"domain,omitempty"`
}

// UpdateTenantRequest represents an update tenant request
type UpdateTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Domain   string    `json:"domain,omitempty"`
}

// ListTenantsRequest represents a list tenants request
type ListTenantsRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit" validate:"min=1,max=100"`
}

// ListTenantsResponse represents a list tenants response
type ListTenantsResponse struct {
	Tenants []*entities.Tenant `json:"tenants"`
	Total   int                `json:"total"`
	Offset  int                `json:"offset"`
	Limit   int                `json:"limit"`
}

// UpdateTenantSettingsRequest represents an update tenant settings request
type UpdateTenantSettingsRequest struct {
	TenantID uuid.UUID                `json:"tenant_id" validate:"required"`
	Settings map[string]interface{}   `json:"settings" validate:"required"`
}

// RegisterTenant registers a new tenant with admin user and subscription
func (uc *TenantUseCase) RegisterTenant(ctx context.Context, req RegisterTenantRequest) (*RegisterTenantResponse, error) {
	// Audit logging
	defer func() {
		auditEvent := ports.AuditEvent{
			Action:    "tenant_register",
			Resource:  "tenant",
			Timestamp: time.Now(),
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		uc.audit.Log(ctx, auditEvent)
	}()

	// Validate inputs
	if req.TenantName == "" {
		return nil, errors.NewValidationError("tenant name is required", "tenant_name cannot be empty")
	}
	if req.AdminEmail == "" {
		return nil, errors.NewValidationError("admin email is required", "admin_email cannot be empty")
	}

	// Check if tenant slug or domain already exists
	slug := generateSlugFromName(req.TenantName)
	exists, err := uc.tenantRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to check tenant slug existence")
		return nil, errors.NewInternalError("failed to validate tenant slug", err)
	}
	if exists {
		return nil, errors.NewValidationError("tenant name already taken", "a tenant with this name already exists")
	}

	if req.Domain != "" {
		exists, err := uc.tenantRepo.ExistsByDomain(ctx, req.Domain)
		if err != nil {
			uc.logger.WithError(err).Error("Failed to check tenant domain existence")
			return nil, errors.NewInternalError("failed to validate tenant domain", err)
		}
		if exists {
			return nil, errors.NewValidationError("domain already taken", "this domain is already in use")
		}
	}

	// Create tenant
	tenant, err := entities.NewTenant(req.TenantName, req.Domain, nil)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to create tenant entity")
		return nil, err
	}

	if err := uc.tenantRepo.Create(ctx, tenant); err != nil {
		uc.logger.WithError(err).WithField("tenant_name", req.TenantName).Error("Failed to create tenant")
		return nil, errors.NewInternalError("failed to create tenant", err)
	}

	// Create subscription
	planType := entities.PlanStarter
	if req.SubscriptionPlan != "" {
		switch req.SubscriptionPlan {
		case "professional":
			planType = entities.PlanProfessional
		case "enterprise":
			planType = entities.PlanEnterprise
		}
	}

	subscription, err := entities.NewTenantSubscription(tenant.ID, planType)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to create subscription entity")
		return nil, err
	}

	if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to create subscription")
		return nil, errors.NewInternalError("failed to create subscription", err)
	}

	// Create admin user
	adminUser, err := entities.NewUser(
		tenant.ID,
		req.AdminEmail, // Use email as username
		req.AdminEmail,
		req.AdminFirstName,
		req.AdminLastName,
		req.AdminPassword,
		entities.RoleAdmin,
	)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to create admin user entity")
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, adminUser); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to create admin user")
		return nil, errors.NewInternalError("failed to create admin user", err)
	}

	// Generate authentication tokens
	tenantContext := entities.NewTenantContext(tenant, subscription)
	tokenPair, err := uc.tenantAuthService.GenerateTenantToken(ctx, adminUser, tenantContext)
	if err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to generate tokens")
		return nil, errors.NewInternalError("failed to generate authentication tokens", err)
	}

	uc.logger.WithFields(logger.Fields{
		"tenant_id":   tenant.ID,
		"tenant_slug": tenant.Slug,
		"admin_email": req.AdminEmail,
	}).Info("Tenant registered successfully")

	return &RegisterTenantResponse{
		Tenant:       tenant,
		AdminUser:    adminUser,
		Subscription: subscription,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessExpiry,
	}, nil
}

// GetTenant retrieves tenant information
func (uc *TenantUseCase) GetTenant(ctx context.Context, req GetTenantRequest) (*entities.Tenant, error) {
	var tenant *entities.Tenant
	var err error

	switch {
	case req.TenantID != nil:
		tenant, err = uc.tenantRepo.GetByID(ctx, *req.TenantID)
	case req.TenantSlug != nil:
		tenant, err = uc.tenantRepo.GetBySlug(ctx, *req.TenantSlug)
	case req.Domain != nil:
		tenant, err = uc.tenantRepo.GetByDomain(ctx, *req.Domain)
	default:
		return nil, errors.NewValidationError("tenant identifier required", "must provide tenant_id, tenant_slug, or domain")
	}

	if err != nil {
		uc.logger.WithError(err).Error("Failed to get tenant")
		return nil, err
	}

	return tenant, nil
}

// UpdateTenant updates tenant information
func (uc *TenantUseCase) UpdateTenant(ctx context.Context, req UpdateTenantRequest, userID uuid.UUID) (*entities.Tenant, error) {
	// Get existing tenant
	tenant, err := uc.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Update tenant
	if err := tenant.UpdateTenant(req.Name, req.Domain); err != nil {
		return nil, err
	}

	// Save changes
	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to update tenant")
		return nil, errors.NewInternalError("failed to update tenant", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "tenant_update",
		Resource:   "tenant",
		ResourceID: tenant.ID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(logger.Fields{
		"tenant_id": tenant.ID,
		"user_id":   userID,
	}).Info("Tenant updated successfully")

	return tenant, nil
}

// ListTenants lists tenants with pagination
func (uc *TenantUseCase) ListTenants(ctx context.Context, req ListTenantsRequest) (*ListTenantsResponse, error) {
	// Set default limit
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	tenants, total, err := uc.tenantRepo.List(ctx, req.Offset, req.Limit)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to list tenants")
		return nil, errors.NewInternalError("failed to list tenants", err)
	}

	return &ListTenantsResponse{
		Tenants: tenants,
		Total:   total,
		Offset:  req.Offset,
		Limit:   req.Limit,
	}, nil
}

// GetTenantSettings retrieves tenant settings
func (uc *TenantUseCase) GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	settings, err := uc.settingRepo.GetSettings(ctx, tenantID)
	if err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant settings")
		return nil, errors.NewInternalError("failed to get tenant settings", err)
	}

	return settings, nil
}

// UpdateTenantSettings updates tenant settings
func (uc *TenantUseCase) UpdateTenantSettings(ctx context.Context, req UpdateTenantSettingsRequest, userID uuid.UUID) error {
	// Validate tenant exists
	_, err := uc.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	// Update each setting
	for key, value := range req.Settings {
		if err := uc.settingRepo.Upsert(ctx, req.TenantID, key, value); err != nil {
			uc.logger.WithError(err).WithFields(logger.Fields{
				"tenant_id": req.TenantID,
				"key":       key,
			}).Error("Failed to update tenant setting")
			return errors.NewInternalError("failed to update tenant setting", err)
		}
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "tenant_settings_update",
		Resource:   "tenant_settings",
		ResourceID: req.TenantID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(logger.Fields{
		"tenant_id": req.TenantID,
		"user_id":   userID,
	}).Info("Tenant settings updated successfully")

	return nil
}

// ActivateTenant activates a tenant
func (uc *TenantUseCase) ActivateTenant(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	tenant, err := uc.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	if err := tenant.ActivateTenant(); err != nil {
		return err
	}

	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to activate tenant")
		return errors.NewInternalError("failed to activate tenant", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "tenant_activate",
		Resource:   "tenant",
		ResourceID: tenant.ID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	return nil
}

// SuspendTenant suspends a tenant
func (uc *TenantUseCase) SuspendTenant(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	tenant, err := uc.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	if err := tenant.SuspendTenant(); err != nil {
		return err
	}

	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to suspend tenant")
		return errors.NewInternalError("failed to suspend tenant", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "tenant_suspend",
		Resource:   "tenant",
		ResourceID: tenant.ID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	return nil
}

// Helper functions

func generateSlugFromName(name string) string {
	return entities.GenerateSlugFromName(name)
}