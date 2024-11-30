package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Update Validator interface to use pointer receiver
type Validator interface {
	Validate() error
}

// Update helper function to properly handle the type constraints
func runValidationTest[T any, PT interface {
	*T
	Validator
}](t *testing.T, tests []TestCase) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			v := tt.Input.(T)
			err := PT(&v).Validate()
			if tt.WantErr {
				assert.Error(t, err)
				if tt.ExpectedErr != nil {
					assert.Equal(t, tt.ExpectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpenseValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid expense",
			Input: ExpenseValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				CategoryID:  validUUID,
				Description: validDescription,
				Date:        validDate,
			},
			WantErr: false,
		},
		{
			Name: "invalid amount",
			Input: ExpenseValidation{
				Amount:      -100.00,
				Currency:    validCurrency,
				CategoryID:  validUUID,
				Description: validDescription,
				Date:        validDate,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidAmount,
		},
		{
			Name: "invalid currency",
			Input: ExpenseValidation{
				Amount:      100.00,
				Currency:    invalidCurrency,
				CategoryID:  validUUID,
				Description: validDescription,
				Date:        validDate,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidCurrency,
		},
		{
			Name: "invalid date passed",
			Input: ExpenseValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				CategoryID:  validUUID,
				Description: validDescription,
				Date:        invalidDate,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidDate,
		},
		{
			Name: "empty description",
			Input: ExpenseValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				CategoryID:  validUUID,
				Description: "",
				Date:        validDate,
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			v := tt.Input.(ExpenseValidation)
			err := (&v).Validate()
			if tt.WantErr {
				assert.Error(t, err)
				if tt.ExpectedErr != nil {
					assert.Equal(t, tt.ExpectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Update test function calls to include pointer type
func TestIncomeValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid income",
			Input: IncomeValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				Source:      "Salary",
				Description: validDescription,
				Date:        validDate,
			},
			WantErr: false,
		},
		{
			Name: "invalid amount",
			Input: IncomeValidation{
				Amount:      -100.00,
				Currency:    validCurrency,
				Source:      "Salary",
				Description: validDescription,
				Date:        validDate,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidAmount,
		},
		{
			Name: "empty source",
			Input: IncomeValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				Source:      "",
				Description: validDescription,
				Date:        validDate,
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "invalid date",
			Input: IncomeValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				Source:      "Salary",
				Description: validDescription,
				Date:        invalidDate,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidDate,
		},
	}

	runValidationTest[IncomeValidation](t, tests)
}

// Update test function calls to include pointer type
func TestSummaryValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid summary",
			Input: SummaryValidation{
				Date:     validDate,
				Currency: validCurrency,
			},
			WantErr: false,
		},
		{
			Name: "invalid date",
			Input: SummaryValidation{
				Date:     invalidDate,
				Currency: validCurrency,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidDate,
		},
		{
			Name: "invalid currency",
			Input: SummaryValidation{
				Date:     validDate,
				Currency: invalidCurrency,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidCurrency,
		},
	}

	runValidationTest[SummaryValidation](t, tests)
}

// Update test function calls to include pointer type
func TestDateRangeQueryValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid query",
			Input: DateRangeQueryValidation{
				StartDate: validDate,
				EndDate:   "2023-12-31T00:00:00Z",
				Limit:     100,
			},
			WantErr: false,
		},
		{
			Name: "invalid start date",
			Input: DateRangeQueryValidation{
				StartDate: invalidDate,
				EndDate:   validDate,
				Limit:     100,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidDate,
		},
		{
			Name: "invalid limit",
			Input: DateRangeQueryValidation{
				StartDate: validDate,
				EndDate:   "2023-12-31T00:00:00Z",
				Limit:     1001,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidLimit,
		},
	}

	runValidationTest[DateRangeQueryValidation](t, tests)
}

// Update test function calls to include pointer type
func TestCategoryValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid category",
			Input: CategoryValidation{
				Name: "Test Category",
			},
			WantErr: false,
		},
		{
			Name: "empty name",
			Input: CategoryValidation{
				Name: "",
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "too long name",
			Input: CategoryValidation{
				Name: longDescription, // Using test fixture
			},
			WantErr: true,
		},
	}

	runValidationTest[CategoryValidation](t, tests)
}

// Update test function calls to include pointer type
func TestUserUpdateValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid update with all fields",
			Input: UserUpdateValidation{
				Name:            "New Name",
				Email:           validEmail,
				CurrentPassword: validPassword,
				NewPassword:     "newpassword123",
			},
			WantErr: false,
		},
		{
			Name: "valid update with only name",
			Input: UserUpdateValidation{
				Name: "New Name",
			},
			WantErr: false,
		},
		{
			Name: "invalid email",
			Input: UserUpdateValidation{
				Email: invalidEmail,
			},
			WantErr: true,
		},
		{
			Name: "missing current password",
			Input: UserUpdateValidation{
				NewPassword: validPassword,
			},
			WantErr: true,
		},
		{
			Name: "short new password",
			Input: UserUpdateValidation{
				CurrentPassword: validPassword,
				NewPassword:     shortPassword,
			},
			WantErr: true,
		},
	}

	runValidationTest[UserUpdateValidation](t, tests)
}
