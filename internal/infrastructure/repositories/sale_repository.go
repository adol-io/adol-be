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

// PostgresSaleRepository implements the SaleRepository interface
type PostgresSaleRepository struct {
	db *sql.DB
}

// NewPostgresSaleRepository creates a new PostgreSQL sale repository
func NewPostgresSaleRepository(db *sql.DB) repositories.SaleRepository {
	return &PostgresSaleRepository{db: db}
}

// Create creates a new sale
func (r *PostgresSaleRepository) Create(ctx context.Context, sale *entities.Sale) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert sale
	query := `
		INSERT INTO sales (id, sale_number, customer_name, customer_email, customer_phone,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, change_amount,
			payment_method, status, notes, created_at, updated_at, completed_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err = tx.ExecContext(ctx, query,
		sale.ID, sale.SaleNumber, sale.CustomerName, sale.CustomerEmail, sale.CustomerPhone,
		sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.PaidAmount, sale.ChangeAmount, sale.PaymentMethod, sale.Status, sale.Notes,
		sale.CreatedAt, sale.UpdatedAt, sale.CompletedAt, sale.CreatedBy)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.NewConflictError(fmt.Sprintf("sale with sale_number '%s' already exists", sale.SaleNumber))
		}
		return fmt.Errorf("failed to insert sale: %w", err)
	}

	// Insert sale items
	if len(sale.Items) > 0 {
		if err := r.insertSaleItems(ctx, tx, sale.ID, sale.Items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves a sale by ID
func (r *PostgresSaleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Sale, error) {
	query := `
		SELECT id, sale_number, customer_name, customer_email, customer_phone,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, change_amount,
			payment_method, status, notes, created_at, updated_at, completed_at, created_by
		FROM sales 
		WHERE id = $1 AND deleted_at IS NULL`

	var sale entities.Sale
	var customerName, customerEmail, customerPhone, notes sql.NullString
	var paymentMethod sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sale.ID, &sale.SaleNumber, &customerName, &customerEmail, &customerPhone,
		&sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount, &sale.TotalAmount,
		&sale.PaidAmount, &sale.ChangeAmount, &paymentMethod, &sale.Status, &notes,
		&sale.CreatedAt, &sale.UpdatedAt, &completedAt, &sale.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("sale")
		}
		return nil, fmt.Errorf("failed to get sale: %w", err)
	}

	// Handle nullable fields
	sale.CustomerName = customerName.String
	sale.CustomerEmail = customerEmail.String
	sale.CustomerPhone = customerPhone.String
	sale.Notes = notes.String
	if paymentMethod.Valid {
		sale.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
	}
	if completedAt.Valid {
		sale.CompletedAt = &completedAt.Time
	}

	// Load sale items
	items, err := r.getSaleItems(ctx, sale.ID)
	if err != nil {
		return nil, err
	}
	sale.Items = items

	return &sale, nil
}

// GetBySaleNumber retrieves a sale by sale number
func (r *PostgresSaleRepository) GetBySaleNumber(ctx context.Context, saleNumber string) (*entities.Sale, error) {
	query := `
		SELECT id, sale_number, customer_name, customer_email, customer_phone,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, change_amount,
			payment_method, status, notes, created_at, updated_at, completed_at, created_by
		FROM sales 
		WHERE sale_number = $1 AND deleted_at IS NULL`

	var sale entities.Sale
	var customerName, customerEmail, customerPhone, notes sql.NullString
	var paymentMethod sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, saleNumber).Scan(
		&sale.ID, &sale.SaleNumber, &customerName, &customerEmail, &customerPhone,
		&sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount, &sale.TotalAmount,
		&sale.PaidAmount, &sale.ChangeAmount, &paymentMethod, &sale.Status, &notes,
		&sale.CreatedAt, &sale.UpdatedAt, &completedAt, &sale.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("sale")
		}
		return nil, fmt.Errorf("failed to get sale: %w", err)
	}

	// Handle nullable fields
	sale.CustomerName = customerName.String
	sale.CustomerEmail = customerEmail.String
	sale.CustomerPhone = customerPhone.String
	sale.Notes = notes.String
	if paymentMethod.Valid {
		sale.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
	}
	if completedAt.Valid {
		sale.CompletedAt = &completedAt.Time
	}

	// Load sale items
	items, err := r.getSaleItems(ctx, sale.ID)
	if err != nil {
		return nil, err
	}
	sale.Items = items

	return &sale, nil
}

// Update updates an existing sale
func (r *PostgresSaleRepository) Update(ctx context.Context, sale *entities.Sale) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update sale
	query := `
		UPDATE sales SET 
			customer_name = $2, customer_email = $3, customer_phone = $4,
			subtotal = $5, tax_amount = $6, discount_amount = $7, total_amount = $8,
			paid_amount = $9, change_amount = $10, payment_method = $11, status = $12,
			notes = $13, updated_at = $14, completed_at = $15
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := tx.ExecContext(ctx, query,
		sale.ID, sale.CustomerName, sale.CustomerEmail, sale.CustomerPhone,
		sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.PaidAmount, sale.ChangeAmount, sale.PaymentMethod, sale.Status,
		sale.Notes, sale.UpdatedAt, sale.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("sale")
	}

	// Delete existing sale items
	if err := r.deleteSaleItems(ctx, tx, sale.ID); err != nil {
		return err
	}

	// Insert updated sale items
	if len(sale.Items) > 0 {
		if err := r.insertSaleItems(ctx, tx, sale.ID, sale.Items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete deletes a sale (soft delete)
func (r *PostgresSaleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sales SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sale: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("sale")
	}

	return nil
}

// List retrieves sales with pagination and filtering
func (r *PostgresSaleRepository) List(ctx context.Context, filter repositories.SaleFilter, pagination utils.PaginationInfo) ([]*entities.Sale, utils.PaginationInfo, error) {
	// Build WHERE clause
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argCount := 0

	if filter.Status != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *filter.Status)
	}

	if filter.PaymentMethod != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argCount))
		args = append(args, *filter.PaymentMethod)
	}

	if filter.CreatedBy != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argCount))
		args = append(args, *filter.CreatedBy)
	}

	if filter.CustomerName != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("customer_name ILIKE $%d", argCount))
		args = append(args, "%"+filter.CustomerName+"%")
	}

	if filter.CustomerEmail != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("customer_email ILIKE $%d", argCount))
		args = append(args, "%"+filter.CustomerEmail+"%")
	}

	if filter.FromDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, *filter.FromDate)
	}

	if filter.ToDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, *filter.ToDate)
	}

	if filter.MinAmount != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("total_amount >= $%d", argCount))
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("total_amount <= $%d", argCount))
		args = append(args, *filter.MaxAmount)
	}

	if filter.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(sale_number ILIKE $%d OR customer_name ILIKE $%d OR customer_email ILIKE $%d)", argCount, argCount, argCount))
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count sales: %w", err)
	}

	// Calculate pagination
	paginationResult := utils.CalculatePagination(pagination.Page, pagination.Limit, total)
	offset := utils.GetOffset(pagination.Page, pagination.Limit)

	// Query with pagination
	query := fmt.Sprintf(`
		SELECT id, sale_number, customer_name, customer_email, customer_phone,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, change_amount,
			payment_method, status, notes, created_at, updated_at, completed_at, created_by
		FROM sales 
		%s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argCount+1, argCount+2)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, paginationResult, fmt.Errorf("failed to query sales: %w", err)
	}
	defer rows.Close()

	var sales []*entities.Sale
	for rows.Next() {
		var sale entities.Sale
		var customerName, customerEmail, customerPhone, notes sql.NullString
		var paymentMethod sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&sale.ID, &sale.SaleNumber, &customerName, &customerEmail, &customerPhone,
			&sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount, &sale.TotalAmount,
			&sale.PaidAmount, &sale.ChangeAmount, &paymentMethod, &sale.Status, &notes,
			&sale.CreatedAt, &sale.UpdatedAt, &completedAt, &sale.CreatedBy)
		if err != nil {
			return nil, paginationResult, fmt.Errorf("failed to scan sale: %w", err)
		}

		// Handle nullable fields
		sale.CustomerName = customerName.String
		sale.CustomerEmail = customerEmail.String
		sale.CustomerPhone = customerPhone.String
		sale.Notes = notes.String
		if paymentMethod.Valid {
			sale.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
		}
		if completedAt.Valid {
			sale.CompletedAt = &completedAt.Time
		}

		// Load sale items for each sale
		items, err := r.getSaleItems(ctx, sale.ID)
		if err != nil {
			return nil, paginationResult, err
		}
		sale.Items = items

		sales = append(sales, &sale)
	}

	if err = rows.Err(); err != nil {
		return nil, paginationResult, fmt.Errorf("failed to iterate sales: %w", err)
	}

	return sales, paginationResult, nil
}

// ExistsBySaleNumber checks if a sale exists by sale number
func (r *PostgresSaleRepository) ExistsBySaleNumber(ctx context.Context, saleNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sales WHERE sale_number = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, saleNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check sale existence: %w", err)
	}

	return exists, nil
}

// GetSalesReport generates sales report for a date range
func (r *PostgresSaleRepository) GetSalesReport(ctx context.Context, fromDate, toDate time.Time) (*repositories.SalesReport, error) {
	// Get basic sales statistics
	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed_sales,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled_sales,
			COALESCE(SUM(CASE WHEN status = 'refunded' THEN 1 ELSE 0 END), 0) as refunded_sales,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN total_amount END), 0) as average_order_value
		FROM sales 
		WHERE created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL`

	var report repositories.SalesReport
	report.FromDate = fromDate
	report.ToDate = toDate

	err := r.db.QueryRowContext(ctx, query, fromDate, toDate).Scan(
		&report.TotalSales, &report.TotalRevenue, &report.CompletedSales,
		&report.CancelledSales, &report.RefundedSales, &report.AverageOrderValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales statistics: %w", err)
	}

	// Get total items sold
	itemsQuery := `
		SELECT COALESCE(SUM(si.quantity), 0)
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at <= $2 AND s.deleted_at IS NULL`

	err = r.db.QueryRowContext(ctx, itemsQuery, fromDate, toDate).Scan(&report.TotalItemsSold)
	if err != nil {
		return nil, fmt.Errorf("failed to get total items sold: %w", err)
	}

	// Get unique customers count
	customersQuery := `
		SELECT COUNT(DISTINCT customer_email)
		FROM sales 
		WHERE created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL AND customer_email IS NOT NULL AND customer_email != ''`

	err = r.db.QueryRowContext(ctx, customersQuery, fromDate, toDate).Scan(&report.UniqueCustomers)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique customers: %w", err)
	}

	// Get payment method statistics
	paymentStats, err := r.getPaymentMethodStats(ctx, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	report.PaymentMethodStats = paymentStats

	// Get daily sales data
	dailySales, err := r.getDailySalesData(ctx, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	report.DailySales = dailySales

	// Calculate profit (requires product cost data)
	report.TotalProfit = decimal.Zero // TODO: Implement profit calculation when product cost tracking is added

	return &report, nil
}

// GetDailySales retrieves daily sales summary
func (r *PostgresSaleRepository) GetDailySales(ctx context.Context, date time.Time) (*repositories.DailySalesReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed_sales,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled_sales,
			COALESCE(SUM(CASE WHEN status = 'refunded' THEN 1 ELSE 0 END), 0) as refunded_sales,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN total_amount END), 0) as average_order_value
		FROM sales 
		WHERE created_at >= $1 AND created_at < $2 AND deleted_at IS NULL`

	var report repositories.DailySalesReport
	report.Date = date

	err := r.db.QueryRowContext(ctx, query, startOfDay, endOfDay).Scan(
		&report.TotalSales, &report.TotalRevenue, &report.CompletedSales,
		&report.CancelledSales, &report.RefundedSales, &report.AverageOrderValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily sales: %w", err)
	}

	// Get total items sold for the day
	itemsQuery := `
		SELECT COALESCE(SUM(si.quantity), 0)
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.deleted_at IS NULL`

	err = r.db.QueryRowContext(ctx, itemsQuery, startOfDay, endOfDay).Scan(&report.TotalItemsSold)
	if err != nil {
		return nil, fmt.Errorf("failed to get total items sold: %w", err)
	}

	// Get top selling products for the day
	topProducts, err := r.getTopSellingProductsForDay(ctx, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	report.TopSellingProducts = topProducts

	return &report, nil
}

// GetTotalSalesByUser retrieves total sales amount by user
func (r *PostgresSaleRepository) GetTotalSalesByUser(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM sales 
		WHERE created_by = $1 AND created_at >= $2 AND created_at <= $3 
			AND status = 'completed' AND deleted_at IS NULL`

	var total decimal.Decimal
	err := r.db.QueryRowContext(ctx, query, userID, fromDate, toDate).Scan(&total)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total sales by user: %w", err)
	}

	return total, nil
}

// Helper functions

// insertSaleItems inserts sale items in a transaction
func (r *PostgresSaleRepository) insertSaleItems(ctx context.Context, tx *sql.Tx, saleID uuid.UUID, items []entities.SaleItem) error {
	query := `
		INSERT INTO sale_items (id, sale_id, product_id, product_sku, product_name, 
			quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			item.ID, saleID, item.ProductID, item.ProductSKU, item.ProductName,
			item.Quantity, item.UnitPrice, item.TotalPrice, item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert sale item: %w", err)
		}
	}

	return nil
}

// deleteSaleItems deletes all sale items for a sale
func (r *PostgresSaleRepository) deleteSaleItems(ctx context.Context, tx *sql.Tx, saleID uuid.UUID) error {
	query := `DELETE FROM sale_items WHERE sale_id = $1`

	_, err := tx.ExecContext(ctx, query, saleID)
	if err != nil {
		return fmt.Errorf("failed to delete sale items: %w", err)
	}

	return nil
}

// getSaleItems retrieves all items for a sale
func (r *PostgresSaleRepository) getSaleItems(ctx context.Context, saleID uuid.UUID) ([]entities.SaleItem, error) {
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

	var items []entities.SaleItem
	for rows.Next() {
		var item entities.SaleItem
		err := rows.Scan(&item.ID, &item.SaleID, &item.ProductID, &item.ProductSKU,
			&item.ProductName, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sale items: %w", err)
	}

	return items, nil
}

// getPaymentMethodStats gets payment method statistics for a date range
func (r *PostgresSaleRepository) getPaymentMethodStats(ctx context.Context, fromDate, toDate time.Time) ([]repositories.PaymentMethodStat, error) {
	query := `
		SELECT 
			payment_method,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM sales 
		WHERE created_at >= $1 AND created_at <= $2 AND status = 'completed' 
			AND deleted_at IS NULL AND payment_method IS NOT NULL
		GROUP BY payment_method
		ORDER BY total_amount DESC`

	rows, err := r.db.QueryContext(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment method stats: %w", err)
	}
	defer rows.Close()

	var stats []repositories.PaymentMethodStat
	var totalRevenue decimal.Decimal

	// First pass: collect data and calculate total
	for rows.Next() {
		var stat repositories.PaymentMethodStat
		err := rows.Scan(&stat.PaymentMethod, &stat.Count, &stat.TotalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment method stat: %w", err)
		}
		stats = append(stats, stat)
		totalRevenue = totalRevenue.Add(stat.TotalAmount)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate payment method stats: %w", err)
	}

	// Second pass: calculate percentages
	for i := range stats {
		if totalRevenue.GreaterThan(decimal.Zero) {
			stats[i].Percentage = stats[i].TotalAmount.Div(totalRevenue).Mul(decimal.NewFromInt(100))
		} else {
			stats[i].Percentage = decimal.Zero
		}
	}

	return stats, nil
}

// getDailySalesData gets daily sales data for a date range
func (r *PostgresSaleRepository) getDailySalesData(ctx context.Context, fromDate, toDate time.Time) ([]repositories.DailySalesData, error) {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue
		FROM sales 
		WHERE created_at >= $1 AND created_at <= $2 AND status = 'completed' AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date`

	rows, err := r.db.QueryContext(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily sales data: %w", err)
	}
	defer rows.Close()

	var dailySales []repositories.DailySalesData
	for rows.Next() {
		var data repositories.DailySalesData
		err := rows.Scan(&data.Date, &data.TotalSales, &data.TotalRevenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily sales data: %w", err)
		}
		dailySales = append(dailySales, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate daily sales data: %w", err)
	}

	return dailySales, nil
}

// getTopSellingProductsForDay gets top selling products for a specific day
func (r *PostgresSaleRepository) getTopSellingProductsForDay(ctx context.Context, startOfDay, endOfDay time.Time) ([]repositories.ProductSalesStats, error) {
	query := `
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
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed' AND s.deleted_at IS NULL
		GROUP BY si.product_id, si.product_sku, si.product_name
		ORDER BY quantity_sold DESC
		LIMIT 10`

	rows, err := r.db.QueryContext(ctx, query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling products: %w", err)
	}
	defer rows.Close()

	var products []repositories.ProductSalesStats
	for rows.Next() {
		var product repositories.ProductSalesStats
		err := rows.Scan(&product.ProductID, &product.ProductSKU, &product.ProductName,
			&product.QuantitySold, &product.TotalRevenue, &product.AveragePrice, &product.SalesCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product sales stats: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product sales stats: %w", err)
	}

	return products, nil
}
