package server

import (
	"context"
	"os"
	"testing"

	"github.com/jorge-dev/centsible/internal/config"
)

var testEnvVars = map[string]string{
	"JWT_SECRET":            "test-secret",
	"CENTSIBLE_DB_HOST":     "localhost",
	"CENTSIBLE_DB_PORT":     "5432",
	"CENTSIBLE_DB_DATABASE": "test_db",
	"CENTSIBLE_DB_USERNAME": "test_user",
	"CENTSIBLE_DB_PASSWORD": "test_password",
	"CENTSIBLE_DB_SCHEMA":   "public",
	"RUN_MIGRATION":         "false",
}

func setupTest(extraVars map[string]string) func() {
	originalEnv := make(map[string]string)

	// Save original env vars
	for k := range testEnvVars {
		originalEnv[k] = os.Getenv(k)
	}
	for k := range extraVars {
		originalEnv[k] = os.Getenv(k)
	}

	// Set test env vars
	for k, v := range testEnvVars {
		os.Setenv(k, v)
	}
	for k, v := range extraVars {
		os.Setenv(k, v)
	}

	return func() {
		// Restore original env vars
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
		config.ResetConfig()
	}
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"PORT":    "8080",
				"APP_ENV": "test",
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PORT":    "-1",
				"APP_ENV": "test",
			},
			wantErr: true,
		},
		{
			name: "non-numeric port",
			envVars: map[string]string{
				"PORT":    "abc",
				"APP_ENV": "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTest(tt.envVars)
			defer cleanup()

			httpServer, serverImpl := NewServer(context.Background())

			if tt.wantErr {
				if httpServer != nil || serverImpl != nil {
					t.Error("NewServer() should return nil when configuration is invalid")
				}
				return
			}

			if httpServer == nil || serverImpl == nil {
				t.Fatal("NewServer() returned nil with valid configuration")
			}

			if serverImpl.port != 8080 {
				t.Errorf("Server port = %v, want %v", serverImpl.port, 8080)
			}

			if _, ok := serverImpl.db.(*mockDB); !ok {
				t.Error("Expected mock database in test environment")
			}
		})
	}
}

func TestGetDB(t *testing.T) {
	mockDB := &mockDB{healthStatus: true}
	s := &Server{
		port: 8080,
		db:   mockDB,
	}

	if got := s.GetDB(); got != mockDB {
		t.Error("GetDB() didn't return expected mock database")
	}
}
