package testUtils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithInitScripts(filepath.Join("..", "database", "init", "01_init.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	// Run migrations after container is ready
	if err := runTestMigrations(connStr); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

// Move seeding to a helper function that accepts a connection
func (p *PostgresContainer) SeedTestData(ctx context.Context, conn *pgx.Conn) error {
	log.Println("Seeding test data...")

	seedPath := filepath.Join("..", "database", "queries", "seed_data.sql")
	log.Printf("Looking for seed file at: %s", seedPath)

	seedContent, err := os.ReadFile(seedPath)
	if err != nil {
		altPath := filepath.Join("internal", "database", "queries", "seed_data.sql")
		log.Printf("First path failed, trying alternative path: %s", altPath)
		seedContent, err = os.ReadFile(altPath)
		if err != nil {
			return fmt.Errorf("error reading seed file from both paths: %v", err)
		}
	}

	// Split by -- name: to get individual named queries
	queries := strings.Split(string(seedContent), "-- name:")

	// Execute queries in order: Users -> Categories -> Income -> Expenses -> Budgets
	for _, query := range queries {
		// Skip empty queries and DeleteSeedData
		if strings.TrimSpace(query) == "" || strings.Contains(query, "DeleteSeedData") {
			continue
		}

		// Extract the actual SQL (everything after :exec)
		if idx := strings.Index(query, ":exec"); idx != -1 {
			queryName := strings.TrimSpace(query[:idx])
			sql := strings.TrimSpace(query[idx+5:])

			log.Printf("Executing seed query: %s", queryName)
			if _, err := conn.Exec(ctx, sql); err != nil {
				return fmt.Errorf("error executing seed query %s: %v", queryName, err)
			}
			log.Printf("Successfully executed: %s", queryName)
		}
	}

	log.Println("Test data seeded successfully")
	return nil
}

func runTestMigrations(connectionString string) error {
	log.Println("Applying test migrations...")
	migration, err := migrate.New(
		"file://../database/migrations",
		connectionString)
	if err != nil {
		return fmt.Errorf("error creating migration: %v", err)
	}
	defer migration.Close()

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Test migrations completed successfully")
	return nil
}
