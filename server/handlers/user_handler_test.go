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
	"github.com/jorge-dev/centsible/server/middleware"
	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	// Add test user
	testUserID := uuid.New()
	mockDB.AddUser(repository.GetUserByIDRow{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name       string
		userID     string
		wantStatus int
		wantBody   bool
	}{
		{
			name:       "Valid user ID",
			userID:     testUserID.String(),
			wantStatus: http.StatusOK,
			wantBody:   true,
		},
		{
			name:       "Invalid user ID format",
			userID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantBody:   false,
		},
		{
			name:       "Non-existent user ID",
			userID:     uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.GetProfile(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody {
				var response UserResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, testUserID.String(), response.ID)
				assert.Equal(t, "Test User", response.Name)
				assert.Equal(t, "test@example.com", response.Email)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	testUserID := uuid.New()
	mockDB.AddUser(repository.GetUserByIDRow{
		ID:        testUserID,
		Name:      "Original Name",
		Email:     "original@example.com",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name       string
		userID     string
		reqBody    UpdateProfileRequest
		wantStatus int
	}{
		{
			name:       "Valid update",
			userID:     testUserID.String(),
			reqBody:    UpdateProfileRequest{Name: "New Name", Email: "new@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			reqBody:    UpdateProfileRequest{Name: "New Name"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/user/profile", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.UpdateProfile(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	testUserID := uuid.New()
	testEmail := "test@example.com"
	mockDB.AddUser(repository.GetUserByIDRow{
		ID:    testUserID,
		Email: testEmail,
	})

	tests := []struct {
		name       string
		userID     string
		email      string
		reqBody    UpdatePasswordRequest
		wantStatus int
	}{
		{
			name:       "Valid password update",
			userID:     testUserID.String(),
			email:      testEmail,
			reqBody:    UpdatePasswordRequest{CurrentPassword: "password", NewPassword: "newpass"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid user ID",
			userID:     "invalid-uuid",
			email:      testEmail,
			reqBody:    UpdatePasswordRequest{CurrentPassword: "current", NewPassword: "newpass"},
			wantStatus: http.StatusBadRequest,
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
			handler.UpdatePassword(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetStats(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	testUserID := uuid.New()
	mockDB.AddUser(repository.GetUserByIDRow{
		ID: testUserID,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/stats", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats repository.GetUserStatsRow
	err := json.NewDecoder(w.Body).Decode(&stats)
	assert.NoError(t, err)
	assert.Equal(t, testUserID, stats.ID)
}

func TestUserRoleOperations(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	testUserID := uuid.New()

	mockDB.AddUser(repository.GetUserByIDRow{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "j@me.com",
		CreatedAt: time.Now(),
	})
	mockDB.SetAdmin(testUserID.String(), true)

	adminRoleId, _ := uuid.Parse("6f8b8ad0-4c23-4f01-ac9d-7444413b21f8")

	adminRoleName := "Admin"

	// userRoleId := uuid.Parse("9fecb503-be55-4adb-a0fe-526660efbf51")
	// userRoleName := "User"

	mockDB.AddUserRole(repository.GetUserRoleRow{
		UserID:   testUserID,
		UserName: "Test User",
		RoleID:   adminRoleId,
		RoleName: adminRoleName,
	})

	t.Run("GetUserRole", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/role", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID.String())
		ctx = context.WithValue(ctx, middleware.RoleIDKey, adminRoleId.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.GetUserRole(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateUserRole", func(t *testing.T) {
		roleID := uuid.New()
		body, _ := json.Marshal(map[string]string{"role_id": roleID.String()})

		req := httptest.NewRequest(http.MethodPut, "/api/user/role", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID.String())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.UpdateUserRole(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestListUsersByRole(t *testing.T) {
	mockDB := repository.NewMockRepository()
	handler := NewUserHandler(mockDB)

	tests := []struct {
		name       string
		roleName   string
		wantStatus int
	}{
		{
			name:       "Valid role",
			roleName:   "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Missing role name",
			roleName:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/users/role/list?role_name="+tt.roleName, nil)
			w := httptest.NewRecorder()
			handler.ListUsersByRole(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
