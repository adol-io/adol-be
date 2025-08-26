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

func TestNewProduct(t *testing.T) {
	t.Run("valid product creation", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.NotEqual(t, uuid.Nil, product.ID)
		assert.Equal(t, "LAPTOP001", product.SKU)
		assert.Equal(t, "Gaming Laptop", product.Name)
		assert.Equal(t, "High-performance gaming laptop", product.Description)
		assert.Equal(t, "Electronics", product.Category)
		assert.True(t, price.Equal(product.Price))
		assert.True(t, cost.Equal(product.Cost))
		assert.Equal(t, ProductStatusActive, product.Status)
		assert.Equal(t, "pcs", product.Unit)
		assert.Equal(t, 5, product.MinStock)
		assert.Equal(t, createdBy, product.CreatedBy)
		assert.WithinDuration(t, time.Now(), product.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), product.UpdatedAt, time.Second)
	})

	t.Run("invalid SKU - empty", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "SKU is required")
	})

	t.Run("invalid SKU - too short", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"AB",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "SKU too short")
	})

	t.Run("invalid name - empty", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "product name is required")
	})

	t.Run("invalid category - empty", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "category is required")
	})

	t.Run("invalid unit - empty", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "unit is required")
	})

	t.Run("invalid price - zero", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.Zero
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid price - negative", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(-10.0)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid cost - negative", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(-10.0)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "invalid cost")
	})

	t.Run("valid cost - zero", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.Zero

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			5,
			createdBy,
		)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.True(t, cost.Equal(product.Cost))
	})

	t.Run("invalid min stock - negative", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			-1,
			createdBy,
		)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "invalid minimum stock")
	})

	t.Run("valid min stock - zero", func(t *testing.T) {
		createdBy := uuid.New()
		price := decimal.NewFromFloat(100.50)
		cost := decimal.NewFromFloat(75.25)

		product, err := NewProduct(
			"LAPTOP001",
			"Gaming Laptop",
			"High-performance gaming laptop",
			"Electronics",
			"pcs",
			price,
			cost,
			0,
			createdBy,
		)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, 0, product.MinStock)
	})
}

func TestProduct_UpdateProduct(t *testing.T) {
	t.Run("valid update", func(t *testing.T) {
		product := createValidProduct(t)
		originalUpdatedAt := product.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		newPrice := decimal.NewFromFloat(150.75)
		newCost := decimal.NewFromFloat(100.50)

		err := product.UpdateProduct(
			"Updated Gaming Laptop",
			"Updated description",
			"Updated Electronics",
			"units",
			newPrice,
			newCost,
			10,
		)

		require.NoError(t, err)
		assert.Equal(t, "Updated Gaming Laptop", product.Name)
		assert.Equal(t, "Updated description", product.Description)
		assert.Equal(t, "Updated Electronics", product.Category)
		assert.Equal(t, "units", product.Unit)
		assert.True(t, newPrice.Equal(product.Price))
		assert.True(t, newCost.Equal(product.Cost))
		assert.Equal(t, 10, product.MinStock)
		assert.True(t, product.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("invalid update - empty name", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateProduct(
			"",
			"Updated description",
			"Updated Electronics",
			"units",
			decimal.NewFromFloat(150.75),
			decimal.NewFromFloat(100.50),
			10,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product name is required")
	})

	t.Run("invalid update - negative price", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateProduct(
			"Updated Gaming Laptop",
			"Updated description",
			"Updated Electronics",
			"units",
			decimal.NewFromFloat(-10.0),
			decimal.NewFromFloat(100.50),
			10,
		)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})
}

func TestProduct_UpdatePrice(t *testing.T) {
	t.Run("valid price update", func(t *testing.T) {
		product := createValidProduct(t)
		originalUpdatedAt := product.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		newPrice := decimal.NewFromFloat(199.99)
		err := product.UpdatePrice(newPrice)

		require.NoError(t, err)
		assert.True(t, newPrice.Equal(product.Price))
		assert.True(t, product.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("invalid price - zero", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdatePrice(decimal.Zero)

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})

	t.Run("invalid price - negative", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdatePrice(decimal.NewFromFloat(-50.0))

		assert.Error(t, err)
		appErr, ok := errors.IsAppError(err)
		assert.True(t, ok)
		assert.Equal(t, errors.ErrorTypeInvalidPrice, appErr.Type)
	})
}

func TestProduct_UpdateCost(t *testing.T) {
	t.Run("valid cost update", func(t *testing.T) {
		product := createValidProduct(t)
		originalUpdatedAt := product.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		newCost := decimal.NewFromFloat(80.0)
		err := product.UpdateCost(newCost)

		require.NoError(t, err)
		assert.True(t, newCost.Equal(product.Cost))
		assert.True(t, product.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("valid cost - zero", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateCost(decimal.Zero)

		require.NoError(t, err)
		assert.True(t, decimal.Zero.Equal(product.Cost))
	})

	t.Run("invalid cost - negative", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateCost(decimal.NewFromFloat(-10.0))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cost")
	})
}

func TestProduct_ChangeStatus(t *testing.T) {
	t.Run("valid status changes", func(t *testing.T) {
		product := createValidProduct(t)
		originalUpdatedAt := product.UpdatedAt

		// Test changing to inactive
		time.Sleep(time.Millisecond)
		err := product.ChangeStatus(ProductStatusInactive)
		require.NoError(t, err)
		assert.Equal(t, ProductStatusInactive, product.Status)
		assert.True(t, product.UpdatedAt.After(originalUpdatedAt))

		// Test changing to discontinued
		inactiveUpdatedAt := product.UpdatedAt
		time.Sleep(time.Millisecond)
		err = product.ChangeStatus(ProductStatusDiscontinued)
		require.NoError(t, err)
		assert.Equal(t, ProductStatusDiscontinued, product.Status)
		assert.True(t, product.UpdatedAt.After(inactiveUpdatedAt))

		// Test changing back to active
		discontinuedUpdatedAt := product.UpdatedAt
		time.Sleep(time.Millisecond)
		err = product.ChangeStatus(ProductStatusActive)
		require.NoError(t, err)
		assert.Equal(t, ProductStatusActive, product.Status)
		assert.True(t, product.UpdatedAt.After(discontinuedUpdatedAt))
	})

	t.Run("invalid status", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.ChangeStatus("invalid_status")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid product status")
	})
}

func TestProduct_UpdateMinStock(t *testing.T) {
	t.Run("valid min stock update", func(t *testing.T) {
		product := createValidProduct(t)
		originalUpdatedAt := product.UpdatedAt

		// Wait a small amount to ensure UpdatedAt changes
		time.Sleep(time.Millisecond)

		err := product.UpdateMinStock(15)

		require.NoError(t, err)
		assert.Equal(t, 15, product.MinStock)
		assert.True(t, product.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("valid min stock - zero", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateMinStock(0)

		require.NoError(t, err)
		assert.Equal(t, 0, product.MinStock)
	})

	t.Run("invalid min stock - negative", func(t *testing.T) {
		product := createValidProduct(t)

		err := product.UpdateMinStock(-5)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid minimum stock")
	})
}

func TestProduct_IsActive(t *testing.T) {
	t.Run("active product", func(t *testing.T) {
		product := createValidProduct(t)
		assert.True(t, product.IsActive())
	})

	t.Run("inactive product", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.ChangeStatus(ProductStatusInactive)
		require.NoError(t, err)
		assert.False(t, product.IsActive())
	})

	t.Run("discontinued product", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.ChangeStatus(ProductStatusDiscontinued)
		require.NoError(t, err)
		assert.False(t, product.IsActive())
	})
}

func TestProduct_GetProfitMargin(t *testing.T) {
	t.Run("normal profit margin calculation", func(t *testing.T) {
		product := createValidProduct(t)
		// Price: 100.50, Cost: 75.25
		// Profit: 100.50 - 75.25 = 25.25
		// Margin: (25.25 / 75.25) * 100 = 33.55%

		margin := product.GetProfitMargin()
		expected := decimal.NewFromFloat(33.55)

		assert.True(t, expected.Equal(margin), "Expected %s, got %s", expected.String(), margin.String())
	})

	t.Run("zero cost - should return zero margin", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.UpdateCost(decimal.Zero)
		require.NoError(t, err)

		margin := product.GetProfitMargin()

		assert.True(t, decimal.Zero.Equal(margin))
	})

	t.Run("high profit margin", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.UpdatePrice(decimal.NewFromFloat(200.0))
		require.NoError(t, err)
		err = product.UpdateCost(decimal.NewFromFloat(50.0))
		require.NoError(t, err)

		// Price: 200, Cost: 50
		// Profit: 200 - 50 = 150
		// Margin: (150 / 50) * 100 = 300%

		margin := product.GetProfitMargin()
		expected := decimal.NewFromFloat(300.0)

		assert.True(t, expected.Equal(margin))
	})

	t.Run("negative profit margin", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.UpdatePrice(decimal.NewFromFloat(50.0))
		require.NoError(t, err)
		err = product.UpdateCost(decimal.NewFromFloat(100.0))
		require.NoError(t, err)

		// Price: 50, Cost: 100
		// Profit: 50 - 100 = -50
		// Margin: (-50 / 100) * 100 = -50%

		margin := product.GetProfitMargin()
		expected := decimal.NewFromFloat(-50.0)

		assert.True(t, expected.Equal(margin))
	})
}

func TestProduct_GetProfitAmount(t *testing.T) {
	t.Run("positive profit", func(t *testing.T) {
		product := createValidProduct(t)
		// Price: 100.50, Cost: 75.25
		// Profit: 100.50 - 75.25 = 25.25

		profit := product.GetProfitAmount()
		expected := decimal.NewFromFloat(25.25)

		assert.True(t, expected.Equal(profit))
	})

	t.Run("zero profit", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.UpdateCost(product.Price)
		require.NoError(t, err)

		profit := product.GetProfitAmount()

		assert.True(t, decimal.Zero.Equal(profit))
	})

	t.Run("negative profit (loss)", func(t *testing.T) {
		product := createValidProduct(t)
		err := product.UpdatePrice(decimal.NewFromFloat(50.0))
		require.NoError(t, err)
		err = product.UpdateCost(decimal.NewFromFloat(75.0))
		require.NoError(t, err)

		profit := product.GetProfitAmount()
		expected := decimal.NewFromFloat(-25.0)

		assert.True(t, expected.Equal(profit))
	})
}

func TestValidateProductStatus(t *testing.T) {
	testCases := []struct {
		name          string
		status        ProductStatus
		expectedError bool
	}{
		{"valid active status", ProductStatusActive, false},
		{"valid inactive status", ProductStatusInactive, false},
		{"valid discontinued status", ProductStatusDiscontinued, false},
		{"invalid status", "invalid", true},
		{"empty status", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProductStatus(tc.status)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid product status")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a valid product for testing
func createValidProduct(t *testing.T) *Product {
	createdBy := uuid.New()
	price := decimal.NewFromFloat(100.50)
	cost := decimal.NewFromFloat(75.25)

	product, err := NewProduct(
		"LAPTOP001",
		"Gaming Laptop",
		"High-performance gaming laptop",
		"Electronics",
		"pcs",
		price,
		cost,
		5,
		createdBy,
	)

	require.NoError(t, err)
	require.NotNil(t, product)

	return product
}
