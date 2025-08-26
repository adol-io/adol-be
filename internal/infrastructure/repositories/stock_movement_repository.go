package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/utils"
)

// PostgreSQLStockMovementRepository implements the StockMovementRepository interface for PostgreSQL
type PostgreSQLStockMovementRepository struct {
	db *sql.DB
}

// NewPostgreSQLStockMovementRepository creates a new PostgreSQL stock movement repository
func NewPostgreSQLStockMovementRepository(db *sql.DB) repositories.StockMovementRepository {
	return &PostgreSQLStockMovementRepository{
		db: db,
	}
}

// Create creates a new stock movement record
func (r *PostgreSQLStockMovementRepository) Create(ctx context.Context, movement *entities.StockMovement) error {
	query := `
		INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		movement.ID,
		movement.ProductID,
		movement.Type,
		movement.Reason,
		movement.Quantity,
		movement.Reference,
		movement.Notes,
		movement.CreatedAt,
		movement.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	return nil
}

// GetByID retrieves a stock movement by ID
func (r *PostgreSQLStockMovementRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.StockMovement, error) {
	query := `
		SELECT id, product_id, type, reason, quantity, reference, notes, created_at, created_by
		FROM stock_movements 
		WHERE id = $1`

	movement := &entities.StockMovement{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&movement.ID,
		&movement.ProductID,
		&movement.Type,
		&movement.Reason,
		&movement.Quantity,
		&movement.Reference,
		&movement.Notes,
		&movement.CreatedAt,
		&movement.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("stock movement")
		}
		return nil, fmt.Errorf("failed to get stock movement by ID: %w", err)
	}

	return movement, nil
}

// List retrieves stock movements with pagination and filtering
func (r *PostgreSQLStockMovementRepository) List(ctx context.Context, filter repositories.StockMovementFilter, pagination utils.PaginationInfo) ([]*entities.StockMovement, utils.PaginationInfo, error) {
	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProductID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("product_id = $%d", argIndex))
		args = append(args, *filter.ProductID)
		argIndex++
	}

	if filter.Type != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Reason != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("reason = $%d", argIndex))
		args = append(args, *filter.Reason)
		argIndex++
	}

	if filter.Reference != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("reference = $%d", argIndex))
		args = append(args, filter.Reference)
		argIndex++
	}

	if filter.CreatedBy != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filter.CreatedBy)
		argIndex++
	}

	if filter.FromDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.ToDate)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDir == "DESC" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filter.OrderBy, direction)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_movements %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count stock movements: %w", err)
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	// Build main query
	query := fmt.Sprintf(`
		SELECT id, product_id, type, reason, quantity, reference, notes, created_at, created_by
		FROM stock_movements 
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to query stock movements: %w", err)
	}
	defer rows.Close()

	var movements []*entities.StockMovement
	for rows.Next() {
		movement := &entities.StockMovement{}
		err := rows.Scan(
			&movement.ID,
			&movement.ProductID,
			&movement.Type,
			&movement.Reason,
			&movement.Quantity,
			&movement.Reference,
			&movement.Notes,
			&movement.CreatedAt,
			&movement.CreatedBy,
		)
		if err != nil {
			return nil, pagination, fmt.Errorf("failed to scan stock movement: %w", err)
		}
		movements = append(movements, movement)
	}

	if err = rows.Err(); err != nil {
		return nil, pagination, fmt.Errorf("failed to iterate stock movements: %w", err)
	}

	// Update pagination info
	resultPagination := utils.PaginationInfo{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalCount: int(total),
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}

	return movements, resultPagination, nil
}

// GetByProductID retrieves stock movements for a specific product
func (r *PostgreSQLStockMovementRepository) GetByProductID(ctx context.Context, productID uuid.UUID, pagination utils.PaginationInfo) ([]*entities.StockMovement, utils.PaginationInfo, error) {
	filter := repositories.StockMovementFilter{
		ProductID: &productID,
	}
	return r.List(ctx, filter, pagination)
}

// GetByReference retrieves stock movements by reference
func (r *PostgreSQLStockMovementRepository) GetByReference(ctx context.Context, reference string) ([]*entities.StockMovement, error) {
	query := `
		SELECT id, product_id, type, reason, quantity, reference, notes, created_at, created_by
		FROM stock_movements 
		WHERE reference = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, reference)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock movements by reference: %w", err)
	}
	defer rows.Close()

	var movements []*entities.StockMovement
	for rows.Next() {
		movement := &entities.StockMovement{}
		err := rows.Scan(
			&movement.ID,
			&movement.ProductID,
			&movement.Type,
			&movement.Reason,
			&movement.Quantity,
			&movement.Reference,
			&movement.Notes,
			&movement.CreatedAt,
			&movement.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock movement: %w", err)
		}
		movements = append(movements, movement)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate stock movements: %w", err)
	}

	return movements, nil
}

// Delete deletes a stock movement record
func (r *PostgreSQLStockMovementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM stock_movements WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete stock movement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("stock movement")
	}

	return nil
}

// BulkCreate creates multiple stock movement records in a transaction
func (r *PostgreSQLStockMovementRepository) BulkCreate(ctx context.Context, movements []*entities.StockMovement) error {
	if len(movements) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, movement := range movements {
		_, err := stmt.ExecContext(ctx,
			movement.ID,
			movement.ProductID,
			movement.Type,
			movement.Reason,
			movement.Quantity,
			movement.Reference,
			movement.Notes,
			movement.CreatedAt,
			movement.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to create stock movement %s: %w", movement.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
