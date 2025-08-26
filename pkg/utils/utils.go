package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GenerateInvoiceNumber generates a unique invoice number
func GenerateInvoiceNumber() string {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	
	// Generate random 4-digit number
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(9999))
	
	return fmt.Sprintf("INV-%04d%02d%02d-%04d", year, month, day, randomNum.Int64())
}

// GenerateSaleNumber generates a unique sale number
func GenerateSaleNumber() string {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	
	// Generate random 3-digit number
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999))
	
	return fmt.Sprintf("SALE-%04d%02d%02d-%02d%02d-%03d", year, month, day, hour, minute, randomNum.Int64())
}

// GenerateReceiptNumber generates a unique receipt number
func GenerateReceiptNumber() string {
	now := time.Now()
	timestamp := now.Unix()
	
	// Generate random 3-digit number
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999))
	
	return fmt.Sprintf("RCP-%d-%03d", timestamp, randomNum.Int64())
}

// NormalizeString normalizes a string by trimming whitespace and converting to lowercase
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// IsValidSKU validates if the SKU format is valid
func IsValidSKU(sku string) bool {
	// SKU should be at least 3 characters and contain only alphanumeric characters and hyphens
	if len(sku) < 3 {
		return false
	}
	
	for _, char := range sku {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}
	
	return true
}

// FormatCurrency formats a decimal value as currency
func FormatCurrency(value float64, currency string) string {
	switch currency {
	case "USD":
		return fmt.Sprintf("$%.2f", value)
	case "EUR":
		return fmt.Sprintf("€%.2f", value)
	case "IDR":
		return fmt.Sprintf("Rp %.0f", value)
	default:
		return fmt.Sprintf("%.2f %s", value, currency)
	}
}

// IsBusinessHours checks if the current time is within business hours
func IsBusinessHours(openHour, closeHour int) bool {
	now := time.Now()
	currentHour := now.Hour()
	
	if openHour <= closeHour {
		// Same day business hours (e.g., 9 AM to 5 PM)
		return currentHour >= openHour && currentHour < closeHour
	} else {
		// Overnight business hours (e.g., 10 PM to 6 AM)
		return currentHour >= openHour || currentHour < closeHour
	}
}

// CalculateAge calculates age from birth date
func CalculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	
	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	
	return age
}

// GetStartOfDay returns the start of the day for a given time
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the end of the day for a given time
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek returns the start of the week (Monday) for a given time
func GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysToSubtract := weekday - 1 // Monday is 1
	return GetStartOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// GetEndOfWeek returns the end of the week (Sunday) for a given time
func GetEndOfWeek(t time.Time) time.Time {
	startOfWeek := GetStartOfWeek(t)
	return GetEndOfDay(startOfWeek.AddDate(0, 0, 6))
}

// GetStartOfMonth returns the start of the month for a given time
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth returns the end of the month for a given time
func GetEndOfMonth(t time.Time) time.Time {
	nextMonth := GetStartOfMonth(t).AddDate(0, 1, 0)
	return nextMonth.Add(-time.Nanosecond)
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// CalculatePagination calculates pagination information
func CalculatePagination(page, limit, totalCount int) PaginationInfo {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1
	
	return PaginationInfo{
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// GetOffset calculates the database offset for pagination
func GetOffset(page, limit int) int {
	return (page - 1) * limit
}