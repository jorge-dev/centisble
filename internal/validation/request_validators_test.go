package validation

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

// Add at the top of file with other test fixtures:
var (
	// ...existing fixtures...
	validDateUTC   = time.Now().UTC().Format(time.RFC3339)
	futureDateUTC  = time.Now().UTC().AddDate(0, 1, 0).Format(time.RFC3339)
	invalidDateUTC = "2023-13-32T25:61:61Z" // intentionally invalid
)

func TestExpenseValidationValidate(t *testing.T) {
	testCases := []TestCase{
		{
			Name: "valid expense",
			Input: ExpenseValidation{
				Amount:      100.00,
				Currency:    validCurrency,
				CategoryID:  validUUID,
				Description: validDescription,
				Date:        validDateUTC,
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
				StartDate: validDateUTC,
				EndDate:   futureDateUTC,
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

func TestBudgetValidationValidate(t *testing.T) {
	tests := []TestCase{
		{
			Name: "valid full budget",
			Input: BudgetValidation{
				Amount:     1000.00,
				Currency:   validCurrency,
				CategoryID: validUUID,
				Type:       "recurring",
				StartDate:  validDate,
				EndDate:    "2023-12-31T00:00:00Z",
				Name:       "Monthly Groceries",
			},
			WantErr: false,
		},
		{
			Name: "missing name in new budget",
			Input: BudgetValidation{
				Amount:     1000.00,
				Currency:   validCurrency,
				CategoryID: validUUID,
				Type:       "recurring",
				StartDate:  validDate,
				EndDate:    "2024-12-31T00:00:00Z",
			},
			WantErr:     true,
			ExpectedErr: ErrEmptyField,
		},
		{
			Name: "name too long",
			Input: BudgetValidation{
				Amount:     1000.00,
				Currency:   validCurrency,
				CategoryID: validUUID,
				Type:       "recurring",
				StartDate:  validDate,
				EndDate:    "2024-12-31T00:00:00Z",
				Name:       strings.Repeat("a", 256),
			},
			WantErr: true,
		},
		{
			Name: "valid partial update - amount only",
			Input: BudgetValidation{
				Amount:          2000.00,
				IsPartialUpdate: true,
			},
			WantErr: false,
		},
		{
			Name: "valid partial update - currency only",
			Input: BudgetValidation{
				Currency:        validCurrency,
				IsPartialUpdate: true,
			},
			WantErr: false,
		},
		{
			Name: "valid partial update - type only",
			Input: BudgetValidation{
				Type:            "one-time",
				IsPartialUpdate: true,
			},
			WantErr: false,
		},
		{
			Name: "invalid amount in partial update",
			Input: BudgetValidation{
				Amount:          -100.00,
				IsPartialUpdate: true,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidAmount,
		},
		{
			Name: "invalid currency in partial update",
			Input: BudgetValidation{
				Currency:        invalidCurrency,
				IsPartialUpdate: true,
			},
			WantErr:     true,
			ExpectedErr: ErrInvalidCurrency,
		},
		{
			Name: "invalid type in partial update",
			Input: BudgetValidation{
				Type:            "invalid-type",
				IsPartialUpdate: true,
			},
			WantErr: true,
		},
		{
			Name: "invalid full budget - missing required fields",
			Input: BudgetValidation{
				Amount:   1000.00,
				Currency: validCurrency,
				// Missing other required fields
			},
			WantErr: true,
		},
		{
			Name: "invalid date range",
			Input: BudgetValidation{
				Amount:     1000.00,
				Currency:   validCurrency,
				CategoryID: validUUID,
				Type:       "recurring",
				StartDate:  "2024-12-31T00:00:00Z",
				EndDate:    validDate, // End date before start date
				Name:       "Monthly Groceries",
			},
			WantErr:     true,
			ExpectedErr: ErrDateRange,
		},
	}

	runValidationTest[BudgetValidation](t, tests)
}

func TestBudgetValidationValidatePartialUpdate(t *testing.T) {
	now := time.Now().UTC()
	future := now.AddDate(0, 1, 0)

	current := CurrentBudget{
		Amount:     1000.00,
		Currency:   "USD",
		CategoryID: validUUID,
		Type:       "recurring",
		StartDate:  now,
		EndDate:    future,
		Name:       "Original Budget",
	}

	tests := []struct {
		name    string
		input   BudgetValidation
		want    CurrentBudget
		wantErr bool
	}{
		{
			name: "valid amount update",
			input: BudgetValidation{
				Amount:          2000.00,
				IsPartialUpdate: true,
			},
			want: CurrentBudget{
				Amount:     2000.00,
				Currency:   current.Currency,
				CategoryID: current.CategoryID,
				Type:       current.Type,
				StartDate:  current.StartDate,
				EndDate:    current.EndDate,
				Name:       current.Name,
			},
			wantErr: false,
		},
		{
			name: "valid multiple field update",
			input: BudgetValidation{
				Amount:          2000.00,
				Currency:        "EUR",
				Type:            "one-time",
				IsPartialUpdate: true,
			},
			want: CurrentBudget{
				Amount:     2000.00,
				Currency:   "EUR",
				CategoryID: current.CategoryID,
				Type:       "one-time",
				StartDate:  current.StartDate,
				EndDate:    current.EndDate,
				Name:       current.Name,
			},
			wantErr: false,
		},
		{
			name: "invalid date range update",
			input: BudgetValidation{
				StartDate:       now.AddDate(0, 2, 0).Format(time.RFC3339),
				EndDate:         now.Format(time.RFC3339), // End before start
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "invalid date format",
			input: BudgetValidation{
				StartDate:       "invalid-date",
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "valid name update",
			input: BudgetValidation{
				Name:            "Updated Budget Name",
				IsPartialUpdate: true,
			},
			want: CurrentBudget{
				Amount:     current.Amount,
				Currency:   current.Currency,
				CategoryID: current.CategoryID,
				Type:       current.Type,
				StartDate:  current.StartDate,
				EndDate:    current.EndDate,
				Name:       "Updated Budget Name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.ValidatePartialUpdate(current)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ...existing code...

func TestExpenseValidation_ValidatePartialUpdate(t *testing.T) {
	now := time.Now().UTC()
	testCategoryId := uuid.New()
	testUpdateDate := now.Add(24 * time.Hour)
	current := CurrentExpense{
		Amount:      100.00,
		Currency:    "USD",
		CategoryID:  uuid.New(),
		Date:        now,
		Description: "Original expense",
	}

	tests := []struct {
		name    string
		update  ExpenseValidation
		want    CurrentExpense
		wantErr bool
	}{
		{
			name: "valid amount update",
			update: ExpenseValidation{
				Amount:          200.00,
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      200.00,
				Currency:    current.Currency,
				CategoryID:  current.CategoryID,
				Date:        current.Date,
				Description: current.Description,
			},
			wantErr: false,
		},
		{
			name: "invalid amount update",
			update: ExpenseValidation{
				Amount:          -50.00,
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "valid currency update",
			update: ExpenseValidation{
				Currency:        "EUR",
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      current.Amount,
				Currency:    "EUR",
				CategoryID:  current.CategoryID,
				Date:        current.Date,
				Description: current.Description,
			},
			wantErr: false,
		},
		{
			name: "invalid currency update",
			update: ExpenseValidation{
				Currency:        "XXX",
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "valid category update",
			update: ExpenseValidation{
				CategoryID:      testCategoryId,
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      current.Amount,
				Currency:    current.Currency,
				CategoryID:  testCategoryId,
				Date:        current.Date,
				Description: current.Description,
			},
			wantErr: false,
		},
		{
			name: "valid date update",
			update: ExpenseValidation{
				Date:            testUpdateDate.Format(time.RFC3339),
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      current.Amount,
				Currency:    current.Currency,
				CategoryID:  current.CategoryID,
				Date:        testUpdateDate.Truncate(time.Second),
				Description: current.Description,
			},
			wantErr: false,
		},
		{
			name: "invalid date update",
			update: ExpenseValidation{
				Date:            "invalid-date",
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "valid description update",
			update: ExpenseValidation{
				Description:     "Updated description",
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      current.Amount,
				Currency:    current.Currency,
				CategoryID:  current.CategoryID,
				Date:        current.Date,
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name: "invalid description update (too long)",
			update: ExpenseValidation{
				Description:     strings.Repeat("a", 1001),
				IsPartialUpdate: true,
			},
			wantErr: true,
		},
		{
			name: "multiple field update",
			update: ExpenseValidation{
				Amount:          150.00,
				Currency:        "EUR",
				Description:     "Multiple update",
				IsPartialUpdate: true,
			},
			want: CurrentExpense{
				Amount:      150.00,
				Currency:    "EUR",
				CategoryID:  current.CategoryID,
				Date:        current.Date,
				Description: "Multiple update",
			},
			wantErr: false,
		},
		{
			name: "no changes",
			update: ExpenseValidation{
				IsPartialUpdate: true,
			},
			want:    current,
			wantErr: false,
		},
		{
			name: "not partial update",
			update: ExpenseValidation{
				Amount:          200.00,
				IsPartialUpdate: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.update.ValidatePartialUpdate(current)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
