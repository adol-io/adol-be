package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
)

type tenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sql.DB) repositories.TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	configJSON, err := json.Marshal(tenant.Configuration)
	if err != nil {
		return errors.NewInternalError("failed to marshal tenant configuration", err)
	}

	query := `
		INSERT INTO tenants (id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Domain,
		tenant.Status,
		configJSON,
		tenant.TrialStart,
		tenant.TrialEnd,
		tenant.CreatedAt,
		tenant.UpdatedAt,
		tenant.CreatedBy,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "tenants_slug_key" {
					return errors.NewValidationError("tenant slug already exists", "slug must be unique")
				}
				if pqErr.Constraint == "tenants_domain_key" {
					return errors.NewValidationError("tenant domain already exists", "domain must be unique")
				}
			}
		}
		return errors.NewInternalError("failed to create tenant", err)
	}

	return nil
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		WHERE id = $1`

	tenant := &entities.Tenant{}
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Domain,
		&tenant.Status,
		&configJSON,
		&tenant.TrialStart,
		&tenant.TrialEnd,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant not found")
		}
		return nil, errors.NewInternalError("failed to get tenant", err)
	}

	if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal tenant configuration", err)
	}

	return tenant, nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*entities.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		WHERE slug = $1`

	tenant := &entities.Tenant{}
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Domain,
		&tenant.Status,
		&configJSON,
		&tenant.TrialStart,
		&tenant.TrialEnd,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant not found")
		}
		return nil, errors.NewInternalError("failed to get tenant by slug", err)
	}

	if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal tenant configuration", err)
	}

	return tenant, nil
}

func (r *tenantRepository) GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		WHERE domain = $1`

	tenant := &entities.Tenant{}
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Domain,
		&tenant.Status,
		&configJSON,
		&tenant.TrialStart,
		&tenant.TrialEnd,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant not found")
		}
		return nil, errors.NewInternalError("failed to get tenant by domain", err)
	}

	if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal tenant configuration", err)
	}

	return tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	configJSON, err := json.Marshal(tenant.Configuration)
	if err != nil {
		return errors.NewInternalError("failed to marshal tenant configuration", err)
	}

	query := `
		UPDATE tenants 
		SET name = $2, slug = $3, domain = $4, status = $5, configuration = $6, 
		    trial_start = $7, trial_end = $8, updated_at = $9
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Domain,
		tenant.Status,
		configJSON,
		tenant.TrialStart,
		tenant.TrialEnd,
		time.Now(),
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "tenants_slug_key" {
					return errors.NewValidationError("tenant slug already exists", "slug must be unique")
				}
				if pqErr.Constraint == "tenants_domain_key" {
					return errors.NewValidationError("tenant domain already exists", "domain must be unique")
				}
			}
		}
		return errors.NewInternalError("failed to update tenant", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant not found")
	}

	return nil
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// This is a soft delete - we just change the status to inactive
	query := `UPDATE tenants SET status = 'inactive', updated_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return errors.NewInternalError("failed to delete tenant", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant not found")
	}

	return nil
}

func (r *tenantRepository) List(ctx context.Context, offset, limit int) ([]*entities.Tenant, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM tenants`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to count tenants", err)
	}

	// Get tenants with pagination
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to list tenants", err)
	}
	defer rows.Close()

	var tenants []*entities.Tenant
	for rows.Next() {
		tenant := &entities.Tenant{}
		var configJSON []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Domain,
			&tenant.Status,
			&configJSON,
			&tenant.TrialStart,
			&tenant.TrialEnd,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.CreatedBy,
		)
		if err != nil {
			return nil, 0, errors.NewInternalError("failed to scan tenant", err)
		}

		if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
			return nil, 0, errors.NewInternalError("failed to unmarshal tenant configuration", err)
		}

		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, errors.NewInternalError("failed to iterate tenants", err)
	}

	return tenants, total, nil
}

func (r *tenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, errors.NewInternalError("failed to check tenant slug existence", err)
	}

	return exists, nil
}

func (r *tenantRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE domain = $1)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, domain).Scan(&exists)
	if err != nil {
		return false, errors.NewInternalError("failed to check tenant domain existence", err)
	}

	return exists, nil
}

func (r *tenantRepository) GetActiveCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tenants WHERE status = 'active'`
	
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.NewInternalError("failed to count active tenants", err)
	}

	return count, nil
}

func (r *tenantRepository) GetTrialTenants(ctx context.Context) ([]*entities.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		WHERE status = 'trial' AND trial_end > NOW()`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.NewInternalError("failed to get trial tenants", err)
	}
	defer rows.Close()

	var tenants []*entities.Tenant
	for rows.Next() {
		tenant := &entities.Tenant{}
		var configJSON []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Domain,
			&tenant.Status,
			&configJSON,
			&tenant.TrialStart,
			&tenant.TrialEnd,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.CreatedBy,
		)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan trial tenant", err)
		}

		if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal tenant configuration", err)
		}

		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalError("failed to iterate trial tenants", err)
	}

	return tenants, nil
}

func (r *tenantRepository) GetExpiredTrialTenants(ctx context.Context) ([]*entities.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, configuration, trial_start, trial_end, created_at, updated_at, created_by
		FROM tenants 
		WHERE status = 'trial' AND trial_end <= NOW()`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.NewInternalError("failed to get expired trial tenants", err)
	}
	defer rows.Close()

	var tenants []*entities.Tenant
	for rows.Next() {
		tenant := &entities.Tenant{}
		var configJSON []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Domain,
			&tenant.Status,
			&configJSON,
			&tenant.TrialStart,
			&tenant.TrialEnd,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.CreatedBy,
		)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan expired trial tenant", err)
		}

		if err := json.Unmarshal(configJSON, &tenant.Configuration); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal tenant configuration", err)
		}

		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalError("failed to iterate expired trial tenants", err)
	}

	return tenants, nil
}