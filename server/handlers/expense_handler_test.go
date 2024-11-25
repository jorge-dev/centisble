package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestCreateExpense_EdgeCases(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    ExpenseRequest
		userID     string
		wantStatus int
	}{
		{
			name:       "Empty request body",
			reqBody:    ExpenseRequest{},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing user context",
			reqBody: ExpenseRequest{
				Amount:   100.50,
				Currency: "USD",
			},
			userID:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Zero amount",
			reqBody: ExpenseRequest{
				Amount:     0,
				Currency:   "USD",
				CategoryID: uuid.New(),
				Date:       time.Now().Format(time.RFC3339),
			},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty currency",
			reqBody: ExpenseRequest{
				Amount:     100.50,
				CategoryID: uuid.New(),
				Date:       time.Now().Format(time.RFC3339),
			},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Very long description",
			reqBody: ExpenseRequest{
				Amount:      100.50,
				Currency:    "USD",
				CategoryID:  uuid.New(),
				Date:        time.Now().Format(time.RFC3339),
				Description: string(make([]byte, 1001)), // Over 1000 characters
			},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid currency code",
			reqBody: ExpenseRequest{
				Amount:      100.50,
				Currency:    "INVALID",
				CategoryID:  uuid.New(),
				Date:        time.Now().Format(time.RFC3339),
				Description: "Test expense",
			},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Negative amount",
			reqBody: ExpenseRequest{
				Amount:      -100.50,
				Currency:    "USD",
				CategoryID:  uuid.New(),
				Date:        time.Now().Format(time.RFC3339),
				Description: "Test expense",
			},
			userID:     suite.testUserID.String(),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/expenses", bytes.NewBuffer(body))
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

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

func TestUpdateExpense_EdgeCases(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		expenseID  string
		reqBody    ExpenseRequest
		setupMock  func()
		wantStatus int
	}{
		{
			name:      "Partial update - only amount",
			expenseID: suite.testExpense.ID.String(),
			reqBody: ExpenseRequest{
				Amount: 299.99,
			},
			setupMock:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:      "Update with same values",
			expenseID: suite.testExpense.ID.String(),
			reqBody: ExpenseRequest{
				Amount:      suite.testExpense.Amount,
				Currency:    suite.testExpense.Currency,
				CategoryID:  suite.testExpense.CategoryID,
				Date:        suite.testExpense.Date.Format(time.RFC3339),
				Description: suite.testExpense.Description,
			},
			setupMock:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:      "Update non-existent expense",
			expenseID: uuid.New().String(),
			reqBody: ExpenseRequest{
				Amount: 150.00,
			},
			setupMock:  nil,
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "Update with invalid category ID",
			expenseID: suite.testExpense.ID.String(),
			reqBody: ExpenseRequest{
				CategoryID: uuid.Nil,
			},
			setupMock:  nil,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

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

func TestGetMonthlyExpenseTotal_EdgeCases(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		dateTime   string
		setupMock  func()
		wantStatus int
		wantEmpty  bool
	}{
		{
			name:       "Month with no expenses",
			dateTime:   time.Now().AddDate(-5, 0, 0).Format(time.RFC3339),
			setupMock:  nil,
			wantStatus: http.StatusOK,
			wantEmpty:  true,
		},
		{
			name:       "Future month",
			dateTime:   time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
			setupMock:  nil,
			wantStatus: http.StatusOK,
			wantEmpty:  true,
		},
		{
			name:       "Invalid month format",
			dateTime:   "2023-13-01T00:00:00Z",
			wantStatus: http.StatusBadRequest,
			wantEmpty:  false,
		},
		{
			name:     "Multiple currencies in same month",
			dateTime: time.Now().Format(time.RFC3339),
			setupMock: func() {
				now := time.Now()
				suite.mockRepo.GetExpenseMock().AddExpense(repository.Expense{
					ID:       uuid.New(),
					UserID:   suite.testUserID,
					Amount:   100.50,
					Currency: "USD",
					Date:     now,
				})
				suite.mockRepo.GetExpenseMock().AddExpense(repository.Expense{
					ID:       uuid.New(),
					UserID:   suite.testUserID,
					Amount:   95.50,
					Currency: "EUR",
					Date:     now,
				})
			},
			wantStatus: http.StatusOK,
			wantEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/api/expenses/monthly/total?date-time="+tt.dateTime, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetMonthlyExpenseTotal(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var totals []repository.GetMonthlyExpenseTotalRow
				err := json.NewDecoder(w.Body).Decode(&totals)
				assert.NoError(t, err)
				if tt.wantEmpty {
					assert.Empty(t, totals)
				} else {
					assert.NotEmpty(t, totals)
				}
			}
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
			name:       "Default limit",
			limit:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Custom limit",
			limit:      "5",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid limit",
			limit:      "invalid",
			wantStatus: http.StatusBadRequest, // Should still work with default limit
		},
		{
			name:       "Negative limit",
			limit:      "-1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Very large limit",
			limit:      "1000000",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Zero limit",
			limit:      "0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Exactly max limit (1000)",
			limit:      "1000",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Just over max limit (1001)",
			limit:      "1001",
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

func TestExpensesByDateRange_EdgeCases(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		startDate  string
		endDate    string
		wantStatus int
	}{
		{
			name:       "Missing dates",
			startDate:  "",
			endDate:    "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "End date before start date",
			startDate:  time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
			endDate:    time.Now().Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Future dates",
			startDate:  time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
			endDate:    time.Now().AddDate(2, 0, 0).Format(time.RFC3339),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Very old dates",
			startDate:  time.Now().AddDate(-10, 0, 0).Format(time.RFC3339),
			endDate:    time.Now().AddDate(-9, 0, 0).Format(time.RFC3339),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/expenses/range?start_date=%s&end_date=%s",
				url.QueryEscape(tt.startDate),
				url.QueryEscape(tt.endDate))
			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetExpensesByDateRange(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetExpensesByDateRange_ValidationChecks(t *testing.T) {
	suite := setupExpenseHandlerTest(t)

	tests := []struct {
		name       string
		startDate  string
		endDate    string
		setupMock  func()
		wantStatus int
	}{
		{
			name:       "Invalid date format - start date",
			startDate:  "2023-13-99",
			endDate:    time.Now().Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid date format - end date",
			startDate:  time.Now().Format(time.RFC3339),
			endDate:    "2023/12/31",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Range too large",
			startDate:  time.Now().AddDate(-2, 0, 0).Format(time.RFC3339),
			endDate:    time.Now().Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Same start and end date",
			startDate:  time.Now().Format(time.RFC3339),
			endDate:    time.Now().Format(time.RFC3339),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			url := fmt.Sprintf("/api/expenses/range?start_date=%s&end_date=%s",
				url.QueryEscape(tt.startDate),
				url.QueryEscape(tt.endDate))
			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetExpensesByDateRange(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
