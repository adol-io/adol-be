package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/pkg/errors"
)

// InvoiceStatus represents invoice status
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusGenerated InvoiceStatus = "generated"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// PaperSize represents paper size for invoice printing
type PaperSize string

const (
	PaperSizeA4      PaperSize = "a4"
	PaperSizeA5      PaperSize = "a5"
	PaperSizeLetter  PaperSize = "letter"
	PaperSizeLegal   PaperSize = "legal"
	PaperSizeReceipt PaperSize = "receipt" // For thermal receipt printers (80mm)
)

// Invoice represents an invoice
type Invoice struct {
	ID              uuid.UUID       `json:"id"`
	InvoiceNumber   string          `json:"invoice_number"`
	SaleID          uuid.UUID       `json:"sale_id"`
	CustomerName    string          `json:"customer_name"`
	CustomerEmail   string          `json:"customer_email,omitempty"`
	CustomerPhone   string          `json:"customer_phone,omitempty"`
	CustomerAddress string          `json:"customer_address,omitempty"`
	Items           []InvoiceItem   `json:"items"`
	Subtotal        decimal.Decimal `json:"subtotal"`
	TaxAmount       decimal.Decimal `json:"tax_amount"`
	DiscountAmount  decimal.Decimal `json:"discount_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	PaidAmount      decimal.Decimal `json:"paid_amount"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	Status          InvoiceStatus   `json:"status"`
	Notes           string          `json:"notes,omitempty"`
	DueDate         *time.Time      `json:"due_date,omitempty"`
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CreatedBy       uuid.UUID       `json:"created_by"`
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	ID          uuid.UUID       `json:"id"`
	InvoiceID   uuid.UUID       `json:"invoice_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	ProductSKU  string          `json:"product_sku"`
	ProductName string          `json:"product_name"`
	Description string          `json:"description,omitempty"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price"`
}

// CompanyInfo represents company information for invoice
type CompanyInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
}

// InvoiceTemplate represents invoice template configuration
type InvoiceTemplate struct {
	PaperSize   PaperSize       `json:"paper_size"`
	CompanyInfo CompanyInfo     `json:"company_info"`
	ShowLogo    bool            `json:"show_logo"`
	LogoPath    string          `json:"logo_path,omitempty"`
	Footer      string          `json:"footer,omitempty"`
	IncludeTax  bool            `json:"include_tax"`
	TaxRate     decimal.Decimal `json:"tax_rate"`
	Currency    string          `json:"currency"`
	Locale      string          `json:"locale"`
}

// NewInvoice creates a new invoice from a sale
func NewInvoice(invoiceNumber string, sale *Sale, createdBy uuid.UUID) (*Invoice, error) {
	if invoiceNumber == "" {
		return nil, errors.NewValidationError("invoice number is required", "invoice_number cannot be empty")
	}
	if sale == nil {
		return nil, errors.NewValidationError("sale is required", "sale cannot be nil")
	}
	if !sale.IsCompleted() {
		return nil, errors.NewValidationError("invalid sale status", "can only create invoice for completed sales")
	}

	now := time.Now()
	invoice := &Invoice{
		ID:             uuid.New(),
		InvoiceNumber:  invoiceNumber,
		SaleID:         sale.ID,
		CustomerName:   sale.CustomerName,
		CustomerEmail:  sale.CustomerEmail,
		CustomerPhone:  sale.CustomerPhone,
		Items:          convertSaleItemsToInvoiceItems(sale.Items),
		Subtotal:       sale.Subtotal,
		TaxAmount:      sale.TaxAmount,
		DiscountAmount: sale.DiscountAmount,
		TotalAmount:    sale.TotalAmount,
		PaidAmount:     sale.PaidAmount,
		PaymentMethod:  sale.PaymentMethod,
		Status:         InvoiceStatusDraft,
		Notes:          sale.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      createdBy,
	}

	return invoice, nil
}

// NewInvoiceItem creates a new invoice item
func NewInvoiceItem(invoiceID, productID uuid.UUID, productSKU, productName, description string, quantity int, unitPrice decimal.Decimal) (*InvoiceItem, error) {
	if quantity <= 0 {
		return nil, errors.NewInvalidQuantityError(quantity)
	}
	if unitPrice.LessThanOrEqual(decimal.Zero) {
		return nil, errors.NewInvalidPriceError(unitPrice.InexactFloat64())
	}
	if productSKU == "" {
		return nil, errors.NewValidationError("product SKU is required", "product_sku cannot be empty")
	}
	if productName == "" {
		return nil, errors.NewValidationError("product name is required", "product_name cannot be empty")
	}

	totalPrice := unitPrice.Mul(decimal.NewFromInt(int64(quantity)))

	item := &InvoiceItem{
		ID:          uuid.New(),
		InvoiceID:   invoiceID,
		ProductID:   productID,
		ProductSKU:  productSKU,
		ProductName: productName,
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  totalPrice,
	}

	return item, nil
}

// UpdateCustomerInfo updates customer information
func (i *Invoice) UpdateCustomerInfo(name, email, phone, address string) error {
	if name == "" {
		return errors.NewValidationError("customer name is required", "customer_name cannot be empty")
	}

	i.CustomerName = name
	i.CustomerEmail = email
	i.CustomerPhone = phone
	i.CustomerAddress = address
	i.UpdatedAt = time.Now()

	return nil
}

// SetDueDate sets the due date for the invoice
func (i *Invoice) SetDueDate(dueDate time.Time) error {
	if dueDate.Before(i.CreatedAt) {
		return errors.NewValidationError("invalid due date", "due date cannot be before invoice creation date")
	}

	i.DueDate = &dueDate
	i.UpdatedAt = time.Now()

	return nil
}

// MarkAsGenerated marks the invoice as generated
func (i *Invoice) MarkAsGenerated() error {
	if i.Status != InvoiceStatusDraft {
		return errors.NewValidationError("invalid invoice status", "only draft invoices can be marked as generated")
	}

	i.Status = InvoiceStatusGenerated
	i.UpdatedAt = time.Now()

	return nil
}

// MarkAsSent marks the invoice as sent
func (i *Invoice) MarkAsSent() error {
	if i.Status != InvoiceStatusGenerated {
		return errors.NewValidationError("invalid invoice status", "only generated invoices can be marked as sent")
	}

	i.Status = InvoiceStatusSent
	i.UpdatedAt = time.Now()

	return nil
}

// MarkAsPaid marks the invoice as paid
func (i *Invoice) MarkAsPaid() error {
	if i.Status == InvoiceStatusCancelled {
		return errors.NewValidationError("invalid invoice status", "cancelled invoices cannot be marked as paid")
	}
	if i.Status == InvoiceStatusPaid {
		return errors.NewValidationError("invalid invoice status", "invoice is already paid")
	}

	i.Status = InvoiceStatusPaid
	now := time.Now()
	i.PaidAt = &now
	i.UpdatedAt = now

	return nil
}

// Cancel cancels the invoice
func (i *Invoice) Cancel() error {
	if i.Status == InvoiceStatusPaid {
		return errors.NewValidationError("invalid invoice status", "paid invoices cannot be cancelled")
	}
	if i.Status == InvoiceStatusCancelled {
		return errors.NewValidationError("invalid invoice status", "invoice is already cancelled")
	}

	i.Status = InvoiceStatusCancelled
	i.UpdatedAt = time.Now()

	return nil
}

// AddNotes adds notes to the invoice
func (i *Invoice) AddNotes(notes string) {
	i.Notes = notes
	i.UpdatedAt = time.Now()
}

// IsDraft checks if the invoice is a draft
func (i *Invoice) IsDraft() bool {
	return i.Status == InvoiceStatusDraft
}

// IsGenerated checks if the invoice is generated
func (i *Invoice) IsGenerated() bool {
	return i.Status == InvoiceStatusGenerated
}

// IsSent checks if the invoice is sent
func (i *Invoice) IsSent() bool {
	return i.Status == InvoiceStatusSent
}

// IsPaid checks if the invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// IsCancelled checks if the invoice is cancelled
func (i *Invoice) IsCancelled() bool {
	return i.Status == InvoiceStatusCancelled
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.DueDate == nil || i.IsPaid() || i.IsCancelled() {
		return false
	}
	return time.Now().After(*i.DueDate)
}

// GetItemCount returns the total number of items in the invoice
func (i *Invoice) GetItemCount() int {
	count := 0
	for _, item := range i.Items {
		count += item.Quantity
	}
	return count
}

// GetPaperSizeDimensions returns paper size dimensions in points (1 point = 1/72 inch)
func GetPaperSizeDimensions(size PaperSize) (width, height float64) {
	switch size {
	case PaperSizeA4:
		return 595, 842 // 210mm x 297mm
	case PaperSizeA5:
		return 420, 595 // 148mm x 210mm
	case PaperSizeLetter:
		return 612, 792 // 8.5" x 11"
	case PaperSizeLegal:
		return 612, 1008 // 8.5" x 14"
	case PaperSizeReceipt:
		return 226, 0 // 80mm width, variable height
	default:
		return 595, 842 // Default to A4
	}
}

// convertSaleItemsToInvoiceItems converts sale items to invoice items
func convertSaleItemsToInvoiceItems(saleItems []SaleItem) []InvoiceItem {
	invoiceItems := make([]InvoiceItem, len(saleItems))
	for i, saleItem := range saleItems {
		invoiceItems[i] = InvoiceItem{
			ID:          uuid.New(),
			ProductID:   saleItem.ProductID,
			ProductSKU:  saleItem.ProductSKU,
			ProductName: saleItem.ProductName,
			Quantity:    saleItem.Quantity,
			UnitPrice:   saleItem.UnitPrice,
			TotalPrice:  saleItem.TotalPrice,
		}
	}
	return invoiceItems
}

// ValidateInvoiceStatus validates invoice status
func ValidateInvoiceStatus(status InvoiceStatus) error {
	switch status {
	case InvoiceStatusDraft, InvoiceStatusGenerated, InvoiceStatusSent, InvoiceStatusPaid, InvoiceStatusCancelled:
		return nil
	default:
		return errors.NewValidationError("invalid invoice status", "status must be one of: draft, generated, sent, paid, cancelled")
	}
}

// ValidatePaperSize validates paper size
func ValidatePaperSize(size PaperSize) error {
	switch size {
	case PaperSizeA4, PaperSizeA5, PaperSizeLetter, PaperSizeLegal, PaperSizeReceipt:
		return nil
	default:
		return errors.NewValidationError("invalid paper size", "paper size must be one of: a4, a5, letter, legal, receipt")
	}
}
