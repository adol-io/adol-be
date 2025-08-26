package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/pkg/errors"
)

// StockMovementType represents the type of stock movement
type StockMovementType string

const (
	StockMovementTypeIn       StockMovementType = "in"       // Stock increase (purchase, return, adjustment)
	StockMovementTypeOut      StockMovementType = "out"      // Stock decrease (sale, damage, adjustment)
	StockMovementTypeReserved StockMovementType = "reserved" // Stock reserved for pending orders
	StockMovementTypeReleased StockMovementType = "released" // Reserved stock released back to available
)

// StockMovementReason represents the reason for stock movement
type StockMovementReason string

const (
	ReasonPurchase    StockMovementReason = "purchase"
	ReasonSale        StockMovementReason = "sale"
	ReasonReturn      StockMovementReason = "return"
	ReasonDamage      StockMovementReason = "damage"
	ReasonExpiry      StockMovementReason = "expiry"
	ReasonAdjustment  StockMovementReason = "adjustment"
	ReasonReservation StockMovementReason = "reservation"
	ReasonRelease     StockMovementReason = "release"
)

// Stock represents current stock levels for a product
type Stock struct {
	ID             uuid.UUID  `json:"id"`
	ProductID      uuid.UUID  `json:"product_id"`
	AvailableQty   int        `json:"available_qty"`
	ReservedQty    int        `json:"reserved_qty"`
	TotalQty       int        `json:"total_qty"` // available + reserved
	ReorderLevel   int        `json:"reorder_level"`
	LastMovementAt *time.Time `json:"last_movement_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// StockMovement represents a stock movement record
type StockMovement struct {
	ID        uuid.UUID           `json:"id"`
	ProductID uuid.UUID           `json:"product_id"`
	Type      StockMovementType   `json:"type"`
	Reason    StockMovementReason `json:"reason"`
	Quantity  int                 `json:"quantity"`
	Reference string              `json:"reference,omitempty"` // Order ID, Invoice ID, etc.
	Notes     string              `json:"notes,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	CreatedBy uuid.UUID           `json:"created_by"`
}

// NewStock creates a new stock record for a product
func NewStock(productID uuid.UUID, initialQty, reorderLevel int) (*Stock, error) {
	if initialQty < 0 {
		return nil, errors.NewValidationError("invalid initial quantity", "initial quantity cannot be negative")
	}
	if reorderLevel < 0 {
		return nil, errors.NewValidationError("invalid reorder level", "reorder level cannot be negative")
	}

	now := time.Now()
	stock := &Stock{
		ID:           uuid.New(),
		ProductID:    productID,
		AvailableQty: initialQty,
		ReservedQty:  0,
		TotalQty:     initialQty,
		ReorderLevel: reorderLevel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if initialQty > 0 {
		stock.LastMovementAt = &now
	}

	return stock, nil
}

// NewStockMovement creates a new stock movement record
func NewStockMovement(productID uuid.UUID, movementType StockMovementType, reason StockMovementReason, quantity int, reference, notes string, createdBy uuid.UUID) (*StockMovement, error) {
	if err := ValidateStockMovementType(movementType); err != nil {
		return nil, err
	}
	if err := ValidateStockMovementReason(reason); err != nil {
		return nil, err
	}
	if quantity <= 0 {
		return nil, errors.NewInvalidQuantityError(quantity)
	}

	movement := &StockMovement{
		ID:        uuid.New(),
		ProductID: productID,
		Type:      movementType,
		Reason:    reason,
		Quantity:  quantity,
		Reference: reference,
		Notes:     notes,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	return movement, nil
}

// AddStock increases available stock
func (s *Stock) AddStock(quantity int, reason StockMovementReason) error {
	if quantity <= 0 {
		return errors.NewInvalidQuantityError(quantity)
	}

	s.AvailableQty += quantity
	s.TotalQty = s.AvailableQty + s.ReservedQty
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastMovementAt = &now

	return nil
}

// RemoveStock decreases available stock
func (s *Stock) RemoveStock(quantity int) error {
	if quantity <= 0 {
		return errors.NewInvalidQuantityError(quantity)
	}
	if s.AvailableQty < quantity {
		return errors.NewInsufficientStockError("product", s.AvailableQty, quantity)
	}

	s.AvailableQty -= quantity
	s.TotalQty = s.AvailableQty + s.ReservedQty
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastMovementAt = &now

	return nil
}

// ReserveStock reserves stock for an order
func (s *Stock) ReserveStock(quantity int) error {
	if quantity <= 0 {
		return errors.NewInvalidQuantityError(quantity)
	}
	if s.AvailableQty < quantity {
		return errors.NewInsufficientStockError("product", s.AvailableQty, quantity)
	}

	s.AvailableQty -= quantity
	s.ReservedQty += quantity
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastMovementAt = &now

	return nil
}

// ReleaseReservedStock releases reserved stock back to available
func (s *Stock) ReleaseReservedStock(quantity int) error {
	if quantity <= 0 {
		return errors.NewInvalidQuantityError(quantity)
	}
	if s.ReservedQty < quantity {
		return errors.NewValidationError("insufficient reserved stock", "not enough reserved stock to release")
	}

	s.ReservedQty -= quantity
	s.AvailableQty += quantity
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastMovementAt = &now

	return nil
}

// ConfirmReservedStock confirms reserved stock (removes from reserved without adding back to available)
func (s *Stock) ConfirmReservedStock(quantity int) error {
	if quantity <= 0 {
		return errors.NewInvalidQuantityError(quantity)
	}
	if s.ReservedQty < quantity {
		return errors.NewValidationError("insufficient reserved stock", "not enough reserved stock to confirm")
	}

	s.ReservedQty -= quantity
	s.TotalQty = s.AvailableQty + s.ReservedQty
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastMovementAt = &now

	return nil
}

// UpdateReorderLevel updates the reorder level
func (s *Stock) UpdateReorderLevel(level int) error {
	if level < 0 {
		return errors.NewValidationError("invalid reorder level", "reorder level cannot be negative")
	}

	s.ReorderLevel = level
	s.UpdatedAt = time.Now()
	return nil
}

// IsLowStock checks if the stock is below reorder level
func (s *Stock) IsLowStock() bool {
	return s.AvailableQty <= s.ReorderLevel
}

// IsOutOfStock checks if the product is out of stock
func (s *Stock) IsOutOfStock() bool {
	return s.AvailableQty == 0
}

// CanFulfillOrder checks if there's enough stock to fulfill an order
func (s *Stock) CanFulfillOrder(quantity int) bool {
	return s.AvailableQty >= quantity
}

// GetStockStatus returns a human-readable stock status
func (s *Stock) GetStockStatus() string {
	if s.IsOutOfStock() {
		return "Out of Stock"
	}
	if s.IsLowStock() {
		return "Low Stock"
	}
	return "In Stock"
}

// ValidateStockMovementType validates stock movement type
func ValidateStockMovementType(movementType StockMovementType) error {
	switch movementType {
	case StockMovementTypeIn, StockMovementTypeOut, StockMovementTypeReserved, StockMovementTypeReleased:
		return nil
	default:
		return errors.NewValidationError("invalid stock movement type", "movement type must be one of: in, out, reserved, released")
	}
}

// ValidateStockMovementReason validates stock movement reason
func ValidateStockMovementReason(reason StockMovementReason) error {
	switch reason {
	case ReasonPurchase, ReasonSale, ReasonReturn, ReasonDamage, ReasonExpiry, ReasonAdjustment, ReasonReservation, ReasonRelease:
		return nil
	default:
		return errors.NewValidationError("invalid stock movement reason", "reason must be one of: purchase, sale, return, damage, expiry, adjustment, reservation, release")
	}
}
