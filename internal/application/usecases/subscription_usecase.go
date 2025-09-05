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
)

// SubscriptionUseCase handles subscription-related operations
type SubscriptionUseCase struct {
	subscriptionRepo repositories.TenantSubscriptionRepository
	tenantRepo       repositories.TenantRepository
	userRepo         repositories.UserRepository
	audit           ports.AuditPort
	logger          logger.Logger
}

// NewSubscriptionUseCase creates a new subscription use case
func NewSubscriptionUseCase(
	subscriptionRepo repositories.TenantSubscriptionRepository,
	tenantRepo repositories.TenantRepository,
	userRepo repositories.UserRepository,
	audit ports.AuditPort,
	logger logger.Logger,
) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
		tenantRepo:       tenantRepo,
		userRepo:         userRepo,
		audit:           audit,
		logger:          logger,
	}
}

// GetSubscriptionRequest represents a get subscription request
type GetSubscriptionRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// UpdateSubscriptionPlanRequest represents an update subscription plan request
type UpdateSubscriptionPlanRequest struct {
	TenantID uuid.UUID                    `json:"tenant_id" validate:"required"`
	PlanType entities.SubscriptionPlanType `json:"plan_type" validate:"required"`
}

// UpdateUsageRequest represents an update usage request
type UpdateUsageRequest struct {
	TenantID uuid.UUID                 `json:"tenant_id" validate:"required"`
	Usage    entities.SubscriptionUsage `json:"usage" validate:"required"`
}

// SubscriptionStatusRequest represents a subscription status change request
type SubscriptionStatusRequest struct {
	TenantID uuid.UUID                 `json:"tenant_id" validate:"required"`
	Status   entities.SubscriptionStatus `json:"status" validate:"required"`
}

// SubscriptionUsageResponse represents subscription usage information
type SubscriptionUsageResponse struct {
	Subscription *entities.TenantSubscription `json:"subscription"`
	Usage        *UsageAnalysis               `json:"usage_analysis"`
	Limits       *LimitAnalysis               `json:"limit_analysis"`
}

// UsageAnalysis provides analysis of current usage
type UsageAnalysis struct {
	Users         int     `json:"users"`
	Products      int     `json:"products"`
	SalesThisMonth int     `json:"sales_this_month"`
	APICallsThisMonth int  `json:"api_calls_this_month"`
	UserUsagePercent     float64 `json:"user_usage_percent"`
	ProductUsagePercent  float64 `json:"product_usage_percent"`
	SalesUsagePercent    float64 `json:"sales_usage_percent"`
	APIUsagePercent      float64 `json:"api_usage_percent"`
}

// LimitAnalysis provides analysis of limits and restrictions
type LimitAnalysis struct {
	CanAddUser     bool   `json:"can_add_user"`
	CanAddProduct  bool   `json:"can_add_product"`
	CanProcessSale bool   `json:"can_process_sale"`
	CanMakeAPICall bool   `json:"can_make_api_call"`
	Warnings       []string `json:"warnings,omitempty"`
}

// GetSubscription retrieves subscription information for a tenant
func (uc *SubscriptionUseCase) GetSubscription(ctx context.Context, req GetSubscriptionRequest) (*entities.TenantSubscription, error) {
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to get subscription")
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscriptionPlan updates the subscription plan for a tenant
func (uc *SubscriptionUseCase) UpdateSubscriptionPlan(ctx context.Context, req UpdateSubscriptionPlanRequest, userID uuid.UUID) (*entities.TenantSubscription, error) {
	// Get current subscription
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Validate tenant exists
	_, err = uc.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Check if it's an upgrade or downgrade
	oldPlan := subscription.PlanType
	
	if uc.isPlanUpgrade(oldPlan, req.PlanType) {
		if err := subscription.UpgradePlan(req.PlanType); err != nil {
			return nil, err
		}
	} else {
		if err := subscription.DowngradePlan(req.PlanType); err != nil {
			return nil, err
		}
	}

	// Update subscription
	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		uc.logger.WithError(err).WithFields(logger.Fields{
			"tenant_id": req.TenantID,
			"old_plan":  oldPlan,
			"new_plan":  req.PlanType,
		}).Error("Failed to update subscription plan")
		return nil, errors.NewInternalError("failed to update subscription plan", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "subscription_plan_update",
		Resource:   "subscription",
		ResourceID: req.TenantID.String(),
		UserID:     userID,
		Details:    map[string]interface{}{
			"old_plan": oldPlan,
			"new_plan": req.PlanType,
		},
		Timestamp: time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(logger.Fields{
		"tenant_id": req.TenantID,
		"user_id":   userID,
		"old_plan":  oldPlan,
		"new_plan":  req.PlanType,
	}).Info("Subscription plan updated successfully")

	return subscription, nil
}

// ActivateSubscription activates a subscription
func (uc *SubscriptionUseCase) ActivateSubscription(ctx context.Context, req SubscriptionStatusRequest, userID uuid.UUID) error {
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	if err := subscription.ActivateSubscription(); err != nil {
		return err
	}

	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to activate subscription")
		return errors.NewInternalError("failed to activate subscription", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "subscription_activate",
		Resource:   "subscription",
		ResourceID: req.TenantID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	return nil
}

// SuspendSubscription suspends a subscription
func (uc *SubscriptionUseCase) SuspendSubscription(ctx context.Context, req SubscriptionStatusRequest, userID uuid.UUID) error {
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	if err := subscription.SuspendSubscription(); err != nil {
		return err
	}

	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to suspend subscription")
		return errors.NewInternalError("failed to suspend subscription", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "subscription_suspend",
		Resource:   "subscription",
		ResourceID: req.TenantID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	return nil
}

// CancelSubscription cancels a subscription
func (uc *SubscriptionUseCase) CancelSubscription(ctx context.Context, req SubscriptionStatusRequest, userID uuid.UUID) error {
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	if err := subscription.CancelSubscription(); err != nil {
		return err
	}

	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to cancel subscription")
		return errors.NewInternalError("failed to cancel subscription", err)
	}

	// Audit logging
	auditEvent := ports.AuditEvent{
		Action:     "subscription_cancel",
		Resource:   "subscription",
		ResourceID: req.TenantID.String(),
		UserID:     userID,
		Timestamp:  time.Now(),
	}
	uc.audit.Log(ctx, auditEvent)

	return nil
}

// UpdateUsage updates the usage statistics for a subscription
func (uc *SubscriptionUseCase) UpdateUsage(ctx context.Context, req UpdateUsageRequest) error {
	// Validate subscription exists
	_, err := uc.subscriptionRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	// Update usage
	if err := uc.subscriptionRepo.UpdateUsage(ctx, req.TenantID, req.Usage); err != nil {
		uc.logger.WithError(err).WithField("tenant_id", req.TenantID).Error("Failed to update subscription usage")
		return errors.NewInternalError("failed to update subscription usage", err)
	}

	return nil
}

// GetUsageAnalysis provides detailed usage analysis for a subscription
func (uc *SubscriptionUseCase) GetUsageAnalysis(ctx context.Context, tenantID uuid.UUID) (*SubscriptionUsageResponse, error) {
	subscription, err := uc.subscriptionRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Calculate usage percentages
	usageAnalysis := &UsageAnalysis{
		Users:         subscription.CurrentUsage.Users,
		Products:      subscription.CurrentUsage.Products,
		SalesThisMonth: subscription.CurrentUsage.SalesThisMonth,
		APICallsThisMonth: subscription.CurrentUsage.APICallsThisMonth,
		UserUsagePercent:     subscription.GetUsagePercentage("users"),
		ProductUsagePercent:  subscription.GetUsagePercentage("products"),
		SalesUsagePercent:    subscription.GetUsagePercentage("sales"),
		APIUsagePercent:      subscription.GetUsagePercentage("api_calls"),
	}

	// Calculate limit analysis
	limitAnalysis := &LimitAnalysis{
		CanAddUser:     subscription.CanAddUser(),
		CanAddProduct:  subscription.CanAddProduct(),
		CanProcessSale: subscription.CanProcessSale(),
		CanMakeAPICall: subscription.CanMakeAPICall(),
		Warnings:       uc.generateUsageWarnings(subscription),
	}

	return &SubscriptionUsageResponse{
		Subscription: subscription,
		Usage:        usageAnalysis,
		Limits:       limitAnalysis,
	}, nil
}

// CollectUsageStatistics collects and updates usage statistics for a tenant
func (uc *SubscriptionUseCase) CollectUsageStatistics(ctx context.Context, tenantID uuid.UUID) error {
	// Count users
	// Note: This would require implementing count methods in repositories
	// For now, we'll create a placeholder
	
	usage := entities.SubscriptionUsage{
		Users:         0, // TODO: Implement user count
		Products:      0, // TODO: Implement product count
		SalesThisMonth: 0, // TODO: Implement sales count for current month
		APICallsThisMonth: 0, // TODO: Implement API calls count for current month
		LastUpdated:   time.Now(),
	}

	return uc.UpdateUsage(ctx, UpdateUsageRequest{
		TenantID: tenantID,
		Usage:    usage,
	})
}

// GetExpiredSubscriptions retrieves subscriptions that have expired
func (uc *SubscriptionUseCase) GetExpiredSubscriptions(ctx context.Context) ([]*entities.TenantSubscription, error) {
	subscriptions, err := uc.subscriptionRepo.GetExpiredSubscriptions(ctx)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to get expired subscriptions")
		return nil, errors.NewInternalError("failed to get expired subscriptions", err)
	}

	return subscriptions, nil
}

// ProcessExpiredSubscriptions processes expired subscriptions
func (uc *SubscriptionUseCase) ProcessExpiredSubscriptions(ctx context.Context) error {
	expiredSubscriptions, err := uc.GetExpiredSubscriptions(ctx)
	if err != nil {
		return err
	}

	for _, subscription := range expiredSubscriptions {
		// Suspend expired subscriptions
		if err := subscription.SuspendSubscription(); err != nil {
			uc.logger.WithError(err).WithField("tenant_id", subscription.TenantID).Error("Failed to suspend expired subscription")
			continue
		}

		if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
			uc.logger.WithError(err).WithField("tenant_id", subscription.TenantID).Error("Failed to update expired subscription")
			continue
		}

		uc.logger.WithField("tenant_id", subscription.TenantID).Info("Expired subscription suspended")
	}

	return nil
}

// Helper methods

// isPlanUpgrade checks if the new plan is an upgrade from the current plan
func (uc *SubscriptionUseCase) isPlanUpgrade(current, new entities.SubscriptionPlanType) bool {
	planHierarchy := map[entities.SubscriptionPlanType]int{
		entities.PlanStarter:      1,
		entities.PlanProfessional: 2,
		entities.PlanEnterprise:   3,
	}

	return planHierarchy[new] > planHierarchy[current]
}

// generateUsageWarnings generates warnings based on usage patterns
func (uc *SubscriptionUseCase) generateUsageWarnings(subscription *entities.TenantSubscription) []string {
	var warnings []string

	// Check if approaching limits (80% threshold)
	if subscription.GetUsagePercentage("users") > 80 {
		warnings = append(warnings, "User limit approaching. Consider upgrading your plan.")
	}

	if subscription.GetUsagePercentage("products") > 80 {
		warnings = append(warnings, "Product limit approaching. Consider upgrading your plan.")
	}

	if subscription.GetUsagePercentage("sales") > 80 {
		warnings = append(warnings, "Monthly sales limit approaching. Consider upgrading your plan.")
	}

	if subscription.GetUsagePercentage("api_calls") > 80 {
		warnings = append(warnings, "Monthly API calls limit approaching. Consider upgrading your plan.")
	}

	return warnings
}