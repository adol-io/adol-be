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

type tenantSettingRepository struct {
	db *sql.DB
}

// NewTenantSettingRepository creates a new tenant setting repository
func NewTenantSettingRepository(db *sql.DB) repositories.TenantSettingRepository {
	return &tenantSettingRepository{db: db}
}

func (r *tenantSettingRepository) Create(ctx context.Context, setting *entities.TenantSetting) error {
	valueJSON, err := json.Marshal(setting.SettingValue)
	if err != nil {
		return errors.NewInternalError("failed to marshal setting value", err)
	}

	query := `
		INSERT INTO tenant_settings (id, tenant_id, setting_key, setting_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = r.db.ExecContext(ctx, query,
		setting.ID,
		setting.TenantID,
		setting.SettingKey,
		valueJSON,
		setting.CreatedAt,
		setting.UpdatedAt,
	)

	if err != nil {
		return errors.NewInternalError("failed to create tenant setting", err)
	}

	return nil
}

func (r *tenantSettingRepository) GetByTenantAndKey(ctx context.Context, tenantID uuid.UUID, key string) (*entities.TenantSetting, error) {
	query := `
		SELECT id, tenant_id, setting_key, setting_value, created_at, updated_at
		FROM tenant_settings 
		WHERE tenant_id = $1 AND setting_key = $2`

	setting := &entities.TenantSetting{}
	var valueJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(
		&setting.ID,
		&setting.TenantID,
		&setting.SettingKey,
		&valueJSON,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("tenant setting not found", "setting")
		}
		return nil, errors.NewInternalError("failed to get tenant setting", err)
	}

	if err := json.Unmarshal(valueJSON, &setting.SettingValue); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal setting value", err)
	}

	return setting, nil
}

func (r *tenantSettingRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entities.TenantSetting, error) {
	query := `
		SELECT id, tenant_id, setting_key, setting_value, created_at, updated_at
		FROM tenant_settings 
		WHERE tenant_id = $1
		ORDER BY setting_key`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get tenant settings", err)
	}
	defer rows.Close()

	var settings []*entities.TenantSetting
	for rows.Next() {
		setting := &entities.TenantSetting{}
		var valueJSON []byte

		err := rows.Scan(
			&setting.ID,
			&setting.TenantID,
			&setting.SettingKey,
			&valueJSON,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan tenant setting", err)
		}

		if err := json.Unmarshal(valueJSON, &setting.SettingValue); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal setting value", err)
		}

		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalError("failed to iterate tenant settings", err)
	}

	return settings, nil
}

func (r *tenantSettingRepository) Update(ctx context.Context, setting *entities.TenantSetting) error {
	valueJSON, err := json.Marshal(setting.SettingValue)
	if err != nil {
		return errors.NewInternalError("failed to marshal setting value", err)
	}

	query := `
		UPDATE tenant_settings 
		SET setting_value = $3, updated_at = $4
		WHERE tenant_id = $1 AND setting_key = $2`

	result, err := r.db.ExecContext(ctx, query,
		setting.TenantID,
		setting.SettingKey,
		valueJSON,
		time.Now(),
	)

	if err != nil {
		return errors.NewInternalError("failed to update tenant setting", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant setting not found", "setting")
	}

	return nil
}

func (r *tenantSettingRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	query := `DELETE FROM tenant_settings WHERE tenant_id = $1 AND setting_key = $2`

	result, err := r.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return errors.NewInternalError("failed to delete tenant setting", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("tenant setting not found", "setting")
	}

	return nil
}

func (r *tenantSettingRepository) Upsert(ctx context.Context, tenantID uuid.UUID, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return errors.NewInternalError("failed to marshal setting value", err)
	}

	query := `
		INSERT INTO tenant_settings (id, tenant_id, setting_key, setting_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, setting_key)
		DO UPDATE SET 
			setting_value = EXCLUDED.setting_value,
			updated_at = EXCLUDED.updated_at`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		uuid.New(),
		tenantID,
		key,
		valueJSON,
		now,
		now,
	)

	if err != nil {
		return errors.NewInternalError("failed to upsert tenant setting", err)
	}

	return nil
}

func (r *tenantSettingRepository) GetSettings(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT setting_key, setting_value
		FROM tenant_settings 
		WHERE tenant_id = $1`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get tenant settings", err)
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key string
		var valueJSON []byte

		err := rows.Scan(&key, &valueJSON)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan tenant setting", err)
		}

		var value interface{}
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			return nil, errors.NewInternalError("failed to unmarshal setting value", err)
		}

		settings[key] = value
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalError("failed to iterate tenant settings", err)
	}

	return settings, nil
}