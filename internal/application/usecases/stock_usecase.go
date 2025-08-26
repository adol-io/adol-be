package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/application/ports"
	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
	"github.com/nicklaros/adol/pkg/utils"
)

// StockUseCase handles stock management operations
type StockUseCase struct {
	stockRepo         repositories.StockRepository
	stockMovementRepo repositories.StockMovementRepository
	productRepo       repositories.ProductRepository
	database          ports.DatabasePort
	audit             ports.AuditPort
	logger            logger.Logger
}

// NewStockUseCase creates a new stock use case
func NewStockUseCase(
	stockRepo repositories.StockRepository,
	stockMovementRepo repositories.StockMovementRepository,
	productRepo repositories.ProductRepository,
	database ports.DatabasePort,
	audit ports.AuditPort,
	logger logger.Logger,
) *StockUseCase {
	return &StockUseCase{
		stockRepo:         stockRepo,
		stockMovementRepo: stockMovementRepo,
		productRepo:       productRepo,
		database:          database,
		audit:             audit,
		logger:            logger,
	}
}

// StockAdjustmentRequest represents stock adjustment request
type StockAdjustmentRequest struct {
	ProductID uuid.UUID                    `json:"product_id" validate:"required"`
	Type      entities.StockMovementType   `json:"type" validate:"required"`
	Reason    entities.StockMovementReason `json:"reason" validate:"required"`
	Quantity  int                          `json:"quantity" validate:"required,min=1"`
	Reference string                       `json:"reference,omitempty"`
	Notes     string                       `json:"notes,omitempty"`
}

// StockResponse represents stock response
type StockResponse struct {
	ID             uuid.UUID  `json:"id"`
	ProductID      uuid.UUID  `json:"product_id"`
	ProductSKU     string     `json:"product_sku"`
	ProductName    string     `json:"product_name"`
	AvailableQty   int        `json:"available_qty"`
	ReservedQty    int        `json:"reserved_qty"`
	TotalQty       int        `json:"total_qty"`
	ReorderLevel   int        `json:"reorder_level"`
	StockStatus    string     `json:"stock_status"`
	LastMovementAt *time.Time `json:"last_movement_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// StockMovementResponse represents stock movement response
type StockMovementResponse struct {
	ID          uuid.UUID                    `json:"id"`
	ProductID   uuid.UUID                    `json:"product_id"`
	ProductSKU  string                       `json:"product_sku"`
	ProductName string                       `json:"product_name"`
	Type        entities.StockMovementType   `json:"type"`
	Reason      entities.StockMovementReason `json:"reason"`
	Quantity    int                          `json:"quantity"`
	Reference   string                       `json:"reference,omitempty"`
	Notes       string                       `json:"notes,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
	CreatedBy   uuid.UUID                    `json:"created_by"`
}

// StockListResponse represents stock list response
type StockListResponse struct {
	Stocks     []*StockResponse     `json:"stocks"`
	Pagination utils.PaginationInfo `json:"pagination"`
}

// StockMovementListResponse represents stock movement list response
type StockMovementListResponse struct {
	Movements  []*StockMovementResponse `json:"movements"`
	Pagination utils.PaginationInfo     `json:"pagination"`
}

// ReserveStockRequest represents reserve stock request
type ReserveStockRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
	Reference string    `json:"reference" validate:"required"`
	Notes     string    `json:"notes,omitempty"`
}

// AdjustStock adjusts stock levels (add or remove)
func (uc *StockUseCase) AdjustStock(ctx context.Context, userID uuid.UUID, req StockAdjustmentRequest) (*StockResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get product to ensure it exists
	product, err := tx.GetProductRepository().GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Get stock record
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	// Store old quantity for audit
	oldQty := stock.AvailableQty

	// Adjust stock based on type
	switch req.Type {
	case entities.StockMovementTypeIn:
		if err := stock.AddStock(req.Quantity, req.Reason); err != nil {
			return nil, err
		}
	case entities.StockMovementTypeOut:
		if err := stock.RemoveStock(req.Quantity); err != nil {
			return nil, err
		}
	default:
		return nil, errors.NewValidationError("invalid stock movement type", "type must be 'in' or 'out' for adjustments")
	}

	// Create stock movement record
	movement, err := entities.NewStockMovement(
		req.ProductID,
		req.Type,
		req.Reason,
		req.Quantity,
		req.Reference,
		req.Notes,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save stock movement
	if err := tx.GetStockMovementRepository().Create(ctx, movement); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to create stock movement")
		return nil, errors.NewInternalError("failed to create stock movement", err)
	}

	// Update stock record
	if err := tx.GetStockRepository().Update(ctx, stock); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to update stock")
		return nil, errors.NewInternalError("failed to update stock", err)
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
		Action:     "adjust_stock",
		Resource:   "stock",
		ResourceID: stock.ID.String(),
		OldValue: map[string]interface{}{
			"available_qty": oldQty,
		},
		NewValue: map[string]interface{}{
			"available_qty": stock.AvailableQty,
			"type":          req.Type,
			"reason":        req.Reason,
			"quantity":      req.Quantity,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"product_id":  req.ProductID,
		"product_sku": product.SKU,
		"type":        req.Type,
		"quantity":    req.Quantity,
		"user_id":     userID,
	}).Info("Stock adjusted successfully")

	return uc.toStockResponse(stock, product), nil
}

// ReserveStock reserves stock for an order
func (uc *StockUseCase) ReserveStock(ctx context.Context, userID uuid.UUID, req ReserveStockRequest) (*StockResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get product to ensure it exists
	product, err := tx.GetProductRepository().GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Get stock record
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	// Reserve stock
	if err := stock.ReserveStock(req.Quantity); err != nil {
		return nil, err
	}

	// Create stock movement record
	movement, err := entities.NewStockMovement(
		req.ProductID,
		entities.StockMovementTypeReserved,
		entities.ReasonReservation,
		req.Quantity,
		req.Reference,
		req.Notes,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save stock movement
	if err := tx.GetStockMovementRepository().Create(ctx, movement); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to create stock movement")
		return nil, errors.NewInternalError("failed to create stock movement", err)
	}

	// Update stock record
	if err := tx.GetStockRepository().Update(ctx, stock); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to update stock")
		return nil, errors.NewInternalError("failed to update stock", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"reference":  req.Reference,
		"user_id":    userID,
	}).Info("Stock reserved successfully")

	return uc.toStockResponse(stock, product), nil
}

// ReleaseReservedStock releases reserved stock back to available
func (uc *StockUseCase) ReleaseReservedStock(ctx context.Context, userID uuid.UUID, req ReserveStockRequest) (*StockResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get product to ensure it exists
	product, err := tx.GetProductRepository().GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Get stock record
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	// Release reserved stock
	if err := stock.ReleaseReservedStock(req.Quantity); err != nil {
		return nil, err
	}

	// Create stock movement record
	movement, err := entities.NewStockMovement(
		req.ProductID,
		entities.StockMovementTypeReleased,
		entities.ReasonRelease,
		req.Quantity,
		req.Reference,
		req.Notes,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save stock movement
	if err := tx.GetStockMovementRepository().Create(ctx, movement); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to create stock movement")
		return nil, errors.NewInternalError("failed to create stock movement", err)
	}

	// Update stock record
	if err := tx.GetStockRepository().Update(ctx, stock); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to update stock")
		return nil, errors.NewInternalError("failed to update stock", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"reference":  req.Reference,
		"user_id":    userID,
	}).Info("Reserved stock released successfully")

	return uc.toStockResponse(stock, product), nil
}

// ConfirmReservedStock confirms reserved stock (used for sales)
func (uc *StockUseCase) ConfirmReservedStock(ctx context.Context, userID uuid.UUID, req ReserveStockRequest) (*StockResponse, error) {
	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Get product to ensure it exists
	product, err := tx.GetProductRepository().GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Get stock record
	stock, err := tx.GetStockRepository().GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	// Confirm reserved stock
	if err := stock.ConfirmReservedStock(req.Quantity); err != nil {
		return nil, err
	}

	// Create stock movement record for the sale
	movement, err := entities.NewStockMovement(
		req.ProductID,
		entities.StockMovementTypeOut,
		entities.ReasonSale,
		req.Quantity,
		req.Reference,
		req.Notes,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save stock movement
	if err := tx.GetStockMovementRepository().Create(ctx, movement); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to create stock movement")
		return nil, errors.NewInternalError("failed to create stock movement", err)
	}

	// Update stock record
	if err := tx.GetStockRepository().Update(ctx, stock); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": req.ProductID,
			"error":      err.Error(),
		}).Error("Failed to update stock")
		return nil, errors.NewInternalError("failed to update stock", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to commit transaction")
		return nil, errors.NewInternalError("failed to commit transaction", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"reference":  req.Reference,
		"user_id":    userID,
	}).Info("Reserved stock confirmed successfully")

	return uc.toStockResponse(stock, product), nil
}

// GetStock retrieves stock information for a product
func (uc *StockUseCase) GetStock(ctx context.Context, productID uuid.UUID) (*StockResponse, error) {
	stock, err := uc.stockRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("stock record")
	}

	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	return uc.toStockResponse(stock, product), nil
}

// ListStock retrieves stock records with pagination and filtering
func (uc *StockUseCase) ListStock(ctx context.Context, filter repositories.StockFilter, pagination utils.PaginationInfo) (*StockListResponse, error) {
	stocks, paginationResult, err := uc.stockRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list stock")
		return nil, errors.NewInternalError("failed to list stock", err)
	}

	stockResponses := make([]*StockResponse, len(stocks))
	for i, stock := range stocks {
		product, err := uc.productRepo.GetByID(ctx, stock.ProductID)
		if err != nil {
			uc.logger.WithFields(map[string]interface{}{
				"stock_id":   stock.ID,
				"product_id": stock.ProductID,
				"error":      err.Error(),
			}).Warn("Failed to get product for stock record")
			continue
		}
		stockResponses[i] = uc.toStockResponse(stock, product)
	}

	return &StockListResponse{
		Stocks:     stockResponses,
		Pagination: paginationResult,
	}, nil
}

// GetLowStockItems retrieves items with low stock
func (uc *StockUseCase) GetLowStockItems(ctx context.Context, pagination utils.PaginationInfo) (*StockListResponse, error) {
	stocks, paginationResult, err := uc.stockRepo.GetLowStockItems(ctx, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get low stock items")
		return nil, errors.NewInternalError("failed to get low stock items", err)
	}

	stockResponses := make([]*StockResponse, len(stocks))
	for i, stock := range stocks {
		product, err := uc.productRepo.GetByID(ctx, stock.ProductID)
		if err != nil {
			uc.logger.WithFields(map[string]interface{}{
				"stock_id":   stock.ID,
				"product_id": stock.ProductID,
				"error":      err.Error(),
			}).Warn("Failed to get product for stock record")
			continue
		}
		stockResponses[i] = uc.toStockResponse(stock, product)
	}

	return &StockListResponse{
		Stocks:     stockResponses,
		Pagination: paginationResult,
	}, nil
}

// GetStockMovements retrieves stock movements with pagination and filtering
func (uc *StockUseCase) GetStockMovements(ctx context.Context, filter repositories.StockMovementFilter, pagination utils.PaginationInfo) (*StockMovementListResponse, error) {
	movements, paginationResult, err := uc.stockMovementRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list stock movements")
		return nil, errors.NewInternalError("failed to list stock movements", err)
	}

	movementResponses := make([]*StockMovementResponse, len(movements))
	for i, movement := range movements {
		product, err := uc.productRepo.GetByID(ctx, movement.ProductID)
		if err != nil {
			uc.logger.WithFields(map[string]interface{}{
				"movement_id": movement.ID,
				"product_id":  movement.ProductID,
				"error":       err.Error(),
			}).Warn("Failed to get product for stock movement")
			continue
		}
		movementResponses[i] = uc.toStockMovementResponse(movement, product)
	}

	return &StockMovementListResponse{
		Movements:  movementResponses,
		Pagination: paginationResult,
	}, nil
}

// GetProductStockMovements retrieves stock movements for a specific product
func (uc *StockUseCase) GetProductStockMovements(ctx context.Context, productID uuid.UUID, pagination utils.PaginationInfo) (*StockMovementListResponse, error) {
	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	movements, paginationResult, err := uc.stockMovementRepo.GetByProductID(ctx, productID, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get product stock movements")
		return nil, errors.NewInternalError("failed to get product stock movements", err)
	}

	movementResponses := make([]*StockMovementResponse, len(movements))
	for i, movement := range movements {
		movementResponses[i] = uc.toStockMovementResponse(movement, product)
	}

	return &StockMovementListResponse{
		Movements:  movementResponses,
		Pagination: paginationResult,
	}, nil
}

// toStockResponse converts stock entity to response
func (uc *StockUseCase) toStockResponse(stock *entities.Stock, product *entities.Product) *StockResponse {
	return &StockResponse{
		ID:             stock.ID,
		ProductID:      stock.ProductID,
		ProductSKU:     product.SKU,
		ProductName:    product.Name,
		AvailableQty:   stock.AvailableQty,
		ReservedQty:    stock.ReservedQty,
		TotalQty:       stock.TotalQty,
		ReorderLevel:   stock.ReorderLevel,
		StockStatus:    stock.GetStockStatus(),
		LastMovementAt: stock.LastMovementAt,
		CreatedAt:      stock.CreatedAt,
		UpdatedAt:      stock.UpdatedAt,
	}
}

// toStockMovementResponse converts stock movement entity to response
func (uc *StockUseCase) toStockMovementResponse(movement *entities.StockMovement, product *entities.Product) *StockMovementResponse {
	return &StockMovementResponse{
		ID:          movement.ID,
		ProductID:   movement.ProductID,
		ProductSKU:  product.SKU,
		ProductName: product.Name,
		Type:        movement.Type,
		Reason:      movement.Reason,
		Quantity:    movement.Quantity,
		Reference:   movement.Reference,
		Notes:       movement.Notes,
		CreatedAt:   movement.CreatedAt,
		CreatedBy:   movement.CreatedBy,
	}
}
