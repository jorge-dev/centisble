package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/database"
)

type Server struct {
	port int
	db   database.Service
}

// GetDB returns the database service
func (s *Server) GetDB() database.Service {
	return s.db
}

func NewServer(ctx context.Context) (*http.Server, *Server) {

	// get the jwt secret from the environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	jwtManager := auth.NewJWTManager(jwtSecret)
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	db := database.New(ctx)
	serverImpl := &Server{
		port: port,
		db:   db,
	}

	if err != nil {
		log.Fatalf("Unable to get database connection: %v", err)
	}

	// Declare Server config
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverImpl.port),
		Handler:      serverImpl.RegisterRoutes(serverImpl.db.GetConnection(), *jwtManager),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer, serverImpl
}
