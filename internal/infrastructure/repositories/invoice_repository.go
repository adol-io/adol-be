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

// PostgresInvoiceRepository implements the InvoiceRepository interface
type PostgresInvoiceRepository struct {
	db *sql.DB
}

// NewPostgresInvoiceRepository creates a new PostgreSQL invoice repository
func NewPostgresInvoiceRepository(db *sql.DB) repositories.InvoiceRepository {
	return &PostgresInvoiceRepository{db: db}
}

// Create creates a new invoice
func (r *PostgresInvoiceRepository) Create(ctx context.Context, invoice *entities.Invoice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert invoice
	query := `
		INSERT INTO invoices (id, invoice_number, sale_id, customer_name, customer_email, 
			customer_phone, customer_address, subtotal, tax_amount, discount_amount, 
			total_amount, paid_amount, payment_method, status, notes, due_date, paid_at,
			created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err = tx.ExecContext(ctx, query,
		invoice.ID, invoice.InvoiceNumber, invoice.SaleID, invoice.CustomerName,
		invoice.CustomerEmail, invoice.CustomerPhone, invoice.CustomerAddress,
		invoice.Subtotal, invoice.TaxAmount, invoice.DiscountAmount, invoice.TotalAmount,
		invoice.PaidAmount, invoice.PaymentMethod, invoice.Status, invoice.Notes,
		invoice.DueDate, invoice.PaidAt, invoice.CreatedAt, invoice.UpdatedAt, invoice.CreatedBy)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.NewConflictError(fmt.Sprintf("invoice with invoice_number '%s' already exists", invoice.InvoiceNumber))
		}
		return fmt.Errorf("failed to insert invoice: %w", err)
	}

	// Insert invoice items
	if len(invoice.Items) > 0 {
		if err := r.insertInvoiceItems(ctx, tx, invoice.ID, invoice.Items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves an invoice by ID
func (r *PostgresInvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error) {
	query := `
		SELECT id, invoice_number, sale_id, customer_name, customer_email, 
			customer_phone, customer_address, subtotal, tax_amount, discount_amount, 
			total_amount, paid_amount, payment_method, status, notes, due_date, paid_at,
			created_at, updated_at, created_by
		FROM invoices 
		WHERE id = $1 AND deleted_at IS NULL`

	var invoice entities.Invoice
	var customerEmail, customerPhone, customerAddress, notes sql.NullString
	var paymentMethod sql.NullString
	var dueDate, paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.SaleID, &invoice.CustomerName,
		&customerEmail, &customerPhone, &customerAddress, &invoice.Subtotal,
		&invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount,
		&invoice.PaidAmount, &paymentMethod, &invoice.Status, &notes, &dueDate, &paidAt,
		&invoice.CreatedAt, &invoice.UpdatedAt, &invoice.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("invoice")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Handle nullable fields
	invoice.CustomerEmail = customerEmail.String
	invoice.CustomerPhone = customerPhone.String
	invoice.CustomerAddress = customerAddress.String
	invoice.Notes = notes.String
	if paymentMethod.Valid {
		invoice.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
	}
	if dueDate.Valid {
		invoice.DueDate = &dueDate.Time
	}
	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}

	// Load invoice items
	items, err := r.getInvoiceItems(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return &invoice, nil
}

// GetByInvoiceNumber retrieves an invoice by invoice number
func (r *PostgresInvoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error) {
	query := `
		SELECT id, invoice_number, sale_id, customer_name, customer_email, 
			customer_phone, customer_address, subtotal, tax_amount, discount_amount, 
			total_amount, paid_amount, payment_method, status, notes, due_date, paid_at,
			created_at, updated_at, created_by
		FROM invoices 
		WHERE invoice_number = $1 AND deleted_at IS NULL`

	var invoice entities.Invoice
	var customerEmail, customerPhone, customerAddress, notes sql.NullString
	var paymentMethod sql.NullString
	var dueDate, paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, invoiceNumber).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.SaleID, &invoice.CustomerName,
		&customerEmail, &customerPhone, &customerAddress, &invoice.Subtotal,
		&invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount,
		&invoice.PaidAmount, &paymentMethod, &invoice.Status, &notes, &dueDate, &paidAt,
		&invoice.CreatedAt, &invoice.UpdatedAt, &invoice.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("invoice")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Handle nullable fields
	invoice.CustomerEmail = customerEmail.String
	invoice.CustomerPhone = customerPhone.String
	invoice.CustomerAddress = customerAddress.String
	invoice.Notes = notes.String
	if paymentMethod.Valid {
		invoice.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
	}
	if dueDate.Valid {
		invoice.DueDate = &dueDate.Time
	}
	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}

	// Load invoice items
	items, err := r.getInvoiceItems(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return &invoice, nil
}

// GetBySaleID retrieves an invoice by sale ID
func (r *PostgresInvoiceRepository) GetBySaleID(ctx context.Context, saleID uuid.UUID) (*entities.Invoice, error) {
	query := `
		SELECT id, invoice_number, sale_id, customer_name, customer_email, 
			customer_phone, customer_address, subtotal, tax_amount, discount_amount, 
			total_amount, paid_amount, payment_method, status, notes, due_date, paid_at,
			created_at, updated_at, created_by
		FROM invoices 
		WHERE sale_id = $1 AND deleted_at IS NULL`

	var invoice entities.Invoice
	var customerEmail, customerPhone, customerAddress, notes sql.NullString
	var paymentMethod sql.NullString
	var dueDate, paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, saleID).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.SaleID, &invoice.CustomerName,
		&customerEmail, &customerPhone, &customerAddress, &invoice.Subtotal,
		&invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount,
		&invoice.PaidAmount, &paymentMethod, &invoice.Status, &notes, &dueDate, &paidAt,
		&invoice.CreatedAt, &invoice.UpdatedAt, &invoice.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("invoice")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Handle nullable fields
	invoice.CustomerEmail = customerEmail.String
	invoice.CustomerPhone = customerPhone.String
	invoice.CustomerAddress = customerAddress.String
	invoice.Notes = notes.String
	if paymentMethod.Valid {
		invoice.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
	}
	if dueDate.Valid {
		invoice.DueDate = &dueDate.Time
	}
	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}

	// Load invoice items
	items, err := r.getInvoiceItems(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return &invoice, nil
}

// Update updates an existing invoice
func (r *PostgresInvoiceRepository) Update(ctx context.Context, invoice *entities.Invoice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update invoice
	query := `
		UPDATE invoices SET 
			customer_name = $2, customer_email = $3, customer_phone = $4, customer_address = $5,
			subtotal = $6, tax_amount = $7, discount_amount = $8, total_amount = $9,
			paid_amount = $10, payment_method = $11, status = $12, notes = $13,
			due_date = $14, paid_at = $15, updated_at = $16
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := tx.ExecContext(ctx, query,
		invoice.ID, invoice.CustomerName, invoice.CustomerEmail, invoice.CustomerPhone,
		invoice.CustomerAddress, invoice.Subtotal, invoice.TaxAmount, invoice.DiscountAmount,
		invoice.TotalAmount, invoice.PaidAmount, invoice.PaymentMethod, invoice.Status,
		invoice.Notes, invoice.DueDate, invoice.PaidAt, invoice.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("invoice")
	}

	// Delete existing invoice items
	if err := r.deleteInvoiceItems(ctx, tx, invoice.ID); err != nil {
		return err
	}

	// Insert updated invoice items
	if len(invoice.Items) > 0 {
		if err := r.insertInvoiceItems(ctx, tx, invoice.ID, invoice.Items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete deletes an invoice (soft delete)
func (r *PostgresInvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE invoices SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("invoice")
	}

	return nil
}

// ExistsByInvoiceNumber checks if an invoice exists by invoice number
func (r *PostgresInvoiceRepository) ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM invoices WHERE invoice_number = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, invoiceNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check invoice existence: %w", err)
	}

	return exists, nil
}

// List retrieves invoices with pagination and filtering
func (r *PostgresInvoiceRepository) List(ctx context.Context, filter repositories.InvoiceFilter, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error) {
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

	if filter.SaleID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("sale_id = $%d", argCount))
		args = append(args, *filter.SaleID)
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

	if filter.DueFromDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argCount))
		args = append(args, *filter.DueFromDate)
	}

	if filter.DueToDate != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argCount))
		args = append(args, *filter.DueToDate)
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

	if filter.Overdue != nil && *filter.Overdue {
		conditions = append(conditions, "due_date < NOW() AND status NOT IN ('paid', 'cancelled')")
	}

	if filter.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(invoice_number ILIKE $%d OR customer_name ILIKE $%d OR customer_email ILIKE $%d)", argCount, argCount, argCount))
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM invoices %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Calculate pagination
	paginationResult := utils.CalculatePagination(pagination.Page, pagination.Limit, total)
	offset := utils.GetOffset(pagination.Page, pagination.Limit)

	// Query with pagination
	query := fmt.Sprintf(`
		SELECT id, invoice_number, sale_id, customer_name, customer_email, 
			customer_phone, customer_address, subtotal, tax_amount, discount_amount, 
			total_amount, paid_amount, payment_method, status, notes, due_date, paid_at,
			created_at, updated_at, created_by
		FROM invoices 
		%s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argCount+1, argCount+2)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, paginationResult, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*entities.Invoice
	for rows.Next() {
		var invoice entities.Invoice
		var customerEmail, customerPhone, customerAddress, notes sql.NullString
		var paymentMethod sql.NullString
		var dueDate, paidAt sql.NullTime

		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.SaleID, &invoice.CustomerName,
			&customerEmail, &customerPhone, &customerAddress, &invoice.Subtotal,
			&invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount,
			&invoice.PaidAmount, &paymentMethod, &invoice.Status, &notes, &dueDate, &paidAt,
			&invoice.CreatedAt, &invoice.UpdatedAt, &invoice.CreatedBy)
		if err != nil {
			return nil, paginationResult, fmt.Errorf("failed to scan invoice: %w", err)
		}

		// Handle nullable fields
		invoice.CustomerEmail = customerEmail.String
		invoice.CustomerPhone = customerPhone.String
		invoice.CustomerAddress = customerAddress.String
		invoice.Notes = notes.String
		if paymentMethod.Valid {
			invoice.PaymentMethod = entities.PaymentMethod(paymentMethod.String)
		}
		if dueDate.Valid {
			invoice.DueDate = &dueDate.Time
		}
		if paidAt.Valid {
			invoice.PaidAt = &paidAt.Time
		}

		// Load invoice items for each invoice
		items, err := r.getInvoiceItems(ctx, invoice.ID)
		if err != nil {
			return nil, paginationResult, err
		}
		invoice.Items = items

		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, paginationResult, fmt.Errorf("failed to iterate invoices: %w", err)
	}

	return invoices, paginationResult, nil
}

// GetOverdueInvoices retrieves overdue invoices
func (r *PostgresInvoiceRepository) GetOverdueInvoices(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error) {
	// Filter for overdue invoices
	filter := repositories.InvoiceFilter{
		Overdue: &[]bool{true}[0], // Create a pointer to true
	}

	return r.List(ctx, filter, pagination)
}

// GetInvoicesByStatus retrieves invoices by status
func (r *PostgresInvoiceRepository) GetInvoicesByStatus(ctx context.Context, status entities.InvoiceStatus, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error) {
	// Filter by status
	filter := repositories.InvoiceFilter{
		Status: &status,
	}

	return r.List(ctx, filter, pagination)
}

// GetInvoiceReport generates invoice report for a date range
func (r *PostgresInvoiceRepository) GetInvoiceReport(ctx context.Context, fromDate, toDate time.Time) (*repositories.InvoiceReport, error) {
	// Get basic invoice statistics
	query := `
		SELECT 
			COUNT(*) as total_invoices,
			COALESCE(SUM(total_amount), 0) as total_amount,
			COALESCE(SUM(paid_amount), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0) as draft_invoices,
			COALESCE(SUM(CASE WHEN status = 'generated' THEN 1 ELSE 0 END), 0) as generated_invoices,
			COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0) as sent_invoices,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0) as paid_invoices,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled_invoices
		FROM invoices 
		WHERE created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL`

	var report repositories.InvoiceReport
	report.FromDate = fromDate
	report.ToDate = toDate

	err := r.db.QueryRowContext(ctx, query, fromDate, toDate).Scan(
		&report.TotalInvoices, &report.TotalAmount, &report.PaidAmount,
		&report.DraftInvoices, &report.GeneratedInvoices, &report.SentInvoices,
		&report.PaidInvoices, &report.CancelledInvoices)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice statistics: %w", err)
	}

	// Calculate outstanding amount
	report.OutstandingAmount = report.TotalAmount.Sub(report.PaidAmount)

	// Get overdue invoices count
	overdueQuery := `
		SELECT COUNT(*)
		FROM invoices 
		WHERE created_at >= $1 AND created_at <= $2 AND due_date < NOW() 
			AND status NOT IN ('paid', 'cancelled') AND deleted_at IS NULL`

	err = r.db.QueryRowContext(ctx, overdueQuery, fromDate, toDate).Scan(&report.OverdueInvoices)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices count: %w", err)
	}

	// Get average payment time
	avgPaymentQuery := `
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (paid_at - created_at))), 0)
		FROM invoices 
		WHERE created_at >= $1 AND created_at <= $2 AND status = 'paid' 
			AND paid_at IS NOT NULL AND deleted_at IS NULL`

	var avgPaymentDays float64
	err = r.db.QueryRowContext(ctx, avgPaymentQuery, fromDate, toDate).Scan(&avgPaymentDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get average payment time: %w", err)
	}
	report.AveragePaymentTime = decimal.NewFromFloat(avgPaymentDays)

	// Get payment method statistics
	paymentStats, err := r.getInvoicePaymentMethodStats(ctx, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	report.PaymentMethodStats = paymentStats

	// Get monthly invoice data
	monthlyInvoices, err := r.getMonthlyInvoiceData(ctx, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	report.MonthlyInvoices = monthlyInvoices

	return &report, nil
}

// Helper functions

// insertInvoiceItems inserts invoice items in a transaction
func (r *PostgresInvoiceRepository) insertInvoiceItems(ctx context.Context, tx *sql.Tx, invoiceID uuid.UUID, items []entities.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (id, invoice_id, product_id, product_sku, product_name, 
			description, quantity, unit_price, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			item.ID, invoiceID, item.ProductID, item.ProductSKU, item.ProductName,
			item.Description, item.Quantity, item.UnitPrice, item.TotalPrice)
		if err != nil {
			return fmt.Errorf("failed to insert invoice item: %w", err)
		}
	}

	return nil
}

// deleteInvoiceItems deletes all invoice items for an invoice
func (r *PostgresInvoiceRepository) deleteInvoiceItems(ctx context.Context, tx *sql.Tx, invoiceID uuid.UUID) error {
	query := `DELETE FROM invoice_items WHERE invoice_id = $1`

	_, err := tx.ExecContext(ctx, query, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	return nil
}

// getInvoiceItems retrieves all items for an invoice
func (r *PostgresInvoiceRepository) getInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]entities.InvoiceItem, error) {
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

	var items []entities.InvoiceItem
	for rows.Next() {
		var item entities.InvoiceItem
		var description sql.NullString
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.ProductID, &item.ProductSKU,
			&item.ProductName, &description, &item.Quantity, &item.UnitPrice, &item.TotalPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}
		item.Description = description.String
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoice items: %w", err)
	}

	return items, nil
}

// getInvoicePaymentMethodStats gets payment method statistics for invoices
func (r *PostgresInvoiceRepository) getInvoicePaymentMethodStats(ctx context.Context, fromDate, toDate time.Time) ([]repositories.PaymentMethodStat, error) {
	query := `
		SELECT 
			payment_method,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM invoices 
		WHERE created_at >= $1 AND created_at <= $2 AND status = 'paid' 
			AND deleted_at IS NULL AND payment_method IS NOT NULL
		GROUP BY payment_method
		ORDER BY total_amount DESC`

	rows, err := r.db.QueryContext(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice payment method stats: %w", err)
	}
	defer rows.Close()

	var stats []repositories.PaymentMethodStat
	var totalRevenue decimal.Decimal

	// First pass: collect data and calculate total
	for rows.Next() {
		var stat repositories.PaymentMethodStat
		err := rows.Scan(&stat.PaymentMethod, &stat.Count, &stat.TotalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice payment method stat: %w", err)
		}
		stats = append(stats, stat)
		totalRevenue = totalRevenue.Add(stat.TotalAmount)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoice payment method stats: %w", err)
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

// getMonthlyInvoiceData gets monthly invoice data for a date range
func (r *PostgresInvoiceRepository) getMonthlyInvoiceData(ctx context.Context, fromDate, toDate time.Time) ([]repositories.MonthlyInvoiceData, error) {
	query := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as total_invoices,
			COALESCE(SUM(total_amount), 0) as total_amount,
			COALESCE(SUM(paid_amount), 0) as paid_amount
		FROM invoices 
		WHERE created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month`

	rows, err := r.db.QueryContext(ctx, query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly invoice data: %w", err)
	}
	defer rows.Close()

	var monthlyData []repositories.MonthlyInvoiceData
	for rows.Next() {
		var data repositories.MonthlyInvoiceData
		err := rows.Scan(&data.Month, &data.TotalInvoices, &data.TotalAmount, &data.PaidAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly invoice data: %w", err)
		}
		monthlyData = append(monthlyData, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate monthly invoice data: %w", err)
	}

	return monthlyData, nil
}
