package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	infraRepos "github.com/nicklaros/adol/internal/infrastructure/repositories"
	"github.com/nicklaros/adol/pkg/utils"
)

func TestStockRepository_Integration(t *testing.T) {
	// Setup test database
	testDB := SetupTestDB(t)
	defer TeardownTestDB(t, testDB)

	// Setup test context and user
	ctx, _ := SetupTestContext(t)
	userID, userCleanup := CreateTestUser(t, testDB.DB)
	defer userCleanup()

	// Create repositories
	stockRepo := infraRepos.NewPostgreSQLStockRepository(testDB.DB)
	stockMovementRepo := infraRepos.NewPostgreSQLStockMovementRepository(testDB.DB)

	t.Run("Create and Get Stock", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Get stock (should be created automatically with product)
		stock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, uuid.MustParse(productID), stock.ProductID)
		assert.Equal(t, 100, stock.AvailableQty) // From CreateTestProduct
		assert.Equal(t, 0, stock.ReservedQty)
		assert.Equal(t, 100, stock.TotalQty)
		assert.Equal(t, 10, stock.ReorderLevel)
	})

	t.Run("Update Stock", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Get initial stock
		stock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)

		// Update stock
		stock.AvailableQty = 150
		stock.ReservedQty = 10
		stock.ReorderLevel = 15
		stock.UpdatedAt = time.Now()

		err = stockRepo.Update(ctx, stock)
		require.NoError(t, err)

		// Get updated stock
		updatedStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, 150, updatedStock.AvailableQty)
		assert.Equal(t, 10, updatedStock.ReservedQty)
		assert.Equal(t, 160, updatedStock.TotalQty) // Calculated field
		assert.Equal(t, 15, updatedStock.ReorderLevel)
	})

	t.Run("Adjust Stock", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Get initial stock
		initialStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		initialQty := initialStock.AvailableQty

		// Adjust stock positively
		adjustment := repositories.StockAdjustment{
			ProductID: uuid.MustParse(productID),
			Quantity:  25,
			Reason:    entities.ReasonAdjustment,
			Reference: "TEST-ADJ-001",
			Notes:     "Test adjustment",
			CreatedBy: uuid.MustParse(userID),
		}

		err = stockRepo.AdjustStock(ctx, adjustment)
		require.NoError(t, err)

		// Verify stock increased
		adjustedStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, initialQty+25, adjustedStock.AvailableQty)

		// Adjust stock negatively
		negativeAdjustment := repositories.StockAdjustment{
			ProductID: uuid.MustParse(productID),
			Quantity:  -15,
			Reason:    entities.ReasonAdjustment,
			Reference: "TEST-ADJ-002",
			Notes:     "Negative adjustment",
			CreatedBy: uuid.MustParse(userID),
		}

		err = stockRepo.AdjustStock(ctx, negativeAdjustment)
		require.NoError(t, err)

		// Verify stock decreased
		finalStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, initialQty+25-15, finalStock.AvailableQty)
	})

	t.Run("Reserve and Release Stock", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Get initial stock
		initialStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		initialAvailable := initialStock.AvailableQty
		initialReserved := initialStock.ReservedQty

		// Reserve stock
		reservation := repositories.StockReservation{
			ProductID: uuid.MustParse(productID),
			Quantity:  20,
			Reference: "SALE-001",
			Notes:     "Test reservation",
			CreatedBy: uuid.MustParse(userID),
		}

		err = stockRepo.ReserveStock(ctx, reservation)
		require.NoError(t, err)

		// Verify reservation
		reservedStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, initialAvailable-20, reservedStock.AvailableQty)
		assert.Equal(t, initialReserved+20, reservedStock.ReservedQty)
		assert.Equal(t, initialAvailable+initialReserved, reservedStock.TotalQty) // Total unchanged

		// Release stock
		release := repositories.StockRelease{
			ProductID: uuid.MustParse(productID),
			Quantity:  20,
			Reference: "SALE-001",
			Notes:     "Test release",
			CreatedBy: uuid.MustParse(userID),
		}

		err = stockRepo.ReleaseReservedStock(ctx, release)
		require.NoError(t, err)

		// Verify release
		releasedStock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)
		assert.Equal(t, initialAvailable, releasedStock.AvailableQty)
		assert.Equal(t, initialReserved, releasedStock.ReservedQty)
	})

	t.Run("Stock Movements", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Create stock movements
		movements := []*entities.StockMovement{
			{
				ID:        uuid.New(),
				ProductID: uuid.MustParse(productID),
				Type:      entities.StockMovementTypeIn,
				Reason:    entities.ReasonPurchase,
				Quantity:  50,
				Reference: "PO-001",
				Notes:     "Purchase order",
				CreatedAt: time.Now(),
				CreatedBy: uuid.MustParse(userID),
			},
			{
				ID:        uuid.New(),
				ProductID: uuid.MustParse(productID),
				Type:      entities.StockMovementTypeOut,
				Reason:    entities.ReasonSale,
				Quantity:  25,
				Reference: "SALE-001",
				Notes:     "Sale transaction",
				CreatedAt: time.Now().Add(time.Hour),
				CreatedBy: uuid.MustParse(userID),
			},
		}

		// Create movements
		for _, movement := range movements {
			err := stockMovementRepo.Create(ctx, movement)
			require.NoError(t, err)
		}

		// Get movements by product
		pagination := utils.PaginationInfo{Page: 1, Limit: 10}
		retrievedMovements, _, err := stockMovementRepo.GetByProductID(ctx, uuid.MustParse(productID), pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrievedMovements), 2)

		// Verify movement details
		foundPurchase := false
		foundSale := false
		for _, movement := range retrievedMovements {
			if movement.Reference == "PO-001" {
				foundPurchase = true
				assert.Equal(t, entities.StockMovementTypeIn, movement.Type)
				assert.Equal(t, entities.ReasonPurchase, movement.Reason)
				assert.Equal(t, 50, movement.Quantity)
			}
			if movement.Reference == "SALE-001" {
				foundSale = true
				assert.Equal(t, entities.StockMovementTypeOut, movement.Type)
				assert.Equal(t, entities.ReasonSale, movement.Reason)
				assert.Equal(t, 25, movement.Quantity)
			}
		}
		assert.True(t, foundPurchase, "Should find purchase movement")
		assert.True(t, foundSale, "Should find sale movement")

		// Test filtering by type
		filter := repositories.StockMovementFilter{
			ProductID: func() *uuid.UUID { id := uuid.MustParse(productID); return &id }(),
			Type:      func() *entities.StockMovementType { t := entities.StockMovementTypeIn; return &t }(),
		}
		inMovements, _, err := stockMovementRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(inMovements), 1)

		// All movements should be of type "in"
		for _, movement := range inMovements {
			assert.Equal(t, entities.StockMovementTypeIn, movement.Type)
		}
	})

	t.Run("Low Stock Detection", func(t *testing.T) {
		// Create test product with low stock
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Update stock to be below reorder level
		stock, err := stockRepo.GetByProductID(ctx, uuid.MustParse(productID))
		require.NoError(t, err)

		stock.AvailableQty = 5  // Below reorder level of 10
		stock.ReorderLevel = 10
		err = stockRepo.Update(ctx, stock)
		require.NoError(t, err)

		// Get low stock items
		pagination := utils.PaginationInfo{Page: 1, Limit: 10}
		lowStockItems, _, err := stockRepo.GetLowStockItems(ctx, pagination)
		require.NoError(t, err)

		// Should find our low stock item
		found := false
		for _, item := range lowStockItems {
			if item.ProductID == uuid.MustParse(productID) {
				found = true
				assert.True(t, item.AvailableQty <= item.ReorderLevel)
				break
			}
		}
		assert.True(t, found, "Should find low stock item")
	})

	t.Run("List Stock with Pagination", func(t *testing.T) {
		// Create multiple test products
		productIDs := make([]string, 3)
		productCleanups := make([]func(), 3)

		for i := 0; i < 3; i++ {
			productID, cleanup := CreateTestProduct(t, testDB.DB, userID)
			productIDs[i] = productID
			productCleanups[i] = cleanup
		}
		defer func() {
			for _, cleanup := range productCleanups {
				cleanup()
			}
		}()

		// List stock with pagination
		pagination := utils.PaginationInfo{Page: 1, Limit: 2}
		filter := repositories.StockFilter{} // Empty filter
		stockList, resultPagination, err := stockRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(stockList), 2)
		assert.Equal(t, 2, resultPagination.Limit)
		assert.Equal(t, 1, resultPagination.Page)
		assert.Greater(t, resultPagination.TotalCount, 0)
	})

	t.Run("Bulk Operations", func(t *testing.T) {
		// Create multiple test products
		productIDs := make([]uuid.UUID, 3)
		productCleanups := make([]func(), 3)

		for i := 0; i < 3; i++ {
			productID, cleanup := CreateTestProduct(t, testDB.DB, userID)
			productIDs[i] = uuid.MustParse(productID)
			productCleanups[i] = cleanup
		}
		defer func() {
			for _, cleanup := range productCleanups {
				cleanup()
			}
		}()

		// Bulk reserve stock
		reservations := make([]repositories.StockReservation, len(productIDs))
		for i, productID := range productIDs {
			reservations[i] = repositories.StockReservation{
				ProductID: productID,
				Quantity:  10,
				Reference: fmt.Sprintf("BULK-SALE-%d", i),
				Notes:     "Bulk reservation test",
				CreatedBy: uuid.MustParse(userID),
			}
		}

		err := stockRepo.BulkReserveStock(ctx, reservations)
		require.NoError(t, err)

		// Verify all reservations
		for _, productID := range productIDs {
			stock, err := stockRepo.GetByProductID(ctx, productID)
			require.NoError(t, err)
			assert.Equal(t, 10, stock.ReservedQty)
			assert.Equal(t, 90, stock.AvailableQty) // 100 - 10
		}

		// Bulk release stock
		releases := make([]repositories.StockRelease, len(productIDs))
		for i, productID := range productIDs {
			releases[i] = repositories.StockRelease{
				ProductID: productID,
				Quantity:  10,
				Reference: fmt.Sprintf("BULK-SALE-%d", i),
				Notes:     "Bulk release test",
				CreatedBy: uuid.MustParse(userID),
			}
		}

		err = stockRepo.BulkReleaseStock(ctx, releases)
		require.NoError(t, err)

		// Verify all releases
		for _, productID := range productIDs {
			stock, err := stockRepo.GetByProductID(ctx, productID)
			require.NoError(t, err)
			assert.Equal(t, 0, stock.ReservedQty)
			assert.Equal(t, 100, stock.AvailableQty) // Back to original
		}
	})

	t.Run("Stock Movement History", func(t *testing.T) {
		// Create test product
		productID, productCleanup := CreateTestProduct(t, testDB.DB, userID)
		defer productCleanup()

		// Create various stock movements over time
		movements := []*entities.StockMovement{
			{
				ID:        uuid.New(),
				ProductID: uuid.MustParse(productID),
				Type:      entities.StockMovementTypeIn,
				Reason:    entities.ReasonPurchase,
				Quantity:  100,
				Reference: "PO-001",
				Notes:     "Initial purchase",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				CreatedBy: uuid.MustParse(userID),
			},
			{
				ID:        uuid.New(),
				ProductID: uuid.MustParse(productID),
				Type:      entities.StockMovementTypeOut,
				Reason:    entities.ReasonSale,
				Quantity:  30,
				Reference: "SALE-001",
				Notes:     "First sale",
				CreatedAt: time.Now().Add(-1 * time.Hour),
				CreatedBy: uuid.MustParse(userID),
			},
			{
				ID:        uuid.New(),
				ProductID: uuid.MustParse(productID),
				Type:      entities.StockMovementTypeOut,
				Reason:    entities.ReasonSale,
				Quantity:  20,
				Reference: "SALE-002",
				Notes:     "Second sale",
				CreatedAt: time.Now(),
				CreatedBy: uuid.MustParse(userID),
			},
		}

		// Create all movements
		for _, movement := range movements {
			err := stockMovementRepo.Create(ctx, movement)
			require.NoError(t, err)
		}

		// Get movement history
		pagination := utils.PaginationInfo{Page: 1, Limit: 10}
		history, _, err := stockMovementRepo.GetByProductID(ctx, uuid.MustParse(productID), pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(history), 3)

		// Verify movements are ordered by creation time (newest first)
		if len(history) >= 2 {
			assert.True(t, history[0].CreatedAt.After(history[1].CreatedAt) || history[0].CreatedAt.Equal(history[1].CreatedAt))
		}

		// Test date range filtering
		fromDate := time.Now().Add(-90 * time.Minute)
		toDate := time.Now().Add(10 * time.Minute)

		filter := repositories.StockMovementFilter{
			ProductID: func() *uuid.UUID { id := uuid.MustParse(productID); return &id }(),
			FromDate:  &fromDate,
			ToDate:    &toDate,
		}

		filteredHistory, _, err := stockMovementRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(filteredHistory), 2) // Should get last 2 movements

		// Verify all movements are within date range
		for _, movement := range filteredHistory {
			assert.True(t, movement.CreatedAt.After(fromDate) || movement.CreatedAt.Equal(fromDate))
			assert.True(t, movement.CreatedAt.Before(toDate) || movement.CreatedAt.Equal(toDate))
		}
	})
}