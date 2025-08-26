package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/utils"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *entities.Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)

	// GetBySKU retrieves a product by SKU
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *entities.Product) error

	// Delete deletes a product (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves products with pagination and filtering
	List(ctx context.Context, filter ProductFilter, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error)

	// GetByCategory retrieves products by category
	GetByCategory(ctx context.Context, category string, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error)

	// ExistsBySKU checks if a product exists by SKU
	ExistsBySKU(ctx context.Context, sku string) (bool, error)

	// GetCategories retrieves all unique categories
	GetCategories(ctx context.Context) ([]string, error)

	// GetLowStockProducts retrieves products with low stock
	GetLowStockProducts(ctx context.Context, pagination utils.PaginationInfo) ([]*entities.Product, utils.PaginationInfo, error)
}

// ProductFilter represents filters for product queries
type ProductFilter struct {
	Category string                  `json:"category,omitempty"`
	Status   *entities.ProductStatus `json:"status,omitempty"`
	Search   string                  `json:"search,omitempty"` // Search in name, description, SKU
	MinPrice *float64                `json:"min_price,omitempty"`
	MaxPrice *float64                `json:"max_price,omitempty"`
	OrderBy  string                  `json:"order_by,omitempty"`
	OrderDir string                  `json:"order_dir,omitempty"` // ASC or DESC
}
