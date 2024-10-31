package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jorge-dev/centsible/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *http.Server {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}
	NewServer := &Server{
		port: port,
		db:   database.New(),
	}

	// log.Printf("Configuring server on port %d", NewServer.port)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// log.Println("Server configuration complete")

	return server
}
