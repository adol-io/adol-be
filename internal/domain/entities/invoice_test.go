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

func TestNewInvoice(t *testing.T) {
	t.Run("valid invoice creation from completed sale", func(t *testing.T) {
		sale := createCompletedSale(t)
		createdBy := uuid.New()
		invoiceNumber := "INV-001"

		invoice, err := NewInvoice(invoiceNumber, sale, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.NotEqual(t, uuid.Nil, invoice.ID)
		assert.Equal(t, invoiceNumber, invoice.InvoiceNumber)
		assert.Equal(t, sale.ID, invoice.SaleID)
		assert.Equal(t, sale.CustomerName, invoice.CustomerName)
		assert.Equal(t, sale.CustomerEmail, invoice.CustomerEmail)
		assert.Equal(t, sale.CustomerPhone, invoice.CustomerPhone)
		assert.Len(t, invoice.Items, len(sale.Items))
		assert.True(t, sale.Subtotal.Equal(invoice.Subtotal))
		assert.True(t, sale.TaxAmount.Equal(invoice.TaxAmount))
		assert.True(t, sale.DiscountAmount.Equal(invoice.DiscountAmount))
		assert.True(t, sale.TotalAmount.Equal(invoice.TotalAmount))
		assert.True(t, sale.PaidAmount.Equal(invoice.PaidAmount))
		assert.Equal(t, sale.PaymentMethod, invoice.PaymentMethod)
		assert.Equal(t, InvoiceStatusDraft, invoice.Status)
		assert.Equal(t, sale.Notes, invoice.Notes)
		assert.Equal(t, createdBy, invoice.CreatedBy)
		assert.WithinDuration(t, time.Now(), invoice.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), invoice.UpdatedAt, time.Second)
		assert.Nil(t, invoice.DueDate)
		assert.Nil(t, invoice.PaidAt)
	})

	t.Run("invalid invoice - empty invoice number", func(t *testing.T) {
		sale := createCompletedSale(t)
		createdBy := uuid.New()

		invoice, err := NewInvoice("", sale, createdBy)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "invoice number is required")
	})

	t.Run("invalid invoice - nil sale", func(t *testing.T) {
		createdBy := uuid.New()

		invoice, err := NewInvoice("INV-001", nil, createdBy)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "sale is required")
	})

	t.Run("invalid invoice - non-completed sale", func(t *testing.T) {
		sale := createValidSale(t) // Pending sale
		createdBy := uuid.New()

		invoice, err := NewInvoice("INV-001", sale, createdBy)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "invalid sale status")
	})
}

func TestNewInvoiceItem(t *testing.T) {
	t.Run("valid invoice item creation", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()
		productSKU := "LAPTOP001"
		productName := "Gaming Laptop"
		description := "High-performance gaming laptop"
		quantity := 2
		unitPrice := decimal.NewFromFloat(999.99)

		item, err := NewInvoiceItem(invoiceID, productID, productSKU, productName, description, quantity, unitPrice)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.NotEqual(t, uuid.Nil, item.ID)
		assert.Equal(t, invoiceID, item.InvoiceID)
		assert.Equal(t, productID, item.ProductID)
		assert.Equal(t, productSKU, item.ProductSKU)
		assert.Equal(t, productName, item.ProductName)
		assert.Equal(t, description, item.Description)
		assert.Equal(t, quantity, item.Quantity)
		assert.True(t, unitPrice.Equal(item.UnitPrice))
		expectedTotal := unitPrice.Mul(decimal.NewFromInt(int64(quantity)))
		assert.True(t, expectedTotal.Equal(item.TotalPrice))
	})

	t.Run("invalid quantity - zero", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "SKU001", "Product", "Description", 0, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid quantity - negative", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "SKU001", "Product", "Description", -5, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidQuantity, appErr.Type)
	})

	t.Run("invalid unit price - zero", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "SKU001", "Product", "Description", 1, decimal.Zero)

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid unit price - negative", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "SKU001", "Product", "Description", 1, decimal.NewFromFloat(-10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid product SKU - empty", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "", "Product", "Description", 1, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "product SKU is required")
	})

	t.Run("invalid product name - empty", func(t *testing.T) {
		invoiceID := uuid.New()
		productID := uuid.New()

		item, err := NewInvoiceItem(invoiceID, productID, "SKU001", "", "Description", 1, decimal.NewFromFloat(10.0))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "product name is required")
	})
}

func TestInvoice_UpdateCustomerInfo(t *testing.T) {
	t.Run("valid customer info update", func(t *testing.T) {
		invoice := createValidInvoice(t)
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		name := "Jane Doe"
		email := "jane@example.com"
		phone := "+1234567890"
		address := "123 Main St, City, State"

		err := invoice.UpdateCustomerInfo(name, email, phone, address)

		require.NoError(t, err)
		assert.Equal(t, name, invoice.CustomerName)
		assert.Equal(t, email, invoice.CustomerEmail)
		assert.Equal(t, phone, invoice.CustomerPhone)
		assert.Equal(t, address, invoice.CustomerAddress)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("invalid customer info - empty name", func(t *testing.T) {
		invoice := createValidInvoice(t)

		err := invoice.UpdateCustomerInfo("", "email@example.com", "+1234567890", "123 Main St")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer name is required")
	})
}

func TestInvoice_SetDueDate(t *testing.T) {
	t.Run("valid due date", func(t *testing.T) {
		invoice := createValidInvoice(t)
		originalUpdatedAt := invoice.UpdatedAt
		dueDate := time.Now().Add(7 * 24 * time.Hour) // 7 days from now

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := invoice.SetDueDate(dueDate)

		require.NoError(t, err)
		assert.NotNil(t, invoice.DueDate)
		assert.WithinDuration(t, dueDate, *invoice.DueDate, time.Second)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("invalid due date - before creation date", func(t *testing.T) {
		invoice := createValidInvoice(t)
		dueDate := invoice.CreatedAt.Add(-1 * time.Hour) // 1 hour before creation

		err := invoice.SetDueDate(dueDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid due date")
	})
}

func TestInvoice_MarkAsGenerated(t *testing.T) {
	t.Run("mark draft invoice as generated", func(t *testing.T) {
		invoice := createValidInvoice(t)
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := invoice.MarkAsGenerated()

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusGenerated, invoice.Status)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("mark non-draft invoice as generated - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusGenerated

		err := invoice.MarkAsGenerated()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})
}

func TestInvoice_MarkAsSent(t *testing.T) {
	t.Run("mark generated invoice as sent", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusGenerated
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := invoice.MarkAsSent()

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusSent, invoice.Status)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("mark non-generated invoice as sent - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)

		err := invoice.MarkAsSent()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})
}

func TestInvoice_MarkAsPaid(t *testing.T) {
	t.Run("mark sent invoice as paid", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusSent
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := invoice.MarkAsPaid()

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusPaid, invoice.Status)
		assert.NotNil(t, invoice.PaidAt)
		assert.WithinDuration(t, time.Now(), *invoice.PaidAt, time.Second)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("mark cancelled invoice as paid - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusCancelled

		err := invoice.MarkAsPaid()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})

	t.Run("mark already paid invoice as paid - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusPaid

		err := invoice.MarkAsPaid()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})
}

func TestInvoice_Cancel(t *testing.T) {
	t.Run("cancel draft invoice", func(t *testing.T) {
		invoice := createValidInvoice(t)
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := invoice.Cancel()

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusCancelled, invoice.Status)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("cancel paid invoice - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusPaid

		err := invoice.Cancel()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})

	t.Run("cancel already cancelled invoice - should fail", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusCancelled

		err := invoice.Cancel()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice status")
	})
}

func TestInvoice_AddNotes(t *testing.T) {
	t.Run("add notes to invoice", func(t *testing.T) {
		invoice := createValidInvoice(t)
		notes := "Customer requested express delivery"
		originalUpdatedAt := invoice.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		invoice.AddNotes(notes)

		assert.Equal(t, notes, invoice.Notes)
		assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))
	})
}

func TestInvoice_StatusChecks(t *testing.T) {
	t.Run("check draft status", func(t *testing.T) {
		invoice := createValidInvoice(t)
		assert.True(t, invoice.IsDraft())
		assert.False(t, invoice.IsGenerated())
		assert.False(t, invoice.IsSent())
		assert.False(t, invoice.IsPaid())
		assert.False(t, invoice.IsCancelled())
	})

	t.Run("check generated status", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusGenerated
		assert.False(t, invoice.IsDraft())
		assert.True(t, invoice.IsGenerated())
		assert.False(t, invoice.IsSent())
		assert.False(t, invoice.IsPaid())
		assert.False(t, invoice.IsCancelled())
	})

	t.Run("check sent status", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusSent
		assert.False(t, invoice.IsDraft())
		assert.False(t, invoice.IsGenerated())
		assert.True(t, invoice.IsSent())
		assert.False(t, invoice.IsPaid())
		assert.False(t, invoice.IsCancelled())
	})

	t.Run("check paid status", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusPaid
		assert.False(t, invoice.IsDraft())
		assert.False(t, invoice.IsGenerated())
		assert.False(t, invoice.IsSent())
		assert.True(t, invoice.IsPaid())
		assert.False(t, invoice.IsCancelled())
	})

	t.Run("check cancelled status", func(t *testing.T) {
		invoice := createValidInvoice(t)
		invoice.Status = InvoiceStatusCancelled
		assert.False(t, invoice.IsDraft())
		assert.False(t, invoice.IsGenerated())
		assert.False(t, invoice.IsSent())
		assert.False(t, invoice.IsPaid())
		assert.True(t, invoice.IsCancelled())
	})
}

func TestInvoice_IsOverdue(t *testing.T) {
	t.Run("invoice without due date - not overdue", func(t *testing.T) {
		invoice := createValidInvoice(t)

		assert.False(t, invoice.IsOverdue())
	})

	t.Run("paid invoice - not overdue", func(t *testing.T) {
		invoice := createValidInvoice(t)
		// Set due date to after creation but before now
		pastDueDate := invoice.CreatedAt.Add(1 * time.Hour)
		err := invoice.SetDueDate(pastDueDate)
		require.NoError(t, err)
		invoice.Status = InvoiceStatusPaid

		assert.False(t, invoice.IsOverdue())
	})

	t.Run("cancelled invoice - not overdue", func(t *testing.T) {
		invoice := createValidInvoice(t)
		// Set due date to after creation but before now
		pastDueDate := invoice.CreatedAt.Add(1 * time.Hour)
		err := invoice.SetDueDate(pastDueDate)
		require.NoError(t, err)
		invoice.Status = InvoiceStatusCancelled

		assert.False(t, invoice.IsOverdue())
	})

	t.Run("unpaid invoice past due date - overdue", func(t *testing.T) {
		invoice := createValidInvoice(t)
		// Set due date to after creation but before now (simulate time passing)
		pastDueDate := invoice.CreatedAt.Add(1 * time.Hour)
		err := invoice.SetDueDate(pastDueDate)
		require.NoError(t, err)
		invoice.Status = InvoiceStatusSent

		// Simulate time passing by manually setting due date to past
		*invoice.DueDate = time.Now().Add(-1 * time.Hour)

		assert.True(t, invoice.IsOverdue())
	})

	t.Run("unpaid invoice before due date - not overdue", func(t *testing.T) {
		invoice := createValidInvoice(t)
		futureDueDate := time.Now().Add(24 * time.Hour)
		err := invoice.SetDueDate(futureDueDate)
		require.NoError(t, err)
		invoice.Status = InvoiceStatusSent

		assert.False(t, invoice.IsOverdue())
	})
}

func TestInvoice_GetItemCount(t *testing.T) {
	t.Run("get item count from invoice", func(t *testing.T) {
		invoice := createValidInvoice(t)

		count := invoice.GetItemCount()

		expectedCount := 0
		for _, item := range invoice.Items {
			expectedCount += item.Quantity
		}
		assert.Equal(t, expectedCount, count)
	})
}

func TestGetPaperSizeDimensions(t *testing.T) {
	testCases := []struct {
		name           string
		paperSize      PaperSize
		expectedWidth  float64
		expectedHeight float64
	}{
		{"A4 dimensions", PaperSizeA4, 595, 842},
		{"A5 dimensions", PaperSizeA5, 420, 595},
		{"Letter dimensions", PaperSizeLetter, 612, 792},
		{"Legal dimensions", PaperSizeLegal, 612, 1008},
		{"Receipt dimensions", PaperSizeReceipt, 226, 0},
		{"Invalid paper size - defaults to A4", "invalid", 595, 842},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width, height := GetPaperSizeDimensions(tc.paperSize)

			assert.Equal(t, tc.expectedWidth, width)
			assert.Equal(t, tc.expectedHeight, height)
		})
	}
}

func TestValidateInvoiceStatus(t *testing.T) {
	testCases := []struct {
		name          string
		status        InvoiceStatus
		expectedError bool
	}{
		{"valid draft status", InvoiceStatusDraft, false},
		{"valid generated status", InvoiceStatusGenerated, false},
		{"valid sent status", InvoiceStatusSent, false},
		{"valid paid status", InvoiceStatusPaid, false},
		{"valid cancelled status", InvoiceStatusCancelled, false},
		{"invalid status", "invalid", true},
		{"empty status", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateInvoiceStatus(tc.status)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid invoice status")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePaperSize(t *testing.T) {
	testCases := []struct {
		name          string
		size          PaperSize
		expectedError bool
	}{
		{"valid A4 size", PaperSizeA4, false},
		{"valid A5 size", PaperSizeA5, false},
		{"valid letter size", PaperSizeLetter, false},
		{"valid legal size", PaperSizeLegal, false},
		{"valid receipt size", PaperSizeReceipt, false},
		{"invalid size", "invalid", true},
		{"empty size", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaperSize(tc.size)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid paper size")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for creating test data

func createValidInvoice(t *testing.T) *Invoice {
	sale := createCompletedSale(t)
	createdBy := uuid.New()

	invoice, err := NewInvoice("INV-001", sale, createdBy)

	require.NoError(t, err)
	require.NotNil(t, invoice)

	return invoice
}

func createCompletedSale(t *testing.T) *Sale {
	sale := createSaleWithItems(t)

	// Apply tax
	err := sale.ApplyTax(decimal.NewFromFloat(10.0))
	require.NoError(t, err)

	// Process payment
	err = sale.ProcessPayment(sale.TotalAmount, PaymentMethodCash)
	require.NoError(t, err)

	// Complete sale
	err = sale.CompleteSale()
	require.NoError(t, err)

	return sale
}
