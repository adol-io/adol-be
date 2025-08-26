package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Domain errors
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeConflict      ErrorType = "CONFLICT"
	ErrorTypeUnauthorized  ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden     ErrorType = "FORBIDDEN"
	ErrorTypeInternal      ErrorType = "INTERNAL_ERROR"
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
	ErrorTypeTimeout       ErrorType = "TIMEOUT"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT"
	
	// Business logic errors
	ErrorTypeInsufficientStock ErrorType = "INSUFFICIENT_STOCK"
	ErrorTypeInvalidPrice      ErrorType = "INVALID_PRICE"
	ErrorTypeInvalidQuantity   ErrorType = "INVALID_QUANTITY"
	ErrorTypeProductNotActive  ErrorType = "PRODUCT_NOT_ACTIVE"
	ErrorTypeUserNotActive     ErrorType = "USER_NOT_ACTIVE"
)

// AppError represents an application error
type AppError struct {
	Type     ErrorType `json:"type"`
	Message  string    `json:"message"`
	Details  string    `json:"details,omitempty"`
	Code     int       `json:"code"`
	Internal error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, message string, internal error) *AppError {
	return &AppError{
		Type:     errorType,
		Message:  message,
		Code:     getHTTPStatusCode(errorType),
		Internal: internal,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Details: details,
		Code:    http.StatusBadRequest,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    http.StatusNotFound,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
		Code:    http.StatusConflict,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string, internal error) *AppError {
	return &AppError{
		Type:     ErrorTypeInternal,
		Message:  message,
		Code:     http.StatusInternalServerError,
		Internal: internal,
	}
}

// NewInsufficientStockError creates an insufficient stock error
func NewInsufficientStockError(productName string, available, requested int) *AppError {
	return &AppError{
		Type:    ErrorTypeInsufficientStock,
		Message: fmt.Sprintf("Insufficient stock for product %s", productName),
		Details: fmt.Sprintf("Available: %d, Requested: %d", available, requested),
		Code:    http.StatusBadRequest,
	}
}

// NewInvalidPriceError creates an invalid price error
func NewInvalidPriceError(price float64) *AppError {
	return &AppError{
		Type:    ErrorTypeInvalidPrice,
		Message: "Invalid price",
		Details: fmt.Sprintf("Price must be greater than 0, got: %.2f", price),
		Code:    http.StatusBadRequest,
	}
}

// NewInvalidQuantityError creates an invalid quantity error
func NewInvalidQuantityError(quantity int) *AppError {
	return &AppError{
		Type:    ErrorTypeInvalidQuantity,
		Message: "Invalid quantity",
		Details: fmt.Sprintf("Quantity must be greater than 0, got: %d", quantity),
		Code:    http.StatusBadRequest,
	}
}

// getHTTPStatusCode returns the appropriate HTTP status code for an error type
func getHTTPStatusCode(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeInvalidPrice, ErrorTypeInvalidQuantity, ErrorTypeInsufficientStock:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden, ErrorTypeUserNotActive, ErrorTypeProductNotActive:
		return http.StatusForbidden
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}