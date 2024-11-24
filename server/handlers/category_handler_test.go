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

type categoryHandlerTestSuite struct {
	mockRepo     *mocks.MockRepository
	handler      *CategoryHandler
	testUser     uuid.UUID
	testCategory repository.Category
}

func (s *categoryHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
}

func setupCategoryHandlerTest(t *testing.T) *categoryHandlerTestSuite {
	suite := &categoryHandlerTestSuite{}
	t.Cleanup(suite.cleanup)

	// Initialize mock repository
	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock

	// Initialize handler with mock queries
	suite.handler = NewCategoryHandler(repo)

	// Set up test data
	suite.testUser = uuid.New()
	suite.testCategory = repository.Category{
		ID:        uuid.New(),
		UserID:    suite.testUser,
		Name:      "Test Category",
		CreatedAt: time.Now(),
	}

	// Add test category to mock
	suite.mockRepo.GetCategoryMock().AddCategory(suite.testCategory)

	return suite
}

func TestCreateCategory(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		reqBody    interface{} // Changed to interface{} to test malformed JSON
		wantStatus int
	}{
		{
			name:       "Valid category creation",
			userID:     suite.testUser.String(),
			reqBody:    createCategoryRequest{Name: "New Category"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			reqBody:    createCategoryRequest{Name: "New Category"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty category name",
			userID:     suite.testUser.String(),
			reqBody:    createCategoryRequest{Name: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Malformed JSON request",
			userID: suite.testUser.String(),
			reqBody: map[string]interface{}{
				"name":    123, // Wrong type for name
				"invalid": "field",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty request body",
			userID:     suite.testUser.String(),
			reqBody:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Very long category name",
			userID: suite.testUser.String(),
			reqBody: createCategoryRequest{
				Name: string(make([]byte, 256)), // Very long string
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Duplicate category name",
			userID: suite.testUser.String(),
			reqBody: createCategoryRequest{
				Name: suite.testCategory.Name,
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.reqBody != nil {
				body, err = json.Marshal(tt.reqBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.CreateCategory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetCategory(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	tests := []struct {
		name       string
		categoryID string
		userID     string
		wantStatus int
	}{
		{
			name:       "Valid category fetch",
			categoryID: suite.testCategory.ID.String(),
			userID:     suite.testUser.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid category ID",
			categoryID: "invalid-uuid",
			userID:     suite.testUser.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent category",
			categoryID: uuid.New().String(),
			userID:     suite.testUser.String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/categories/"+tt.categoryID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.categoryID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetCategory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	tests := []struct {
		name       string
		categoryID string
		userID     string
		reqBody    interface{}
		wantStatus int
	}{
		{
			name:       "Valid category update",
			categoryID: suite.testCategory.ID.String(),
			userID:     suite.testUser.String(),
			reqBody:    updateCategoryRequest{Name: "Updated Category"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid category ID",
			categoryID: "invalid-uuid",
			userID:     suite.testUser.String(),
			reqBody:    updateCategoryRequest{Name: "Updated Category"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty request body",
			categoryID: suite.testCategory.ID.String(),
			userID:     suite.testUser.String(),
			reqBody:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Wrong user ID for category",
			categoryID: suite.testCategory.ID.String(),
			userID:     uuid.New().String(), // Different user
			reqBody: updateCategoryRequest{
				Name: "Updated Name",
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.reqBody != nil {
				body, err = json.Marshal(tt.reqBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/categories/"+tt.categoryID, bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.categoryID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdateCategory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	tests := []struct {
		name       string
		categoryID string
		userID     string
		wantStatus int
	}{
		{
			name:       "Valid category deletion",
			categoryID: suite.testCategory.ID.String(),
			userID:     suite.testUser.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid category ID",
			categoryID: "invalid-uuid",
			userID:     suite.testUser.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent category",
			categoryID: uuid.New().String(),
			userID:     suite.testUser.String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/categories/"+tt.categoryID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.categoryID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.DeleteCategory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListCategories(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		setupFunc  func()
		wantStatus int
		wantCount  int
	}{
		{
			name:       "Empty category list",
			userID:     uuid.New().String(),
			setupFunc:  func() {},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:   "Multiple categories",
			userID: suite.testUser.String(),
			setupFunc: func() {
				// Add multiple categories
				for i := 0; i < 3; i++ {
					cat := repository.Category{
						ID:        uuid.New(),
						UserID:    suite.testUser,
						Name:      fmt.Sprintf("Category %d", i),
						CreatedAt: time.Now(),
					}
					suite.mockRepo.GetCategoryMock().AddCategory(cat)
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  4, // 3 new + 1 existing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.ListCategories(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code == http.StatusOK {
				var categories []repository.Category
				err := json.NewDecoder(w.Body).Decode(&categories)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(categories))
			}
		})
	}
}

func TestGetCategoryStats(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/categories/"+suite.testCategory.ID.String()+"/stats", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", suite.testCategory.ID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, suite.testUser.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.GetCategoryStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats repository.GetCategoryUsageRow
	err := json.NewDecoder(w.Body).Decode(&stats)
	assert.NoError(t, err)
}

func TestGetMostUsedCategories(t *testing.T) {
	suite := setupCategoryHandlerTest(t)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/categories/most-used"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetMostUsedCategories(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
