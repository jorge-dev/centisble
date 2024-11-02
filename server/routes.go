package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/handlers"
	customMiddleware "github.com/jorge-dev/centsible/server/middleware"
)

func (s *Server) RegisterRoutes(conn *pgx.Conn, jwtManager auth.JWTManager) http.Handler {
	queries := repository.New(conn)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", s.HelloWorldHandler)
		r.Get("/health", s.healthHandler)

	})

	// Auth routes
	r.Group(func(r chi.Router) {
		authHandler := handlers.NewAuthHandler(queries, &jwtManager)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Private routes
	r.Group(func(r chi.Router) {
		authMiddleware := customMiddleware.NewAuthMiddleware(&jwtManager)
		r.Use(authMiddleware.AuthRequired)
		// Add private
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
