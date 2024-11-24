package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

type expenseHandlerTestSuite struct {
	mockRepo    *mocks.MockRepository
	handler     *ExpenseHandler
	testExpense repository.Expense
	testUserID  uuid.UUID
}

func (s *expenseHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
}

func setupExpenseHandlerTest(t *testing.T) *expenseHandlerTestSuite {
	suite := &expenseHandlerTestSuite{}
	t.Cleanup(suite.cleanup)

	// Initialize mock repository
	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock

	// Initialize handler with mock
	suite.handler = NewExpenseHandler(repo)

	// Set up test data
	suite.testUserID = uuid.New()
	categoryID := uuid.New()
	now := time.Now()

	suite.testExpense = repository.Expense{
		ID:          uuid.New(),
		UserID:      suite.testUserID,
		Amount:      100.50,
		Currency:    "USD",
		CategoryID:  categoryID,
		Date:        now,
		Description: "Test expense",
		CreatedAt:   now,
	}

	// Add test expense to mock
	suite.mockRepo.GetExpenseMock().AddExpense(suite.testExpense)

	return suite
}

func TestCreateExpense(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    ExpenseRequest
		wantStatus int
	}{
		{
			name: "Valid expense",
			reqBody: ExpenseRequest{
				Amount:      100.50,
				Currency:    "USD",
				CategoryID:  uuid.New(),
				Date:        time.Now().Format(time.RFC3339),
				Description: "Test expense",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid date format",
			reqBody: ExpenseRequest{
				Amount:      100.50,
				Currency:    "USD",
				CategoryID:  uuid.New(),
				Date:        "invalid-date",
				Description: "Test expense",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/expenses", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.CreateExpense(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetExpenseByID(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		expenseID  string
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid expense ID",
			expenseID:  suite.testExpense.ID.String(),
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid expense ID",
			expenseID:  "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Non-existent expense ID",
			expenseID:  uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/expenses/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.expenseID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetExpenseByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody {
				var response repository.Expense
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, suite.testExpense.ID, response.ID)
			}
		})
	}
}

func TestUpdateExpense(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		expenseID  string
		reqBody    ExpenseRequest
		wantStatus int
	}{
		{
			name:      "Valid update",
			expenseID: suite.testExpense.ID.String(),
			reqBody: ExpenseRequest{
				Amount:      200.50,
				Currency:    "EUR",
				CategoryID:  suite.testExpense.CategoryID,
				Date:        time.Now().Format(time.RFC3339),
				Description: "Updated description",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "Invalid expense ID",
			expenseID: "invalid-uuid",
			reqBody: ExpenseRequest{
				Amount: 200.50,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/expenses/{id}", bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.expenseID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdateExpense(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteExpense(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		expenseID  string
		wantStatus int
	}{
		{
			name:       "Valid delete",
			expenseID:  suite.testExpense.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid expense ID",
			expenseID:  "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/expenses/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.expenseID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.DeleteExpense(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListExpenses(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/expenses", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.ListExpenses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var expenses []repository.Expense
	err := json.NewDecoder(w.Body).Decode(&expenses)
	assert.NoError(t, err)
	assert.NotEmpty(t, expenses)
}

func TestGetMonthlyExpenseTotal(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		dateTime   string
		wantStatus int
	}{
		{
			name:       "Valid date",
			dateTime:   time.Now().Format(time.RFC3339),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid date format",
			dateTime:   "invalid-date",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/expenses/monthly/total?date-time="+tt.dateTime, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetMonthlyExpenseTotal(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetExpenseTotalsByCategory(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/expenses/category/totals", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.GetExpenseTotalsByCategory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var totals []repository.GetExpenseTotalsByCategoryRow
	err := json.NewDecoder(w.Body).Decode(&totals)
	assert.NoError(t, err)
}

func TestGetRecentExpenses(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		limit      string
		wantStatus int
	}{
		{
			name:       "Valid limit",
			limit:      "5",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid limit",
			limit:      "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/expenses/recent?limit="+tt.limit, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetRecentExpenses(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
