package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
)

// PostgresInvoiceItemRepository implements the InvoiceItemRepository interface
type PostgresInvoiceItemRepository struct {
	db *sql.DB
}

// NewPostgresInvoiceItemRepository creates a new PostgreSQL invoice item repository
func NewPostgresInvoiceItemRepository(db *sql.DB) repositories.InvoiceItemRepository {
	return &PostgresInvoiceItemRepository{db: db}
}

// Create creates a new invoice item
func (r *PostgresInvoiceItemRepository) Create(ctx context.Context, item *entities.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (id, invoice_id, product_id, product_sku, product_name, 
			description, quantity, unit_price, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.InvoiceID, item.ProductID, item.ProductSKU, item.ProductName,
		item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)
	if err != nil {
		return fmt.Errorf("failed to create invoice item: %w", err)
	}

	return nil
}

// GetByID retrieves an invoice item by ID
func (r *PostgresInvoiceItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, product_id, product_sku, product_name, 
			description, quantity, unit_price, total_price
		FROM invoice_items 
		WHERE id = $1`

	var item entities.InvoiceItem
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductSKU, &item.ProductName,
		&description, &item.Quantity, &item.UnitPrice, &item.TotalPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("invoice item")
		}
		return nil, fmt.Errorf("failed to get invoice item: %w", err)
	}

	item.Description = description.String
	return &item, nil
}

// GetByInvoiceID retrieves all items for an invoice
func (r *PostgresInvoiceItemRepository) GetByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*entities.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, product_id, product_sku, product_name, 
			description, quantity, unit_price, total_price
		FROM invoice_items 
		WHERE invoice_id = $1 
		ORDER BY product_name`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice items: %w", err)
	}
	defer rows.Close()

	var items []*entities.InvoiceItem
	for rows.Next() {
		var item entities.InvoiceItem
		var description sql.NullString
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductSKU,
			&item.ProductName, &description, &item.Quantity, &item.UnitPrice, &item.TotalPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}
		item.Description = description.String
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoice items: %w", err)
	}

	return items, nil
}

// Update updates an invoice item
func (r *PostgresInvoiceItemRepository) Update(ctx context.Context, item *entities.InvoiceItem) error {
	query := `
		UPDATE invoice_items SET 
			product_id = $2, product_sku = $3, product_name = $4, description = $5,
			quantity = $6, unit_price = $7, total_price = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.ProductSKU, item.ProductName, item.Description,
		item.Quantity, item.UnitPrice, item.TotalPrice)
	if err != nil {
		return fmt.Errorf("failed to update invoice item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("invoice item")
	}

	return nil
}

// Delete deletes an invoice item
func (r *PostgresInvoiceItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invoice_items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("invoice item")
	}

	return nil
}

// BulkCreate creates multiple invoice items in a transaction
func (r *PostgresInvoiceItemRepository) BulkCreate(ctx context.Context, items []*entities.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO invoice_items (id, invoice_id, product_id, product_sku, product_name, 
			description, quantity, unit_price, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			item.ID, item.InvoiceID, item.ProductID, item.ProductSKU, item.ProductName,
			item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)
		if err != nil {
			return fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	return tx.Commit()
}

// BulkUpdate updates multiple invoice items in a transaction
func (r *PostgresInvoiceItemRepository) BulkUpdate(ctx context.Context, items []*entities.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE invoice_items SET 
			product_id = $2, product_sku = $3, product_name = $4, description = $5,
			quantity = $6, unit_price = $7, total_price = $8
		WHERE id = $1`

	for _, item := range items {
		result, err := tx.ExecContext(ctx, query,
			item.ID, item.ProductID, item.ProductSKU, item.ProductName, item.Description,
			item.Quantity, item.UnitPrice, item.TotalPrice)
		if err != nil {
			return fmt.Errorf("failed to update invoice item: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return errors.NewNotFoundError("invoice item")
		}
	}

	return tx.Commit()
}

// DeleteByInvoiceID deletes all items for an invoice
func (r *PostgresInvoiceItemRepository) DeleteByInvoiceID(ctx context.Context, invoiceID uuid.UUID) error {
	query := `DELETE FROM invoice_items WHERE invoice_id = $1`

	_, err := r.db.ExecContext(ctx, query, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	return nil
}
