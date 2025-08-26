package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicklaros/adol/pkg/errors"
)

func TestNewSale(t *testing.T) {
	t.Run("valid sale creation", func(t *testing.T) {
		createdBy := uuid.New()
		saleNumber := "SALE-001"
		customerName := "John Doe"
		customerEmail := "john@example.com"
		customerPhone := "+1234567890"

		sale, err := NewSale(saleNumber, customerName, customerEmail, customerPhone, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, sale)
		assert.NotEqual(t, uuid.Nil, sale.ID)
		assert.Equal(t, saleNumber, sale.SaleNumber)
		assert.Equal(t, customerName, sale.CustomerName)
		assert.Equal(t, customerEmail, sale.CustomerEmail)
		assert.Equal(t, customerPhone, sale.CustomerPhone)
		assert.Empty(t, sale.Items)
		assert.True(t, decimal.Zero.Equal(sale.Subtotal))
		assert.True(t, decimal.Zero.Equal(sale.TaxAmount))
		assert.True(t, decimal.Zero.Equal(sale.DiscountAmount))
		assert.True(t, decimal.Zero.Equal(sale.TotalAmount))
		assert.True(t, decimal.Zero.Equal(sale.PaidAmount))
		assert.True(t, decimal.Zero.Equal(sale.ChangeAmount))
		assert.Equal(t, SaleStatusPending, sale.Status)
		assert.Equal(t, createdBy, sale.CreatedBy)
		assert.WithinDuration(t, time.Now(), sale.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), sale.UpdatedAt, time.Second)
		assert.Nil(t, sale.CompletedAt)
	})

	t.Run("valid sale creation with minimal info", func(t *testing.T) {
		createdBy := uuid.New()
		saleNumber := "SALE-002"

		sale, err := NewSale(saleNumber, "", "", "", createdBy)

		require.NoError(t, err)
		assert.NotNil(t, sale)
		assert.Equal(t, saleNumber, sale.SaleNumber)
		assert.Empty(t, sale.CustomerName)
		assert.Empty(t, sale.CustomerEmail)
		assert.Empty(t, sale.CustomerPhone)
	})

	t.Run("invalid sale - empty sale number", func(t *testing.T) {
		createdBy := uuid.New()

		sale, err := NewSale("", "John Doe", "john@example.com", "+1234567890", createdBy)

		assert.Error(t, err)
		assert.Nil(t, sale)
		assert.Contains(t, err.Error(), "sale number is required")
	})
}

func TestNewSaleItem(t *testing.T) {
	t.Run("valid sale item creation", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()
		productSKU := "LAPTOP001"
		productName := "Gaming Laptop"
		quantity := 2
		unitPrice := decimal.NewFromFloat(999.99)

		item, err := NewSaleItem(saleID, productID, productSKU, productName, quantity, unitPrice)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.NotEqual(t, uuid.Nil, item.ID)
		assert.Equal(t, saleID, item.SaleID)
		assert.Equal(t, productID, item.ProductID)
		assert.Equal(t, productSKU, item.ProductSKU)
		assert.Equal(t, productName, item.ProductName)
		assert.Equal(t, quantity, item.Quantity)
		assert.True(t, unitPrice.Equal(item.UnitPrice))
		expectedTotal := unitPrice.Mul(decimal.NewFromInt(int64(quantity)))
		assert.True(t, expectedTotal.Equal(item.TotalPrice))
		assert.WithinDuration(t, time.Now(), item.CreatedAt, time.Second)
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "SKU001", "Product", 0, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "SKU001", "Product", -5, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid unit price - zero", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "SKU001", "Product", 1, decimal.Zero)

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid unit price - negative", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "SKU001", "Product", 1, decimal.NewFromFloat(-10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid product SKU - empty", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "", "Product", 1, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "product SKU is required")
	})

	t.Run("invalid product name - empty", func(t *testing.T) {
		saleID := uuid.New()
		productID := uuid.New()

		item, err := NewSaleItem(saleID, productID, "SKU001", "", 1, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "product name is required")
	})
}

func TestSale_AddItem(t *testing.T) {
	t.Run("add new item to sale", func(t *testing.T) {
		sale := createValidSale(t)
		item := createValidSaleItem(t, sale.ID)
		originalUpdatedAt := sale.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := sale.AddItem(item)

		require.NoError(t, err)
		assert.Len(t, sale.Items, 1)
		assert.Equal(t, item.ProductID, sale.Items[0].ProductID)
		assert.Equal(t, item.Quantity, sale.Items[0].Quantity)
		assert.True(t, item.TotalPrice.Equal(sale.Subtotal))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("add existing product - should combine quantities", func(t *testing.T) {
		sale := createValidSale(t)
		item1 := createValidSaleItem(t, sale.ID)
		item2 := createValidSaleItem(t, sale.ID)
		item2.ProductID = item1.ProductID // Same product

		err := sale.AddItem(item1)
		require.NoError(t, err)

		err = sale.AddItem(item2)
		require.NoError(t, err)

		assert.Len(t, sale.Items, 1) // Should still be 1 item
		assert.Equal(t, item1.Quantity+item2.Quantity, sale.Items[0].Quantity)
		expectedTotal := item1.UnitPrice.Mul(decimal.NewFromInt(int64(sale.Items[0].Quantity)))
		assert.True(t, expectedTotal.Equal(sale.Items[0].TotalPrice))
	})

	t.Run("add nil item", func(t *testing.T) {
		sale := createValidSale(t)

		err := sale.AddItem(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "item is required")
	})
}

func TestSale_RemoveItem(t *testing.T) {
	t.Run("remove existing item", func(t *testing.T) {
		sale := createValidSale(t)
		item := createValidSaleItem(t, sale.ID)

		err := sale.AddItem(item)
		require.NoError(t, err)

		originalUpdatedAt := sale.UpdatedAt
		time.Sleep(time.Millisecond)

		err = sale.RemoveItem(item.ProductID)

		require.NoError(t, err)
		assert.Len(t, sale.Items, 0)
		assert.True(t, decimal.Zero.Equal(sale.Subtotal))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("remove non-existing item", func(t *testing.T) {
		sale := createValidSale(t)

		err := sale.RemoveItem(uuid.New())

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
	})
}

func TestSale_UpdateItemQuantity(t *testing.T) {
	t.Run("update existing item quantity", func(t *testing.T) {
		sale := createValidSale(t)
		item := createValidSaleItem(t, sale.ID)

		err := sale.AddItem(item)
		require.NoError(t, err)

		originalUpdatedAt := sale.UpdatedAt
		time.Sleep(time.Millisecond)

		newQuantity := 5
		err = sale.UpdateItemQuantity(item.ProductID, newQuantity)

		require.NoError(t, err)
		assert.Equal(t, newQuantity, sale.Items[0].Quantity)
		expectedTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(newQuantity)))
		assert.True(t, expectedTotal.Equal(sale.Items[0].TotalPrice))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("update non-existing item", func(t *testing.T) {
		sale := createValidSale(t)

		err := sale.UpdateItemQuantity(uuid.New(), 5)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		sale := createValidSale(t)
		item := createValidSaleItem(t, sale.ID)

		err := sale.AddItem(item)
		require.NoError(t, err)

		err = sale.UpdateItemQuantity(item.ProductID, 0)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		sale := createValidSale(t)
		item := createValidSaleItem(t, sale.ID)

		err := sale.AddItem(item)
		require.NoError(t, err)

		err = sale.UpdateItemQuantity(item.ProductID, -2)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})
}

func TestSale_ApplyDiscount(t *testing.T) {
	t.Run("apply valid discount", func(t *testing.T) {
		sale := createSaleWithItems(t)
		originalUpdatedAt := sale.UpdatedAt
		discountAmount := decimal.NewFromFloat(50.0)

		time.Sleep(time.Millisecond)

		err := sale.ApplyDiscount(discountAmount)

		require.NoError(t, err)
		assert.True(t, discountAmount.Equal(sale.DiscountAmount))
		expectedTotal := sale.Subtotal.Sub(discountAmount).Add(sale.TaxAmount)
		assert.True(t, expectedTotal.Equal(sale.TotalAmount))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("apply zero discount", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ApplyDiscount(decimal.Zero)

		require.NoError(t, err)
		assert.True(t, decimal.Zero.Equal(sale.DiscountAmount))
	})

	t.Run("apply negative discount", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ApplyDiscount(decimal.NewFromFloat(-10.0))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid discount")
	})

	t.Run("apply discount greater than subtotal", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ApplyDiscount(sale.Subtotal.Add(decimal.NewFromFloat(100.0)))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid discount")
	})
}

func TestSale_ApplyTax(t *testing.T) {
	t.Run("apply valid tax", func(t *testing.T) {
		sale := createSaleWithItems(t)
		originalUpdatedAt := sale.UpdatedAt
		taxPercentage := decimal.NewFromFloat(10.0)

		time.Sleep(time.Millisecond)

		err := sale.ApplyTax(taxPercentage)

		require.NoError(t, err)
		expectedTax := sale.Subtotal.Sub(sale.DiscountAmount).Mul(taxPercentage).Div(decimal.NewFromInt(100))
		assert.True(t, expectedTax.Equal(sale.TaxAmount))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("apply zero tax", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ApplyTax(decimal.Zero)

		require.NoError(t, err)
		assert.True(t, decimal.Zero.Equal(sale.TaxAmount))
	})

	t.Run("apply negative tax", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ApplyTax(decimal.NewFromFloat(-5.0))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tax")
	})
}

func TestSale_ProcessPayment(t *testing.T) {
	t.Run("process valid payment", func(t *testing.T) {
		sale := createSaleWithItems(t)
		paidAmount := sale.TotalAmount.Add(decimal.NewFromFloat(10.0))
		paymentMethod := PaymentMethodCash
		originalUpdatedAt := sale.UpdatedAt

		time.Sleep(time.Millisecond)

		err := sale.ProcessPayment(paidAmount, paymentMethod)

		require.NoError(t, err)
		assert.True(t, paidAmount.Equal(sale.PaidAmount))
		assert.Equal(t, paymentMethod, sale.PaymentMethod)
		expectedChange := paidAmount.Sub(sale.TotalAmount)
		assert.True(t, expectedChange.Equal(sale.ChangeAmount))
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("process exact payment", func(t *testing.T) {
		sale := createSaleWithItems(t)
		paidAmount := sale.TotalAmount

		err := sale.ProcessPayment(paidAmount, PaymentMethodCard)

		require.NoError(t, err)
		assert.True(t, decimal.Zero.Equal(sale.ChangeAmount))
	})

	t.Run("insufficient payment", func(t *testing.T) {
		sale := createSaleWithItems(t)
		paidAmount := sale.TotalAmount.Sub(decimal.NewFromFloat(10.0))

		err := sale.ProcessPayment(paidAmount, PaymentMethodCash)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient payment")
	})

	t.Run("invalid payment method", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.ProcessPayment(sale.TotalAmount, "invalid_method")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment method")
	})
}

func TestSale_CompleteSale(t *testing.T) {
	t.Run("complete valid sale", func(t *testing.T) {
		sale := createSaleWithItems(t)
		err := sale.ProcessPayment(sale.TotalAmount, PaymentMethodCash)
		require.NoError(t, err)

		originalUpdatedAt := sale.UpdatedAt
		time.Sleep(time.Millisecond)

		err = sale.CompleteSale()

		require.NoError(t, err)
		assert.Equal(t, SaleStatusCompleted, sale.Status)
		assert.NotNil(t, sale.CompletedAt)
		assert.WithinDuration(t, time.Now(), *sale.CompletedAt, time.Second)
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("complete sale with wrong status", func(t *testing.T) {
		sale := createSaleWithItems(t)
		sale.Status = SaleStatusCompleted

		err := sale.CompleteSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sale status")
	})

	t.Run("complete empty sale", func(t *testing.T) {
		sale := createValidSale(t)

		err := sale.CompleteSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty sale")
	})

	t.Run("complete sale without payment", func(t *testing.T) {
		sale := createSaleWithItems(t)

		err := sale.CompleteSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incomplete payment")
	})
}

func TestSale_CancelSale(t *testing.T) {
	t.Run("cancel pending sale", func(t *testing.T) {
		sale := createValidSale(t)
		originalUpdatedAt := sale.UpdatedAt

		time.Sleep(time.Millisecond)

		err := sale.CancelSale()

		require.NoError(t, err)
		assert.Equal(t, SaleStatusCancelled, sale.Status)
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("cancel completed sale - should fail", func(t *testing.T) {
		sale := createValidSale(t)
		sale.Status = SaleStatusCompleted

		err := sale.CancelSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sale status")
	})

	t.Run("cancel already cancelled sale", func(t *testing.T) {
		sale := createValidSale(t)
		sale.Status = SaleStatusCancelled

		err := sale.CancelSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sale status")
	})
}

func TestSale_RefundSale(t *testing.T) {
	t.Run("refund completed sale", func(t *testing.T) {
		sale := createValidSale(t)
		sale.Status = SaleStatusCompleted
		originalUpdatedAt := sale.UpdatedAt

		time.Sleep(time.Millisecond)

		err := sale.RefundSale()

		require.NoError(t, err)
		assert.Equal(t, SaleStatusRefunded, sale.Status)
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("refund non-completed sale", func(t *testing.T) {
		sale := createValidSale(t)

		err := sale.RefundSale()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sale status")
	})
}

func TestSale_AddNotes(t *testing.T) {
	t.Run("add notes to sale", func(t *testing.T) {
		sale := createValidSale(t)
		notes := "Customer requested express delivery"
		originalUpdatedAt := sale.UpdatedAt

		time.Sleep(time.Millisecond)

		sale.AddNotes(notes)

		assert.Equal(t, notes, sale.Notes)
		assert.True(t, sale.UpdatedAt.After(originalUpdatedAt))
	})
}

func TestSale_GetItemCount(t *testing.T) {
	t.Run("get item count from sale", func(t *testing.T) {
		sale := createSaleWithItems(t)

		count := sale.GetItemCount()

		expectedCount := 0
		for _, item := range sale.Items {
			expectedCount += item.Quantity
		}
		assert.Equal(t, expectedCount, count)
	})

	t.Run("get item count from empty sale", func(t *testing.T) {
		sale := createValidSale(t)

		count := sale.GetItemCount()

		assert.Equal(t, 0, count)
	})
}

func TestSale_StatusChecks(t *testing.T) {
	t.Run("check completed status", func(t *testing.T) {
		sale := createValidSale(t)
		assert.False(t, sale.IsCompleted())

		sale.Status = SaleStatusCompleted
		assert.True(t, sale.IsCompleted())
	})

	t.Run("check cancelled status", func(t *testing.T) {
		sale := createValidSale(t)
		assert.False(t, sale.IsCancelled())

		sale.Status = SaleStatusCancelled
		assert.True(t, sale.IsCancelled())
	})

	t.Run("check refunded status", func(t *testing.T) {
		sale := createValidSale(t)
		assert.False(t, sale.IsRefunded())

		sale.Status = SaleStatusRefunded
		assert.True(t, sale.IsRefunded())
	})
}

func TestValidatePaymentMethod(t *testing.T) {
	testCases := []struct {
		name          string
		method        PaymentMethod
		expectedError bool
	}{
		{"valid cash method", PaymentMethodCash, false},
		{"valid card method", PaymentMethodCard, false},
		{"valid digital wallet method", PaymentMethodDigitalWallet, false},
		{"valid bank transfer method", PaymentMethodBankTransfer, false},
		{"invalid method", "invalid", true},
		{"empty method", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaymentMethod(tc.method)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid payment method")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSaleStatus(t *testing.T) {
	testCases := []struct {
		name          string
		status        SaleStatus
		expectedError bool
	}{
		{"valid pending status", SaleStatusPending, false},
		{"valid completed status", SaleStatusCompleted, false},
		{"valid cancelled status", SaleStatusCancelled, false},
		{"valid refunded status", SaleStatusRefunded, false},
		{"invalid status", "invalid", true},
		{"empty status", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSaleStatus(tc.status)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid sale status")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for creating test data

func createValidSale(t *testing.T) *Sale {
	createdBy := uuid.New()

	sale, err := NewSale("SALE-001", "John Doe", "john@example.com", "+1234567890", createdBy)

	require.NoError(t, err)
	require.NotNil(t, sale)

	return sale
}

func createValidSaleItem(t *testing.T, saleID uuid.UUID) *SaleItem {
	productID := uuid.New()
	productSKU := "LAPTOP001"
	productName := "Gaming Laptop"
	quantity := 2
	unitPrice := decimal.NewFromFloat(999.99)

	item, err := NewSaleItem(saleID, productID, productSKU, productName, quantity, unitPrice)

	require.NoError(t, err)
	require.NotNil(t, item)

	return item
}

func createSaleWithItems(t *testing.T) *Sale {
	sale := createValidSale(t)

	// Add first item
	item1 := createValidSaleItem(t, sale.ID)
	err := sale.AddItem(item1)
	require.NoError(t, err)

	// Add second item with different product
	item2, err := NewSaleItem(
		sale.ID,
		uuid.New(),
		"MOUSE001",
		"Gaming Mouse",
		1,
		decimal.NewFromFloat(79.99),
	)
	require.NoError(t, err)
	err = sale.AddItem(item2)
	require.NoError(t, err)

	return sale
}
