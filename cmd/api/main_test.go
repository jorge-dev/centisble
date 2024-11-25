package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jorge-dev/centsible/server"
)

// TODO: Refactor this code to make the db mock better ad more reusable
// MockDB implements database.Service interface for testing
type MockDB struct {
	closed bool
	mu     sync.Mutex
}

func NewMockDB() *MockDB {
	return &MockDB{}
}

func (m *MockDB) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockDB) GetConnection() *pgx.Conn {
	return nil // For testing purposes, we return nil as we don't need a real connection
}

func (m *MockDB) Health() map[string]string {
	return map[string]string{
		"status":  "up",
		"message": "Mock DB is healthy",
	}
}

// Add other required database.Service interface methods with mock implementations

func setupTestEnv(t *testing.T) func() {
	// Store original env values
	originalEnv := map[string]string{
		"PORT":                  os.Getenv("PORT"),
		"APP_ENV":               os.Getenv("APP_ENV"),
		"RUN_MIGRATION":         os.Getenv("RUN_MIGRATION"),
		"JWT_SECRET":            os.Getenv("JWT_SECRET"),
		"CENTSIBLE_DB_HOST":     os.Getenv("CENTSIBLE_DB_HOST"),
		"CENTSIBLE_DB_PORT":     os.Getenv("CENTSIBLE_DB_PORT"),
		"CENTSIBLE_DB_DATABASE": os.Getenv("CENTSIBLE_DB_DATABASE"),
		"CENTSIBLE_DB_USERNAME": os.Getenv("CENTSIBLE_DB_USERNAME"),
		"CENTSIBLE_DB_PASSWORD": os.Getenv("CENTSIBLE_DB_PASSWORD"),
		"CENTSIBLE_DB_SCHEMA":   os.Getenv("CENTSIBLE_DB_SCHEMA"),
	}

	// Set test values
	testEnv := map[string]string{
		"PORT":                  "8080",
		"APP_ENV":               "test",
		"RUN_MIGRATION":         "false",
		"JWT_SECRET":            "test-secret",
		"CENTSIBLE_DB_HOST":     "localhost",
		"CENTSIBLE_DB_PORT":     "5432",
		"CENTSIBLE_DB_DATABASE": "centsible_test",
		"CENTSIBLE_DB_USERNAME": "test_user",
		"CENTSIBLE_DB_PASSWORD": "test_password",
		"CENTSIBLE_DB_SCHEMA":   "public",
	}

	// Set test environment
	for key, value := range testEnv {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Could not set test environment %s: %v", key, err)
		}
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

func TestMain(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a cancellable context for the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channels to coordinate test flow
	serverReady := make(chan struct{})
	serverError := make(chan error, 1)
	mockDB := NewMockDB()

	// Start the server in a goroutine
	go func() {
		httpServer, serverImpl := server.NewServer(ctx)
		if httpServer == nil || serverImpl == nil {
			serverError <- fmt.Errorf("server initialization failed")
			return
		}

		// Override the DB with our mock
		serverImpl.SetDB(mockDB)

		// Signal that the server is ready
		close(serverReady)

		// Run graceful shutdown in another goroutine
		go func() {
			gracefulShutdown(ctx, httpServer, mockDB)
		}()

		// Start the server
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	// Wait for server to be ready or timeout after 1 second
	// This is enough time for local server initialization
	select {
	case <-serverReady:
		// Server is ready, continue with test
	case err := <-serverError:
		t.Fatalf("Server failed to start: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Small pause to ensure server is fully running
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown
	cancel()

	// Wait up to 1 second for shutdown to complete
	// This should be plenty of time for a test server to shutdown
	select {
	case err := <-serverError:
		t.Fatalf("Server error during shutdown: %v", err)
	case <-time.After(1 * time.Second):
		// Shutdown completed successfully
	}
}

func TestGracefulShutdown(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	httpServer, serverImpl := server.NewServer(ctx)
	if httpServer == nil || serverImpl == nil {
		t.Fatal("Server initialization failed")
	}

	mockDB := NewMockDB()
	serverImpl.SetDB(mockDB)

	// Create a context with cancel
	shutdownCtx, cancel := context.WithCancel(context.Background())

	// Start graceful shutdown in a goroutine
	go func() {
		gracefulShutdown(shutdownCtx, httpServer, mockDB)
	}()

	// Trigger shutdown after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Give some time for shutdown to complete
	time.Sleep(1 * time.Second)

	// Verify that the mock DB was properly closed
	if !mockDB.closed {
		t.Error("Database was not closed during shutdown")
	}
}
