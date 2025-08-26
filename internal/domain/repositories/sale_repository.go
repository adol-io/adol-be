package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/utils"
)

// SaleRepository defines the interface for sale data access
type SaleRepository interface {
	// Create creates a new sale
	Create(ctx context.Context, sale *entities.Sale) error

	// GetByID retrieves a sale by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Sale, error)

	// GetBySaleNumber retrieves a sale by sale number
	GetBySaleNumber(ctx context.Context, saleNumber string) (*entities.Sale, error)

	// Update updates an existing sale
	Update(ctx context.Context, sale *entities.Sale) error

	// Delete deletes a sale (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves sales with pagination and filtering
	List(ctx context.Context, filter SaleFilter, pagination utils.PaginationInfo) ([]*entities.Sale, utils.PaginationInfo, error)

	// GetSalesReport generates sales report for a date range
	GetSalesReport(ctx context.Context, fromDate, toDate time.Time) (*SalesReport, error)

	// GetDailySales retrieves daily sales summary
	GetDailySales(ctx context.Context, date time.Time) (*DailySalesReport, error)

	// GetTotalSalesByUser retrieves total sales amount by user
	GetTotalSalesByUser(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (decimal.Decimal, error)

	// ExistsBySaleNumber checks if a sale exists by sale number
	ExistsBySaleNumber(ctx context.Context, saleNumber string) (bool, error)
}

// SaleItemRepository defines the interface for sale item data access
type SaleItemRepository interface {
	// Create creates a new sale item
	Create(ctx context.Context, item *entities.SaleItem) error

	// GetByID retrieves a sale item by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.SaleItem, error)

	// GetBySaleID retrieves all items for a sale
	GetBySaleID(ctx context.Context, saleID uuid.UUID) ([]*entities.SaleItem, error)

	// Update updates a sale item
	Update(ctx context.Context, item *entities.SaleItem) error

	// Delete deletes a sale item
	Delete(ctx context.Context, id uuid.UUID) error

	// BulkCreate creates multiple sale items in a transaction
	BulkCreate(ctx context.Context, items []*entities.SaleItem) error

	// BulkUpdate updates multiple sale items in a transaction
	BulkUpdate(ctx context.Context, items []*entities.SaleItem) error

	// DeleteBySaleID deletes all items for a sale
	DeleteBySaleID(ctx context.Context, saleID uuid.UUID) error

	// GetTopSellingProducts retrieves top selling products by quantity or revenue
	GetTopSellingProducts(ctx context.Context, fromDate, toDate time.Time, limit int, byRevenue bool) ([]*ProductSalesStats, error)
}

// SaleFilter represents filters for sale queries
type SaleFilter struct {
	Status        *entities.SaleStatus    `json:"status,omitempty"`
	PaymentMethod *entities.PaymentMethod `json:"payment_method,omitempty"`
	CreatedBy     *uuid.UUID              `json:"created_by,omitempty"`
	CustomerName  string                  `json:"customer_name,omitempty"`
	CustomerEmail string                  `json:"customer_email,omitempty"`
	FromDate      *time.Time              `json:"from_date,omitempty"`
	ToDate        *time.Time              `json:"to_date,omitempty"`
	MinAmount     *decimal.Decimal        `json:"min_amount,omitempty"`
	MaxAmount     *decimal.Decimal        `json:"max_amount,omitempty"`
	Search        string                  `json:"search,omitempty"` // Search in sale_number, customer_name, customer_email
	OrderBy       string                  `json:"order_by,omitempty"`
	OrderDir      string                  `json:"order_dir,omitempty"` // ASC or DESC
}

// SalesReport represents a sales report for a date range
type SalesReport struct {
	FromDate           time.Time           `json:"from_date"`
	ToDate             time.Time           `json:"to_date"`
	TotalSales         int                 `json:"total_sales"`
	TotalRevenue       decimal.Decimal     `json:"total_revenue"`
	TotalProfit        decimal.Decimal     `json:"total_profit"`
	CompletedSales     int                 `json:"completed_sales"`
	CancelledSales     int                 `json:"cancelled_sales"`
	RefundedSales      int                 `json:"refunded_sales"`
	AverageOrderValue  decimal.Decimal     `json:"average_order_value"`
	TotalItemsSold     int                 `json:"total_items_sold"`
	UniqueCustomers    int                 `json:"unique_customers"`
	PaymentMethodStats []PaymentMethodStat `json:"payment_method_stats"`
	DailySales         []DailySalesData    `json:"daily_sales"`
}

// DailySalesReport represents daily sales summary
type DailySalesReport struct {
	Date               time.Time           `json:"date"`
	TotalSales         int                 `json:"total_sales"`
	TotalRevenue       decimal.Decimal     `json:"total_revenue"`
	CompletedSales     int                 `json:"completed_sales"`
	CancelledSales     int                 `json:"cancelled_sales"`
	RefundedSales      int                 `json:"refunded_sales"`
	AverageOrderValue  decimal.Decimal     `json:"average_order_value"`
	TotalItemsSold     int                 `json:"total_items_sold"`
	TopSellingProducts []ProductSalesStats `json:"top_selling_products"`
}

// PaymentMethodStat represents payment method statistics
type PaymentMethodStat struct {
	PaymentMethod entities.PaymentMethod `json:"payment_method"`
	Count         int                    `json:"count"`
	TotalAmount   decimal.Decimal        `json:"total_amount"`
	Percentage    decimal.Decimal        `json:"percentage"`
}

// DailySalesData represents daily sales data point
type DailySalesData struct {
	Date         time.Time       `json:"date"`
	TotalSales   int             `json:"total_sales"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
}

// ProductSalesStats represents product sales statistics
type ProductSalesStats struct {
	ProductID    uuid.UUID       `json:"product_id"`
	ProductSKU   string          `json:"product_sku"`
	ProductName  string          `json:"product_name"`
	QuantitySold int             `json:"quantity_sold"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	AveragePrice decimal.Decimal `json:"average_price"`
	SalesCount   int             `json:"sales_count"`
}
