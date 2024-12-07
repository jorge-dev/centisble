package validation

import (
	"fmt"
	"log"
	"strings"
	"time"

	currencyValidator "github.com/bojanz/currency"
	"github.com/google/uuid"
)

// ExpenseValidation validates expense-related requests
type ExpenseValidation struct {
	Amount          float64
	Currency        string
	CategoryID      uuid.UUID
	Description     string
	Date            string
	IsPartialUpdate bool
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

// Add this new type
type CurrentExpense struct {
	Amount      float64
	Currency    string
	CategoryID  uuid.UUID
	Date        time.Time
	Description string
}

// Add this new method
func (v *ExpenseValidation) ValidatePartialUpdate(current CurrentExpense) (CurrentExpense, error) {
	result := CurrentExpense{
		Amount:      current.Amount,
		Currency:    current.Currency,
		CategoryID:  current.CategoryID,
		Date:        current.Date,
		Description: current.Description,
	}

	if !v.IsPartialUpdate {
		return CurrentExpense{}, fmt.Errorf("not a partial update")
	}

	// Handle individual field updates
	if v.Amount != 0 {
		if v.Amount <= 0 {
			return CurrentExpense{}, ErrInvalidAmount
		}
		result.Amount = v.Amount

	}

	if v.Currency != "" {
		if !currencyValidator.IsValid(v.Currency) {
			return CurrentExpense{}, ErrInvalidCurrency
		}
		result.Currency = v.Currency
	}

	if v.CategoryID != uuid.Nil {
		result.CategoryID = v.CategoryID
	}

	if v.Date != "" {
		date, err := ValidateDate(v.Date)
		if err != nil {
			return CurrentExpense{}, err
		}
		result.Date = date
	}

	if v.Description != "" {
		if err := (&TextValidator{
			Text:     v.Description,
			MinLen:   1,
			MaxLen:   1000,
			Required: true,
		}).Validate(); err != nil {
			return CurrentExpense{}, err
		}
		result.Description = v.Description
	}

	return result, nil
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
	Amount          float64
	Currency        string
	CategoryID      uuid.UUID
	Type            string
	StartDate       string
	EndDate         string
	AlertThreshold  *float64
	IsPartialUpdate bool
	Name            string // Add name field
}

func (v *BudgetValidation) ValidateAlertThreshold() error {
	if v.AlertThreshold != nil {
		if *v.AlertThreshold < 0 || *v.AlertThreshold > 100 {
			return fmt.Errorf("alert threshold must be between 0 and 100")
		}
	}
	return nil
}

func (v *BudgetValidation) Validate() error {
	if !v.IsPartialUpdate {
		// Validate name for new budgets
		if err := (&TextValidator{
			Text:     v.Name,
			MinLen:   1,
			MaxLen:   255,
			Required: true,
		}).Validate(); err != nil {
			return err
		}

		// Full validation for new budgets
		if err := (&MoneyValidator{Amount: v.Amount, Currency: v.Currency}).Validate(); err != nil {
			return err
		}
		if v.CategoryID == uuid.Nil {
			return ErrInvalidUUID
		}
		if v.Type != "recurring" && v.Type != "one-time" {
			return fmt.Errorf("type must be either 'recurring' or 'one-time'")
		}

		start, err := ValidateDate(v.StartDate)
		if err != nil {
			return err
		}

		end, err := ValidateDate(v.EndDate)
		if err != nil {
			return err
		}

		if err := (&DateRangeValidator{
			StartDate: start,
			EndDate:   end,
		}).Validate(); err != nil {
			return err
		}
	} else {
		// Validate name if provided in partial update
		if v.Name != "" {
			if err := (&TextValidator{
				Text:     v.Name,
				MinLen:   1,
				MaxLen:   255,
				Required: false,
			}).Validate(); err != nil {
				return err
			}
		}

		// Partial update validation - validate amount and currency separately
		if v.Amount != 0 {
			if v.Amount <= 0 {
				return ErrInvalidAmount
			}
		}
		if v.Currency != "" {
			if !currencyValidator.IsValid(v.Currency) {
				return ErrInvalidCurrency
			}
		}
		if v.CategoryID != uuid.Nil {
			// Validate category ID if provided
			if _, err := ValidateUUID(v.CategoryID.String()); err != nil {
				return err
			}
		}
		// Validate type if provided
		if v.Type != "" && v.Type != "recurring" && v.Type != "one-time" {
			return fmt.Errorf("type must be either 'recurring' or 'one-time'")
		}

		var start time.Time
		var end time.Time

		if v.StartDate != "" {
			startDate, err := ValidateDate(v.StartDate)
			if err != nil {
				return err
			}
			start = startDate
		}

		if v.EndDate != "" {
			endDate, err := ValidateDate(v.EndDate)
			if err != nil {
				return err
			}
			end = endDate
		}
		log.Println("start date", start)

		if !start.IsZero() && !end.IsZero() {
			if err := (&DateRangeValidator{
				StartDate: start,
				EndDate:   end,
			}).Validate(); err != nil {
				return err
			}
		}

	}

	// Validate alert threshold if present
	if err := v.ValidateAlertThreshold(); err != nil {
		return err
	}

	return nil
}

// Add this new type
type CurrentBudget struct {
	Amount     float64
	Currency   string
	CategoryID uuid.UUID
	Type       string
	StartDate  time.Time
	EndDate    time.Time
	Name       string // Add name field
}

// Add this new method
func (v *BudgetValidation) ValidatePartialUpdate(current CurrentBudget) (CurrentBudget, error) {
	result := CurrentBudget{
		Amount:     current.Amount,
		Currency:   current.Currency,
		CategoryID: current.CategoryID,
		Type:       current.Type,
		StartDate:  current.StartDate,
		EndDate:    current.EndDate,
		Name:       current.Name, // Copy existing name
	}

	// Handle date updates, considering both current and new dates
	var newStartDate, newEndDate time.Time
	var err error

	if v.StartDate != "" {
		newStartDate, err = time.Parse(time.RFC3339, v.StartDate)
		if err != nil {
			return CurrentBudget{}, fmt.Errorf("invalid start date format")
		}
		result.StartDate = newStartDate
	} else {
		newStartDate = current.StartDate
	}

	if v.EndDate != "" {
		newEndDate, err = time.Parse(time.RFC3339, v.EndDate)
		if err != nil {
			return CurrentBudget{}, fmt.Errorf("invalid end date format")
		}
		result.EndDate = newEndDate
	} else {
		newEndDate = current.EndDate
	}

	// Validate the date range using both current and new dates
	if err := (&DateRangeValidator{
		StartDate: newStartDate,
		EndDate:   newEndDate,
	}).Validate(); err != nil {
		return CurrentBudget{}, err
	}

	// Handle other fields
	if v.Amount != 0 {
		result.Amount = v.Amount
	}
	if v.Currency != "" {
		result.Currency = v.Currency
	}
	if v.CategoryID != uuid.Nil {
		result.CategoryID = v.CategoryID
	}
	if v.Type != "" {
		result.Type = v.Type
	}

	// Handle name update
	if v.Name != "" {
		result.Name = v.Name
	}

	return result, nil
}

// AuthValidation validates authentication-related requests
type AuthValidation struct {
	Name     string
	Email    string
	Password string
	IsLogin  bool // Add this field to distinguish between login and registration
}

func (v *AuthValidation) Validate() error {
	// Validate name (required for registration, not for login)
	if !v.IsLogin {
		if err := (&TextValidator{
			Text:     v.Name,
			MinLen:   2,
			MaxLen:   100,
			Required: true, // Required for registration
		}).Validate(); err != nil {
			return err
		}
	}

	// Email validation (always required)
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

	// Validate password (always required)
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

type UserProfileValidation struct {
	Name  string
	Email string
}

func (v *UserProfileValidation) Validate() error {
	if v.Name != "" {
		if err := (&TextValidator{
			Text:     v.Name,
			MinLen:   UserNameMinLength,
			MaxLen:   UserNameMaxLength,
			Required: false,
		}).Validate(); err != nil {
			return err
		}
	}

	if v.Email != "" {
		if err := (&TextValidator{
			Text:     v.Email,
			MinLen:   EmailMinLength,
			MaxLen:   EmailMaxLength,
			Required: false,
		}).Validate(); err != nil {
			return err
		}
		if !strings.Contains(v.Email, "@") {
			return fmt.Errorf("invalid email format")
		}
	}

	return nil
}

type PasswordUpdateValidation struct {
	CurrentPassword string
	NewPassword     string
}

func (v *PasswordUpdateValidation) Validate() error {
	if v.CurrentPassword == "" {
		return fmt.Errorf("current password is required")
	}

	if v.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}

	return (&TextValidator{
		Text:     v.NewPassword,
		MinLen:   PasswordMinLength,
		MaxLen:   PasswordMaxLength,
		Required: true,
	}).Validate()
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

type CurrentIncome struct {
	Amount      float64
	Currency    string
	Source      string
	Date        time.Time
	Description string
}

func (v *IncomeValidation) ValidatePartialUpdate(current CurrentIncome) (CurrentIncome, error) {
	result := CurrentIncome{
		Amount:      current.Amount,
		Currency:    current.Currency,
		Source:      current.Source,
		Date:        current.Date,
		Description: current.Description,
	}

	if v.Amount != 0 {
		if v.Amount <= 0 {
			return CurrentIncome{}, ErrInvalidAmount
		}
		result.Amount = v.Amount
	}

	if v.Currency != "" {
		if !currencyValidator.IsValid(v.Currency) {
			return CurrentIncome{}, ErrInvalidCurrency
		}
		result.Currency = v.Currency
	}

	if v.Source != "" {
		if err := (&TextValidator{
			Text:     v.Source,
			MinLen:   1,
			MaxLen:   255,
			Required: true,
		}).Validate(); err != nil {
			return CurrentIncome{}, err
		}
		result.Source = v.Source
	}

	if v.Date != "" {
		date, err := ValidateDate(v.Date)
		if err != nil {
			return CurrentIncome{}, err
		}
		result.Date = date
	}

	if v.Description != "" {
		if err := (&TextValidator{
			Text:     v.Description,
			MinLen:   0,
			MaxLen:   1000,
			Required: false,
		}).Validate(); err != nil {
			return CurrentIncome{}, err
		}
		result.Description = v.Description
	}

	return result, nil
}

// SummaryValidation validates summary-related requests
type SummaryValidation struct {
	Date       string
	Currency   string
	ParsedDate time.Time // Add this field
}

func (v *SummaryValidation) Validate() error {
	// Default to current date if empty
	if v.Date == "" {
		v.ParsedDate = time.Now()
	} else {
		// Parse date
		date, err := ValidateDate(v.Date)
		if err != nil {
			return err
		}
		v.ParsedDate = date
	}

	// Validate currency if provided
	if v.Currency != "" && !currencyValidator.IsValid(v.Currency) {
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
