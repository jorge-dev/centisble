package handlers

import (
	"bytes"
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

type userHandlerTestSuite struct {
	mockRepo *mocks.MockRepository
	handler  *UserHandler
	testUser struct {
		ID        uuid.UUID
		Name      string
		Email     string
		CreatedAt time.Time
	}
	adminRoleID   uuid.UUID
	adminRoleName string
}

// Add cleanup method to the suite
func (s *userHandlerTestSuite) cleanup() {
	// Reset mock repository state
	s.mockRepo.Reset()
	// Clear test user data
	s.testUser = struct {
		ID        uuid.UUID
		Name      string
		Email     string
		CreatedAt time.Time
	}{}
}

func setupUserHandlerTest(t *testing.T) *userHandlerTestSuite {
	suite := &userHandlerTestSuite{}

	// Register cleanup to run after each test
	t.Cleanup(suite.cleanup)

	// Initialize mock repository
	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock

	// Initialize handler
	suite.handler = NewUserHandler(repo)

	// Set up test user data
	suite.testUser.ID = uuid.New()
	suite.testUser.Name = "Test User"
	suite.testUser.Email = "test@example.com"
	suite.testUser.CreatedAt = time.Now()

	// Add test user to mock
	suite.mockRepo.GetUserMock().AddUser(repository.GetUserByIDRow{
		ID:        suite.testUser.ID,
		Name:      suite.testUser.Name,
		Email:     suite.testUser.Email,
		CreatedAt: suite.testUser.CreatedAt,
	})

	// Set up role data
	suite.adminRoleID = uuid.MustParse("6f8b8ad0-4c23-4f01-ac9d-7444413b21f8")
	suite.adminRoleName = "Admin"

	// Set up user role
	suite.mockRepo.GetUserMock().SetAdmin(suite.testUser.ID.String(), true)
	suite.mockRepo.GetUserMock().AddUserRole(repository.GetUserRoleRow{
		UserID:   suite.testUser.ID,
		UserName: suite.testUser.Name,
		RoleID:   suite.adminRoleID,
		RoleName: suite.adminRoleName,
	})

	return suite
}

func TestGetProfile(t *testing.T) {
	suite := setupUserHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		setupCtx   func(context.Context) context.Context
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid user ID",
			userID:     suite.testUser.ID.String(),
			setupCtx:   func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid user ID format",
			userID:     "invalid-uuid",
			setupCtx:   func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Non-existent user ID",
			userID:     uuid.New().String(),
			setupCtx:   func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusNotFound,
			wantBody:   false,
		},
		{
			name:       "Missing UserID in context",
			userID:     "",
			setupCtx:   func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusUnauthorized,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
			ctx := tt.setupCtx(req.Context())
			if tt.userID != "" {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.userID)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.GetProfile(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody {
				var response UserResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, suite.testUser.ID.String(), response.ID)
				assert.Equal(t, suite.testUser.Name, response.Name)
				assert.Equal(t, suite.testUser.Email, response.Email)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	suite := setupUserHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		reqBody    UpdateProfileRequest
		wantStatus int
	}{
		{
			name:       "Valid update",
			userID:     suite.testUser.ID.String(),
			reqBody:    UpdateProfileRequest{Name: "New Name", Email: "new@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			reqBody:    UpdateProfileRequest{Name: "New Name"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty request body",
			userID:     suite.testUser.ID.String(),
			reqBody:    UpdateProfileRequest{},
			wantStatus: http.StatusOK, // Should use existing values
		},
		{
			name:       "Partial update - name only",
			userID:     suite.testUser.ID.String(),
			reqBody:    UpdateProfileRequest{Name: "New Name"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Partial update - email only",
			userID:     suite.testUser.ID.String(),
			reqBody:    UpdateProfileRequest{Email: "newemail@example.com"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/user/profile", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdateProfile(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	suite := setupUserHandlerTest(t)

	tests := []struct {
		name       string
		userID     string
		email      string
		reqBody    UpdatePasswordRequest
		wantStatus int
	}{
		{
			name:       "Valid password update",
			userID:     suite.testUser.ID.String(),
			email:      suite.testUser.Email,
			reqBody:    UpdatePasswordRequest{CurrentPassword: "password", NewPassword: "newpass"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			email:      suite.testUser.Email,
			reqBody:    UpdatePasswordRequest{CurrentPassword: "current", NewPassword: "newpass"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid current password",
			userID:     suite.testUser.ID.String(),
			email:      suite.testUser.Email,
			reqBody:    UpdatePasswordRequest{CurrentPassword: "wrongpass", NewPassword: "newpass"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Missing email in context",
			userID:     suite.testUser.ID.String(),
			email:      "",
			reqBody:    UpdatePasswordRequest{CurrentPassword: "password", NewPassword: "newpass"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/user/password", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			ctx = context.WithValue(ctx, middleware.EmailKey, tt.email)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdatePassword(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetStats(t *testing.T) {
	suite := setupUserHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user/stats", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats repository.GetUserStatsRow
	err := json.NewDecoder(w.Body).Decode(&stats)
	assert.NoError(t, err)
	assert.Equal(t, suite.testUser.ID, stats.ID)
}

func TestUserRoleOperations(t *testing.T) {
	suite := setupUserHandlerTest(t)

	t.Run("GetUserRole", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/role", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
		ctx = context.WithValue(ctx, middleware.RoleIDKey, suite.adminRoleID.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		suite.handler.GetUserRole(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateUserRole", func(t *testing.T) {
		roleID := uuid.New()
		body, _ := json.Marshal(map[string]string{"role_id": roleID.String()})

		req := httptest.NewRequest(http.MethodPut, "/api/user/role", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, suite.testUser.ID.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		suite.handler.UpdateUserRole(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUpdateUserRole(t *testing.T) {
	suite := setupUserHandlerTest(t)

	// Add a non-admin user to the mock
	nonAdminID := uuid.New()
	suite.mockRepo.GetUserMock().AddUser(repository.GetUserByIDRow{
		ID:    nonAdminID,
		Name:  "Non Admin",
		Email: "nonadmin@example.com",
	})
	suite.mockRepo.GetUserMock().SetAdmin(nonAdminID.String(), false)

	tests := []struct {
		name       string
		userID     string
		roleID     string
		wantStatus int
	}{
		{
			name:       "Non-admin attempting update",
			userID:     nonAdminID.String(),
			roleID:     uuid.New().String(),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Empty role ID",
			userID:     suite.testUser.ID.String(),
			roleID:     "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"role_id": tt.roleID})
			req := httptest.NewRequest(http.MethodPut, "/api/user/role", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			suite.handler.UpdateUserRole(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListUsersByRole(t *testing.T) {
	suite := setupUserHandlerTest(t)

	tests := []struct {
		name       string
		roleName   string
		wantStatus int
	}{
		{
			name:       "Valid role",
			roleName:   "Admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Missing role name",
			roleName:   "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent role",
			roleName:   "nonexistent_role",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/role/list?role_name="+tt.roleName, nil)
			w := httptest.NewRecorder()
			suite.handler.ListUsersByRole(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
