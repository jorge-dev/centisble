package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jorge-dev/centsible/server"
)

func gracefulShutdown(apiServer *http.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // Listen for the interrupt signal

	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	// Set a timeout to allow ongoing requests to finish within 5 seconds
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(timeoutCtx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	} else {
		log.Println("Server exited cleanly.")
	}
}

func main() {
	server := server.NewServer()

	// Run graceful shutdown in a separate goroutine
	go func() {
		gracefulShutdown(server)
	}()

	log.Printf("Server is configured to listen on %s", server.Addr)

	// Start the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %s", err)
	}

	log.Println("Graceful shutdown complete.")
}
