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

type incomeHandlerTestSuite struct {
	mockRepo   *mocks.MockRepository
	handler    *IncomeHandler
	testIncome repository.Income
	testUserID uuid.UUID
}

func (s *incomeHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
}

func setupIncomeHandlerTest(t *testing.T) *incomeHandlerTestSuite {
	suite := &incomeHandlerTestSuite{}
	t.Cleanup(suite.cleanup)

	// Initialize mock repository
	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock

	// Initialize handler with mock
	suite.handler = NewIncomeHandler(repo)

	// Set up test data
	suite.testUserID = uuid.New()
	now := time.Now()

	suite.testIncome = repository.Income{
		ID:          uuid.New(),
		UserID:      suite.testUserID,
		Amount:      1000.50,
		Currency:    "USD",
		Source:      "Salary",
		Date:        now,
		Description: "Monthly salary",
		CreatedAt:   now,
	}

	// Add test income to mock
	suite.mockRepo.GetIncomeMock().AddIncome(suite.testIncome)

	return suite
}

func TestCreateIncome(t *testing.T) {
	suite := setupIncomeHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    CreateIncomeRequest
		wantStatus int
	}{
		{
			name: "Valid income",
			reqBody: CreateIncomeRequest{
				Amount:      1000.50,
				Currency:    "USD",
				Source:      "Salary",
				Date:        time.Now(),
				Description: "Monthly salary",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid amount",
			reqBody: CreateIncomeRequest{
				Amount:      -100,
				Currency:    "USD",
				Source:      "Salary",
				Date:        time.Now(),
				Description: "Monthly salary",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing currency",
			reqBody: CreateIncomeRequest{
				Amount:      1000.50,
				Source:      "Salary",
				Date:        time.Now(),
				Description: "Monthly salary",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/income", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.CreateIncome(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetIncomeByID(t *testing.T) {
	suite := setupIncomeHandlerTest(t)

	tests := []struct {
		name       string
		incomeID   string
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid income ID",
			incomeID:   suite.testIncome.ID.String(),
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid income ID",
			incomeID:   "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Non-existent income ID",
			incomeID:   uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/income/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.incomeID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetIncomeByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody {
				var response repository.Income
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, suite.testIncome.ID, response.ID)
			}
		})
	}
}

func TestUpdateIncome(t *testing.T) {
	suite := setupIncomeHandlerTest(t)

	tests := []struct {
		name       string
		incomeID   string
		reqBody    CreateIncomeRequest
		wantStatus int
	}{
		{
			name:     "Valid update",
			incomeID: suite.testIncome.ID.String(),
			reqBody: CreateIncomeRequest{
				Amount:      2000.50,
				Currency:    "EUR",
				Source:      "Bonus",
				Date:        time.Now(),
				Description: "Updated description",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "Invalid amount",
			incomeID: suite.testIncome.ID.String(),
			reqBody: CreateIncomeRequest{
				Amount: -100,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/income/{id}", bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.incomeID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdateIncome(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteIncome(t *testing.T) {
	suite := setupIncomeHandlerTest(t)

	tests := []struct {
		name       string
		incomeID   string
		wantStatus int
	}{
		{
			name:       "Valid delete",
			incomeID:   suite.testIncome.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid income ID",
			incomeID:   "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent income ID",
			incomeID:   uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/income/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.incomeID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUserID.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.DeleteIncome(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListIncome(t *testing.T) {
	suite := setupIncomeHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/income", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUserID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.GetIncomeList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var incomes []repository.Income
	err := json.NewDecoder(w.Body).Decode(&incomes)
	assert.NoError(t, err)
	assert.NotEmpty(t, incomes)
}
