package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nicklaros/adol/pkg/errors"
)

// SaleStatus represents sale status
type SaleStatus string

const (
	SaleStatusPending   SaleStatus = "pending"
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusCancelled SaleStatus = "cancelled"
	SaleStatusRefunded  SaleStatus = "refunded"
)

// PaymentMethod represents payment method
type PaymentMethod string

const (
	PaymentMethodCash          PaymentMethod = "cash"
	PaymentMethodCard          PaymentMethod = "card"
	PaymentMethodDigitalWallet PaymentMethod = "digital_wallet"
	PaymentMethodBankTransfer  PaymentMethod = "bank_transfer"
)

// Sale represents a sales transaction
type Sale struct {
	ID             uuid.UUID       `json:"id"`
	SaleNumber     string          `json:"sale_number"`
	CustomerName   string          `json:"customer_name,omitempty"`
	CustomerEmail  string          `json:"customer_email,omitempty"`
	CustomerPhone  string          `json:"customer_phone,omitempty"`
	Items          []SaleItem      `json:"items"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	PaidAmount     decimal.Decimal `json:"paid_amount"`
	ChangeAmount   decimal.Decimal `json:"change_amount"`
	PaymentMethod  PaymentMethod   `json:"payment_method"`
	Status         SaleStatus      `json:"status"`
	Notes          string          `json:"notes,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      uuid.UUID       `json:"created_by"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

// SaleItem represents an item in a sale
type SaleItem struct {
	ID          uuid.UUID       `json:"id"`
	SaleID      uuid.UUID       `json:"sale_id"`
	ProductID   uuid.UUID       `json:"product_id"`
	ProductSKU  string          `json:"product_sku"`
	ProductName string          `json:"product_name"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price"`
	CreatedAt   time.Time       `json:"created_at"`
}

// NewSale creates a new sale
func NewSale(saleNumber, customerName, customerEmail, customerPhone string, createdBy uuid.UUID) (*Sale, error) {
	if saleNumber == "" {
		return nil, errors.NewValidationError("sale number is required", "sale_number cannot be empty")
	}

	now := time.Now()
	sale := &Sale{
		ID:             uuid.New(),
		SaleNumber:     saleNumber,
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		CustomerPhone:  customerPhone,
		Items:          make([]SaleItem, 0),
		Subtotal:       decimal.Zero,
		TaxAmount:      decimal.Zero,
		DiscountAmount: decimal.Zero,
		TotalAmount:    decimal.Zero,
		PaidAmount:     decimal.Zero,
		ChangeAmount:   decimal.Zero,
		Status:         SaleStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      createdBy,
	}

	return sale, nil
}

// NewSaleItem creates a new sale item
func NewSaleItem(saleID, productID uuid.UUID, productSKU, productName string, quantity int, unitPrice decimal.Decimal) (*SaleItem, error) {
	if quantity <= 0 {
		return nil, errors.NewInvalidQuantityError(quantity)
	}
	if unitPrice.LessThanOrEqual(decimal.Zero) {
		return nil, errors.NewInvalidPriceError(unitPrice.InexactFloat64())
	}
	if productSKU == "" {
		return nil, errors.NewValidationError("product SKU is required", "product_sku cannot be empty")
	}
	if productName == "" {
		return nil, errors.NewValidationError("product name is required", "product_name cannot be empty")
	}

	totalPrice := unitPrice.Mul(decimal.NewFromInt(int64(quantity)))

	item := &SaleItem{
		ID:          uuid.New(),
		SaleID:      saleID,
		ProductID:   productID,
		ProductSKU:  productSKU,
		ProductName: productName,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  totalPrice,
		CreatedAt:   time.Now(),
	}

	return item, nil
}

// AddItem adds an item to the sale
func (s *Sale) AddItem(item *SaleItem) error {
	if item == nil {
		return errors.NewValidationError("item is required", "sale item cannot be nil")
	}

	// Check if item with same product already exists
	for i, existingItem := range s.Items {
		if existingItem.ProductID == item.ProductID {
			// Update existing item
			s.Items[i].Quantity += item.Quantity
			s.Items[i].TotalPrice = s.Items[i].UnitPrice.Mul(decimal.NewFromInt(int64(s.Items[i].Quantity)))
			s.Items[i].CreatedAt = time.Now()
			s.UpdatedAt = time.Now()
			s.recalculateAmounts()
			return nil
		}
	}

	// Add new item
	s.Items = append(s.Items, *item)
	s.UpdatedAt = time.Now()
	s.recalculateAmounts()
	return nil
}

// RemoveItem removes an item from the sale
func (s *Sale) RemoveItem(productID uuid.UUID) error {
	for i, item := range s.Items {
		if item.ProductID == productID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			s.UpdatedAt = time.Now()
			s.recalculateAmounts()
			return nil
		}
	}
	return errors.NewNotFoundError("sale item")
}

// UpdateItemQuantity updates the quantity of an item
func (s *Sale) UpdateItemQuantity(productID uuid.UUID, newQuantity int) error {
	if newQuantity <= 0 {
		return errors.NewInvalidQuantityError(newQuantity)
	}

	for i, item := range s.Items {
		if item.ProductID == productID {
			s.Items[i].Quantity = newQuantity
			s.Items[i].TotalPrice = item.UnitPrice.Mul(decimal.NewFromInt(int64(newQuantity)))
			s.UpdatedAt = time.Now()
			s.recalculateAmounts()
			return nil
		}
	}
	return errors.NewNotFoundError("sale item")
}

// ApplyDiscount applies a discount to the sale
func (s *Sale) ApplyDiscount(discountAmount decimal.Decimal) error {
	if discountAmount.LessThan(decimal.Zero) {
		return errors.NewValidationError("invalid discount", "discount amount cannot be negative")
	}
	if discountAmount.GreaterThan(s.Subtotal) {
		return errors.NewValidationError("invalid discount", "discount amount cannot be greater than subtotal")
	}

	s.DiscountAmount = discountAmount
	s.UpdatedAt = time.Now()
	s.recalculateAmounts()
	return nil
}

// ApplyTax applies tax to the sale
func (s *Sale) ApplyTax(taxPercentage decimal.Decimal) error {
	if taxPercentage.LessThan(decimal.Zero) {
		return errors.NewValidationError("invalid tax", "tax percentage cannot be negative")
	}

	taxableAmount := s.Subtotal.Sub(s.DiscountAmount)
	s.TaxAmount = taxableAmount.Mul(taxPercentage).Div(decimal.NewFromInt(100))
	s.UpdatedAt = time.Now()
	s.recalculateAmounts()
	return nil
}

// ProcessPayment processes payment for the sale
func (s *Sale) ProcessPayment(paidAmount decimal.Decimal, paymentMethod PaymentMethod) error {
	if err := ValidatePaymentMethod(paymentMethod); err != nil {
		return err
	}

	if paidAmount.LessThan(s.TotalAmount) {
		return errors.NewValidationError("insufficient payment", "paid amount is less than total amount")
	}

	s.PaidAmount = paidAmount
	s.ChangeAmount = paidAmount.Sub(s.TotalAmount)
	s.PaymentMethod = paymentMethod
	s.UpdatedAt = time.Now()

	return nil
}

// CompleteSale completes the sale
func (s *Sale) CompleteSale() error {
	if s.Status != SaleStatusPending {
		return errors.NewValidationError("invalid sale status", "only pending sales can be completed")
	}
	if len(s.Items) == 0 {
		return errors.NewValidationError("empty sale", "sale must have at least one item")
	}
	if s.PaidAmount.LessThan(s.TotalAmount) {
		return errors.NewValidationError("incomplete payment", "sale payment is incomplete")
	}

	s.Status = SaleStatusCompleted
	now := time.Now()
	s.CompletedAt = &now
	s.UpdatedAt = now

	return nil
}

// CancelSale cancels the sale
func (s *Sale) CancelSale() error {
	if s.Status == SaleStatusCompleted {
		return errors.NewValidationError("invalid sale status", "completed sales cannot be cancelled")
	}
	if s.Status == SaleStatusCancelled {
		return errors.NewValidationError("invalid sale status", "sale is already cancelled")
	}

	s.Status = SaleStatusCancelled
	s.UpdatedAt = time.Now()

	return nil
}

// RefundSale refunds the sale
func (s *Sale) RefundSale() error {
	if s.Status != SaleStatusCompleted {
		return errors.NewValidationError("invalid sale status", "only completed sales can be refunded")
	}

	s.Status = SaleStatusRefunded
	s.UpdatedAt = time.Now()

	return nil
}

// AddNotes adds notes to the sale
func (s *Sale) AddNotes(notes string) {
	s.Notes = notes
	s.UpdatedAt = time.Now()
}

// GetItemCount returns the total number of items in the sale
func (s *Sale) GetItemCount() int {
	count := 0
	for _, item := range s.Items {
		count += item.Quantity
	}
	return count
}

// IsCompleted checks if the sale is completed
func (s *Sale) IsCompleted() bool {
	return s.Status == SaleStatusCompleted
}

// IsCancelled checks if the sale is cancelled
func (s *Sale) IsCancelled() bool {
	return s.Status == SaleStatusCancelled
}

// IsRefunded checks if the sale is refunded
func (s *Sale) IsRefunded() bool {
	return s.Status == SaleStatusRefunded
}

// recalculateAmounts recalculates subtotal and total amounts
func (s *Sale) recalculateAmounts() {
	s.Subtotal = decimal.Zero
	for _, item := range s.Items {
		s.Subtotal = s.Subtotal.Add(item.TotalPrice)
	}

	s.TotalAmount = s.Subtotal.Sub(s.DiscountAmount).Add(s.TaxAmount)
}

// ValidatePaymentMethod validates payment method
func ValidatePaymentMethod(method PaymentMethod) error {
	switch method {
	case PaymentMethodCash, PaymentMethodCard, PaymentMethodDigitalWallet, PaymentMethodBankTransfer:
		return nil
	default:
		return errors.NewValidationError("invalid payment method", "payment method must be one of: cash, card, digital_wallet, bank_transfer")
	}
}

// ValidateSaleStatus validates sale status
func ValidateSaleStatus(status SaleStatus) error {
	switch status {
	case SaleStatusPending, SaleStatusCompleted, SaleStatusCancelled, SaleStatusRefunded:
		return nil
	default:
		return errors.NewValidationError("invalid sale status", "status must be one of: pending, completed, cancelled, refunded")
	}
}
