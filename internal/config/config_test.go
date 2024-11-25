package config

import (
	"os"
	"testing"
)

func setupTestEnv() func() {
	// Save original env vars
	originalEnv := map[string]string{
		"PORT":                  os.Getenv("PORT"),
		"APP_ENV":               os.Getenv("APP_ENV"),
		"CENTSIBLE_DB_HOST":     os.Getenv("CENTSIBLE_DB_HOST"),
		"CENTSIBLE_DB_PORT":     os.Getenv("CENTSIBLE_DB_PORT"),
		"CENTSIBLE_DB_DATABASE": os.Getenv("CENTSIBLE_DB_DATABASE"),
		"CENTSIBLE_DB_USERNAME": os.Getenv("CENTSIBLE_DB_USERNAME"),
		"CENTSIBLE_DB_PASSWORD": os.Getenv("CENTSIBLE_DB_PASSWORD"),
		"CENTSIBLE_DB_SCHEMA":   os.Getenv("CENTSIBLE_DB_SCHEMA"),
		"RUN_MIGRATION":         os.Getenv("RUN_MIGRATION"),
		"JWT_SECRET":            os.Getenv("JWT_SECRET"),
	}

	// Return cleanup function
	return func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

func TestGet(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	// Set test environment variables
	testEnv := map[string]string{
		"PORT":                  "9090",
		"APP_ENV":               "test",
		"CENTSIBLE_DB_HOST":     "testhost",
		"CENTSIBLE_DB_PORT":     "5432",
		"CENTSIBLE_DB_DATABASE": "testdb",
		"CENTSIBLE_DB_USERNAME": "testuser",
		"CENTSIBLE_DB_PASSWORD": "testpass",
		"CENTSIBLE_DB_SCHEMA":   "public",
		"RUN_MIGRATION":         "true",
		"JWT_SECRET":            "test-secret",
	}

	for key, value := range testEnv {
		os.Setenv(key, value)
	}

	config := Get()

	// Test configuration values
	if config.Port != 9090 {
		t.Errorf("Expected Port to be 9090, got %d", config.Port)
	}

	if config.AppEnv != "test" {
		t.Errorf("Expected AppEnv to be test, got %s", config.AppEnv)
	}

	if config.Database.Host != "testhost" {
		t.Errorf("Expected DB Host to be testhost, got %s", config.Database.Host)
	}

	if !config.Database.RunMigration {
		t.Error("Expected RunMigration to be true")
	}

	if config.JWT.Secret != "test-secret" {
		t.Errorf("Expected JWT Secret to be test-secret, got %s", config.JWT.Secret)
	}
}

func TestLoadPort(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	tests := []struct {
		name     string
		portEnv  string
		expected int
	}{
		{"Default Port", "", 8080},
		{"Custom Port", "3000", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.portEnv != "" {
				os.Setenv("PORT", tt.portEnv)
			} else {
				os.Unsetenv("PORT")
			}

			port := loadPort()
			if port != tt.expected {
				t.Errorf("Expected port %d, got %d", tt.expected, port)
			}
		})
	}
}

func TestLoadEnvWithDefault(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{"Use Environment Value", "TEST_KEY", "test-value", "default", "test-value"},
		{"Use Default Value", "EMPTY_KEY", "", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			result := loadEnvWithDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
