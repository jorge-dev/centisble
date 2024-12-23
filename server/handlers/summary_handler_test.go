package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/internal/repository/mocks"
	"github.com/jorge-dev/centsible/server/middleware"
	"github.com/stretchr/testify/assert"
)

type summaryHandlerTestSuite struct {
	mockRepo *mocks.MockRepository
	handler  *SummaryHandler
	testUser struct {
		ID uuid.UUID
	}
}

func (s *summaryHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
	s.testUser = struct {
		ID uuid.UUID
	}{}
}

func setupSummaryHandlerTest(t *testing.T) *summaryHandlerTestSuite {
	suite := &summaryHandlerTestSuite{}

	t.Cleanup(suite.cleanup)

	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock
	suite.handler = NewSummaryHandler(repo)
	suite.testUser.ID = uuid.New()

	// Setup mock data
	topCategories, _ := json.Marshal([]TopCategory{
		{
			CategoryID:   uuid.New().String(),
			CategoryName: "Food",
			UsageCount:   10,
			TotalSpent:   500.00,
		},
	})

	monthlySummary := []repository.GetMonthlySummaryRow{
		{
			Currency:      "USD",
			TotalIncome:   1000.00,
			TotalExpenses: 500.00,
			TotalSavings:  500.00,
			TopCategories: topCategories,
		},
	}

	monthlyTrend, _ := json.Marshal([]MonthlyTrend{
		{
			Month:        time.Now(),
			CategoryName: "Food",
			Amount:       500.00,
		},
	})

	yearlySummary := []repository.GetYearlySummaryRow{
		{
			Currency:      "USD",
			TotalIncome:   12000.00,
			TotalExpenses: 6000.00,
			TotalSavings:  6000.00,
			TopCategories: topCategories,
			MonthlyTrend:  monthlyTrend,
		},
	}

	// Add test data to mock
	suite.mockRepo.GetSummaryMock().AddMonthlySummary(suite.testUser.ID, time.Now(), monthlySummary)
	suite.mockRepo.GetSummaryMock().AddYearlySummary(suite.testUser.ID, time.Now(), yearlySummary)

	return suite
}

func TestGetMonthlySummary(t *testing.T) {
	suite := setupSummaryHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		date       string
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid request - current month",
			userID:     suite.testUser.ID.String(),
			date:       "",
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Valid request - specific date",
			userID:     suite.testUser.ID.String(),
			date:       time.Now().Format(time.RFC3339),
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			date:       "",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Invalid date format",
			userID:     suite.testUser.ID.String(),
			date:       "invalid-date",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/summary/monthly"
			if tt.date != "" {
				url += "?date=" + tt.date
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetMonthlySummary(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody {
				var response []MonthlySummaryResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response)
				assert.Equal(t, "USD", response[0].Currency)
				assert.NotEmpty(t, response[0].TopCategories)
				assert.Equal(t, "Food", response[0].TopCategories[0].CategoryName)
			}
		})
	}
}

func TestGetYearlySummary(t *testing.T) {
	suite := setupSummaryHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		date       string
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid request - current year",
			userID:     suite.testUser.ID.String(),
			date:       "",
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Valid request - specific year",
			userID:     suite.testUser.ID.String(),
			date:       time.Now().Format(time.RFC3339),
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			date:       "",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Invalid date format",
			userID:     suite.testUser.ID.String(),
			date:       "invalid-date",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/summary/yearly"
			if tt.date != "" {
				url += "?date=" + tt.date
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetYearlySummary(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody {
				var response []YearlySummaryResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response)
				assert.Equal(t, "USD", response[0].Currency)
				assert.NotEmpty(t, response[0].TopCategories)
				assert.NotEmpty(t, response[0].MonthlyTrend)
				assert.Equal(t, "Food", response[0].TopCategories[0].CategoryName)
				assert.Equal(t, "Food", response[0].MonthlyTrend[0].CategoryName)
			}
		})
	}
}
