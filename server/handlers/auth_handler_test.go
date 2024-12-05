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
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/internal/repository/mocks"
	"github.com/jorge-dev/centsible/server/middleware"
	"github.com/stretchr/testify/assert"
)

type authHandlerTestSuite struct {
	mockRepo   *mocks.MockRepository
	handler    *AuthHandler
	jwtManager *auth.JWTManager
	testUser   struct {
		ID           uuid.UUID
		Name         string
		Email        string
		Password     string
		RoleID       uuid.UUID
		PasswordHash string
		CreatedAt    time.Time
	}
	validTokenPair *auth.TokenPair
}

func (s *authHandlerTestSuite) cleanup() {
	s.mockRepo.Reset()
	s.testUser = struct {
		ID           uuid.UUID
		Name         string
		Email        string
		Password     string
		RoleID       uuid.UUID
		PasswordHash string
		CreatedAt    time.Time
	}{}
	s.validTokenPair = nil
}

func setupAuthHandlerTest(t *testing.T) *authHandlerTestSuite {
	suite := &authHandlerTestSuite{}

	t.Cleanup(suite.cleanup)

	// Initialize mock repository
	repo := mocks.NewMockRepository()
	mock, ok := repo.(*mocks.MockRepository)
	if !ok {
		t.Fatal("could not cast to MockRepository")
	}
	suite.mockRepo = mock

	// Initialize JWT manager with test secret
	suite.jwtManager = auth.NewJWTManager("test_secret")

	// Initialize handler
	suite.handler = NewAuthHandler(repo, suite.jwtManager)

	// Set up test user data
	suite.testUser.ID = uuid.New()
	suite.testUser.Name = "Test User"
	suite.testUser.Email = "test@example.com"
	suite.testUser.Password = "password"
	suite.testUser.RoleID = uuid.New()
	suite.testUser.CreatedAt = time.Now()

	// Hash the test password
	var err error
	suite.testUser.PasswordHash, err = auth.HashPassword(suite.testUser.Password)
	if err != nil {
		t.Fatal("failed to hash test password")
	}

	// Add test user to mock
	suite.mockRepo.GetUserMock().AddUser(repository.GetUserByIDRow{
		ID:        suite.testUser.ID,
		Name:      suite.testUser.Name,
		Email:     suite.testUser.Email,
		CreatedAt: suite.testUser.CreatedAt,
	})

	suite.mockRepo.GetUserMock().SetEmailExists(suite.testUser.Email, true)

	return suite
}

func TestRegister(t *testing.T) {
	suite := setupAuthHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    RegisterRequest
		wantStatus int
	}{
		{
			name: "Valid registration",
			reqBody: RegisterRequest{
				Name:     "New User",
				Email:    "new@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid email format",
			reqBody: RegisterRequest{
				Name:     "New User",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty name",
			reqBody: RegisterRequest{
				Name:     "",
				Email:    "new@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Short password",
			reqBody: RegisterRequest{
				Name:     "New User",
				Email:    "new@example.com",
				Password: "short",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Existing email",
			reqBody: RegisterRequest{
				Name:     "Another User",
				Email:    "test@example.com", // Same as testUser
				Password: "password123",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.handler.Register(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var response AuthResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.TokenPair.AccessToken)
				assert.NotEmpty(t, response.TokenPair.RefreshToken)
				assert.NotEmpty(t, response.User.ID)
				assert.Equal(t, tt.reqBody.Name, response.User.Name)
				assert.Equal(t, tt.reqBody.Email, response.User.Email)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	suite := setupAuthHandlerTest(t)

	tests := []struct {
		name       string
		reqBody    LoginRequest
		setContext func(context.Context) context.Context
		wantStatus int
	}{
		{
			name: "Valid login",
			reqBody: LoginRequest{
				Email:    suite.testUser.Email,
				Password: suite.testUser.Password,
			},
			// Fix: Convert UUID to string when setting context value
			setContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, suite.testUser.ID.String())
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid email",
			reqBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: suite.testUser.Password,
			},
			setContext: func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Wrong password",
			reqBody: LoginRequest{
				Email:    suite.testUser.Email,
				Password: "wrongpassword",
			},
			setContext: func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid email format",
			reqBody: LoginRequest{
				Email:    "invalid-email",
				Password: suite.testUser.Password,
			},
			setContext: func(ctx context.Context) context.Context { return ctx },

			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty password",
			reqBody: LoginRequest{
				Email: suite.testUser.Email,
			},
			setContext: func(ctx context.Context) context.Context { return ctx },
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := tt.setContext(req.Context())
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			suite.handler.Login(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var response AuthResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.TokenPair.AccessToken)
				assert.NotEmpty(t, response.TokenPair.RefreshToken)
				assert.Equal(t, suite.testUser.Email, response.User.Email)
			}
		})
	}
}

func TestSignout(t *testing.T) {
	suite := setupAuthHandlerTest(t)

	tests := []struct {
		name       string
		setupAuth  func() string
		wantStatus int
	}{
		{
			name: "Valid token",
			setupAuth: func() string {
				tokenPair, _ := suite.jwtManager.GenerateTokenPair(
					suite.testUser.ID.String(),
					suite.testUser.Email,
					suite.testUser.RoleID.String(),
				)
				return "Bearer " + tokenPair.AccessToken
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "No token provided",
			setupAuth: func() string {
				return ""
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid token",
			setupAuth: func() string {
				return "Bearer invalid.token.here"
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/signout", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}
			w := httptest.NewRecorder()

			suite.handler.Signout(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "Successfully signed out", response["message"])
			}
		})
	}
}
