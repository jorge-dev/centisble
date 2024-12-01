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
		{
			name: "Missing required fields",
			reqBody: CreateBudgetRequest{
				Amount: 1000.00,
				// Missing other required fields
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid budget type",
			reqBody: CreateBudgetRequest{
				Amount:     1000.00,
				Currency:   "USD",
				CategoryID: uuid.New(),
				Type:       "invalid-type", // Should be "recurring" or "one-time"
				StartDate:  time.Now().Format(time.RFC3339),
				EndDate:    time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "End date before start date",
			reqBody: CreateBudgetRequest{
				Amount:     1000.00,
				Currency:   "USD",
				CategoryID: uuid.New(),
				Type:       "recurring",
				StartDate:  time.Now().Format(time.RFC3339),
				EndDate:    time.Now().AddDate(0, -1, 0).Format(time.RFC3339), // End date before start date
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
		setupMock  func()
	}{
		{
			name:     "Valid full update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Amount:     2000.00,
				Currency:   "EUR",
				CategoryID: uuid.New(),
				Type:       "recurring",
				StartDate:  time.Now().Format(time.RFC3339),
				EndDate:    time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusOK,
			setupMock:  func() {},
		},
		{
			name:     "Valid partial update - amount only",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Amount: 2500.00,
			},
			wantStatus: http.StatusOK,
			setupMock:  func() {},
		},
		{
			name:     "Valid partial update - start dates= only",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				StartDate: time.Now().Format(time.RFC3339),
			},
			wantStatus: http.StatusOK,
			setupMock:  func() {},
		},
		{
			name:     "Valid partial update - end dates only",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				EndDate: time.Now().AddDate(0, 3, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusOK,
			setupMock:  func() {},
		},
		{
			name:     "Invalid partial update - end date before start date",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				EndDate: time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Valid partial update - type only",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Type: "one-time",
			},
			wantStatus: http.StatusOK,
			setupMock:  func() {},
		},
		{
			name:       "Invalid budget ID",
			budgetID:   "invalid-uuid",
			reqBody:    CreateBudgetRequest{},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Invalid amount in partial update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Amount: -100.00,
			},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Invalid currency in partial update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Currency: "INVALID",
			},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Invalid date range in partial update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				StartDate: time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
				EndDate:   time.Now().Format(time.RFC3339), // End date before start date
			},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Invalid budget type in partial update",
			budgetID: suite.testBudget.ID.String(),
			reqBody: CreateBudgetRequest{
				Type: "invalid-type",
			},
			wantStatus: http.StatusBadRequest,
			setupMock:  func() {},
		},
		{
			name:     "Budget not found",
			budgetID: uuid.New().String(), // Non-existent budget ID
			reqBody: CreateBudgetRequest{
				Amount: 2000.00,
			},
			wantStatus: http.StatusNotFound,
			setupMock: func() {
				// Clear existing budgets and don't add the test budget
				suite.mockRepo.Reset()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/budgets/%s", tt.budgetID), bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.budgetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String()))

			w := httptest.NewRecorder()
			suite.handler.UpdateBudget(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			// For successful updates, verify the response contains a budget
			if tt.wantStatus == http.StatusOK {
				var updatedBudget repository.Budget
				err := json.NewDecoder(w.Body).Decode(&updatedBudget)
				assert.NoError(t, err)
				assert.NotEmpty(t, updatedBudget)
			}
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

func TestGetRecurringBudgets(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		setupMock  func()
		userID     string
		wantStatus int
	}{
		{
			name:       "Valid request",
			setupMock:  func() {},
			userID:     suite.testUser.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid user ID",
			setupMock:  func() {},
			userID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "No recurring budgets found",
			setupMock: func() {
				suite.mockRepo.Reset() // Clear all budgets
			},
			userID:     suite.testUser.ID.String(),
			wantStatus: http.StatusOK, // Should return empty array
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req := httptest.NewRequest(http.MethodGet, "/api/budgets/recurring", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetRecurringBudgets(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetBudgetsByCategory(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		categoryID string
		setupMock  func()
		wantStatus int
	}{
		{
			name:       "Valid category ID",
			categoryID: uuid.New().String(),
			setupMock:  func() {},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid category ID",
			categoryID: "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Category with no budgets",
			categoryID: uuid.New().String(),
			setupMock: func() {
				suite.mockRepo.Reset() // Clear all budgets
			},
			wantStatus: http.StatusOK, // Should return empty array
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req := httptest.NewRequest(http.MethodGet, "/api/budgets/category/"+tt.categoryID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("categoryId", tt.categoryID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUser.ID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetBudgetsByCategory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetOneTimeBudgets(t *testing.T) {
	suite := setupBudgetHandlerTest(t)

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
	}{
		{
			name:       "Valid request",
			setupMock:  func() {},
			wantStatus: http.StatusOK,
		},
		{
			name: "No one-time budgets",
			setupMock: func() {
				suite.mockRepo.Reset() // Clear all budgets
			},
			wantStatus: http.StatusOK, // Should return empty array
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req := httptest.NewRequest(http.MethodGet, "/api/budgets/one-time", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetOneTimeBudgets(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
