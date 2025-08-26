package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/utils"
)

// PostgreSQLProductRepository implements the ProductRepository interface for PostgreSQL
type PostgreSQLProductRepository struct {
	db *sql.DB
}

// NewPostgreSQLProductRepository creates a new PostgreSQL product repository
func NewPostgreSQLProductRepository(db *sql.DB) repositories.ProductRepository {
	return &PostgreSQLProductRepository{
		db: db,
	}
}

// Create creates a new product
func (r *PostgreSQLProductRepository) Create(ctx context.Context, product *entities.Product) error {
	query := `
		INSERT INTO products (id, sku, name, description, category, price, cost, status, unit, min_stock, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
		product.Cost,
		product.Status,
		product.Unit,
		product.MinStock,
		product.CreatedAt,
		product.UpdatedAt,
		product.CreatedBy,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "sku") {
					return errors.NewConflictError("SKU already exists")
				}
				return errors.NewConflictError("product already exists")
			}
		}
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *PostgreSQLProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, cost, status, unit, min_stock, created_at, updated_at, created_by
		FROM products 
		WHERE id = $1 AND deleted_at IS NULL`

	product := &entities.Product{}
	var priceStr, costStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Category,
		&priceStr,
		&costStr,
		&product.Status,
		&product.Unit,
		&product.MinStock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("product")
		}
		return nil, fmt.Errorf("failed to get product by ID: %w", err)
	}

	// Parse decimal values
	if product.Price, err = decimal.NewFromString(priceStr); err != nil {
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}
	if product.Cost, err = decimal.NewFromString(costStr); err != nil {
		return nil, fmt.Errorf("failed to parse cost: %w", err)
	}

	return product, nil
}

// GetBySKU retrieves a product by SKU
func (r *PostgreSQLProductRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, cost, status, unit, min_stock, created_at, updated_at, created_by
		FROM products 
		WHERE sku = $1 AND deleted_at IS NULL`

	product := &entities.Product{}
	var priceStr, costStr string

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Category,
		&priceStr,
		&costStr,
		&product.Status,
		&product.Unit,
		&product.MinStock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("product")
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	// Parse decimal values
	if product.Price, err = decimal.NewFromString(priceStr); err != nil {
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}
	if product.Cost, err = decimal.NewFromString(costStr); err != nil {
		return nil, fmt.Errorf("failed to parse cost: %w", err)
	}

	return product, nil
}

// Update updates an existing product
func (r *PostgreSQLProductRepository) Update(ctx context.Context, product *entities.Product) error {
	query := `
		UPDATE products 
		SET sku = $2, name = $3, description = $4, category = $5, price = $6, cost = $7, 
		    status = $8, unit = $9, min_stock = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
		product.Cost,
		product.Status,
		product.Unit,
		product.MinStock,
		product.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "sku") {
					return errors.NewConflictError("SKU already exists")
				}
				return errors.NewConflictError("product already exists")
			}
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("product")
	}

	return nil
}

// Delete deletes a product (soft delete)
func (r *PostgreSQLProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products 
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("product")
	}

	return nil
}

// List retrieves products with pagination and filtering
func (r *PostgreSQLProductRepository) List(ctx context.Context, filter repositories.ProductFilter, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error) {
	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	whereConditions = append(whereConditions, "deleted_at IS NULL")

	if filter.Category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, filter.Category)
		argIndex++
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		searchCondition := fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR sku ILIKE $%d)", argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count products: %w", err)
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	// Build main query
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category, price, cost, status, unit, min_stock, created_at, updated_at, created_by
		FROM products 
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		product := &entities.Product{}
		var priceStr, costStr string

		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Category,
			&priceStr,
			&costStr,
			&product.Status,
			&product.Unit,
			&product.MinStock,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.CreatedBy,
		)
		if err != nil {
			return nil, pagination, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse decimal values
		if product.Price, err = decimal.NewFromString(priceStr); err != nil {
			return nil, pagination, fmt.Errorf("failed to parse price: %w", err)
		}
		if product.Cost, err = decimal.NewFromString(costStr); err != nil {
			return nil, pagination, fmt.Errorf("failed to parse cost: %w", err)
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, pagination, fmt.Errorf("failed to iterate products: %w", err)
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

	return products, resultPagination, nil
}

// GetByCategory retrieves products by category
func (r *PostgreSQLProductRepository) GetByCategory(ctx context.Context, category string, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error) {
	filter := repositories.ProductFilter{
		Category: category,
	}
	return r.List(ctx, filter, pagination)
}

// ExistsBySKU checks if a product exists by SKU
func (r *PostgreSQLProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, sku).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if product exists by SKU: %w", err)
	}

	return exists, nil
}

// GetCategories retrieves all unique categories
func (r *PostgreSQLProductRepository) GetCategories(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT category 
		FROM products 
		WHERE deleted_at IS NULL AND category IS NOT NULL AND category != ''
		ORDER BY category`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate categories: %w", err)
	}

	return categories, nil
}

// GetLowStockProducts retrieves products with low stock
func (r *PostgreSQLProductRepository) GetLowStockProducts(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error) {
	// This requires joining with stock table to compare available quantity with min_stock
	// Count total records
	countQuery := `
		SELECT COUNT(*) 
		FROM products p
		JOIN stock s ON p.id = s.product_id
		WHERE p.deleted_at IS NULL AND s.available_qty <= p.min_stock`

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count low stock products: %w", err)
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	// Main query with JOIN
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.category, p.price, p.cost, p.status, 
		       p.unit, p.min_stock, p.created_at, p.updated_at, p.created_by
		FROM products p
		JOIN stock s ON p.id = s.product_id
		WHERE p.deleted_at IS NULL AND s.available_qty <= p.min_stock
		ORDER BY s.available_qty ASC, p.name ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, pagination.Limit, offset)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to query low stock products: %w", err)
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		product := &entities.Product{}
		var priceStr, costStr string

		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Category,
			&priceStr,
			&costStr,
			&product.Status,
			&product.Unit,
			&product.MinStock,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.CreatedBy,
		)
		if err != nil {
			return nil, pagination, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse decimal values
		if product.Price, err = decimal.NewFromString(priceStr); err != nil {
			return nil, pagination, fmt.Errorf("failed to parse price: %w", err)
		}
		if product.Cost, err = decimal.NewFromString(costStr); err != nil {
			return nil, pagination, fmt.Errorf("failed to parse cost: %w", err)
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, pagination, fmt.Errorf("failed to iterate low stock products: %w", err)
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

	return products, resultPagination, nil
}
