package validation

import (
	"fmt"
	"time"

	currencyValidator "github.com/bojanz/currency"
	"github.com/google/uuid"
)

// Common validation errors
var (
	ErrEmptyField      = fmt.Errorf("field cannot be empty")
	ErrInvalidAmount   = fmt.Errorf("amount must be greater than 0")
	ErrInvalidCurrency = fmt.Errorf("invalid currency code")
	ErrInvalidUUID     = fmt.Errorf("invalid UUID")
	ErrInvalidDate     = fmt.Errorf("invalid date format")
	ErrDateRange       = fmt.Errorf("end date must be after start date")
	ErrInvalidLimit    = fmt.Errorf("limit must be between 1 and 1000")
	ErrDateRangeYear   = fmt.Errorf("date range must not exceed 1 year")
)

// MoneyValidator validates amount and currency
type MoneyValidator struct {
	Amount   float64
	Currency string
}

const (
	CategoryNameMaxLength = 255
)

func (m *MoneyValidator) Validate() error {
	if m.Amount <= 0 {
		return ErrInvalidAmount
	}
	if m.Currency == "" || !currencyValidator.IsValid(m.Currency) {
		return ErrInvalidCurrency
	}
	return nil
}

// DateRangeValidator validates date ranges
type DateRangeValidator struct {
	StartDate time.Time
	EndDate   time.Time
}

func (d *DateRangeValidator) Validate() error {
	if d.StartDate.IsZero() || d.EndDate.IsZero() {
		return ErrInvalidDate
	}
	if d.EndDate.Before(d.StartDate) {
		return ErrDateRange
	}
	// Check if the date range is not more than 1 year
	if d.EndDate.Sub(d.StartDate) > 365*24*time.Hour {
		return ErrDateRangeYear

	}
	return nil
}

// PaginationValidator validates limit and offset parameters
type PaginationValidator struct {
	Limit  int32
	Offset int32
}

func (p *PaginationValidator) Validate() error {
	if p.Limit <= 0 || p.Limit > 1000 {
		return ErrInvalidLimit
	}
	return nil
}

// TextValidator validates text fields with length constraints
type TextValidator struct {
	Text     string
	MinLen   int
	MaxLen   int
	Required bool
}

func (t *TextValidator) Validate() error {
	if t.Required && t.Text == "" {
		return ErrEmptyField
	}
	if t.Text != "" {
		if len(t.Text) < t.MinLen {
			return fmt.Errorf("text length must be at least %d characters", t.MinLen)
		}
		if len(t.Text) > t.MaxLen {
			return fmt.Errorf("text length must not exceed %d characters", t.MaxLen)
		}
	}
	return nil
}

// Common validation functions
func ValidateUUID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return uid, nil
}

func ValidateDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, ErrEmptyField
	}
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return date, nil
}

// ValidateRole validates if a role ID is valid
func ValidateRole(roleID string) (uuid.UUID, error) {
	if roleID == "" {
		return uuid.Nil, fmt.Errorf("role ID is required")
	}
	return ValidateUUID(roleID)
}
