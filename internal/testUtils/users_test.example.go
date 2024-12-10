package testUtils

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jorge-dev/centsible/internal/repository"
)

var parallel = false

func TestUsersRepository(t *testing.T) {
	checkParallel(t)
	ctx := context.Background()

	// Create test container
	container, err := CreatePostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Terminate(ctx)

	// Create pgx connection
	conn, err := pgx.Connect(ctx, container.ConnectionString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	container.SeedTestData(ctx, conn)

	// Initialize repository with pgx connection
	queries := repository.New(conn)

	t.Run("Check seeded user exist", func(t *testing.T) {
		t.Skip("Skipping check seeded user exist")
		// Test with known seeded email from seed_data.sql
		exists, err := queries.CheckEmailExists(ctx, "john.doe@example.com")
		if err != nil {
			t.Fatalf("Failed to check email existence: %v", err)
		}
		if !exists {
			t.Error("Expected seeded user john.doe@example.com to exist")
		}
	})

	t.Run("Get user by email", func(t *testing.T) {
		t.Skip("Skipping get user by email")
		user, err := queries.GetUserByEmail(ctx, "john.doe@example.com")
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}
		if user.Name != "John Doe" {
			t.Errorf("Expected user name 'John Doe', got '%s'", user.Name)
		}
	})

	t.Run("List users by role", func(t *testing.T) {
		t.Skip("Skipping list users by role")
		users, err := queries.ListUsersByRole(ctx, "Admin")
		if err != nil {
			t.Fatalf("Failed to list users by role: %v", err)
		}
		if len(users) == 0 {
			t.Error("Expected at least one admin user")
		}
		// Verify the first admin user
		if users[0].Role != "Admin" {
			t.Errorf("Expected role 'Admin', got '%s'", users[0].Role)
		}
	})

	t.Run("Get user stats", func(t *testing.T) {
		t.Skip("Skipping get user stats")
		// First get a user ID
		user, err := queries.GetUserByEmail(ctx, "john.doe@example.com")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		stats, err := queries.GetUserStats(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get user stats: %v", err)
		}

		// Verify stats fields exist
		if stats.ID != user.ID {
			t.Errorf("Expected user ID %v, got %v", user.ID, stats.ID)
		}
		if stats.Name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%s'", stats.Name)
		}
		if stats.TotalIncomeRecords != 6 {
			t.Errorf("Expected 6 income records, got %d", stats.TotalIncomeRecords)
		}
	})

	t.Run("Update user role", func(t *testing.T) {
		t.Skip("Skipping update user role")
		// First get a user ID
		user, err := queries.GetUserByEmail(ctx, "john.doe@example.com")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}
		uid := user.ID

		// get the role Id for the role "User: with war sql query and not the one from the repository
		row := conn.QueryRow(ctx, "SELECT id FROM roles WHERE name = 'User'")

		var roleID uuid.UUID
		err = row.Scan(&roleID)
		if err != nil {
			t.Fatalf("Failed to get role ID: %v", err)
		}

		// Update user role
		_, err = queries.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
			RoleID: roleID,
			UserID: uid,
		})
		if err != nil {
			t.Fatalf("Failed to update user role: %v", err)
		}

		user, err = queries.GetUserByEmail(ctx, "john.doe@example.com")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.RoleID.String() != roleID.String() {
			t.Errorf("Expected role 'User', got '%s'", user.RoleID)
		}
	})
}

func checkParallel(t *testing.T) {
	if parallel {
		t.Parallel()
	}
}
