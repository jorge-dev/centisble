package validation

import (
	"fmt"
	"strings"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
)

// ExpenseValidation validates expense-related requests
type ExpenseValidation struct {
	Amount      float64
	Currency    string
	CategoryID  uuid.UUID
	Description string
	Date        string
}

func (v *ExpenseValidation) Validate() error {
	// Validate money
	if err := (&MoneyValidator{Amount: v.Amount, Currency: v.Currency}).Validate(); err != nil {
		return err
	}

	// Validate category
	if v.CategoryID == uuid.Nil {
		return ErrInvalidUUID
	}

	// Validate description
	if err := (&TextValidator{
		Text:     v.Description,
		MinLen:   1,
		MaxLen:   1000,
		Required: true,
	}).Validate(); err != nil {
		return err
	}

	// Validate date
	if _, err := ValidateDate(v.Date); err != nil {
		return err
	}

	return nil
}

// CategoryValidation validates category-related requests
type CategoryValidation struct {
	Name string
}

func (v *CategoryValidation) Validate() error {
	return (&TextValidator{
		Text:     v.Name,
		MinLen:   1,
		MaxLen:   255,
		Required: true,
	}).Validate()
}

// BudgetValidation validates budget-related requests
type BudgetValidation struct {
	Amount     float64
	Currency   string
	CategoryID uuid.UUID
	Type       string
	StartDate  string
	EndDate    string
}

func (v *BudgetValidation) Validate() error {
	// Validate money
	if err := (&MoneyValidator{Amount: v.Amount, Currency: v.Currency}).Validate(); err != nil {
		return err
	}

	// Validate category
	if v.CategoryID == uuid.Nil {
		return ErrInvalidUUID
	}

	// Validate type
	if v.Type != "recurring" && v.Type != "one-time" {
		return fmt.Errorf("type must be either 'recurring' or 'one-time'")
	}

	// Validate dates
	start, err := ValidateDate(v.StartDate)
	if err != nil {
		return err
	}
	end, err := ValidateDate(v.EndDate)
	if err != nil {
		return err
	}

	return (&DateRangeValidator{
		StartDate: start,
		EndDate:   end,
	}).Validate()
}

// AuthValidation validates authentication-related requests
type AuthValidation struct {
	Name     string
	Email    string
	Password string
}

func (v *AuthValidation) Validate() error {
	// Validate name (only for registration)
	if v.Name != "" {
		if err := (&TextValidator{
			Text:     v.Name,
			MinLen:   2,
			MaxLen:   100,
			Required: true,
		}).Validate(); err != nil {
			return err
		}
	}

	// Validate email
	if err := (&TextValidator{
		Text:     v.Email,
		MinLen:   5,
		MaxLen:   255,
		Required: true,
	}).Validate(); err != nil {
		return err
	}

	// Basic email format validation
	if !strings.Contains(v.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	return (&TextValidator{
		Text:     v.Password,
		MinLen:   6,
		MaxLen:   100,
		Required: true,
	}).Validate()
}

// UserUpdateValidation validates user update requests
type UserUpdateValidation struct {
	Name            string
	Email           string
	CurrentPassword string
	NewPassword     string
}

func (v *UserUpdateValidation) Validate() error {
	if v.Name != "" {
		if err := (&TextValidator{
			Text:     v.Name,
			MinLen:   2,
			MaxLen:   100,
			Required: false,
		}).Validate(); err != nil {
			return err
		}
	}

	if v.Email != "" {
		if err := (&TextValidator{
			Text:     v.Email,
			MinLen:   5,
			MaxLen:   255,
			Required: false,
		}).Validate(); err != nil {
			return err
		}
		if !strings.Contains(v.Email, "@") {
			return fmt.Errorf("invalid email format")
		}
	}

	if v.NewPassword != "" {
		if v.CurrentPassword == "" {
			return fmt.Errorf("current password is required when setting new password")
		}
		if err := (&TextValidator{
			Text:     v.NewPassword,
			MinLen:   6,
			MaxLen:   100,
			Required: true,
		}).Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IncomeValidation validates income-related requests
type IncomeValidation struct {
	Amount      float64
	Currency    string
	Source      string
	Date        string
	Description string
}

func (v *IncomeValidation) Validate() error {
	// Validate money
	if err := (&MoneyValidator{Amount: v.Amount, Currency: v.Currency}).Validate(); err != nil {
		return err
	}

	// Validate source
	if err := (&TextValidator{
		Text:     v.Source,
		MinLen:   1,
		MaxLen:   255,
		Required: true,
	}).Validate(); err != nil {
		return err
	}

	// Validate description (optional)
	if v.Description != "" {
		if err := (&TextValidator{
			Text:     v.Description,
			MinLen:   0,
			MaxLen:   1000,
			Required: false,
		}).Validate(); err != nil {
			return err
		}
	}

	// Validate date
	if _, err := ValidateDate(v.Date); err != nil {
		return err
	}

	return nil
}

// SummaryValidation validates summary-related requests
type SummaryValidation struct {
	Date     string
	Currency string
}

func (v *SummaryValidation) Validate() error {
	// Validate date
	if _, err := ValidateDate(v.Date); err != nil {
		return err
	}

	// Validate currency if provided
	if v.Currency != "" && !currency.IsValid(v.Currency) {
		return ErrInvalidCurrency
	}

	return nil
}

// DateRangeQueryValidation validates date range query parameters
type DateRangeQueryValidation struct {
	StartDate string
	EndDate   string
	Limit     int32
}

func (v *DateRangeQueryValidation) Validate() error {
	// Validate dates
	start, err := ValidateDate(v.StartDate)
	if err != nil {
		return err
	}
	end, err := ValidateDate(v.EndDate)
	if err != nil {
		return err
	}

	// Validate date range
	if err := (&DateRangeValidator{
		StartDate: start,
		EndDate:   end,
	}).Validate(); err != nil {
		return err
	}

	// Validate limit if provided
	if v.Limit != 0 {
		if err := (&PaginationValidator{Limit: v.Limit}).Validate(); err != nil {
			return err
		}
	}

	return nil
}
