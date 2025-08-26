package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	infraRepos "github.com/nicklaros/adol/internal/infrastructure/repositories"
	"github.com/nicklaros/adol/pkg/pointer"
	"github.com/nicklaros/adol/pkg/utils"
)

func TestProductRepository_Integration(t *testing.T) {
	// Setup test database
	testDB := SetupTestDB(t)
	defer TeardownTestDB(t, testDB)

	// Setup test context and user
	ctx, _ := SetupTestContext(t)
	userID, userCleanup := CreateTestUser(t, testDB.DB)
	defer userCleanup()

	// Create repository
	productRepo := infraRepos.NewPostgreSQLProductRepository(testDB.DB)

	t.Run("Create and Get Product", func(t *testing.T) {
		// Create test product
		product := &entities.Product{
			ID:          uuid.New(),
			SKU:         "TEST-PRODUCT-001",
			Name:        "Test Product",
			Description: "Test product description",
			Category:    "Electronics",
			Price:       decimal.NewFromFloat(99.99),
			Cost:        decimal.NewFromFloat(50.00),
			Unit:        "piece",
			MinStock:    10,
			Status:      entities.ProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   uuid.MustParse(userID),
		}

		// Create product
		err := productRepo.Create(ctx, product)
		require.NoError(t, err)

		// Get product by ID
		retrievedProduct, err := productRepo.GetByID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.ID, retrievedProduct.ID)
		assert.Equal(t, product.SKU, retrievedProduct.SKU)
		assert.Equal(t, product.Name, retrievedProduct.Name)
		assert.Equal(t, product.Description, retrievedProduct.Description)
		assert.Equal(t, product.Category, retrievedProduct.Category)
		assert.True(t, product.Price.Equal(retrievedProduct.Price))
		assert.True(t, product.Cost.Equal(retrievedProduct.Cost))
		assert.Equal(t, product.Unit, retrievedProduct.Unit)
		assert.Equal(t, product.MinStock, retrievedProduct.MinStock)
		assert.Equal(t, product.Status, retrievedProduct.Status)
		assert.Equal(t, product.CreatedBy, retrievedProduct.CreatedBy)

		// Get product by SKU
		productBySKU, err := productRepo.GetBySKU(ctx, product.SKU)
		require.NoError(t, err)
		assert.Equal(t, product.ID, productBySKU.ID)

		// Cleanup
		err = productRepo.Delete(ctx, product.ID)
		require.NoError(t, err)
	})

	t.Run("Update Product", func(t *testing.T) {
		// Create test product
		product := &entities.Product{
			ID:          uuid.New(),
			SKU:         "UPDATE-PRODUCT-001",
			Name:        "Update Product",
			Description: "Original description",
			Category:    "Category1",
			Price:       decimal.NewFromFloat(10.00),
			Cost:        decimal.NewFromFloat(5.00),
			Unit:        "piece",
			MinStock:    5,
			Status:      entities.ProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   uuid.MustParse(userID),
		}

		// Create product
		err := productRepo.Create(ctx, product)
		require.NoError(t, err)

		// Update product
		product.Name = "Updated Product Name"
		product.Description = "Updated description"
		product.Category = "Category2"
		product.Price = decimal.NewFromFloat(15.00)
		product.Cost = decimal.NewFromFloat(7.50)
		product.MinStock = 8
		product.UpdatedAt = time.Now()

		err = productRepo.Update(ctx, product)
		require.NoError(t, err)

		// Get updated product
		updatedProduct, err := productRepo.GetByID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Product Name", updatedProduct.Name)
		assert.Equal(t, "Updated description", updatedProduct.Description)
		assert.Equal(t, "Category2", updatedProduct.Category)
		assert.True(t, decimal.NewFromFloat(15.00).Equal(updatedProduct.Price))
		assert.True(t, decimal.NewFromFloat(7.50).Equal(updatedProduct.Cost))
		assert.Equal(t, 8, updatedProduct.MinStock)

		// Cleanup
		err = productRepo.Delete(ctx, product.ID)
		require.NoError(t, err)
	})

	t.Run("List Products with Filtering", func(t *testing.T) {
		// Create multiple test products
		products := make([]*entities.Product, 3)
		for i := 0; i < 3; i++ {
			products[i] = &entities.Product{
				ID:          uuid.New(),
				SKU:         fmt.Sprintf("LIST-PRODUCT-%03d", i),
				Name:        fmt.Sprintf("List Product %d", i),
				Description: fmt.Sprintf("Description %d", i),
				Category:    "TestCategory",
				Price:       decimal.NewFromFloat(float64(10 + i)),
				Cost:        decimal.NewFromFloat(float64(5 + i)),
				Unit:        "piece",
				MinStock:    5,
				Status:      entities.ProductStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				CreatedBy:   uuid.MustParse(userID),
			}

			err := productRepo.Create(ctx, products[i])
			require.NoError(t, err)
		}

		// Test basic list with pagination
		filter := repositories.ProductFilter{}
		pagination := utils.PaginationInfo{
			Page:  1,
			Limit: 2,
		}

		productList, resultPagination, err := productRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(productList), 2)
		assert.Equal(t, 2, resultPagination.Limit)
		assert.Equal(t, 1, resultPagination.Page)

		// Test filtering by category
		filter.Category = "TestCategory"
		filteredProducts, _, err := productRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(filteredProducts), 2) // At least 2 from our test data

		// Test filtering by status
		filter = repositories.ProductFilter{
			Status: pointer.Of(entities.ProductStatusActive),
		}
		activeProducts, _, err := productRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activeProducts), 2)

		// Test search functionality
		filter = repositories.ProductFilter{
			Search: "List Product",
		}
		searchResults, _, err := productRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(searchResults), 2) // Should find our test products

		// Cleanup
		for _, product := range products {
			err = productRepo.Delete(ctx, product.ID)
			require.NoError(t, err)
		}
	})

	t.Run("Get Categories", func(t *testing.T) {
		// Create products with different categories
		categories := []string{"Electronics", "Books", "Clothing"}
		products := make([]*entities.Product, len(categories))

		for i, category := range categories {
			products[i] = &entities.Product{
				ID:          uuid.New(),
				SKU:         fmt.Sprintf("CAT-PRODUCT-%03d", i),
				Name:        fmt.Sprintf("Category Product %d", i),
				Description: "Test description",
				Category:    category,
				Price:       decimal.NewFromFloat(10.00),
				Cost:        decimal.NewFromFloat(5.00),
				Unit:        "piece",
				MinStock:    5,
				Status:      entities.ProductStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				CreatedBy:   uuid.MustParse(userID),
			}

			err := productRepo.Create(ctx, products[i])
			require.NoError(t, err)
		}

		// Get categories
		retrievedCategories, err := productRepo.GetCategories(ctx)
		require.NoError(t, err)

		// Verify all our categories are present
		for _, expectedCategory := range categories {
			found := false
			for _, retrievedCategory := range retrievedCategories {
				if retrievedCategory == expectedCategory {
					found = true
					break
				}
			}
			assert.True(t, found, "Category %s should be in the list", expectedCategory)
		}

		// Cleanup
		for _, product := range products {
			err = productRepo.Delete(ctx, product.ID)
			require.NoError(t, err)
		}
	})

	t.Run("Get By Category", func(t *testing.T) {
		// Create products in specific category
		testCategory := "TestCategorySpecific"
		products := make([]*entities.Product, 2)

		for i := 0; i < 2; i++ {
			products[i] = &entities.Product{
				ID:          uuid.New(),
				SKU:         fmt.Sprintf("BYCAT-PRODUCT-%03d", i),
				Name:        fmt.Sprintf("ByCategory Product %d", i),
				Description: "Test description",
				Category:    testCategory,
				Price:       decimal.NewFromFloat(10.00),
				Cost:        decimal.NewFromFloat(5.00),
				Unit:        "piece",
				MinStock:    5,
				Status:      entities.ProductStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				CreatedBy:   uuid.MustParse(userID),
			}

			err := productRepo.Create(ctx, products[i])
			require.NoError(t, err)
		}

		// Get products by category
		pagination := utils.PaginationInfo{Page: 1, Limit: 10}
		categoryProducts, _, err := productRepo.GetByCategory(ctx, testCategory, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(categoryProducts), 2)

		// Verify all returned products are in the correct category
		for _, product := range categoryProducts {
			assert.Equal(t, testCategory, product.Category)
		}

		// Cleanup
		for _, product := range products {
			err = productRepo.Delete(ctx, product.ID)
			require.NoError(t, err)
		}
	})

	t.Run("SKU Uniqueness", func(t *testing.T) {
		// Create first product
		product1 := &entities.Product{
			ID:          uuid.New(),
			SKU:         "UNIQUE-SKU-001",
			Name:        "First Product",
			Description: "First description",
			Category:    "Category1",
			Price:       decimal.NewFromFloat(10.00),
			Cost:        decimal.NewFromFloat(5.00),
			Unit:        "piece",
			MinStock:    5,
			Status:      entities.ProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   uuid.MustParse(userID),
		}

		err := productRepo.Create(ctx, product1)
		require.NoError(t, err)

		// Try to create product with same SKU
		product2 := &entities.Product{
			ID:          uuid.New(),
			SKU:         "UNIQUE-SKU-001", // Same SKU
			Name:        "Second Product",
			Description: "Second description",
			Category:    "Category2",
			Price:       decimal.NewFromFloat(15.00),
			Cost:        decimal.NewFromFloat(7.50),
			Unit:        "piece",
			MinStock:    5,
			Status:      entities.ProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   uuid.MustParse(userID),
		}

		err = productRepo.Create(ctx, product2)
		assert.Error(t, err) // Should fail due to unique SKU constraint

		// Test ExistsBySKU
		exists, err := productRepo.ExistsBySKU(ctx, "UNIQUE-SKU-001")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = productRepo.ExistsBySKU(ctx, "NON-EXISTENT-SKU")
		require.NoError(t, err)
		assert.False(t, exists)

		// Cleanup
		err = productRepo.Delete(ctx, product1.ID)
		require.NoError(t, err)
	})

	t.Run("Soft Delete", func(t *testing.T) {
		// Create test product
		product := &entities.Product{
			ID:          uuid.New(),
			SKU:         "DELETE-PRODUCT-001",
			Name:        "Delete Product",
			Description: "To be deleted",
			Category:    "DeleteCategory",
			Price:       decimal.NewFromFloat(10.00),
			Cost:        decimal.NewFromFloat(5.00),
			Unit:        "piece",
			MinStock:    5,
			Status:      entities.ProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   uuid.MustParse(userID),
		}

		err := productRepo.Create(ctx, product)
		require.NoError(t, err)

		// Verify product exists
		retrievedProduct, err := productRepo.GetByID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.ID, retrievedProduct.ID)

		// Soft delete product
		err = productRepo.Delete(ctx, product.ID)
		require.NoError(t, err)

		// Verify product is no longer accessible via GetByID
		_, err = productRepo.GetByID(ctx, product.ID)
		assert.Error(t, err) // Should not be found

		// Verify product doesn't appear in lists
		filter := repositories.ProductFilter{
			Category: "DeleteCategory",
		}
		pagination := utils.PaginationInfo{Page: 1, Limit: 10}
		products, _, err := productRepo.List(ctx, filter, pagination)
		require.NoError(t, err)

		// Should not find our deleted product
		found := false
		for _, p := range products {
			if p.ID == product.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "Deleted product should not appear in list")

	})
}
