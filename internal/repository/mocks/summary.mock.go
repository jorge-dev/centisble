package mocks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jorge-dev/centsible/internal/repository"
)

type SummaryMock struct {
	monthlySummaries map[string][]repository.GetMonthlySummaryRow
	yearlySummaries  map[string][]repository.GetYearlySummaryRow
}

func NewSummaryMock() *SummaryMock {
	return &SummaryMock{
		monthlySummaries: make(map[string][]repository.GetMonthlySummaryRow),
		yearlySummaries:  make(map[string][]repository.GetYearlySummaryRow),
	}
}

// Helper methods for setting up test data
func (m *SummaryMock) AddMonthlySummary(userID uuid.UUID, date time.Time, summary []repository.GetMonthlySummaryRow) {
	key := userID.String() + date.Format("2006-01")
	m.monthlySummaries[key] = summary
}

func (m *SummaryMock) AddYearlySummary(userID uuid.UUID, date time.Time, summary []repository.GetYearlySummaryRow) {
	key := userID.String() + date.Format("2006")
	m.yearlySummaries[key] = summary
}

// Implementation of summary-related Repository interface methods
func (m *SummaryMock) GetMonthlySummary(ctx context.Context, arg repository.GetMonthlySummaryParams) ([]repository.GetMonthlySummaryRow, error) {
	key := arg.UserID.String() + arg.Date.Format("2006-01")
	summary, exists := m.monthlySummaries[key]
	if !exists {
		// Return default mock data if no specific data was set
		topCategories, _ := json.Marshal([]map[string]interface{}{
			{
				"category_id":   uuid.New(),
				"category_name": "Food",
				"usage_count":   10,
				"total_spent":   500.00,
			},
		})

		return []repository.GetMonthlySummaryRow{
			{
				Currency:      pgtype.Text{String: "USD", Valid: true},
				TotalIncome:   1000.00,
				TotalExpenses: 500.00,
				TotalSavings:  500.00,
				TopCategories: topCategories,
			},
		}, nil
	}
	return summary, nil
}

func (m *SummaryMock) GetYearlySummary(ctx context.Context, arg repository.GetYearlySummaryParams) ([]repository.GetYearlySummaryRow, error) {
	key := arg.UserID.String() + arg.Date.Format("2006")
	summary, exists := m.yearlySummaries[key]
	if !exists {
		// Return default mock data if no specific data was set
		topCategories, _ := json.Marshal([]map[string]interface{}{
			{
				"category_id":   uuid.New(),
				"category_name": "Food",
				"usage_count":   120,
				"total_spent":   6000.00,
			},
		})

		monthlyTrend, _ := json.Marshal([]map[string]interface{}{
			{
				"month":         time.Now().Format("2006-01-02"),
				"category_name": "Food",
				"amount":        500.00,
			},
		})

		return []repository.GetYearlySummaryRow{
			{
				Currency:      pgtype.Text{String: "USD", Valid: true},
				TotalIncome:   12000.00,
				TotalExpenses: 6000.00,
				TotalSavings:  6000.00,
				TopCategories: topCategories,
				MonthlyTrend:  monthlyTrend,
			},
		}, nil
	}
	return summary, nil
}
