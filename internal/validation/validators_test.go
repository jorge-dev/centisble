package validation

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMoneyValidatorValidate(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		wantErr  error
	}{
		{"valid money", 100.00, "USD", nil},
		{"zero amount", 0, "USD", ErrInvalidAmount},
		{"negative amount", -100.00, "USD", ErrInvalidAmount},
		{"invalid currency", 100.00, "XXX", ErrInvalidCurrency},
		{"empty currency", 100.00, "", ErrInvalidCurrency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &MoneyValidator{Amount: tt.amount, Currency: tt.currency}
			err := v.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDateRangeValidatorValidate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		wantErr   error
	}{
		{"valid range", now, now.Add(24 * time.Hour), nil},
		{"zero dates", time.Time{}, time.Time{}, ErrInvalidDate},
		{"end before start", now, now.Add(-24 * time.Hour), ErrDateRange},
		{"same date", now, now, nil},
		{"more than 1 year", now, now.Add(366 * 24 * time.Hour), ErrDateRangeYear},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &DateRangeValidator{StartDate: tt.startDate, EndDate: tt.endDate}
			err := v.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestPaginationValidatorValidate(t *testing.T) {
	tests := []struct {
		name    string
		limit   int32
		offset  int32
		wantErr error
	}{
		{"valid pagination", 10, 0, nil},
		{"zero limit", 0, 0, ErrInvalidLimit},
		{"negative limit", -1, 0, ErrInvalidLimit},
		{"exceed max limit", 1001, 0, ErrInvalidLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &PaginationValidator{Limit: tt.limit, Offset: tt.offset}
			err := v.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestTextValidatorValidate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minLen   int
		maxLen   int
		required bool
		wantErr  bool
	}{
		{"valid text", "test", 1, 10, true, false},
		{"empty required", "", 1, 10, true, true},
		{"empty not required", "", 1, 10, false, false},
		{"too short", "a", 2, 10, true, true},
		{"too long", "toolongtext", 1, 5, true, true},
		{"exact min length", "ab", 2, 10, true, false},
		{"exact max length", "abcde", 1, 5, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &TextValidator{
				Text:     tt.text,
				MinLen:   tt.minLen,
				MaxLen:   tt.maxLen,
				Required: tt.required,
			}
			err := v.Validate()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestValidateUUID(t *testing.T) {
	validUUID := uuid.New()
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid UUID", validUUID.String(), false},
		{"invalid UUID", "invalid-uuid", true},
		{"empty UUID", "", true},
		{"malformed UUID", "123e4567-e89b-12d3-a456-invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateUUID(tt.id)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{"valid date", "2023-01-01T00:00:00Z", false},
		{"invalid format", "2023-13-45", true},
		{"empty date", "", true},
		{"invalid time", "2023-01-01T25:00:00Z", true},
		{"valid with timezone", "2023-01-01T00:00:00+02:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateDate(tt.dateStr)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestValidateRole(t *testing.T) {
	validRole := uuid.New()
	tests := []struct {
		name    string
		roleID  string
		wantErr bool
	}{
		{"valid role", validRole.String(), false},
		{"empty role", "", true},
		{"invalid role", "invalid-role", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateRole(tt.roleID)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
