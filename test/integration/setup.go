package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/nicklaros/adol/internal/infrastructure/config"
	"github.com/nicklaros/adol/pkg/logger"
)

// TestDB represents a test database connection
type TestDB struct {
	DB   *sql.DB
	Name string
}

// SetupTestDB creates a test database and applies migrations
func SetupTestDB(t *testing.T) *TestDB {
	// Load test configuration
	cfg := getTestConfig()

	// Create a unique database name for this test
	testDBName := fmt.Sprintf("adol_test_%d", os.Getpid())

	// Connect to postgres to create test database
	adminConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)

	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer adminDB.Close()

	// Create test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	testConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, testDBName, cfg.Database.SSLMode)

	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := testDB.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Apply migrations
	driver, err := postgres.WithInstance(testDB, &postgres.Config{})
	if err != nil {
		t.Fatalf("Failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://../../migrations", testDBName, driver)
	if err != nil {
		t.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDB{
		DB:   testDB,
		Name: testDBName,
	}
}

// TeardownTestDB cleans up the test database
func TeardownTestDB(t *testing.T, testDB *TestDB) {
	// Close test database connection
	testDB.DB.Close()

	// Load test configuration
	cfg := getTestConfig()

	// Connect to postgres to drop test database
	adminConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)

	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		t.Logf("Failed to connect to postgres for cleanup: %v", err)
		return
	}
	defer adminDB.Close()

	// Terminate connections to test database
	_, err = adminDB.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='%s' AND pid <> pg_backend_pid()", testDB.Name))
	if err != nil {
		t.Logf("Failed to terminate connections: %v", err)
	}

	// Drop test database
	_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDB.Name))
	if err != nil {
		t.Logf("Failed to drop test database: %v", err)
	}
}

// getTestConfig returns test configuration
func getTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
			User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
			Password: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
			SSLMode:  getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
		},
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestContext creates a test context with logger
func SetupTestContext(t *testing.T) (context.Context, logger.Logger) {
	ctx := context.Background()
	
	// Create test logger
	testLogger := logger.NewLogger()
	
	return ctx, testLogger
}

// CreateTestUser creates a test user for integration tests
func CreateTestUser(t *testing.T, db *sql.DB) (userID string, cleanup func()) {
	userID = "550e8400-e29b-41d4-a716-446655440001" // Fixed UUID for testing
	
	// Insert test user
	query := `
		INSERT INTO users (id, username, email, first_name, last_name, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	
	_, err := db.Exec(query,
		userID,
		"testuser",
		"test@example.com",
		"Test",
		"User",
		"$2a$10$hash", // Mock password hash
		"admin",
		"active",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	cleanup = func() {
		_, err := db.Exec("DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			t.Logf("Failed to cleanup test user: %v", err)
		}
	}
	
	return userID, cleanup
}

// CreateTestProduct creates a test product for integration tests
func CreateTestProduct(t *testing.T, db *sql.DB, createdBy string) (productID string, cleanup func()) {
	productID = "550e8400-e29b-41d4-a716-446655440002" // Fixed UUID for testing
	
	// Insert test product
	query := `
		INSERT INTO products (id, sku, name, description, category, price, cost, unit, min_stock, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())`
	
	_, err := db.Exec(query,
		productID,
		"TEST-SKU-001",
		"Test Product",
		"Test Description",
		"Test Category",
		"10.99",
		"5.00",
		"piece",
		10,
		"active",
		createdBy,
	)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}
	
	// Create stock entry
	stockQuery := `
		INSERT INTO stock (product_id, available_qty, reserved_qty, reorder_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`
	
	_, err = db.Exec(stockQuery, productID, 100, 0, 10)
	if err != nil {
		t.Fatalf("Failed to create test stock: %v", err)
	}
	
	cleanup = func() {
		_, err := db.Exec("DELETE FROM stock WHERE product_id = $1", productID)
		if err != nil {
			t.Logf("Failed to cleanup test stock: %v", err)
		}
		_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
		if err != nil {
			t.Logf("Failed to cleanup test product: %v", err)
		}
	}
	
	return productID, cleanup
}