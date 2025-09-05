package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/pkg/errors"
)

// ProductStatus represents product status
type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Price       decimal.Decimal `json:"price"`
	Cost        decimal.Decimal `json:"cost"`
	Status      ProductStatus   `json:"status"`
	Unit        string          `json:"unit"` // e.g., "pcs", "kg", "ltr"
	MinStock    int             `json:"min_stock"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   uuid.UUID       `json:"created_by"`
}

// NewProduct creates a new product
func NewProduct(tenantID uuid.UUID, sku, name, description, category, unit string, price, cost decimal.Decimal, minStock int, createdBy uuid.UUID) (*Product, error) {
	if err := validateProductInput(sku, name, category, unit, price, cost, minStock); err != nil {
		return nil, err
	}

	now := time.Now()
	product := &Product{
		ID:          uuid.New(),
		TenantID:    tenantID,
		SKU:         sku,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
		Cost:        cost,
		Status:      ProductStatusActive,
		Unit:        unit,
		MinStock:    minStock,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
	}

	return product, nil
}

// UpdateProduct updates product information
func (p *Product) UpdateProduct(name, description, category, unit string, price, cost decimal.Decimal, minStock int) error {
	if err := validateProductUpdateInput(name, category, unit, price, cost, minStock); err != nil {
		return err
	}

	p.Name = name
	p.Description = description
	p.Category = category
	p.Price = price
	p.Cost = cost
	p.Unit = unit
	p.MinStock = minStock
	p.UpdatedAt = time.Now()

	return nil
}

// UpdatePrice updates the product price
func (p *Product) UpdatePrice(newPrice decimal.Decimal) error {
	if newPrice.LessThanOrEqual(decimal.Zero) {
		return errors.NewInvalidPriceError(newPrice.InexactFloat64())
	}

	p.Price = newPrice
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateCost updates the product cost
func (p *Product) UpdateCost(newCost decimal.Decimal) error {
	if newCost.LessThan(decimal.Zero) {
		return errors.NewValidationError("invalid cost", "cost cannot be negative")
	}

	p.Cost = newCost
	p.UpdatedAt = time.Now()
	return nil
}

// ChangeStatus changes the product status
func (p *Product) ChangeStatus(status ProductStatus) error {
	if err := ValidateProductStatus(status); err != nil {
		return err
	}

	p.Status = status
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateMinStock updates the minimum stock level
func (p *Product) UpdateMinStock(minStock int) error {
	if minStock < 0 {
		return errors.NewValidationError("invalid minimum stock", "minimum stock cannot be negative")
	}

	p.MinStock = minStock
	p.UpdatedAt = time.Now()
	return nil
}

// IsActive checks if the product is active
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

// GetProfitMargin calculates the profit margin percentage
func (p *Product) GetProfitMargin() decimal.Decimal {
	if p.Cost.IsZero() {
		return decimal.Zero
	}

	profit := p.Price.Sub(p.Cost)
	margin := profit.Div(p.Cost).Mul(decimal.NewFromInt(100))
	return margin.Round(2)
}

// GetProfitAmount calculates the profit amount per unit
func (p *Product) GetProfitAmount() decimal.Decimal {
	return p.Price.Sub(p.Cost)
}

// ValidateProductStatus validates if the status is valid
func ValidateProductStatus(status ProductStatus) error {
	switch status {
	case ProductStatusActive, ProductStatusInactive, ProductStatusDiscontinued:
		return nil
	default:
		return errors.NewValidationError("invalid product status", "status must be one of: active, inactive, discontinued")
	}
}

// Helper functions

func validateProductInput(sku, name, category, unit string, price, cost decimal.Decimal, minStock int) error {
	if sku == "" {
		return errors.NewValidationError("SKU is required", "sku cannot be empty")
	}
	if len(sku) < 3 {
		return errors.NewValidationError("SKU too short", "sku must be at least 3 characters long")
	}
	if name == "" {
		return errors.NewValidationError("product name is required", "name cannot be empty")
	}
	if category == "" {
		return errors.NewValidationError("category is required", "category cannot be empty")
	}
	if unit == "" {
		return errors.NewValidationError("unit is required", "unit cannot be empty")
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return errors.NewInvalidPriceError(price.InexactFloat64())
	}
	if cost.LessThan(decimal.Zero) {
		return errors.NewValidationError("invalid cost", "cost cannot be negative")
	}
	if minStock < 0 {
		return errors.NewValidationError("invalid minimum stock", "minimum stock cannot be negative")
	}
	return nil
}

func validateProductUpdateInput(name, category, unit string, price, cost decimal.Decimal, minStock int) error {
	if name == "" {
		return errors.NewValidationError("product name is required", "name cannot be empty")
	}
	if category == "" {
		return errors.NewValidationError("category is required", "category cannot be empty")
	}
	if unit == "" {
		return errors.NewValidationError("unit is required", "unit cannot be empty")
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return errors.NewInvalidPriceError(price.InexactFloat64())
	}
	if cost.LessThan(decimal.Zero) {
		return errors.NewValidationError("invalid cost", "cost cannot be negative")
	}
	if minStock < 0 {
		return errors.NewValidationError("invalid minimum stock", "minimum stock cannot be negative")
	}
	return nil
}
