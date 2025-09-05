package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/errors"
)

type tenantSubscriptionRepository struct {
	db *sql.DB
}

// NewTenantSubscriptionRepository creates a new tenant subscription repository
func NewTenantSubscriptionRepository(db *sql.DB) repositories.TenantSubscriptionRepository {
	return &tenantSubscriptionRepository{db: db}
}

func (r *tenantSubscriptionRepository) Create(ctx context.Context, subscription *entities.TenantSubscription) error {
	featuresJSON, err := json.Marshal(subscription.Features)
	if err != nil {
		return errors.NewInternalError("failed to marshal subscription features", err)
	}

	limitsJSON, err := json.Marshal(subscription.UsageLimits)
	if err != nil {
		return errors.NewInternalError("failed to marshal usage limits", err)
	}

	usageJSON, err := json.Marshal(subscription.CurrentUsage)
	if err != nil {
		return errors.NewInternalError("failed to marshal current usage", err)
	}

	query := `
		INSERT INTO tenant_subscriptions (id, tenant_id, plan_type, status, billing_start, billing_end, monthly_fee, features, usage_limits, current_usage, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = r.db.ExecContext(ctx, query,
		subscription.ID,
		subscription.TenantID,
		subscription.PlanType,
		subscription.Status,
		subscription.BillingStart,
		subscription.BillingEnd,
		subscription.MonthlyFee,
		featuresJSON,
		limitsJSON,
		usageJSON,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return errors.NewInternalError("failed to create tenant subscription", err)
	}

	return nil
}

func (r *tenantSubscriptionRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entities.TenantSubscription, error) {
	query := `
		SELECT id, tenant_id, plan_type, status, billing_start, billing_end, monthly_fee, features, usage_limits, current_usage, created_at, updated_at
		FROM tenant_subscriptions 
		WHERE tenant_id = $1`

	subscription := &entities.TenantSubscription{}
	var featuresJSON, limitsJSON, usageJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&subscription.ID,
		&subscription.TenantID,
		&subscription.PlanType,
		&subscription.Status,
		&subscription.BillingStart,
		&subscription.BillingEnd,
		&subscription.MonthlyFee,
		&featuresJSON,
		&limitsJSON,
		&usageJSON,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant subscription not found", "subscription")
		}
		return nil, errors.NewInternalError("failed to get tenant subscription", err)
	}

	if err := json.Unmarshal(featuresJSON, &subscription.Features); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal subscription features", err)
	}

	if err := json.Unmarshal(limitsJSON, &subscription.UsageLimits); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal usage limits", err)
	}

	if err := json.Unmarshal(usageJSON, &subscription.CurrentUsage); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal current usage", err)
	}

	return subscription, nil
}

func (r *tenantSubscriptionRepository) Update(ctx context.Context, subscription *entities.TenantSubscription) error {
	featuresJSON, err := json.Marshal(subscription.Features)
	if err != nil {
		return errors.NewInternalError("failed to marshal subscription features", err)
	}

	limitsJSON, err := json.Marshal(subscription.UsageLimits)
	if err != nil {
		return errors.NewInternalError("failed to marshal usage limits", err)
	}

	usageJSON, err := json.Marshal(subscription.CurrentUsage)
	if err != nil {
		return errors.NewInternalError("failed to marshal current usage", err)
	}

	query := `
		UPDATE tenant_subscriptions 
		SET plan_type = $2, status = $3, billing_start = $4, billing_end = $5, monthly_fee = $6, 
		    features = $7, usage_limits = $8, current_usage = $9, updated_at = $10
		WHERE tenant_id = $1`

	result, err := r.db.ExecContext(ctx, query,
		subscription.TenantID,
		subscription.PlanType,
		subscription.Status,
		subscription.BillingStart,
		subscription.BillingEnd,
		subscription.MonthlyFee,
		featuresJSON,
		limitsJSON,
		usageJSON,
		time.Now(),
	)

	if err != nil {
		return errors.NewInternalError("failed to update tenant subscription", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant subscription not found", "subscription")
	}

	return nil
}

func (r *tenantSubscriptionRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	query := `DELETE FROM tenant_subscriptions WHERE tenant_id = $1`

	result, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return errors.NewInternalError("failed to delete tenant subscription", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant subscription not found", "subscription")
	}

	return nil
}

func (r *tenantSubscriptionRepository) GetActiveSubscriptions(ctx context.Context) ([]*entities.TenantSubscription, error) {
	query := `
		SELECT id, tenant_id, plan_type, status, billing_start, billing_end, monthly_fee, features, usage_limits, current_usage, created_at, updated_at
		FROM tenant_subscriptions 
		WHERE status = 'active'
		ORDER BY created_at DESC`

	return r.querySubscriptions(ctx, query)
}

func (r *tenantSubscriptionRepository) GetSubscriptionsByPlan(ctx context.Context, planType entities.SubscriptionPlanType) ([]*entities.TenantSubscription, error) {
	query := `
		SELECT id, tenant_id, plan_type, status, billing_start, billing_end, monthly_fee, features, usage_limits, current_usage, created_at, updated_at
		FROM tenant_subscriptions 
		WHERE plan_type = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, planType)
	if err != nil {
		return nil, errors.NewInternalError("failed to get subscriptions by plan", err)
	}
	defer rows.Close()

	return r.scanSubscriptions(rows)
}

func (r *tenantSubscriptionRepository) GetExpiredSubscriptions(ctx context.Context) ([]*entities.TenantSubscription, error) {
	query := `
		SELECT id, tenant_id, plan_type, status, billing_start, billing_end, monthly_fee, features, usage_limits, current_usage, created_at, updated_at
		FROM tenant_subscriptions 
		WHERE billing_end IS NOT NULL AND billing_end <= NOW()
		ORDER BY billing_end ASC`

	return r.querySubscriptions(ctx, query)
}

func (r *tenantSubscriptionRepository) UpdateUsage(ctx context.Context, tenantID uuid.UUID, usage entities.SubscriptionUsage) error {
	usageJSON, err := json.Marshal(usage)
	if err != nil {
		return errors.NewInternalError("failed to marshal usage", err)
	}

	query := `
		UPDATE tenant_subscriptions 
		SET current_usage = $2, updated_at = $3
		WHERE tenant_id = $1`

	result, err := r.db.ExecContext(ctx, query, tenantID, usageJSON, time.Now())
	if err != nil {
		return errors.NewInternalError("failed to update subscription usage", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant subscription not found", "subscription")
	}

	return nil
}

func (r *tenantSubscriptionRepository) GetUsageByTenant(ctx context.Context, tenantID uuid.UUID) (*entities.SubscriptionUsage, error) {
	query := `SELECT current_usage FROM tenant_subscriptions WHERE tenant_id = $1`

	var usageJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&usageJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant subscription not found", "subscription")
		}
		return nil, errors.NewInternalError("failed to get subscription usage", err)
	}

	var usage entities.SubscriptionUsage
	if err := json.Unmarshal(usageJSON, &usage); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal usage", err)
	}

	return &usage, nil
}

// Helper methods

func (r *tenantSubscriptionRepository) querySubscriptions(ctx context.Context, query string, args ...interface{}) ([]*entities.TenantSubscription, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewInternalError("failed to query subscriptions", err)
	}
	defer rows.Close()

	return r.scanSubscriptions(rows)
}

func (r *tenantSubscriptionRepository) scanSubscriptions(rows *sql.Rows) ([]*entities.TenantSubscription, error) {
	var subscriptions []*entities.TenantSubscription

	for rows.Next() {
		subscription := &entities.TenantSubscription{}
		var featuresJSON, limitsJSON, usageJSON []byte

		err := rows.Scan(
			&subscription.ID,
			&subscription.TenantID,
			&subscription.PlanType,
			&subscription.Status,
			&subscription.BillingStart,
			&subscription.BillingEnd,
			&subscription.MonthlyFee,
			&featuresJSON,
			&limitsJSON,
			&usageJSON,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan subscription", err)
		}

		if err := json.Unmarshal(featuresJSON, &subscription.Features); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal subscription features", err)
		}

		if err := json.Unmarshal(limitsJSON, &subscription.UsageLimits); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal usage limits", err)
		}

		if err := json.Unmarshal(usageJSON, &subscription.CurrentUsage); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal current usage", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewInternalError("failed to iterate subscriptions", err)
	}

	return subscriptions, nil
}