package server

import (
	"github.com/jackc/pgx/v5"
	"github.com/jorge-dev/centsible/internal/database"
)

// Mock database service
type mockDB struct {
	database.Service
	healthStatus bool
}

func newMockDB() *mockDB {
	return &mockDB{
		healthStatus: true,
	}
}

func (m *mockDB) Health() map[string]string {
	if m.healthStatus {
		return map[string]string{"status": "up", "message": "Database connection is healthy"}
	}
	return map[string]string{"status": "down", "message": "Database connection failed"}
}

func (m *mockDB) GetConnection() *pgx.Conn {
	if !m.healthStatus {
		return nil
	}
	return &pgx.Conn{} // Return empty connection for testing
}
