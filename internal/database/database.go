package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close(ctx context.Context) error

	// GetConnection returns the underlying database connection.
	GetConnection() *pgx.Conn
}

type dbService struct {
	conn *pgx.Conn
}

var (
	database     = os.Getenv("CENTSIBLE_DB_DATABASE")
	password     = os.Getenv("CENTSIBLE_DB_PASSWORD")
	username     = os.Getenv("CENTSIBLE_DB_USERNAME")
	port         = os.Getenv("CENTSIBLE_DB_PORT")
	host         = os.Getenv("CENTSIBLE_DB_HOST")
	schema       = os.Getenv("CENTSIBLE_DB_SCHEMA")
	dbInstance   *dbService
	runMigration = os.Getenv("RUN_MIGRATION") == "true" // Add this line
)

func New(ctx context.Context) Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)

	connection, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	// defer connection.Close(ctx)
	// db, err := sql.Conn("pgx", connStr)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Test the connection
	if err := connection.Ping(ctx); err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	dbInstance = &dbService{
		conn: connection,
	}

	// Only run migrations if flag is set
	if runMigration {
		if err := runMigrations(connStr); err != nil {
			log.Printf("Warning: failed to run migrations: %v", err)
			// Don't fatal here, just warn
		}
	}

	return dbInstance
}

// Add this helper function
func runMigrations(connectionString string) error {
	// driver, err := postgres.WithInstance(conn, &postgres.Config{})
	// if err != nil {
	// 	return fmt.Errorf("error creating driver: %v", err)
	// }

	log.Println("Applying migrations...")
	migration, err := migrate.New(
		"file://internal/database/migrations", // Update this line
		connectionString)
	if err != nil {
		return fmt.Errorf("error creating migration: %v", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *dbService) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// try to query the database
	_, err := s.conn.Exec(ctx, "SELECT 1")
	if err != nil {
		log.Printf("db down: %v", err)
	} else {
		log.Printf("db up")
	}

	// Ping the database
	err = s.conn.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err)
		return stats
	}

	// Database is up, add a basic health message
	stats["status"] = "up"
	stats["message"] = "The database connection is healthy."

	// Additional health indicators
	stats["connection_info"] = "Using single connection (pgx.Conn)"

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *dbService) Close(ctx context.Context) error {
	log.Printf("Disconnected from database: %s", database)
	return s.conn.Close(ctx)
}

func (s *dbService) GetConnection() *pgx.Conn {
	return s.conn
}
