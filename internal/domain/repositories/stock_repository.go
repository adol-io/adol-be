package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/utils"
)

// StockRepository defines the interface for stock data access
type StockRepository interface {
	// Create creates a new stock record
	Create(ctx context.Context, stock *entities.Stock) error

	// GetByID retrieves stock by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Stock, error)

	// GetByProductID retrieves stock by product ID
	GetByProductID(ctx context.Context, productID uuid.UUID) (*entities.Stock, error)

	// Update updates stock information
	Update(ctx context.Context, stock *entities.Stock) error

	// Delete deletes a stock record
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves stock records with pagination and filtering
	List(ctx context.Context, filter StockFilter, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error)

	// GetLowStockItems retrieves items with stock below reorder level
	GetLowStockItems(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error)

	// GetOutOfStockItems retrieves items that are out of stock
	GetOutOfStockItems(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Stock, utils.PaginationInfo, error)

	// BulkUpdateStock updates multiple stock records in a transaction
	BulkUpdateStock(ctx context.Context, stocks []*entities.Stock) error

	// AdjustStock adjusts stock quantity (positive or negative)
	AdjustStock(ctx context.Context, adjustment StockAdjustment) error

	// ReserveStock reserves stock for an order
	ReserveStock(ctx context.Context, reservation StockReservation) error

	// ReleaseReservedStock releases reserved stock back to available
	ReleaseReservedStock(ctx context.Context, release StockRelease) error

	// BulkReserveStock reserves stock for multiple products in a transaction
	BulkReserveStock(ctx context.Context, reservations []StockReservation) error

	// BulkReleaseStock releases reserved stock for multiple products in a transaction
	BulkReleaseStock(ctx context.Context, releases []StockRelease) error
}

// StockMovementRepository defines the interface for stock movement data access
type StockMovementRepository interface {
	// Create creates a new stock movement record
	Create(ctx context.Context, movement *entities.StockMovement) error

	// GetByID retrieves a stock movement by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.StockMovement, error)

	// List retrieves stock movements with pagination and filtering
	List(ctx context.Context, filter StockMovementFilter, pagination utils.PaginationInfo) ([]*entities.StockMovement, utils.PaginationInfo, error)

	// GetByProductID retrieves stock movements for a specific product
	GetByProductID(ctx context.Context, productID uuid.UUID, pagination utils.PaginationInfo) ([]*entities.StockMovement, utils.PaginationInfo, error)

	// GetByReference retrieves stock movements by reference (e.g., sale ID, invoice ID)
	GetByReference(ctx context.Context, reference string) ([]*entities.StockMovement, error)

	// Delete deletes a stock movement record
	Delete(ctx context.Context, id uuid.UUID) error

	// BulkCreate creates multiple stock movement records in a transaction
	BulkCreate(ctx context.Context, movements []*entities.StockMovement) error
}

// StockFilter represents filters for stock queries
type StockFilter struct {
	ProductID  *uuid.UUID `json:"product_id,omitempty"`
	LowStock   *bool      `json:"low_stock,omitempty"`    // Filter for items below reorder level
	OutOfStock *bool      `json:"out_of_stock,omitempty"` // Filter for items with zero stock
	Search     string     `json:"search,omitempty"`       // Search in product name/SKU
	OrderBy    string     `json:"order_by,omitempty"`
	OrderDir   string     `json:"order_dir,omitempty"` // ASC or DESC
}

// StockMovementFilter represents filters for stock movement queries
type StockMovementFilter struct {
	ProductID *uuid.UUID                    `json:"product_id,omitempty"`
	Type      *entities.StockMovementType   `json:"type,omitempty"`
	Reason    *entities.StockMovementReason `json:"reason,omitempty"`
	Reference string                        `json:"reference,omitempty"`
	CreatedBy *uuid.UUID                    `json:"created_by,omitempty"`
	FromDate  *time.Time                    `json:"from_date,omitempty"`
	ToDate    *time.Time                    `json:"to_date,omitempty"`
	OrderBy   string                        `json:"order_by,omitempty"`
	OrderDir  string                        `json:"order_dir,omitempty"` // ASC or DESC
}

// StockAdjustment represents a stock adjustment operation
type StockAdjustment struct {
	ProductID uuid.UUID                   `json:"product_id"`
	Quantity  int                         `json:"quantity"` // Can be positive or negative
	Reason    entities.StockMovementReason `json:"reason"`
	Reference string                      `json:"reference,omitempty"`
	Notes     string                      `json:"notes,omitempty"`
	CreatedBy uuid.UUID                   `json:"created_by"`
}

// StockReservation represents a stock reservation operation
type StockReservation struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Reference string    `json:"reference,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedBy uuid.UUID `json:"created_by"`
}

// StockRelease represents a stock release operation
type StockRelease struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Reference string    `json:"reference,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedBy uuid.UUID `json:"created_by"`
}
