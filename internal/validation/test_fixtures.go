package validation

import (
	"time"

	"github.com/google/uuid"
)

var (
	validUUID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testTime  = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
)

// Test constants
const (
	validDate   = "2023-01-01T00:00:00Z"
	invalidDate = "2023-13-45"

	validCurrency   = "USD"
	invalidCurrency = "XXX"

	validEmail   = "test@example.com"
	invalidEmail = "invalid-email"

	validPassword = "password123"
	shortPassword = "123"

	validDescription = "Test description"
	//make this be 255 characters long
	longDescription = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et 
	dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea 
	commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 
	Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
	`
)

// Common test cases
type TestCase struct {
	Name        string
	Input       interface{}
	WantErr     bool
	ExpectedErr error
}

// TestFixtureExamples demonstrates usage of test fixtures.
// This ensures all test fixtures are referenced at least once.
var TestFixtureExamples = struct {
	UUID        uuid.UUID
	Time        time.Time
	Date        string
	Currency    string
	Email       string
	Password    string
	Description string
}{
	UUID:        validUUID,
	Time:        testTime,
	Date:        validDate,
	Currency:    validCurrency,
	Email:       validEmail,
	Password:    validPassword,
	Description: validDescription,
}
