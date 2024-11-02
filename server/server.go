package server

import (
	"context"
	"fmt"
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

func NewServer(ctx context.Context) (*http.Server, *Server) {
	cfg := config.Get()

	if cfg.AppEnv == "local" {
		fmt.Println("Running in local mode")
	}

	db := database.New(ctx)
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
