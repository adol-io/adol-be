package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
)

// PostgresSaleItemRepository implements the SaleItemRepository interface
type PostgresSaleItemRepository struct {
	db *sql.DB
}

// NewPostgresSaleItemRepository creates a new PostgreSQL sale item repository
func NewPostgresSaleItemRepository(db *sql.DB) repositories.SaleItemRepository {
	return &PostgresSaleItemRepository{db: db}
}

// Create creates a new sale item
func (r *PostgresSaleItemRepository) Create(ctx context.Context, item *entities.SaleItem) error {
	query := `
		INSERT INTO sale_items (id, sale_id, product_id, product_sku, product_name, 
			quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.SaleID, item.ProductID, item.ProductSKU, item.ProductName,
		item.Quantity, item.UnitPrice, item.TotalPrice, item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create sale item: %w", err)
	}

	return nil
}

// GetByID retrieves a sale item by ID
func (r *PostgresSaleItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, product_sku, product_name, 
			quantity, unit_price, total_price, created_at
		FROM sale_items 
		WHERE id = $1`

	var item entities.SaleItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.SaleID, &item.ProductID, &item.ProductSKU, &item.ProductName,
		&item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("sale item")
		}
		return nil, fmt.Errorf("failed to get sale item: %w", err)
	}

	return &item, nil
}

// GetBySaleID retrieves all items for a sale
func (r *PostgresSaleItemRepository) GetBySaleID(ctx context.Context, saleID uuid.UUID) ([]*entities.SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, product_sku, product_name, 
			quantity, unit_price, total_price, created_at
		FROM sale_items 
		WHERE sale_id = $1 
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, saleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sale items: %w", err)
	}
	defer rows.Close()

	var items []*entities.SaleItem
	for rows.Next() {
		var item entities.SaleItem
		err := rows.Scan(&item.ID, &item.SaleID, &item.ProductID, &item.ProductSKU,
			&item.ProductName, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sale items: %w", err)
	}

	return items, nil
}

// Update updates a sale item
func (r *PostgresSaleItemRepository) Update(ctx context.Context, item *entities.SaleItem) error {
	query := `
		UPDATE sale_items SET 
			product_id = $2, product_sku = $3, product_name = $4,
			quantity = $5, unit_price = $6, total_price = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.ProductSKU, item.ProductName,
		item.Quantity, item.UnitPrice, item.TotalPrice)
	if err != nil {
		return fmt.Errorf("failed to update sale item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("sale item")
	}

	return nil
}

// Delete deletes a sale item
func (r *PostgresSaleItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sale_items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sale item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("sale item")
	}

	return nil
}

// BulkCreate creates multiple sale items in a transaction
func (r *PostgresSaleItemRepository) BulkCreate(ctx context.Context, items []*entities.SaleItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO sale_items (id, sale_id, product_id, product_sku, product_name, 
			quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			item.ID, item.SaleID, item.ProductID, item.ProductSKU, item.ProductName,
			item.Quantity, item.UnitPrice, item.TotalPrice, item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create sale item: %w", err)
		}
	}

	return tx.Commit()
}

// BulkUpdate updates multiple sale items in a transaction
func (r *PostgresSaleItemRepository) BulkUpdate(ctx context.Context, items []*entities.SaleItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE sale_items SET 
			product_id = $2, product_sku = $3, product_name = $4,
			quantity = $5, unit_price = $6, total_price = $7
		WHERE id = $1`

	for _, item := range items {
		result, err := tx.ExecContext(ctx, query,
			item.ID, item.ProductID, item.ProductSKU, item.ProductName,
			item.Quantity, item.UnitPrice, item.TotalPrice)
		if err != nil {
			return fmt.Errorf("failed to update sale item: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return errors.NewNotFoundError("sale item")
		}
	}

	return tx.Commit()
}

// DeleteBySaleID deletes all items for a sale
func (r *PostgresSaleItemRepository) DeleteBySaleID(ctx context.Context, saleID uuid.UUID) error {
	query := `DELETE FROM sale_items WHERE sale_id = $1`

	_, err := r.db.ExecContext(ctx, query, saleID)
	if err != nil {
		return fmt.Errorf("failed to delete sale items: %w", err)
	}

	return nil
}

// GetTopSellingProducts retrieves top selling products by quantity or revenue
func (r *PostgresSaleItemRepository) GetTopSellingProducts(ctx context.Context, fromDate, toDate time.Time, limit int, byRevenue bool) ([]*repositories.ProductSalesStats, error) {
	orderBy := "quantity_sold DESC"
	if byRevenue {
		orderBy = "total_revenue DESC"
	}

	query := fmt.Sprintf(`
		SELECT 
			si.product_id,
			si.product_sku,
			si.product_name,
			SUM(si.quantity) as quantity_sold,
			SUM(si.total_price) as total_revenue,
			AVG(si.unit_price) as average_price,
			COUNT(DISTINCT s.id) as sales_count
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at <= $2 
			AND s.status = 'completed' AND s.deleted_at IS NULL
		GROUP BY si.product_id, si.product_sku, si.product_name
		ORDER BY %s
		LIMIT $3`, orderBy)

	rows, err := r.db.QueryContext(ctx, query, fromDate, toDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling products: %w", err)
	}
	defer rows.Close()

	var products []*repositories.ProductSalesStats
	for rows.Next() {
		var product repositories.ProductSalesStats
		err := rows.Scan(&product.ProductID, &product.ProductSKU, &product.ProductName,
			&product.QuantitySold, &product.TotalRevenue, &product.AveragePrice, &product.SalesCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product sales stats: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product sales stats: %w", err)
	}

	return products, nil
}
