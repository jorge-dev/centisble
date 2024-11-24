package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
)

type IncomeMock struct {
	incomes map[string]repository.Income
}

func NewIncomeMock() *IncomeMock {
	return &IncomeMock{
		incomes: make(map[string]repository.Income),
	}
}

// Helper methods for setting up test data
func (m *IncomeMock) AddIncome(income repository.Income) {
	m.incomes[income.ID.String()] = income
}

func (m *IncomeMock) CreateIncome(ctx context.Context, arg repository.CreateIncomeParams) (repository.Income, error) {
	income := repository.Income{
		ID:          arg.ID,
		UserID:      arg.UserID,
		Amount:      arg.Amount,
		Currency:    arg.Currency,
		Source:      arg.Source,
		Date:        arg.Date,
		Description: arg.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   nil,
		DeletedAt:   nil,
	}
	m.incomes[income.ID.String()] = income
	return income, nil
}

func (m *IncomeMock) DeleteIncome(ctx context.Context, arg repository.DeleteIncomeParams) (int64, error) {
	if income, exists := m.incomes[arg.ID.String()]; exists && income.UserID == arg.UserID {
		now := time.Now()
		income.DeletedAt = &now
		m.incomes[arg.ID.String()] = income
		return 1, nil
	}
	return 0, nil
}

func (m *IncomeMock) GetIncomeByID(ctx context.Context, arg repository.GetIncomeByIDParams) (repository.Income, error) {
	if income, exists := m.incomes[arg.ID.String()]; exists && income.UserID == arg.UserID && income.DeletedAt == nil {
		return income, nil
	}
	return repository.Income{}, ErrRecordNotFound
}

func (m *IncomeMock) ListIncome(ctx context.Context, userID uuid.UUID) ([]repository.Income, error) {
	var incomes []repository.Income
	for _, income := range m.incomes {
		if income.UserID == userID && income.DeletedAt == nil {
			incomes = append(incomes, income)
		}
	}
	return incomes, nil
}

func (m *IncomeMock) UpdateIncome(ctx context.Context, arg repository.UpdateIncomeParams) (repository.Income, error) {
	if income, exists := m.incomes[arg.ID.String()]; exists && income.UserID == arg.UserID && income.DeletedAt == nil {
		now := time.Now()
		income.Amount = arg.Amount
		income.Currency = arg.Currency
		income.Source = arg.Source
		income.Date = arg.Date
		income.Description = arg.Description
		income.UpdatedAt = &now
		m.incomes[arg.ID.String()] = income
		return income, nil
	}
	return repository.Income{}, ErrRecordNotFound
}

func (m *IncomeMock) GetIncomeByDateRange(ctx context.Context, arg repository.GetIncomeByDateRangeParams) ([]repository.Income, error) {
	var incomes []repository.Income
	for _, income := range m.incomes {
		if income.UserID == arg.UserID && income.DeletedAt == nil &&
			!income.Date.Before(arg.StartDate) && !income.Date.After(arg.EndDate) {
			incomes = append(incomes, income)
		}
	}
	return incomes, nil
}

func (m *IncomeMock) GetIncomeBySource(ctx context.Context, arg repository.GetIncomeBySourceParams) ([]repository.Income, error) {
	var incomes []repository.Income
	for _, income := range m.incomes {
		if income.UserID == arg.UserID && income.Source == arg.Source && income.DeletedAt == nil {
			incomes = append(incomes, income)
		}
	}
	return incomes, nil
}

func (m *IncomeMock) GetIncomeSummaryBySource(ctx context.Context, userID uuid.UUID) ([]repository.GetIncomeSummaryBySourceRow, error) {
	summary := make(map[string]map[string]*repository.GetIncomeSummaryBySourceRow)

	for _, income := range m.incomes {
		if income.UserID == userID && income.DeletedAt == nil {
			if _, exists := summary[income.Source]; !exists {
				summary[income.Source] = make(map[string]*repository.GetIncomeSummaryBySourceRow)
			}
			if _, exists := summary[income.Source][income.Currency]; !exists {
				summary[income.Source][income.Currency] = &repository.GetIncomeSummaryBySourceRow{
					Source:   income.Source,
					Currency: income.Currency,
				}
			}
			s := summary[income.Source][income.Currency]
			s.TransactionCount++
			s.TotalAmount += int64(income.Amount)
			s.AverageAmount = float64(s.TotalAmount) / float64(s.TransactionCount)
		}
	}

	var result []repository.GetIncomeSummaryBySourceRow
	for _, currencies := range summary {
		for _, s := range currencies {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *IncomeMock) GetMonthlyIncomeTotal(ctx context.Context, arg repository.GetMonthlyIncomeTotalParams) (repository.GetMonthlyIncomeTotalRow, error) {
	var total float64
	var currency string

	for _, income := range m.incomes {
		if income.UserID == arg.UserID && income.DeletedAt == nil &&
			income.Date.Year() == arg.Date.Year() && income.Date.Month() == arg.Date.Month() {
			total += income.Amount
			currency = income.Currency
		}
	}

	return repository.GetMonthlyIncomeTotalRow{
		TotalAmount: total,
		Currency:    currency,
	}, nil
}

func (m *IncomeMock) GetRecentIncome(ctx context.Context, arg repository.GetRecentIncomeParams) ([]repository.Income, error) {
	var incomes []repository.Income
	for _, income := range m.incomes {
		if income.UserID == arg.UserID && income.DeletedAt == nil {
			incomes = append(incomes, income)
		}
	}

	// Sort by date descending and limit results
	if int32(len(incomes)) > arg.Limit {
		incomes = incomes[:arg.Limit]
	}
	return incomes, nil
}
