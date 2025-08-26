package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/utils"
)

// InvoiceRepository defines the interface for invoice data access
type InvoiceRepository interface {
	// Create creates a new invoice
	Create(ctx context.Context, invoice *entities.Invoice) error

	// GetByID retrieves an invoice by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error)

	// GetByInvoiceNumber retrieves an invoice by invoice number
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)

	// GetBySaleID retrieves an invoice by sale ID
	GetBySaleID(ctx context.Context, saleID uuid.UUID) (*entities.Invoice, error)

	// Update updates an existing invoice
	Update(ctx context.Context, invoice *entities.Invoice) error

	// Delete deletes an invoice (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves invoices with pagination and filtering
	List(ctx context.Context, filter InvoiceFilter, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error)

	// GetOverdueInvoices retrieves overdue invoices
	GetOverdueInvoices(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error)

	// GetInvoiceReport generates invoice report for a date range
	GetInvoiceReport(ctx context.Context, fromDate, toDate time.Time) (*InvoiceReport, error)

	// ExistsByInvoiceNumber checks if an invoice exists by invoice number
	ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error)

	// GetInvoicesByStatus retrieves invoices by status
	GetInvoicesByStatus(ctx context.Context, status entities.InvoiceStatus, pagination utils.PaginationInfo) ([]*entities.Invoice, utils.PaginationInfo, error)
}

// InvoiceItemRepository defines the interface for invoice item data access
type InvoiceItemRepository interface {
	// Create creates a new invoice item
	Create(ctx context.Context, item *entities.InvoiceItem) error

	// GetByID retrieves an invoice item by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.InvoiceItem, error)

	// GetByInvoiceID retrieves all items for an invoice
	GetByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*entities.InvoiceItem, error)

	// Update updates an invoice item
	Update(ctx context.Context, item *entities.InvoiceItem) error

	// Delete deletes an invoice item
	Delete(ctx context.Context, id uuid.UUID) error

	// BulkCreate creates multiple invoice items in a transaction
	BulkCreate(ctx context.Context, items []*entities.InvoiceItem) error

	// BulkUpdate updates multiple invoice items in a transaction
	BulkUpdate(ctx context.Context, items []*entities.InvoiceItem) error

	// DeleteByInvoiceID deletes all items for an invoice
	DeleteByInvoiceID(ctx context.Context, invoiceID uuid.UUID) error
}

// InvoiceFilter represents filters for invoice queries
type InvoiceFilter struct {
	Status        *entities.InvoiceStatus `json:"status,omitempty"`
	PaymentMethod *entities.PaymentMethod `json:"payment_method,omitempty"`
	CreatedBy     *uuid.UUID              `json:"created_by,omitempty"`
	CustomerName  string                  `json:"customer_name,omitempty"`
	CustomerEmail string                  `json:"customer_email,omitempty"`
	SaleID        *uuid.UUID              `json:"sale_id,omitempty"`
	FromDate      *time.Time              `json:"from_date,omitempty"`
	ToDate        *time.Time              `json:"to_date,omitempty"`
	DueFromDate   *time.Time              `json:"due_from_date,omitempty"`
	DueToDate     *time.Time              `json:"due_to_date,omitempty"`
	MinAmount     *decimal.Decimal        `json:"min_amount,omitempty"`
	MaxAmount     *decimal.Decimal        `json:"max_amount,omitempty"`
	Overdue       *bool                   `json:"overdue,omitempty"`
	Search        string                  `json:"search,omitempty"` // Search in invoice_number, customer_name, customer_email
	OrderBy       string                  `json:"order_by,omitempty"`
	OrderDir      string                  `json:"order_dir,omitempty"` // ASC or DESC
}

// InvoiceReport represents an invoice report for a date range
type InvoiceReport struct {
	FromDate           time.Time            `json:"from_date"`
	ToDate             time.Time            `json:"to_date"`
	TotalInvoices      int                  `json:"total_invoices"`
	TotalAmount        decimal.Decimal      `json:"total_amount"`
	PaidAmount         decimal.Decimal      `json:"paid_amount"`
	OutstandingAmount  decimal.Decimal      `json:"outstanding_amount"`
	DraftInvoices      int                  `json:"draft_invoices"`
	GeneratedInvoices  int                  `json:"generated_invoices"`
	SentInvoices       int                  `json:"sent_invoices"`
	PaidInvoices       int                  `json:"paid_invoices"`
	CancelledInvoices  int                  `json:"cancelled_invoices"`
	OverdueInvoices    int                  `json:"overdue_invoices"`
	AveragePaymentTime decimal.Decimal      `json:"average_payment_time"` // in days
	PaymentMethodStats []PaymentMethodStat  `json:"payment_method_stats"`
	MonthlyInvoices    []MonthlyInvoiceData `json:"monthly_invoices"`
}

// MonthlyInvoiceData represents monthly invoice data point
type MonthlyInvoiceData struct {
	Month         time.Time       `json:"month"`
	TotalInvoices int             `json:"total_invoices"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	PaidAmount    decimal.Decimal `json:"paid_amount"`
}
