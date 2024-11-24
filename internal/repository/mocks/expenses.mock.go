package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
)

type ExpenseMock struct {
	expenses map[string]repository.Expense
}

func NewExpenseMock() *ExpenseMock {
	return &ExpenseMock{
		expenses: make(map[string]repository.Expense),
	}
}

// Helper method for setting up test data
func (m *ExpenseMock) AddExpense(expense repository.Expense) {
	m.expenses[expense.ID.String()] = expense
}

func (m *ExpenseMock) CreateExpense(ctx context.Context, arg repository.CreateExpenseParams) (repository.Expense, error) {
	now := time.Now()
	expense := repository.Expense{
		ID:          arg.ID,
		UserID:      arg.UserID,
		Amount:      arg.Amount,
		Currency:    arg.Currency,
		CategoryID:  arg.CategoryID,
		Date:        arg.Date,
		Description: arg.Description,
		CreatedAt:   now,
	}
	m.expenses[expense.ID.String()] = expense
	return expense, nil
}

func (m *ExpenseMock) DeleteExpense(ctx context.Context, arg repository.DeleteExpenseParams) (int64, error) {
	key := arg.ID.String()
	expense, exists := m.expenses[key]
	if !exists || expense.UserID != arg.UserID {
		return 0, nil
	}
	delete(m.expenses, key)
	return 1, nil
}

func (m *ExpenseMock) GetExpenseByID(ctx context.Context, arg repository.GetExpenseByIDParams) (repository.Expense, error) {
	expense, exists := m.expenses[arg.ID.String()]
	if !exists || expense.UserID != arg.UserID {
		return repository.Expense{}, ErrRecordNotFound
	}
	return expense, nil
}

func (m *ExpenseMock) GetExpenseTotalsByCategory(ctx context.Context, userID uuid.UUID) ([]repository.GetExpenseTotalsByCategoryRow, error) {
	totals := make(map[uuid.UUID]repository.GetExpenseTotalsByCategoryRow)

	for _, expense := range m.expenses {
		if expense.UserID == userID {
			total, exists := totals[expense.CategoryID]
			if !exists {
				total = repository.GetExpenseTotalsByCategoryRow{
					CategoryID:   expense.CategoryID,
					CategoryName: "Test Category",
					Currency:     expense.Currency,
				}
			}
			total.TransactionCount++
			total.TotalAmount += expense.Amount
			totals[expense.CategoryID] = total
		}
	}

	result := make([]repository.GetExpenseTotalsByCategoryRow, 0, len(totals))
	for _, total := range totals {
		result = append(result, total)
	}
	return result, nil
}

func (m *ExpenseMock) GetExpensesByCategory(ctx context.Context, arg repository.GetExpensesByCategoryParams) ([]repository.Expense, error) {
	var result []repository.Expense
	for _, expense := range m.expenses {
		if expense.UserID == arg.UserID && expense.CategoryID == arg.CategoryID {
			result = append(result, expense)
		}
	}
	return result, nil
}

func (m *ExpenseMock) GetExpensesByDateRange(ctx context.Context, arg repository.GetExpensesByDateRangeParams) ([]repository.Expense, error) {
	var result []repository.Expense
	for _, expense := range m.expenses {
		if expense.UserID == arg.UserID &&
			!expense.Date.Before(arg.StartDate) &&
			!expense.Date.After(arg.EndDate) {
			result = append(result, expense)
		}
	}
	return result, nil
}

func (m *ExpenseMock) GetMonthlyExpenseTotal(ctx context.Context, arg repository.GetMonthlyExpenseTotalParams) ([]repository.GetMonthlyExpenseTotalRow, error) {
	totals := make(map[string]float64)

	targetMonth := arg.Date.Format("2006-01")
	for _, expense := range m.expenses {
		if expense.UserID == arg.UserID && expense.Date.Format("2006-01") == targetMonth {
			totals[expense.Currency] += expense.Amount
		}
	}

	result := make([]repository.GetMonthlyExpenseTotalRow, 0, len(totals))
	for currency, amount := range totals {
		result = append(result, repository.GetMonthlyExpenseTotalRow{
			Currency:    currency,
			TotalAmount: amount,
		})
	}
	return result, nil
}

func (m *ExpenseMock) GetRecentExpenses(ctx context.Context, arg repository.GetRecentExpensesParams) ([]repository.Expense, error) {
	var result []repository.Expense
	for _, expense := range m.expenses {
		if expense.UserID == arg.UserID {
			result = append(result, expense)
		}
	}
	// Sort by date and limit results
	if int32(len(result)) > arg.Limit {
		result = result[:arg.Limit]
	}
	return result, nil
}

func (m *ExpenseMock) ListExpenses(ctx context.Context, userID uuid.UUID) ([]repository.Expense, error) {
	var result []repository.Expense
	for _, expense := range m.expenses {
		if expense.UserID == userID {
			result = append(result, expense)
		}
	}
	return result, nil
}

func (m *ExpenseMock) UpdateExpense(ctx context.Context, arg repository.UpdateExpenseParams) (repository.Expense, error) {
	expense, exists := m.expenses[arg.ID.String()]
	if !exists || expense.UserID != arg.UserID {
		return repository.Expense{}, ErrRecordNotFound
	}

	expense.Amount = arg.Amount
	expense.Currency = arg.Currency
	expense.CategoryID = arg.CategoryID
	expense.Date = arg.Date
	expense.Description = arg.Description
	now := time.Now()
	expense.UpdatedAt = &now

	m.expenses[arg.ID.String()] = expense
	return expense, nil
}
