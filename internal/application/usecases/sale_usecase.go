package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/internal/application/ports"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
	"github.com/nicklaros/adol/pkg/utils"
)

// SaleUseCase handles sales management operations
type SaleUseCase struct {
	saleRepo          repositories.SaleRepository
	saleItemRepo      repositories.SaleItemRepository
	productRepo       repositories.ProductRepository
	stockRepo         repositories.StockRepository
	stockMovementRepo repositories.StockMovementRepository
	database          ports.DatabasePort
	audit             ports.AuditPort
	logger            logger.Logger
}

// NewSaleUseCase creates a new sale use case
func NewSaleUseCase(
	saleRepo repositories.SaleRepository,
	saleItemRepo repositories.SaleItemRepository,
	productRepo repositories.ProductRepository,
	stockRepo repositories.StockRepository,
	stockMovementRepo repositories.StockMovementRepository,
	database ports.DatabasePort,
	audit ports.AuditPort,
	logger logger.Logger,
) *SaleUseCase {
	return &SaleUseCase{
		saleRepo:          saleRepo,
		saleItemRepo:      saleItemRepo,
		productRepo:       productRepo,
		stockRepo:         stockRepo,
		stockMovementRepo: stockMovementRepo,
		database:          database,
		audit:             audit,
		logger:            logger,
	}
}

// CreateSaleRequest represents create sale request
type CreateSaleRequest struct {
	CustomerName  string `json:"customer_name,omitempty"`
	CustomerEmail string `json:"customer_email,omitempty"`
	CustomerPhone string `json:"customer_phone,omitempty"`
}

// AddSaleItemRequest represents add sale item request
type AddSaleItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

// UpdateSaleItemRequest represents update sale item request
type UpdateSaleItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

// CompleteSaleRequest represents complete sale request
type CompleteSaleRequest struct {
	PaidAmount     decimal.Decimal        `json:"paid_amount" validate:"required"`
	PaymentMethod  entities.PaymentMethod `json:"payment_method" validate:"required"`
	DiscountAmount decimal.Decimal        `json:"discount_amount,omitempty"`
	TaxPercentage  decimal.Decimal        `json:"tax_percentage,omitempty"`
	Notes          string                 `json:"notes,omitempty"`
}

// SaleResponse represents sale response
type SaleResponse struct {
	ID             uuid.UUID              `json:"id"`
	SaleNumber     string                 `json:"sale_number"`
	CustomerName   string                 `json:"customer_name,omitempty"`
	CustomerEmail  string                 `json:"customer_email,omitempty"`
	CustomerPhone  string                 `json:"customer_phone,omitempty"`
	Items          []*SaleItemResponse    `json:"items"`
	Subtotal       decimal.Decimal        `json:"subtotal"`
	TaxAmount      decimal.Decimal        `json:"tax_amount"`
	DiscountAmount decimal.Decimal        `json:"discount_amount"`
	TotalAmount    decimal.Decimal        `json:"total_amount"`
	PaidAmount     decimal.Decimal        `json:"paid_amount"`
	ChangeAmount   decimal.Decimal        `json:"change_amount"`
	PaymentMethod  entities.PaymentMethod `json:"payment_method,omitempty"`
	Status         entities.SaleStatus    `json:"status"`
	Notes          string                 `json:"notes,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      uuid.UUID              `json:"created_by"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
}

// SaleItemResponse represents sale item response
type SaleItemResponse struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"product_id"`
	ProductSKU  string          `json:"product_sku"`
	ProductName string          `json:"product_name"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SaleListResponse represents sale list response
type SaleListResponse struct {
	Sales      []*SaleResponse      `json:"sales"`
	Pagination utils.PaginationInfo `json:"pagination"`
}

// CreateSale creates a new sale
func (uc *SaleUseCase) CreateSale(ctx context.Context, userID uuid.UUID, req CreateSaleRequest) (*SaleResponse, error) {
	// Generate sale number
	saleNumber := utils.GenerateSaleNumber()

	// Create sale entity
	sale, err := entities.NewSale(
		saleNumber,
		req.CustomerName,
		req.CustomerEmail,
		req.CustomerPhone,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save sale
	if err := uc.saleRepo.Create(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_number": saleNumber,
			"user_id":     userID,
			"error":       err.Error(),
		}).Error("Failed to create sale")
		return nil, errors.NewInternalError("failed to create sale", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "create",
		Resource:   "sale",
		ResourceID: sale.ID.String(),
		NewValue: map[string]interface{}{
			"sale_number":    sale.SaleNumber,
			"customer_name":  sale.CustomerName,
			"customer_email": sale.CustomerEmail,
			"customer_phone": sale.CustomerPhone,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":     sale.ID,
		"sale_number": saleNumber,
		"user_id":     userID,
	}).Info("Sale created successfully")

	return uc.toSaleResponse(sale), nil
}

// GetSale retrieves a sale by ID
func (uc *SaleUseCase) GetSale(ctx context.Context, saleID uuid.UUID) (*SaleResponse, error) {
	sale, err := uc.saleRepo.GetByID(ctx, saleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	return uc.toSaleResponse(sale), nil
}

// GetSaleBySaleNumber retrieves a sale by sale number
func (uc *SaleUseCase) GetSaleBySaleNumber(ctx context.Context, saleNumber string) (*SaleResponse, error) {
	sale, err := uc.saleRepo.GetBySaleNumber(ctx, saleNumber)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	return uc.toSaleResponse(sale), nil
}

// AddSaleItem adds an item to a sale
func (uc *SaleUseCase) AddSaleItem(ctx context.Context, userID, saleID uuid.UUID, req AddSaleItemRequest) (*SaleResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get sale
	sale, err := tx.GetSaleRepository().GetByID(ctx, saleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	// Check if sale is still pending
	if sale.Status != entities.SaleStatusPending {
		return nil, errors.NewValidationError("invalid sale status", "can only modify pending sales")
	}

	// Get product
	product, err := tx.GetProductRepository().GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Check if product is active
	if !product.IsActive() {
		return nil, errors.NewValidationError("product not active", "cannot add inactive product to sale")
	}

	// Check stock availability
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	if !stock.CanFulfillOrder(req.Quantity) {
		return nil, errors.NewInsufficientStockError(product.Name, stock.AvailableQty, req.Quantity)
	}

	// Create sale item
	saleItem, err := entities.NewSaleItem(
		saleID,
		req.ProductID,
		product.SKU,
		product.Name,
		req.Quantity,
		product.Price,
	)
	if err != nil {
		return nil, err
	}

	// Add item to sale
	if err := sale.AddItem(saleItem); err != nil {
		return nil, err
	}

	// Save sale item
	if err := tx.GetSaleItemRepository().Create(ctx, saleItem); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id":    saleID,
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to create sale item")
		return nil, errors.NewInternalError("failed to create sale item", err)
	}

	// Update sale
	if err := tx.GetSaleRepository().Update(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to update sale")
		return nil, errors.NewInternalError("failed to update sale", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":    saleID,
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"user_id":    userID,
	}).Info("Sale item added successfully")

	return uc.toSaleResponse(sale), nil
}

// UpdateSaleItem updates the quantity of a sale item
func (uc *SaleUseCase) UpdateSaleItem(ctx context.Context, userID, saleID uuid.UUID, req UpdateSaleItemRequest) (*SaleResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get sale
	sale, err := tx.GetSaleRepository().GetByID(ctx, saleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	// Check if sale is still pending
	if sale.Status != entities.SaleStatusPending {
		return nil, errors.NewValidationError("invalid sale status", "can only modify pending sales")
	}

	// Check stock availability
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	if !stock.CanFulfillOrder(req.Quantity) {
		product, _ := tx.GetProductRepository().GetByID(ctx, req.ProductID)
		productName := req.ProductID.String()
		if product != nil {
			productName = product.Name
		}
		return nil, errors.NewInsufficientStockError(productName, stock.AvailableQty, req.Quantity)
	}

	// Update sale item quantity
	if err := sale.UpdateItemQuantity(req.ProductID, req.Quantity); err != nil {
		return nil, err
	}

	// Update sale items in database
	if err := tx.GetSaleItemRepository().BulkUpdate(ctx, convertSaleItemsToEntities(sale.Items)); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to update sale items")
		return nil, errors.NewInternalError("failed to update sale items", err)
	}

	// Update sale
	if err := tx.GetSaleRepository().Update(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to update sale")
		return nil, errors.NewInternalError("failed to update sale", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":    saleID,
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"user_id":    userID,
	}).Info("Sale item updated successfully")

	return uc.toSaleResponse(sale), nil
}

// RemoveSaleItem removes an item from a sale
func (uc *SaleUseCase) RemoveSaleItem(ctx context.Context, userID, saleID, productID uuid.UUID) (*SaleResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get sale
	sale, err := tx.GetSaleRepository().GetByID(ctx, saleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	// Check if sale is still pending
	if sale.Status != entities.SaleStatusPending {
		return nil, errors.NewValidationError("invalid sale status", "can only modify pending sales")
	}

	// Remove item from sale
	if err := sale.RemoveItem(productID); err != nil {
		return nil, err
	}

	// Delete sale item from database
	saleItems, err := tx.GetSaleItemRepository().GetBySaleID(ctx, saleID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get sale items", err)
	}

	for _, item := range saleItems {
		if item.ProductID == productID {
			if err := tx.GetSaleItemRepository().Delete(ctx, item.ID); err != nil {
				uc.logger.WithFields(map[string]interface{}{
					"sale_item_id": item.ID,
					"error":        err.Error(),
				}).Error("Failed to delete sale item")
				return nil, errors.NewInternalError("failed to delete sale item", err)
			}
			break
		}
	}

	// Update sale
	if err := tx.GetSaleRepository().Update(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to update sale")
		return nil, errors.NewInternalError("failed to update sale", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":    saleID,
		"product_id": productID,
		"user_id":    userID,
	}).Info("Sale item removed successfully")

	return uc.toSaleResponse(sale), nil
}

// CompleteSale completes a sale with payment
func (uc *SaleUseCase) CompleteSale(ctx context.Context, userID, saleID uuid.UUID, req CompleteSaleRequest) (*SaleResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get sale with items
	sale, err := tx.GetSaleRepository().GetByID(ctx, saleID)
	if err != nil {
		return nil, errors.NewNotFoundError("sale")
	}

	// Apply discount if provided
	if req.DiscountAmount.GreaterThan(decimal.Zero) {
		if err := sale.ApplyDiscount(req.DiscountAmount); err != nil {
			return nil, err
		}
	}

	// Apply tax if provided
	if req.TaxPercentage.GreaterThan(decimal.Zero) {
		if err := sale.ApplyTax(req.TaxPercentage); err != nil {
			return nil, err
		}
	}

	// Process payment
	if err := sale.ProcessPayment(req.PaidAmount, req.PaymentMethod); err != nil {
		return nil, err
	}

	// Add notes if provided
	if req.Notes != "" {
		sale.AddNotes(req.Notes)
	}

	// Complete the sale
	if err := sale.CompleteSale(); err != nil {
		return nil, err
	}

	// Update stock for each item
	for _, item := range sale.Items {
		stock, err := tx.GetStockRepository().GetByProductID(ctx, item.ProductID)
		if err != nil {
			return nil, errors.NewNotFoundError("stock record")
		}

		// Remove stock
		if err := stock.RemoveStock(item.Quantity); err != nil {
			return nil, err
		}

		// Create stock movement
		movement, err := entities.NewStockMovement(
			item.ProductID,
			entities.StockMovementTypeOut,
			entities.ReasonSale,
			item.Quantity,
			sale.SaleNumber,
			"Sale completion",
			userID,
		)
		if err != nil {
			return nil, err
		}

		// Save stock movement
		if err := tx.GetStockMovementRepository().Create(ctx, movement); err != nil {
			uc.logger.WithFields(map[string]interface{}{
				"product_id": item.ProductID,
				"error":      err.Error(),
			}).Error("Failed to create stock movement")
			return nil, errors.NewInternalError("failed to create stock movement", err)
		}

		// Update stock
		if err := tx.GetStockRepository().Update(ctx, stock); err != nil {
			uc.logger.WithFields(map[string]interface{}{
				"product_id": item.ProductID,
				"error":      err.Error(),
			}).Error("Failed to update stock")
			return nil, errors.NewInternalError("failed to update stock", err)
		}
	}

	// Update sale
	if err := tx.GetSaleRepository().Update(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to update sale")
		return nil, errors.NewInternalError("failed to update sale", err)
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
		Action:     "complete",
		Resource:   "sale",
		ResourceID: saleID.String(),
		NewValue: map[string]interface{}{
			"total_amount":   sale.TotalAmount,
			"paid_amount":    sale.PaidAmount,
			"payment_method": sale.PaymentMethod,
			"status":         sale.Status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":      saleID,
		"sale_number":  sale.SaleNumber,
		"total_amount": sale.TotalAmount,
		"user_id":      userID,
	}).Info("Sale completed successfully")

	return uc.toSaleResponse(sale), nil
}

// CancelSale cancels a sale
func (uc *SaleUseCase) CancelSale(ctx context.Context, userID, saleID uuid.UUID) error {
	// Get sale
	sale, err := uc.saleRepo.GetByID(ctx, saleID)
	if err != nil {
		return errors.NewNotFoundError("sale")
	}

	// Cancel sale
	if err := sale.CancelSale(); err != nil {
		return err
	}

	// Update sale
	if err := uc.saleRepo.Update(ctx, sale); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sale_id": saleID,
			"error":   err.Error(),
		}).Error("Failed to cancel sale")
		return errors.NewInternalError("failed to cancel sale", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "cancel",
		Resource:   "sale",
		ResourceID: saleID.String(),
		NewValue: map[string]interface{}{
			"status": sale.Status,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"sale_id":     saleID,
		"sale_number": sale.SaleNumber,
		"user_id":     userID,
	}).Info("Sale cancelled successfully")

	return nil
}

// ListSales retrieves sales with pagination and filtering
func (uc *SaleUseCase) ListSales(ctx context.Context, filter repositories.SaleFilter, pagination utils.PaginationInfo) (*SaleListResponse, error) {
	sales, paginationResult, err := uc.saleRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list sales")
		return nil, errors.NewInternalError("failed to list sales", err)
	}

	saleResponses := make([]*SaleResponse, len(sales))
	for i, sale := range sales {
		saleResponses[i] = uc.toSaleResponse(sale)
	}

	return &SaleListResponse{
		Sales:      saleResponses,
		Pagination: paginationResult,
	}, nil
}

// toSaleResponse converts sale entity to response
func (uc *SaleUseCase) toSaleResponse(sale *entities.Sale) *SaleResponse {
	items := make([]*SaleItemResponse, len(sale.Items))
	for i, item := range sale.Items {
		items[i] = &SaleItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			CreatedAt:   item.CreatedAt,
		}
	}

	return &SaleResponse{
		ID:             sale.ID,
		SaleNumber:     sale.SaleNumber,
		CustomerName:   sale.CustomerName,
		CustomerEmail:  sale.CustomerEmail,
		CustomerPhone:  sale.CustomerPhone,
		Items:          items,
		Subtotal:       sale.Subtotal,
		TaxAmount:      sale.TaxAmount,
		DiscountAmount: sale.DiscountAmount,
		TotalAmount:    sale.TotalAmount,
		PaidAmount:     sale.PaidAmount,
		ChangeAmount:   sale.ChangeAmount,
		PaymentMethod:  sale.PaymentMethod,
		Status:         sale.Status,
		Notes:          sale.Notes,
		CreatedAt:      sale.CreatedAt,
		UpdatedAt:      sale.UpdatedAt,
		CreatedBy:      sale.CreatedBy,
		CompletedAt:    sale.CompletedAt,
	}
}

// convertSaleItemsToEntities converts sale items to entities
func convertSaleItemsToEntities(items []entities.SaleItem) []*entities.SaleItem {
	entities := make([]*entities.SaleItem, len(items))
	for i := range items {
		entities[i] = &items[i]
	}
	return entities
}
