package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicklaros/adol/pkg/errors"
)

func TestNewStock(t *testing.T) {
	t.Run("valid stock creation", func(t *testing.T) {
		productID := uuid.New()
		initialQty := 50
		reorderLevel := 10

		stock, err := NewStock(productID, initialQty, reorderLevel)

		require.NoError(t, err)
		assert.NotNil(t, stock)
		assert.NotEqual(t, uuid.Nil, stock.ID)
		assert.Equal(t, productID, stock.ProductID)
		assert.Equal(t, initialQty, stock.AvailableQty)
		assert.Equal(t, 0, stock.ReservedQty)
		assert.Equal(t, initialQty, stock.TotalQty)
		assert.Equal(t, reorderLevel, stock.ReorderLevel)
		assert.NotNil(t, stock.LastMovementAt)
		assert.WithinDuration(t, time.Now(), stock.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), stock.UpdatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), *stock.LastMovementAt, time.Second)
	})

	t.Run("valid stock creation with zero initial quantity", func(t *testing.T) {
		productID := uuid.New()
		initialQty := 0
		reorderLevel := 5

		stock, err := NewStock(productID, initialQty, reorderLevel)

		require.NoError(t, err)
		assert.NotNil(t, stock)
		assert.Equal(t, 0, stock.AvailableQty)
		assert.Equal(t, 0, stock.TotalQty)
		assert.Nil(t, stock.LastMovementAt) // No movement time set for zero initial quantity
	})

	t.Run("invalid initial quantity - negative", func(t *testing.T) {
		productID := uuid.New()

		stock, err := NewStock(productID, -10, 5)

		assert.Error(t, err)
		assert.Nil(t, stock)
		assert.Contains(t, err.Error(), "invalid initial quantity")
	})

	t.Run("invalid reorder level - negative", func(t *testing.T) {
		productID := uuid.New()

		stock, err := NewStock(productID, 10, -5)

		assert.Error(t, err)
		assert.Nil(t, stock)
		assert.Contains(t, err.Error(), "invalid reorder level")
	})

	t.Run("valid with zero reorder level", func(t *testing.T) {
		productID := uuid.New()

		stock, err := NewStock(productID, 10, 0)

		require.NoError(t, err)
		assert.NotNil(t, stock)
		assert.Equal(t, 0, stock.ReorderLevel)
	})
}

func TestNewStockMovement(t *testing.T) {
	t.Run("valid stock movement creation", func(t *testing.T) {
		productID := uuid.New()
		createdBy := uuid.New()
		movementType := StockMovementTypeIn
		reason := ReasonPurchase
		quantity := 25
		reference := "PO-001"
		notes := "Initial purchase"

		movement, err := NewStockMovement(productID, movementType, reason, quantity, reference, notes, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, movement)
		assert.NotEqual(t, uuid.Nil, movement.ID)
		assert.Equal(t, productID, movement.ProductID)
		assert.Equal(t, movementType, movement.Type)
		assert.Equal(t, reason, movement.Reason)
		assert.Equal(t, quantity, movement.Quantity)
		assert.Equal(t, reference, movement.Reference)
		assert.Equal(t, notes, movement.Notes)
		assert.Equal(t, createdBy, movement.CreatedBy)
		assert.WithinDuration(t, time.Now(), movement.CreatedAt, time.Second)
	})

	t.Run("invalid movement type", func(t *testing.T) {
		productID := uuid.New()
		createdBy := uuid.New()

		movement, err := NewStockMovement(productID, "invalid", ReasonPurchase, 25, "PO-001", "Notes", createdBy)

		assert.Error(t, err)
		assert.Nil(t, movement)
		assert.Contains(t, err.Error(), "invalid stock movement type")
	})

	t.Run("invalid movement reason", func(t *testing.T) {
		productID := uuid.New()
		createdBy := uuid.New()

		movement, err := NewStockMovement(productID, StockMovementTypeIn, "invalid", 25, "PO-001", "Notes", createdBy)

		assert.Error(t, err)
		assert.Nil(t, movement)
		assert.Contains(t, err.Error(), "invalid stock movement reason")
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		productID := uuid.New()
		createdBy := uuid.New()

		movement, err := NewStockMovement(productID, StockMovementTypeIn, ReasonPurchase, 0, "PO-001", "Notes", createdBy)

		assert.Error(t, err)
		assert.Nil(t, movement)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		productID := uuid.New()
		createdBy := uuid.New()

		movement, err := NewStockMovement(productID, StockMovementTypeIn, ReasonPurchase, -10, "PO-001", "Notes", createdBy)

		assert.Error(t, err)
		assert.Nil(t, movement)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_AddStock(t *testing.T) {
	t.Run("valid stock addition", func(t *testing.T) {
		stock := createValidStock(t)
		originalQty := stock.AvailableQty
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := stock.AddStock(25, ReasonPurchase)

		require.NoError(t, err)
		assert.Equal(t, originalQty+25, stock.AvailableQty)
		assert.Equal(t, stock.AvailableQty+stock.ReservedQty, stock.TotalQty)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
		assert.NotNil(t, stock.LastMovementAt)
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.AddStock(0, ReasonPurchase)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.AddStock(-10, ReasonPurchase)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_RemoveStock(t *testing.T) {
	t.Run("valid stock removal", func(t *testing.T) {
		stock := createValidStock(t)
		originalQty := stock.AvailableQty
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := stock.RemoveStock(15)

		require.NoError(t, err)
		assert.Equal(t, originalQty-15, stock.AvailableQty)
		assert.Equal(t, stock.AvailableQty+stock.ReservedQty, stock.TotalQty)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
		assert.NotNil(t, stock.LastMovementAt)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.RemoveStock(stock.AvailableQty + 10)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInsufficientStock, appErr.Type)
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.RemoveStock(0)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.RemoveStock(-5)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_ReserveStock(t *testing.T) {
	t.Run("valid stock reservation", func(t *testing.T) {
		stock := createValidStock(t)
		originalAvailable := stock.AvailableQty
		originalReserved := stock.ReservedQty
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		reserveQty := 20
		err := stock.ReserveStock(reserveQty)

		require.NoError(t, err)
		assert.Equal(t, originalAvailable-reserveQty, stock.AvailableQty)
		assert.Equal(t, originalReserved+reserveQty, stock.ReservedQty)
		assert.Equal(t, stock.AvailableQty+stock.ReservedQty, stock.TotalQty)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
		assert.NotNil(t, stock.LastMovementAt)
	})

	t.Run("insufficient available stock", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReserveStock(stock.AvailableQty + 10)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInsufficientStock, appErr.Type)
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReserveStock(0)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReserveStock(-5)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_ReleaseReservedStock(t *testing.T) {
	t.Run("valid release of reserved stock", func(t *testing.T) {
		stock := createValidStock(t)

		// First reserve some stock
		reserveQty := 20
		err := stock.ReserveStock(reserveQty)
		require.NoError(t, err)

		originalAvailable := stock.AvailableQty
		originalReserved := stock.ReservedQty
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		releaseQty := 10
		err = stock.ReleaseReservedStock(releaseQty)

		require.NoError(t, err)
		assert.Equal(t, originalAvailable+releaseQty, stock.AvailableQty)
		assert.Equal(t, originalReserved-releaseQty, stock.ReservedQty)
		assert.Equal(t, stock.AvailableQty+stock.ReservedQty, stock.TotalQty)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
		assert.NotNil(t, stock.LastMovementAt)
	})

	t.Run("insufficient reserved stock", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReleaseReservedStock(10)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient reserved stock")
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReleaseReservedStock(0)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ReleaseReservedStock(-5)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_ConfirmReservedStock(t *testing.T) {
	t.Run("valid confirmation of reserved stock", func(t *testing.T) {
		stock := createValidStock(t)

		// First reserve some stock
		reserveQty := 20
		err := stock.ReserveStock(reserveQty)
		require.NoError(t, err)

		originalAvailable := stock.AvailableQty
		originalReserved := stock.ReservedQty
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		confirmQty := 15
		err = stock.ConfirmReservedStock(confirmQty)

		require.NoError(t, err)
		assert.Equal(t, originalAvailable, stock.AvailableQty) // Available should not change
		assert.Equal(t, originalReserved-confirmQty, stock.ReservedQty)
		assert.Equal(t, stock.AvailableQty+stock.ReservedQty, stock.TotalQty)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
		assert.NotNil(t, stock.LastMovementAt)
	})

	t.Run("insufficient reserved stock", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ConfirmReservedStock(10)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient reserved stock")
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ConfirmReservedStock(0)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.ConfirmReservedStock(-5)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestStock_UpdateReorderLevel(t *testing.T) {
	t.Run("valid reorder level update", func(t *testing.T) {
		stock := createValidStock(t)
		originalUpdatedAt := stock.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		newLevel := 20
		err := stock.UpdateReorderLevel(newLevel)

		require.NoError(t, err)
		assert.Equal(t, newLevel, stock.ReorderLevel)
		assert.True(t, stock.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("valid reorder level - zero", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.UpdateReorderLevel(0)

		require.NoError(t, err)
		assert.Equal(t, 0, stock.ReorderLevel)
	})

	t.Run("invalid reorder level - negative", func(t *testing.T) {
		stock := createValidStock(t)

		err := stock.UpdateReorderLevel(-5)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid reorder level")
	})
}

func TestStock_IsLowStock(t *testing.T) {
	t.Run("stock above reorder level", func(t *testing.T) {
		stock := createValidStock(t) // Available: 50, Reorder: 10

		assert.False(t, stock.IsLowStock())
	})

	t.Run("stock at reorder level", func(t *testing.T) {
		stock := createValidStock(t)

		// Remove stock to reach reorder level
		err := stock.RemoveStock(stock.AvailableQty - stock.ReorderLevel)
		require.NoError(t, err)

		assert.True(t, stock.IsLowStock())
	})

	t.Run("stock below reorder level", func(t *testing.T) {
		stock := createValidStock(t)

		// Remove stock to go below reorder level
		err := stock.RemoveStock(stock.AvailableQty - stock.ReorderLevel + 1)
		require.NoError(t, err)

		assert.True(t, stock.IsLowStock())
	})
}

func TestStock_IsOutOfStock(t *testing.T) {
	t.Run("stock available", func(t *testing.T) {
		stock := createValidStock(t)

		assert.False(t, stock.IsOutOfStock())
	})

	t.Run("stock depleted", func(t *testing.T) {
		stock := createValidStock(t)

		// Remove all available stock
		err := stock.RemoveStock(stock.AvailableQty)
		require.NoError(t, err)

		assert.True(t, stock.IsOutOfStock())
	})
}

func TestStock_CanFulfillOrder(t *testing.T) {
	t.Run("sufficient stock", func(t *testing.T) {
		stock := createValidStock(t)

		assert.True(t, stock.CanFulfillOrder(25))
		assert.True(t, stock.CanFulfillOrder(stock.AvailableQty))
	})

	t.Run("insufficient stock", func(t *testing.T) {
		stock := createValidStock(t)

		assert.False(t, stock.CanFulfillOrder(stock.AvailableQty+1))
	})

	t.Run("exact stock amount", func(t *testing.T) {
		stock := createValidStock(t)

		assert.True(t, stock.CanFulfillOrder(stock.AvailableQty))
	})
}

func TestStock_GetStockStatus(t *testing.T) {
	t.Run("in stock status", func(t *testing.T) {
		stock := createValidStock(t)

		status := stock.GetStockStatus()

		assert.Equal(t, "In Stock", status)
	})

	t.Run("low stock status", func(t *testing.T) {
		stock := createValidStock(t)

		// Remove stock to reach reorder level
		err := stock.RemoveStock(stock.AvailableQty - stock.ReorderLevel)
		require.NoError(t, err)

		status := stock.GetStockStatus()

		assert.Equal(t, "Low Stock", status)
	})

	t.Run("out of stock status", func(t *testing.T) {
		stock := createValidStock(t)

		// Remove all available stock
		err := stock.RemoveStock(stock.AvailableQty)
		require.NoError(t, err)

		status := stock.GetStockStatus()

		assert.Equal(t, "Out of Stock", status)
	})
}

func TestValidateStockMovementType(t *testing.T) {
	testCases := []struct {
		name          string
		movementType  StockMovementType
		expectedError bool
	}{
		{"valid in type", StockMovementTypeIn, false},
		{"valid out type", StockMovementTypeOut, false},
		{"valid reserved type", StockMovementTypeReserved, false},
		{"valid released type", StockMovementTypeReleased, false},
		{"invalid type", "invalid", true},
		{"empty type", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStockMovementType(tc.movementType)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid stock movement type")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStockMovementReason(t *testing.T) {
	testCases := []struct {
		name          string
		reason        StockMovementReason
		expectedError bool
	}{
		{"valid purchase reason", ReasonPurchase, false},
		{"valid sale reason", ReasonSale, false},
		{"valid return reason", ReasonReturn, false},
		{"valid damage reason", ReasonDamage, false},
		{"valid expiry reason", ReasonExpiry, false},
		{"valid adjustment reason", ReasonAdjustment, false},
		{"valid reservation reason", ReasonReservation, false},
		{"valid release reason", ReasonRelease, false},
		{"invalid reason", "invalid", true},
		{"empty reason", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStockMovementReason(tc.reason)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid stock movement reason")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a valid stock for testing
func createValidStock(t *testing.T) *Stock {
	productID := uuid.New()

	stock, err := NewStock(productID, 50, 10) // 50 available, 10 reorder level

	require.NoError(t, err)
	require.NotNil(t, stock)

	return stock
}
