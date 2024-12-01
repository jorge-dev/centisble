package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Update Validator interface to use pointer receiver
type Validator interface {
	Validate() error
}

// Update helper function to properly handle the type constraints
func runValidationTest[TypeToValidate any, ValidatorPointerType interface {
	*TypeToValidate
	Validator
}](t *testing.T, testCases []TestCase) {
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			valueToValidate := testCase.Input.(TypeToValidate)
			validationErr := ValidatorPointerType(&valueToValidate).Validate()

			if testCase.WantErr {
				assert.Error(t, validationErr)
				if testCase.ExpectedErr != nil {
					assert.Equal(t, testCase.ExpectedErr, validationErr)
				}
			} else {
				assert.NoError(t, validationErr)
			}
		})
	}
}

func TestExpenseValidationValidate(t *testing.T) {
	testCases := []TestCase{
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

	runValidationTest[ExpenseValidation](t, testCases)
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

// Add AuthValidation test
func TestAuthValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid registration",
			Input: AuthValidation{
				Name:     validName,
				Email:    validEmail,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr: false,
		},
		{
			Name: "valid login",
			Input: AuthValidation{
				Email:    validEmail,
				Password: validPassword,
				IsLogin:  true,
			},
			WantErr: false,
		},
		{
			Name: "missing name in registration",
			Input: AuthValidation{
				Email:    validEmail,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "invalid email format",
			Input: AuthValidation{
				Name:     validName,
				Email:    invalidEmail,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr: true,
		},
		{
			Name: "empty email",
			Input: AuthValidation{
				Name:     validName,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "empty password",
			Input: AuthValidation{
				Name:    validName,
				Email:   validEmail,
				IsLogin: false,
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "short password",
			Input: AuthValidation{
				Name:     validName,
				Email:    validEmail,
				Password: shortPassword,
				IsLogin:  false,
			},
			WantErr: true,
		},
		{
			Name: "name too short (registration)",
			Input: AuthValidation{
				Name:     "A",
				Email:    validEmail,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr: true,
		},
		{
			Name: "name too long (registration)",
			Input: AuthValidation{
				Name:     strings.Repeat("a", 101),
				Email:    validEmail,
				Password: validPassword,
				IsLogin:  false,
			},
			WantErr: true,
		},
	}

	runValidationTest[AuthValidation](t, tests)
}

// test for auth validation
