package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/jorge-dev/centsible/internal/config"
)

//go:embed migrations/*.sql
var migrations embed.FS

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

var dbInstance *dbService

func New(ctx context.Context) Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	cfg := config.Get()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Schema,
	)

	connection, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	// Test the connection
	if err := connection.Ping(ctx); err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	dbInstance = &dbService{
		conn: connection,
	}

	// Only run migrations if flag is set
	if cfg.Database.RunMigration {
		if err := runMigrations(connStr); err != nil {
			log.Printf("Warning: failed to run migrations: %v", err)
		}
	}

	return dbInstance
}

// Add this helper function
func runMigrations(connectionString string) error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Printf("Failed to create migration source: %v", err)
		return fmt.Errorf("error creating migration source: %v", err)
	}

	log.Println("Applying migrations...")
	migration, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
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

	// Ping the database
	err := s.conn.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["message"] = "The database connection is unhealthy."
		stats["error"] = "db not responding"
		log.Printf("db down: %v", err)
		return stats
	}

	// try to query the database
	_, err = s.conn.Exec(ctx, "SELECT 1")
	if err != nil {
		log.Printf("db not responding after query: %v", err)
		stats["status"] = "degraded"
		stats["message"] = "The database connection is unhealthy."
		stats["error"] = fmt.Sprintf("Cant connect to database: %v", err)
	}

	// Database is up, add a basic health message
	stats["status"] = "up"
	stats["message"] = "The database connection is healthy."

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *dbService) Close(ctx context.Context) error {
	cfg := config.Get()
	log.Printf("Disconnected from database: %s", cfg.Database.Database)
	return s.conn.Close(ctx)
}

func (s *dbService) GetConnection() *pgx.Conn {
	return s.conn
}
