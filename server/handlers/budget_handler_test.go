package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/internal/repository/mocks"
	"github.com/jorge-dev/centsible/server/middleware"
	"github.com/stretchr/testify/assert"
)

type budgetHandlerTestSuite struct {
	mockRepo   *mocks.MockRepository
	handler    *BudgetHandler
	testBudget repository.Budget
	testUser   struct {
		ID uuid.UUID
	}
}

func (s *budgetHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
	s.testBudget = repository.Budget{}
	s.testUser.ID = uuid.Nil
}

func setupBudgetHandlerTest(t *testing.T) *budgetHandlerTestSuite {
	suite := &budgetHandlerTestSuite{}
	t.Cleanup(suite.cleanup)

	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock
	suite.handler = NewBudgetHandler(repo)

	// Setup test data
	suite.testUser.ID = uuid.New()
	suite.testBudget = repository.Budget{
		ID:         uuid.New(),
		UserID:     suite.testUser.ID,
		Amount:     1000.00,
		Currency:   "USD",
		CategoryID: uuid.New(),
		Type:       "recurring",
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 1, 0),
		CreatedAt:  time.Now(),
	}

	// Add test budget to mock
	suite.mockRepo.GetBudgetMock().AddBudget(suite.testBudget)

	return suite
}

func TestCreateBudget(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    CreateBudgetRequest
		wantStatus int
	}{
		{
			name: "Valid budget creation",
			reqBody: CreateBudgetRequest{
				Amount:     1000.00,
				Currency:   "USD",
				CategoryID: uuid.New(),
				Type:       "recurring",
				StartDate:  time.Now().Format(time.RFC3339),
				EndDate:    time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid date format",
			reqBody: CreateBudgetRequest{
				Amount:     1000.00,
				Currency:   "USD",
				CategoryID: uuid.New(),
				Type:       "recurring",
				StartDate:  "invalid-date",
				EndDate:    "invalid-date",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/budgets", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.CreateBudget(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetBudgetUsage(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		budgetID   string
		wantStatus int
	}{
		{
			name:       "Valid budget ID",
			budgetID:   suite.testBudget.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid budget ID",
			budgetID:   "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/budgets/%s/usage", tt.budgetID), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.budgetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String()))

			w := httptest.NewRecorder()
			suite.handler.GetBudgetUsage(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetBudgetsNearLimit(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name           string
		alertThreshold string
		wantStatus     int
	}{
		{
			name:           "Valid threshold",
			alertThreshold: "80",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "Invalid threshold",
			alertThreshold: "invalid",
			wantStatus:     http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/budgets/alerts?alert_threshold="+tt.alertThreshold, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetBudgetsNearLimit(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateBudget(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		budgetID   string
		reqBody    CreateBudgetRequest
		wantStatus int
	}{
		{
			name:     "Valid update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Amount:    2000.00,
				Currency:  "EUR",
				StartDate: time.Now().Format(time.RFC3339),
				EndDate:   time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid budget ID",
			budgetID:   "invalid-uuid",
			reqBody:    CreateBudgetRequest{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/budgets/%s", tt.budgetID), bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.budgetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String()))

			w := httptest.NewRecorder()
			suite.handler.UpdateBudget(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteBudget(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		budgetID   string
		wantStatus int
	}{
		{
			name:       "Valid deletion",
			budgetID:   suite.testBudget.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid budget ID",
			budgetID:   "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/budgets/%s", tt.budgetID), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.budgetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String()))

			w := httptest.NewRecorder()
			suite.handler.DeleteBudget(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListBudgets(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/budgets", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.ListBudgets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var budgets []repository.Budget
	err := json.NewDecoder(w.Body).Decode(&budgets)
	assert.NoError(t, err)
}

// Additional test functions can be added for:
// - GetRecurringBudgets
// - GetOneTimeBudgets
// - GetBudgetsByCategory
