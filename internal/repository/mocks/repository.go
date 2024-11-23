package mocks

import (
	"github.com/jorge-dev/centsible/internal/repository"
)

// MockRepository combines all domain-specific mocks
type MockRepository struct {
	*UserMock
	*BudgetMock
	*CategoryMock
}

// NewMockRepository creates a new composite mock repository
func NewMockRepository() repository.Repository {
	return &MockRepository{
		UserMock:     NewUserMock(),
		BudgetMock:   NewBudgetMock(),
		CategoryMock: NewCategoryMock(),
	}
}

// Ensure MockRepository implements Repository interface
var _ repository.Repository = (*MockRepository)(nil)

// Helper functions for testing
func (m *MockRepository) Reset() {
	m.UserMock = NewUserMock()
	m.BudgetMock = NewBudgetMock()
	m.CategoryMock = NewCategoryMock()
}

// GetUserMock returns the underlying UserMock for testing helpers
func (m *MockRepository) GetUserMock() *UserMock {
	return m.UserMock
}

// GetBudgetMock returns the underlying BudgetMock for testing helpers
func (m *MockRepository) GetBudgetMock() *BudgetMock {
	return m.BudgetMock
}

// GetCategoryMock returns the underlying CategoryMock for testing helpers
func (m *MockRepository) GetCategoryMock() *CategoryMock {
	return m.CategoryMock
}
