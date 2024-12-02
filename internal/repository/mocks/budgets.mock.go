package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
)

type BudgetMock struct {
	budgets map[string]repository.Budget
}

func NewBudgetMock() *BudgetMock {
	return &BudgetMock{
		budgets: make(map[string]repository.Budget),
	}
}

// Helper methods for setting up test data
func (m *BudgetMock) AddBudget(budget repository.Budget) {
	m.budgets[budget.ID.String()] = budget
}

func (m *BudgetMock) CreateBudget(ctx context.Context, arg repository.CreateBudgetParams) (repository.Budget, error) {
	now := time.Now()
	budget := repository.Budget{
		ID:         arg.ID,
		UserID:     arg.UserID,
		Amount:     arg.Amount,
		Currency:   arg.Currency,
		CategoryID: arg.CategoryID,
		Type:       arg.Type,
		StartDate:  arg.StartDate,
		EndDate:    arg.EndDate,
		CreatedAt:  now,
		UpdatedAt:  &now,
		Name:       arg.Name,
	}
	m.budgets[budget.ID.String()] = budget
	return budget, nil
}

func (m *BudgetMock) DeleteBudget(ctx context.Context, arg repository.DeleteBudgetParams) (int64, error) {
	key := arg.ID.String()
	if budget, exists := m.budgets[key]; exists && budget.UserID == arg.UserID {
		now := time.Now()
		budget.DeletedAt = &now
		m.budgets[key] = budget
		return 1, nil
	}
	return 0, nil
}

func (m *BudgetMock) GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]repository.Budget, error) {
	var result []repository.Budget
	now := time.Now()
	for _, budget := range m.budgets {
		if budget.UserID == userID && budget.DeletedAt == nil &&
			!budget.StartDate.After(now) &&
			(budget.EndDate.After(now) || budget.EndDate.IsZero()) {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (m *BudgetMock) GetBudgetByID(ctx context.Context, arg repository.GetBudgetByIDParams) (repository.Budget, error) {
	if budget, exists := m.budgets[arg.ID.String()]; exists && budget.UserID == arg.UserID && budget.DeletedAt == nil {
		return budget, nil
	}
	return repository.Budget{}, ErrRecordNotFound
}

func (m *BudgetMock) GetBudgetUsage(ctx context.Context, arg repository.GetBudgetUsageParams) (repository.GetBudgetUsageRow, error) {
	budget, exists := m.budgets[arg.BudgetID.String()]
	if !exists || budget.UserID != arg.UserID || budget.DeletedAt != nil {
		return repository.GetBudgetUsageRow{}, ErrRecordNotFound
	}

	// Mock spent amount and usage percentage
	spentAmount := budget.Amount * 0.75 // Mock 75% usage
	usagePercentage := 75.0

	return repository.GetBudgetUsageRow{
		Budget:          budget,
		SpentAmount:     spentAmount,
		UsagePercentage: usagePercentage,
	}, nil
}

func (m *BudgetMock) GetBudgetsByCategory(ctx context.Context, arg repository.GetBudgetsByCategoryParams) ([]repository.Budget, error) {
	var result []repository.Budget
	for _, budget := range m.budgets {
		if budget.UserID == arg.UserID && budget.CategoryID == arg.CategoryID && budget.DeletedAt == nil {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (m *BudgetMock) GetBudgetsNearLimit(ctx context.Context, arg repository.GetBudgetsNearLimitParams) ([]repository.GetBudgetsNearLimitRow, error) {
	var result []repository.GetBudgetsNearLimitRow
	for _, budget := range m.budgets {
		if budget.UserID == arg.UserID && budget.DeletedAt == nil {
			spentAmount := budget.Amount * 0.8 // Mock 80% usage
			usagePercentage := 80.0
			if usagePercentage >= arg.Threshold {
				result = append(result, repository.GetBudgetsNearLimitRow{
					Budget:          budget,
					SpentAmount:     spentAmount,
					UsagePercentage: usagePercentage,
				})
			}
		}
	}
	return result, nil
}

func (m *BudgetMock) GetOneTimeBudgets(ctx context.Context, userID uuid.UUID) ([]repository.Budget, error) {
	var result []repository.Budget
	for _, budget := range m.budgets {
		if budget.UserID == userID && budget.Type == "one-time" && budget.DeletedAt == nil {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (m *BudgetMock) GetRecurringBudgets(ctx context.Context, userID uuid.UUID) ([]repository.Budget, error) {
	var result []repository.Budget
	for _, budget := range m.budgets {
		if budget.UserID == userID && budget.Type == "recurring" && budget.DeletedAt == nil {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (m *BudgetMock) ListBudgets(ctx context.Context, userID uuid.UUID) ([]repository.Budget, error) {
	var result []repository.Budget
	for _, budget := range m.budgets {
		if budget.UserID == userID && budget.DeletedAt == nil {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (m *BudgetMock) UpdateBudget(ctx context.Context, arg repository.UpdateBudgetParams) (repository.Budget, error) {
	if budget, exists := m.budgets[arg.ID.String()]; exists && budget.UserID == arg.UserID && budget.DeletedAt == nil {
		now := time.Now()
		budget.Amount = arg.Amount
		budget.Currency = arg.Currency
		budget.CategoryID = arg.CategoryID
		budget.Type = arg.Type
		budget.StartDate = arg.StartDate
		budget.EndDate = arg.EndDate
		budget.Name = arg.Name
		budget.UpdatedAt = &now
		m.budgets[arg.ID.String()] = budget
		return budget, nil
	}
	return repository.Budget{}, ErrRecordNotFound
}
