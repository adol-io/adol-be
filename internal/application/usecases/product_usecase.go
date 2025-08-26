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

// ProductUseCase handles product management operations
type ProductUseCase struct {
	productRepo repositories.ProductRepository
	stockRepo   repositories.StockRepository
	database    ports.DatabasePort
	audit       ports.AuditPort
	logger      logger.Logger
}

// NewProductUseCase creates a new product use case
func NewProductUseCase(
	productRepo repositories.ProductRepository,
	stockRepo repositories.StockRepository,
	database ports.DatabasePort,
	audit ports.AuditPort,
	logger logger.Logger,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
		stockRepo:   stockRepo,
		database:    database,
		audit:       audit,
		logger:      logger,
	}
}

// CreateProductRequest represents create product request
type CreateProductRequest struct {
	SKU          string          `json:"sku" validate:"required,min=3"`
	Name         string          `json:"name" validate:"required"`
	Description  string          `json:"description"`
	Category     string          `json:"category" validate:"required"`
	Price        decimal.Decimal `json:"price" validate:"required"`
	Cost         decimal.Decimal `json:"cost" validate:"required"`
	Unit         string          `json:"unit" validate:"required"`
	MinStock     int             `json:"min_stock" validate:"min=0"`
	InitialStock int             `json:"initial_stock" validate:"min=0"`
}

// UpdateProductRequest represents update product request
type UpdateProductRequest struct {
	Name        string                  `json:"name,omitempty"`
	Description string                  `json:"description,omitempty"`
	Category    string                  `json:"category,omitempty"`
	Price       *decimal.Decimal        `json:"price,omitempty"`
	Cost        *decimal.Decimal        `json:"cost,omitempty"`
	Unit        string                  `json:"unit,omitempty"`
	MinStock    *int                    `json:"min_stock,omitempty"`
	Status      *entities.ProductStatus `json:"status,omitempty"`
}

// ProductResponse represents product response
type ProductResponse struct {
	ID             uuid.UUID              `json:"id"`
	SKU            string                 `json:"sku"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Price          decimal.Decimal        `json:"price"`
	Cost           decimal.Decimal        `json:"cost"`
	Status         entities.ProductStatus `json:"status"`
	Unit           string                 `json:"unit"`
	MinStock       int                    `json:"min_stock"`
	AvailableStock int                    `json:"available_stock,omitempty"`
	ReservedStock  int                    `json:"reserved_stock,omitempty"`
	TotalStock     int                    `json:"total_stock,omitempty"`
	ProfitMargin   decimal.Decimal        `json:"profit_margin"`
	ProfitAmount   decimal.Decimal        `json:"profit_amount"`
	StockStatus    string                 `json:"stock_status,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      uuid.UUID              `json:"created_by"`
}

// ProductListResponse represents product list response
type ProductListResponse struct {
	Products   []*ProductResponse   `json:"products"`
	Pagination utils.PaginationInfo `json:"pagination"`
}

// CreateProduct creates a new product with initial stock
func (uc *ProductUseCase) CreateProduct(ctx context.Context, userID uuid.UUID, req CreateProductRequest) (*ProductResponse, error) {
	// Validate SKU format
	if !utils.IsValidSKU(req.SKU) {
		return nil, errors.NewValidationError("invalid SKU format", "SKU must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Start transaction
	tx, err := uc.database.BeginTransaction(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to begin transaction")
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// Check if SKU already exists
	exists, err := tx.GetProductRepository().ExistsBySKU(ctx, req.SKU)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to check SKU existence")
		return nil, errors.NewInternalError("failed to check SKU", err)
	}
	if exists {
		return nil, errors.NewConflictError("SKU already exists")
	}

	// Create product entity
	product, err := entities.NewProduct(
		req.SKU,
		req.Name,
		req.Description,
		req.Category,
		req.Unit,
		req.Price,
		req.Cost,
		req.MinStock,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Save product
	if err := tx.GetProductRepository().Create(ctx, product); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"sku":   req.SKU,
			"name":  req.Name,
			"error": err.Error(),
		}).Error("Failed to create product")
		return nil, errors.NewInternalError("failed to create product", err)
	}

	// Create initial stock record
	stock, err := entities.NewStock(product.ID, req.InitialStock, req.MinStock)
	if err != nil {
		return nil, err
	}

	if err := tx.GetStockRepository().Create(ctx, stock); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": product.ID,
			"error":      err.Error(),
		}).Error("Failed to create initial stock")
		return nil, errors.NewInternalError("failed to create initial stock", err)
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
		Resource:   "product",
		ResourceID: product.ID.String(),
		NewValue: map[string]interface{}{
			"sku":           product.SKU,
			"name":          product.Name,
			"category":      product.Category,
			"price":         product.Price,
			"cost":          product.Cost,
			"initial_stock": req.InitialStock,
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"product_id": product.ID,
		"sku":        product.SKU,
		"user_id":    userID,
	}).Info("Product created successfully")

	// Return response with stock information
	response := uc.toProductResponse(product)
	response.AvailableStock = stock.AvailableQty
	response.ReservedStock = stock.ReservedQty
	response.TotalStock = stock.TotalQty
	response.StockStatus = stock.GetStockStatus()

	return response, nil
}

// GetProduct retrieves a product by ID
func (uc *ProductUseCase) GetProduct(ctx context.Context, productID uuid.UUID) (*ProductResponse, error) {
	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	response := uc.toProductResponse(product)

	// Get stock information
	if stock, err := uc.stockRepo.GetByProductID(ctx, productID); err == nil {
		response.AvailableStock = stock.AvailableQty
		response.ReservedStock = stock.ReservedQty
		response.TotalStock = stock.TotalQty
		response.StockStatus = stock.GetStockStatus()
	}

	return response, nil
}

// GetProductBySKU retrieves a product by SKU
func (uc *ProductUseCase) GetProductBySKU(ctx context.Context, sku string) (*ProductResponse, error) {
	product, err := uc.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	response := uc.toProductResponse(product)

	// Get stock information
	if stock, err := uc.stockRepo.GetByProductID(ctx, product.ID); err == nil {
		response.AvailableStock = stock.AvailableQty
		response.ReservedStock = stock.ReservedQty
		response.TotalStock = stock.TotalQty
		response.StockStatus = stock.GetStockStatus()
	}

	return response, nil
}

// UpdateProduct updates an existing product
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, userID, productID uuid.UUID, req UpdateProductRequest) (*ProductResponse, error) {
	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("product")
	}

	// Store old values for audit log
	oldValue := map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"price":       product.Price,
		"cost":        product.Cost,
		"unit":        product.Unit,
		"min_stock":   product.MinStock,
		"status":      product.Status,
	}

	// Update product fields
	if req.Name != "" || req.Description != "" || req.Category != "" || req.Unit != "" || req.Price != nil || req.Cost != nil || req.MinStock != nil {
		name := product.Name
		description := product.Description
		category := product.Category
		unit := product.Unit
		price := product.Price
		cost := product.Cost
		minStock := product.MinStock

		if req.Name != "" {
			name = req.Name
		}
		if req.Description != "" {
			description = req.Description
		}
		if req.Category != "" {
			category = req.Category
		}
		if req.Unit != "" {
			unit = req.Unit
		}
		if req.Price != nil {
			price = *req.Price
		}
		if req.Cost != nil {
			cost = *req.Cost
		}
		if req.MinStock != nil {
			minStock = *req.MinStock
		}

		if err := product.UpdateProduct(name, description, category, unit, price, cost, minStock); err != nil {
			return nil, err
		}
	}

	// Update status if provided
	if req.Status != nil {
		if err := product.ChangeStatus(*req.Status); err != nil {
			return nil, err
		}
	}

	// Save product
	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Failed to update product")
		return nil, errors.NewInternalError("failed to update product", err)
	}

	// New values for audit log
	newValue := map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"price":       product.Price,
		"cost":        product.Cost,
		"unit":        product.Unit,
		"min_stock":   product.MinStock,
		"status":      product.Status,
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "update",
		Resource:   "product",
		ResourceID: productID.String(),
		OldValue:   oldValue,
		NewValue:   newValue,
		Timestamp:  time.Now(),
		Success:    true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"product_id": productID,
		"user_id":    userID,
	}).Info("Product updated successfully")

	response := uc.toProductResponse(product)

	// Get stock information
	if stock, err := uc.stockRepo.GetByProductID(ctx, productID); err == nil {
		response.AvailableStock = stock.AvailableQty
		response.ReservedStock = stock.ReservedQty
		response.TotalStock = stock.TotalQty
		response.StockStatus = stock.GetStockStatus()
	}

	return response, nil
}

// DeleteProduct deletes a product (soft delete)
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, userID, productID uuid.UUID) error {
	// Get product to ensure it exists
	product, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return errors.NewNotFoundError("product")
	}

	// Check if product has stock movements or sales
	// This would require additional repository methods to check dependencies
	// For now, we'll just deactivate the product instead of hard delete

	// Deactivate product instead of deleting
	if err := product.ChangeStatus(entities.ProductStatusDiscontinued); err != nil {
		return err
	}

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.WithFields(map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Failed to discontinue product")
		return errors.NewInternalError("failed to discontinue product", err)
	}

	// Audit log
	auditEvent := ports.AuditEvent{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "delete",
		Resource:   "product",
		ResourceID: productID.String(),
		OldValue: map[string]interface{}{
			"sku":      product.SKU,
			"name":     product.Name,
			"category": product.Category,
			"status":   "discontinued",
		},
		Timestamp: time.Now(),
		Success:   true,
	}
	uc.audit.Log(ctx, auditEvent)

	uc.logger.WithFields(map[string]interface{}{
		"product_id": productID,
		"user_id":    userID,
	}).Info("Product discontinued successfully")

	return nil
}

// ListProducts retrieves products with pagination and filtering
func (uc *ProductUseCase) ListProducts(ctx context.Context, filter repositories.ProductFilter, pagination utils.PaginationInfo) (*ProductListResponse, error) {
	products, paginationResult, err := uc.productRepo.List(ctx, filter, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to list products")
		return nil, errors.NewInternalError("failed to list products", err)
	}

	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		response := uc.toProductResponse(product)

		// Get stock information for each product
		if stock, err := uc.stockRepo.GetByProductID(ctx, product.ID); err == nil {
			response.AvailableStock = stock.AvailableQty
			response.ReservedStock = stock.ReservedQty
			response.TotalStock = stock.TotalQty
			response.StockStatus = stock.GetStockStatus()
		}

		productResponses[i] = response
	}

	return &ProductListResponse{
		Products:   productResponses,
		Pagination: paginationResult,
	}, nil
}

// GetProductsByCategory retrieves products by category
func (uc *ProductUseCase) GetProductsByCategory(ctx context.Context, category string, pagination utils.PaginationInfo) (*ProductListResponse, error) {
	products, paginationResult, err := uc.productRepo.GetByCategory(ctx, category, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get products by category")
		return nil, errors.NewInternalError("failed to get products by category", err)
	}

	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		response := uc.toProductResponse(product)

		// Get stock information for each product
		if stock, err := uc.stockRepo.GetByProductID(ctx, product.ID); err == nil {
			response.AvailableStock = stock.AvailableQty
			response.ReservedStock = stock.ReservedQty
			response.TotalStock = stock.TotalQty
			response.StockStatus = stock.GetStockStatus()
		}

		productResponses[i] = response
	}

	return &ProductListResponse{
		Products:   productResponses,
		Pagination: paginationResult,
	}, nil
}

// GetCategories retrieves all product categories
func (uc *ProductUseCase) GetCategories(ctx context.Context) ([]string, error) {
	categories, err := uc.productRepo.GetCategories(ctx)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get categories")
		return nil, errors.NewInternalError("failed to get categories", err)
	}

	return categories, nil
}

// GetLowStockProducts retrieves products with low stock
func (uc *ProductUseCase) GetLowStockProducts(ctx context.Context, pagination utils.PaginationInfo) (*ProductListResponse, error) {
	products, paginationResult, err := uc.productRepo.GetLowStockProducts(ctx, pagination)
	if err != nil {
		uc.logger.WithField("error", err.Error()).Error("Failed to get low stock products")
		return nil, errors.NewInternalError("failed to get low stock products", err)
	}

	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		response := uc.toProductResponse(product)

		// Get stock information for each product
		if stock, err := uc.stockRepo.GetByProductID(ctx, product.ID); err == nil {
			response.AvailableStock = stock.AvailableQty
			response.ReservedStock = stock.ReservedQty
			response.TotalStock = stock.TotalQty
			response.StockStatus = stock.GetStockStatus()
		}

		productResponses[i] = response
	}

	return &ProductListResponse{
		Products:   productResponses,
		Pagination: paginationResult,
	}, nil
}

// toProductResponse converts product entity to response
func (uc *ProductUseCase) toProductResponse(product *entities.Product) *ProductResponse {
	return &ProductResponse{
		ID:           product.ID,
		SKU:          product.SKU,
		Name:         product.Name,
		Description:  product.Description,
		Category:     product.Category,
		Price:        product.Price,
		Cost:         product.Cost,
		Status:       product.Status,
		Unit:         product.Unit,
		MinStock:     product.MinStock,
		ProfitMargin: product.GetProfitMargin(),
		ProfitAmount: product.GetProfitAmount(),
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
		CreatedBy:    product.CreatedBy,
	}
}
