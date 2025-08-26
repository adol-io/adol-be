package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/internal/application/ports"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/internal/domain/services"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
	"github.com/nicklaros/adol/pkg/utils"
)

// InvoiceUseCase handles invoice management operations
type InvoiceUseCase struct {
	invoiceRepo     repositories.InvoiceRepository
	invoiceItemRepo repositories.InvoiceItemRepository
	saleRepo        repositories.SaleRepository
	pdfService      services.InvoicePDFService
	emailService    services.EmailService
	printService    services.PrintService
	database        ports.DatabasePort
	audit           ports.AuditPort
	logger          logger.Logger
}

// NewInvoiceUseCase creates a new invoice use case
func NewInvoiceUseCase(
	invoiceRepo repositories.InvoiceRepository,
	invoiceItemRepo repositories.InvoiceItemRepository,
	saleRepo repositories.SaleRepository,
	pdfService services.InvoicePDFService,
	emailService services.EmailService,
	printService services.PrintService,
	database ports.DatabasePort,
	audit ports.AuditPort,
	logger logger.Logger,
) *InvoiceUseCase {
	return &InvoiceUseCase{
		invoiceRepo:     invoiceRepo,
		invoiceItemRepo: invoiceItemRepo,
		saleRepo:        saleRepo,
		pdfService:      pdfService,
		emailService:    emailService,
		printService:    printService,
		database:        database,
		audit:           audit,
		logger:          logger,
	}
}

// CreateInvoiceRequest represents create invoice request
type CreateInvoiceRequest struct {
	SaleID          uuid.UUID `json:"sale_id" validate:"required"`
	CustomerAddress string     `json:"customer_address,omitempty"`
	DueDate         *time.Time  `json:"due_date,omitempty"`
	Notes           string    `json:"notes,omitempty"`
} 

// GenerateInvoicePDFRequest represents generate invoice PDF request
type GenerateInvoicePDFRequest struct {
	InvoiceID uuid.UUID             `json:"invoice_id" validate:"required"`
	PaperSize entities.PaperSize    `json:"paper_size,omitempty"`
	Template  *entities.InvoiceTemplate `json:"template,omitempty"`
}

// SendInvoiceEmailRequest represents send invoice email request
type SendInvoiceEmailRequest struct {
	InvoiceID   uuid.UUID             `json:"invoice_id" validate:"required"`
	EmailTo     string                `json:"email_to" validate:"required,email"`
	Subject     string                `json:"subject,omitempty"`
	Message     string                `json:"message,omitempty"`
	PaperSize   entities.PaperSize    `json:"paper_size,omitempty"`
	Template    *entities.InvoiceTemplate `json:"template,omitempty"`
}

// PrintInvoiceRequest represents print invoice request
type PrintInvoiceRequest struct {
	InvoiceID   uuid.UUID             `json:"invoice_id" validate:"required"`
	PrinterName string                `json:"printer_name,omitempty"`
	PaperSize   entities.PaperSize    `json:"paper_size,omitempty"`
	Template    *entities.InvoiceTemplate `json:"template,omitempty"`
}

// InvoiceResponse represents invoice response
type InvoiceResponse struct {
	ID              uuid.UUID                 `json:"id"`
	InvoiceNumber   string                    `json:"invoice_number"`
	SaleID          uuid.UUID                 `json:"sale_id"`
	CustomerName    string                    `json:"customer_name"`
	CustomerEmail   string                    `json:"customer_email,omitempty"`
	CustomerPhone   string                    `json:"customer_phone,omitempty"`
	CustomerAddress string                    `json:"customer_address,omitempty"`
	Items           []*InvoiceItemResponse    `json:"items"`
	Subtotal        decimal.Decimal           `json:"subtotal"`
	TaxAmount       decimal.Decimal           `json:"tax_amount"`
	DiscountAmount  decimal.Decimal           `json:"discount_amount"`
	TotalAmount     decimal.Decimal           `json:"total_amount"`
	PaidAmount      decimal.Decimal           `json:"paid_amount"`
	PaymentMethod   entities.PaymentMethod    `json:"payment_method"`
	Status          entities.InvoiceStatus    `json:"status"`
	Notes           string                    `json:"notes,omitempty"`
	DueDate         *time.Time                `json:"due_date,omitempty"`
	PaidAt          *time.Time                `json:"paid_at,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	CreatedBy       uuid.UUID                 `json:"created_by"`
}

// InvoiceItemResponse represents invoice item response
type InvoiceItemResponse struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"product_id"`
	ProductSKU  string          `json:"product_sku"`
	ProductName string          `json:"product_name"`
	Description string          `json:"description,omitempty"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price"`
}

// InvoiceListResponse represents invoice list response
type InvoiceListResponse struct {
	Invoices   []*InvoiceResponse   `json:"invoices"`
	Pagination utils.PaginationInfo `json:"pagination"`
}

// CreateInvoice creates an invoice from a completed sale
func (uc *InvoiceUseCase) CreateInvoice(ctx context.Context, userID uuid.UUID, req CreateInvoiceRequest) (*InvoiceResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get sale
	sale, err := tx.GetSaleRepository().GetByID(ctx, req.SaleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	// Check if sale is completed
	if !sale.IsCompleted() {
		return nil, errors.NewValidationError("invalid sale status", "can only create invoice for completed sales")
	}

	// Check if invoice already exists for this sale
	if existingInvoice, err := tx.GetInvoiceRepository().GetBySaleID(ctx, req.SaleID); err == nil {
		return uc.toInvoiceResponse(existingInvoice), nil
	}

	// Generate invoice number
	invoiceNumber := utils.GenerateInvoiceNumber()

	// Create invoice entity
	invoice, err := entities.NewInvoice(invoiceNumber, sale, userID)
	if err != nil {
		return nil, err
	}

	// Update customer address if provided
	if req.CustomerAddress != "" {
		if err := invoice.UpdateCustomerInfo(
			invoice.CustomerName,
			invoice.CustomerEmail,
			invoice.CustomerPhone,
			req.CustomerAddress,
		); err != nil {
			return nil, err
		}
	}

	// Set due date if provided
	if req.DueDate != nil {
		if err := invoice.SetDueDate(*req.DueDate); err != nil {
			return nil, err
		}
	}

	// Add notes if provided
	if req.Notes != "" {
		invoice.AddNotes(req.Notes)
	}

	// Save invoice
	if err := tx.GetInvoiceRepository().Create(ctx, invoice); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_number": invoiceNumber,
			"sale_id":        req.SaleID,
			"error":          err.Error(),
		}).Error("Failed to create invoice")
		return nil, errors.NewInternalError("failed to create invoice", err)
	}

	// Save invoice items
	invoiceItems := make([]*entities.InvoiceItem, len(invoice.Items))
	for i := range invoice.Items {
		invoice.Items[i].InvoiceID = invoice.ID
		invoiceItems[i] = &invoice.Items[i]
	}

	if err := tx.GetInvoiceItemRepository().BulkCreate(ctx, invoiceItems); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"error":      err.Error(),
		}).Error("Failed to create invoice items")
		return nil, errors.NewInternalError("failed to create invoice items", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "create",
		Resource:   "invoice",
		ResourceID: invoice.ID.String(),
		NewValue: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
			"sale_id":        invoice.SaleID,
			"total_amount":   invoice.TotalAmount,
			"customer_name":  invoice.CustomerName,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoiceNumber,
		"sale_id":        req.SaleID,
		"user_id":        userID,
	}).Info("Invoice created successfully")

	return uc.toInvoiceResponse(invoice), nil
}

// GetInvoice retrieves an invoice by ID
func (uc *InvoiceUseCase) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*InvoiceResponse, error) {
	invoice, err := uc.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, errors.NewNotFoundError("invoice")
	}

	return uc.toInvoiceResponse(invoice), nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (uc *InvoiceUseCase) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*InvoiceResponse, error) {
	invoice, err := uc.invoiceRepo.GetByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		return nil, errors.NewNotFoundError("invoice")
	}

	return uc.toInvoiceResponse(invoice), nil
}

// GenerateInvoicePDF generates a PDF for an invoice
func (uc *InvoiceUseCase) GenerateInvoicePDF(ctx context.Context, req GenerateInvoicePDFRequest) ([]byte, error) {
	// Get invoice
	invoice, err := uc.invoiceRepo.GetByID(ctx, req.InvoiceID)
	if err != nil {
		return nil, errors.NewNotFoundError("invoice")
	}

	// Use provided template or get default
	template := req.Template
	if template == nil {
		paperSize := req.PaperSize
		if paperSize == "" {
			paperSize = entities.PaperSizeA4
		}
		template = uc.pdfService.GetDefaultTemplate(paperSize)
	}

	// Generate PDF
	pdfData, err := uc.pdfService.GenerateInvoicePDF(ctx, invoice, template)
	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": req.InvoiceID,
			"error":      err.Error(),
		}).Error("Failed to generate invoice PDF")
		return nil, errors.NewInternalError("failed to generate PDF", err)
	}

	// Mark invoice as generated if it's still a draft
	if invoice.IsDraft() {
		if err := invoice.MarkAsGenerated(); err == nil {
			uc.invoiceRepo.Update(ctx, invoice)
		}
	}

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     req.InvoiceID,
		"invoice_number": invoice.InvoiceNumber,
		"paper_size":     template.PaperSize,
	}).Info("Invoice PDF generated successfully")

	return pdfData, nil
}

// SendInvoiceEmail sends an invoice via email
func (uc *InvoiceUseCase) SendInvoiceEmail(ctx context.Context, userID uuid.UUID, req SendInvoiceEmailRequest) error {
	// Get invoice
	invoice, err := uc.invoiceRepo.GetByID(ctx, req.InvoiceID)
	if err != nil {
		return errors.NewNotFoundError("invoice")
	}

	// Use provided template or get default
	template := req.Template
	if template == nil {
		paperSize := req.PaperSize
		if paperSize == "" {
			paperSize = entities.PaperSizeA4
		}
		template = uc.pdfService.GetDefaultTemplate(paperSize)
	}

	// Generate PDF
	pdfData, err := uc.pdfService.GenerateInvoicePDF(ctx, invoice, template)
	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": req.InvoiceID,
			"error":      err.Error(),
		}).Error("Failed to generate invoice PDF for email")
		return errors.NewInternalError("failed to generate PDF", err)
	}

	// Send email
	if err := uc.emailService.SendInvoiceEmail(ctx, invoice, req.EmailTo, pdfData); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": req.InvoiceID,
			"email_to":   req.EmailTo,
			"error":      err.Error(),
		}).Error("Failed to send invoice email")
		return errors.NewInternalError("failed to send email", err)
	}

	// Mark invoice as sent
	if invoice.IsGenerated() {
		if err := invoice.MarkAsSent(); err == nil {
			uc.invoiceRepo.Update(ctx, invoice)
		}
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "send_email",
		Resource:   "invoice",
		ResourceID: invoice.ID.String(),
		NewValue: map[string]interface{}{
			"email_to": req.EmailTo,
			"status":   "sent",
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     req.InvoiceID,
		"invoice_number": invoice.InvoiceNumber,
		"email_to":       req.EmailTo,
		"user_id":        userID,
	}).Info("Invoice email sent successfully")

	return nil
}

// PrintInvoice prints an invoice
func (uc *InvoiceUseCase) PrintInvoice(ctx context.Context, userID uuid.UUID, req PrintInvoiceRequest) error {
	// Get invoice
	invoice, err := uc.invoiceRepo.GetByID(ctx, req.InvoiceID)
	if err != nil {
		return errors.NewNotFoundError("invoice")
	}

	// Use provided template or get default
	template := req.Template
	if template == nil {
		paperSize := req.PaperSize
		if paperSize == "" {
			paperSize = entities.PaperSizeA4
		}
		template = uc.pdfService.GetDefaultTemplate(paperSize)
	}

	// Print invoice
	if template.PaperSize == entities.PaperSizeReceipt {
		err = uc.printService.PrintReceipt(ctx, invoice, template, req.PrinterName)
	} else {
		err = uc.printService.PrintInvoice(ctx, invoice, template, req.PrinterName)
	}

	if err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id":   req.InvoiceID,
			"printer_name": req.PrinterName,
			"error":        err.Error(),
		}).Error("Failed to print invoice")
		return errors.NewInternalError("failed to print invoice", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "print",
		Resource:   "invoice",
		ResourceID: invoice.ID.String(),
		NewValue: map[string]interface{}{
			"printer_name": req.PrinterName,
			"paper_size":   template.PaperSize,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     req.InvoiceID,
		"invoice_number": invoice.InvoiceNumber,
		"printer_name":   req.PrinterName,
		"user_id":        userID,
	}).Info("Invoice printed successfully")

	return nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (uc *InvoiceUseCase) MarkInvoiceAsPaid(ctx context.Context, userID, invoiceID uuid.UUID) error {
	// Get invoice
	invoice, err := uc.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return errors.NewNotFoundError("invoice")
	}

	// Mark as paid
	if err := invoice.MarkAsPaid(); err != nil {
		return err
	}

	// Update invoice
	if err := uc.invoiceRepo.Update(ctx, invoice); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": invoiceID,
			"error":      err.Error(),
		}).Error("Failed to update invoice")
		return errors.NewInternalError("failed to update invoice", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "mark_paid",
		Resource:   "invoice",
		ResourceID: invoiceID.String(),
		NewValue: map[string]interface{}{
			"status":  invoice.Status,
			"paid_at": invoice.PaidAt,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoiceID,
		"invoice_number": invoice.InvoiceNumber,
		"user_id":        userID,
	}).Info("Invoice marked as paid successfully")

	return nil
}

// CancelInvoice cancels an invoice
func (uc *InvoiceUseCase) CancelInvoice(ctx context.Context, userID, invoiceID uuid.UUID) error {
	// Get invoice
	invoice, err := uc.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return errors.NewNotFoundError("invoice")
	}

	// Cancel invoice
	if err := invoice.Cancel(); err != nil {
		return err
	}

	// Update invoice
	if err := uc.invoiceRepo.Update(ctx, invoice); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"invoice_id": invoiceID,
			"error":      err.Error(),
		}).Error("Failed to cancel invoice")
		return errors.NewInternalError("failed to cancel invoice", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "cancel",
		Resource:   "invoice",
		ResourceID: invoiceID.String(),
		NewValue: map[string]interface{}{
			"status": invoice.Status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoiceID,
		"invoice_number": invoice.InvoiceNumber,
		"user_id":        userID,
	}).Info("Invoice cancelled successfully")

	return nil
}

// ListInvoices retrieves invoices with pagination and filtering
func (uc *InvoiceUseCase) ListInvoices(ctx context.Context, filter repositories.InvoiceFilter, pagination utils.PaginationInfo) (*InvoiceListResponse, error) {
	invoices, paginationResult, err := uc.invoiceRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list invoices")
		return nil, errors.NewInternalError("failed to list invoices", err)
	}

	invoiceResponses := make([]*InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = uc.toInvoiceResponse(invoice)
	}

	return &InvoiceListResponse{
		Invoices:   invoiceResponses,
		Pagination: paginationResult,
	}, nil
}

// GetOverdueInvoices retrieves overdue invoices
func (uc *InvoiceUseCase) GetOverdueInvoices(ctx context.Context, pagination utils.PaginationInfo) (*InvoiceListResponse, error) {
	invoices, paginationResult, err := uc.invoiceRepo.GetOverdueInvoices(ctx, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get overdue invoices")
		return nil, errors.NewInternalError("failed to get overdue invoices", err)
	}

	invoiceResponses := make([]*InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = uc.toInvoiceResponse(invoice)
	}

	return &InvoiceListResponse{
		Invoices:   invoiceResponses,
		Pagination: paginationResult,
	}, nil
}

// toInvoiceResponse converts invoice entity to response
func (uc *InvoiceUseCase) toInvoiceResponse(invoice *entities.Invoice) *InvoiceResponse {
	items := make([]*InvoiceItemResponse, len(invoice.Items))
	for i, item := range invoice.Items {
		items[i] = &InvoiceItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}

	return &InvoiceResponse{
		ID:              invoice.ID,
		InvoiceNumber:   invoice.InvoiceNumber,
		SaleID:          invoice.SaleID,
		CustomerName:    invoice.CustomerName,
		CustomerEmail:   invoice.CustomerEmail,
		CustomerPhone:   invoice.CustomerPhone,
		CustomerAddress: invoice.CustomerAddress,
		Items:           items,
		Subtotal:        invoice.Subtotal,
		TaxAmount:       invoice.TaxAmount,
		DiscountAmount:  invoice.DiscountAmount,
		TotalAmount:     invoice.TotalAmount,
		PaidAmount:      invoice.PaidAmount,
		PaymentMethod:   invoice.PaymentMethod,
		Status:          invoice.Status,
		Notes:           invoice.Notes,
		DueDate:         invoice.DueDate,
		PaidAt:          invoice.PaidAt,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		CreatedBy:       invoice.CreatedBy,
	}
}
