package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/utils"
)

// PostgreSQLStockRepository implements the StockRepository interface for PostgreSQL
type PostgreSQLStockRepository struct {
	db *sql.DB
}

// NewPostgreSQLStockRepository creates a new PostgreSQL stock repository
func NewPostgreSQLStockRepository(db *sql.DB) repositories.StockRepository {
	return &PostgreSQLStockRepository{
		db: db,
	}
}

// Create creates a new stock record
func (r *PostgreSQLStockRepository) Create(ctx context.Context, stock *entities.Stock) error {
	query := `
		INSERT INTO stock (id, product_id, available_qty, reserved_qty, total_qty, reorder_level, last_movement_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		stock.ID,
		stock.ProductID,
		stock.AvailableQty,
		stock.ReservedQty,
		stock.TotalQty,
		stock.ReorderLevel,
		stock.LastMovementAt,
		stock.CreatedAt,
		stock.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	return nil
}

// GetByID retrieves stock by ID
func (r *PostgreSQLStockRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Stock, error) {
	query := `
		SELECT id, product_id, available_qty, reserved_qty, total_qty, reorder_level, 
		       last_movement_at, created_at, updated_at
		FROM stock 
		WHERE id = $1`

	stock := &entities.Stock{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&stock.ID,
		&stock.ProductID,
		&stock.AvailableQty,
		&stock.ReservedQty,
		&stock.TotalQty,
		&stock.ReorderLevel,
		&stock.LastMovementAt,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("stock")
		}
		return nil, fmt.Errorf("failed to get stock by ID: %w", err)
	}

	return stock, nil
}

// GetByProductID retrieves stock by product ID
func (r *PostgreSQLStockRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*entities.Stock, error) {
	query := `
		SELECT id, product_id, available_qty, reserved_qty, total_qty, reorder_level, 
		       last_movement_at, created_at, updated_at
		FROM stock 
		WHERE product_id = $1`

	stock := &entities.Stock{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&stock.ID,
		&stock.ProductID,
		&stock.AvailableQty,
		&stock.ReservedQty,
		&stock.TotalQty,
		&stock.ReorderLevel,
		&stock.LastMovementAt,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("stock")
		}
		return nil, fmt.Errorf("failed to get stock by product ID: %w", err)
	}

	return stock, nil
}

// Update updates stock information
func (r *PostgreSQLStockRepository) Update(ctx context.Context, stock *entities.Stock) error {
	query := `
		UPDATE stock 
		SET available_qty = $2, reserved_qty = $3, total_qty = $4, reorder_level = $5, 
		    last_movement_at = $6, updated_at = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		stock.ID,
		stock.AvailableQty,
		stock.ReservedQty,
		stock.TotalQty,
		stock.ReorderLevel,
		stock.LastMovementAt,
		stock.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("stock")
	}

	return nil
}

// Delete deletes a stock record
func (r *PostgreSQLStockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM stock WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("stock")
	}

	return nil
}

// List retrieves stock records with pagination and filtering
func (r *PostgreSQLStockRepository) List(ctx context.Context, filter repositories.StockFilter, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error) {
	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProductID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("s.product_id = $%d", argIndex))
		args = append(args, *filter.ProductID)
		argIndex++
	}

	if filter.LowStock != nil && *filter.LowStock {
		whereConditions = append(whereConditions, "s.available_qty <= s.reorder_level")
	}

	if filter.OutOfStock != nil && *filter.OutOfStock {
		whereConditions = append(whereConditions, "s.available_qty = 0")
	}

	if filter.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "s.created_at DESC"
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDir == "DESC" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("s.%s %s", filter.OrderBy, direction)
	}

	// Count total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM stock s
		JOIN products p ON s.product_id = p.id
		%s`, whereClause)

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count stock: %w", err)
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	// Build main query
	query := fmt.Sprintf(`
		SELECT s.id, s.product_id, s.available_qty, s.reserved_qty, s.total_qty, s.reorder_level, 
		       s.last_movement_at, s.created_at, s.updated_at
		FROM stock s
		JOIN products p ON s.product_id = p.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to query stock: %w", err)
	}
	defer rows.Close()

	var stocks []*entities.Stock
	for rows.Next() {
		stock := &entities.Stock{}
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.AvailableQty,
			&stock.ReservedQty,
			&stock.TotalQty,
			&stock.ReorderLevel,
			&stock.LastMovementAt,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, pagination, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, stock)
	}

	if err = rows.Err(); err != nil {
		return nil, pagination, fmt.Errorf("failed to iterate stock: %w", err)
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

	return stocks, resultPagination, nil
}

// GetLowStockItems retrieves items with stock below reorder level
func (r *PostgreSQLStockRepository) GetLowStockItems(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error) {
	filter := repositories.StockFilter{
		LowStock: &[]bool{true}[0],
	}
	return r.List(ctx, filter, pagination)
}

// GetOutOfStockItems retrieves items that are out of stock
func (r *PostgreSQLStockRepository) GetOutOfStockItems(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error) {
	filter := repositories.StockFilter{
		OutOfStock: &[]bool{true}[0],
	}
	return r.List(ctx, filter, pagination)
}

// BulkUpdateStock updates multiple stock records in a transaction
func (r *PostgreSQLStockRepository) BulkUpdateStock(ctx context.Context, stocks []*entities.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE stock 
		SET available_qty = $2, reserved_qty = $3, total_qty = $4, reorder_level = $5, 
		    last_movement_at = $6, updated_at = $7
		WHERE id = $1`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stock := range stocks {
		_, err := stmt.ExecContext(ctx,
			stock.ID,
			stock.AvailableQty,
			stock.ReservedQty,
			stock.TotalQty,
			stock.ReorderLevel,
			stock.LastMovementAt,
			stock.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock %s: %w", stock.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AdjustStock adjusts stock quantity (positive or negative)
func (r *PostgreSQLStockRepository) AdjustStock(ctx context.Context, adjustment repositories.StockAdjustment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current stock
	stock, err := r.GetByProductID(ctx, adjustment.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Apply adjustment
	if adjustment.Quantity > 0 {
		err = stock.AddStock(adjustment.Quantity, adjustment.Reason)
	} else {
		err = stock.RemoveStock(-adjustment.Quantity)
	}
	if err != nil {
		return err
	}

	// Update stock
	updateQuery := `
		UPDATE stock 
		SET available_qty = $2, total_qty = $3, last_movement_at = $4, updated_at = $5
		WHERE product_id = $1`

	_, err = tx.ExecContext(ctx, updateQuery,
		adjustment.ProductID,
		stock.AvailableQty,
		stock.TotalQty,
		stock.LastMovementAt,
		stock.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	// Create stock movement record
	movementType := entities.StockMovementTypeIn
	if adjustment.Quantity < 0 {
		movementType = entities.StockMovementTypeOut
	}

	movementQuery := `
		INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(),
		adjustment.ProductID,
		movementType,
		adjustment.Reason,
		int(math.Abs(float64(adjustment.Quantity))),
		adjustment.Reference,
		adjustment.Notes,
		time.Now(),
		adjustment.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReserveStock reserves stock for an order
func (r *PostgreSQLStockRepository) ReserveStock(ctx context.Context, reservation repositories.StockReservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current stock
	stock, err := r.GetByProductID(ctx, reservation.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Reserve stock
	err = stock.ReserveStock(reservation.Quantity)
	if err != nil {
		return err
	}

	// Update stock
	updateQuery := `
		UPDATE stock 
		SET available_qty = $2, reserved_qty = $3, last_movement_at = $4, updated_at = $5
		WHERE product_id = $1`

	_, err = tx.ExecContext(ctx, updateQuery,
		reservation.ProductID,
		stock.AvailableQty,
		stock.ReservedQty,
		stock.LastMovementAt,
		stock.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	// Create stock movement record
	movementQuery := `
		INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(),
		reservation.ProductID,
		entities.StockMovementTypeReserved,
		entities.ReasonReservation,
		reservation.Quantity,
		reservation.Reference,
		reservation.Notes,
		time.Now(),
		reservation.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReleaseReservedStock releases reserved stock back to available
func (r *PostgreSQLStockRepository) ReleaseReservedStock(ctx context.Context, release repositories.StockRelease) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current stock
	stock, err := r.GetByProductID(ctx, release.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Release reserved stock
	err = stock.ReleaseReservedStock(release.Quantity)
	if err != nil {
		return err
	}

	// Update stock
	updateQuery := `
		UPDATE stock 
		SET available_qty = $2, reserved_qty = $3, last_movement_at = $4, updated_at = $5
		WHERE product_id = $1`

	_, err = tx.ExecContext(ctx, updateQuery,
		release.ProductID,
		stock.AvailableQty,
		stock.ReservedQty,
		stock.LastMovementAt,
		stock.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	// Create stock movement record
	movementQuery := `
		INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(),
		release.ProductID,
		entities.StockMovementTypeReleased,
		entities.ReasonRelease,
		release.Quantity,
		release.Reference,
		release.Notes,
		time.Now(),
		release.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BulkReserveStock reserves stock for multiple products in a transaction
func (r *PostgreSQLStockRepository) BulkReserveStock(ctx context.Context, reservations []repositories.StockReservation) error {
	if len(reservations) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, reservation := range reservations {
		// Get current stock
		stock, err := r.GetByProductID(ctx, reservation.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get stock for product %s: %w", reservation.ProductID, err)
		}

		// Reserve stock
		err = stock.ReserveStock(reservation.Quantity)
		if err != nil {
			return fmt.Errorf("failed to reserve stock for product %s: %w", reservation.ProductID, err)
		}

		// Update stock
		updateQuery := `
			UPDATE stock 
			SET available_qty = $2, reserved_qty = $3, last_movement_at = $4, updated_at = $5
			WHERE product_id = $1`

		_, err = tx.ExecContext(ctx, updateQuery,
			reservation.ProductID,
			stock.AvailableQty,
			stock.ReservedQty,
			stock.LastMovementAt,
			stock.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock for product %s: %w", reservation.ProductID, err)
		}

		// Create stock movement record
		movementQuery := `
			INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = tx.ExecContext(ctx, movementQuery,
			uuid.New(),
			reservation.ProductID,
			entities.StockMovementTypeReserved,
			entities.ReasonReservation,
			reservation.Quantity,
			reservation.Reference,
			reservation.Notes,
			time.Now(),
			reservation.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to create stock movement for product %s: %w", reservation.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BulkReleaseStock releases reserved stock for multiple products in a transaction
func (r *PostgreSQLStockRepository) BulkReleaseStock(ctx context.Context, releases []repositories.StockRelease) error {
	if len(releases) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, release := range releases {
		// Get current stock
		stock, err := r.GetByProductID(ctx, release.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get stock for product %s: %w", release.ProductID, err)
		}

		// Release reserved stock
		err = stock.ReleaseReservedStock(release.Quantity)
		if err != nil {
			return fmt.Errorf("failed to release stock for product %s: %w", release.ProductID, err)
		}

		// Update stock
		updateQuery := `
			UPDATE stock 
			SET available_qty = $2, reserved_qty = $3, last_movement_at = $4, updated_at = $5
			WHERE product_id = $1`

		_, err = tx.ExecContext(ctx, updateQuery,
			release.ProductID,
			stock.AvailableQty,
			stock.ReservedQty,
			stock.LastMovementAt,
			stock.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock for product %s: %w", release.ProductID, err)
		}

		// Create stock movement record
		movementQuery := `
			INSERT INTO stock_movements (id, product_id, type, reason, quantity, reference, notes, created_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = tx.ExecContext(ctx, movementQuery,
			uuid.New(),
			release.ProductID,
			entities.StockMovementTypeReleased,
			entities.ReasonRelease,
			release.Quantity,
			release.Reference,
			release.Notes,
			time.Now(),
			release.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to create stock movement for product %s: %w", release.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}