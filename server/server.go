package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/config"
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

// Add this method to your Server struct
func (s *Server) SetDB(db database.Service) {
	s.db = db
}

func NewServer(ctx context.Context) (*http.Server, *Server) {
	cfg := config.Get()

	// Validate required configuration
	if cfg.Port <= 0 {
		log.Printf("Invalid port configuration: %d", cfg.Port)
		return nil, nil
	}

	if cfg.AppEnv == "local" {
		fmt.Println("Running in local mode")
	}

	var db database.Service
	if cfg.AppEnv == "test" {
		db = newMockDB()
	} else {
		db = database.New(ctx)
	}

	serverImpl := &Server{
		port: cfg.Port,
		db:   db,
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret)

	// Declare Server config
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverImpl.port),
		Handler:      serverImpl.RegisterRoutes(serverImpl.db.GetConnection(), *jwtManager, cfg.AppEnv),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer, serverImpl
}
