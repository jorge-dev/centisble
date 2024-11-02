package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jorge-dev/centsible/internal/config"
	"github.com/jorge-dev/centsible/internal/database"
	"github.com/jorge-dev/centsible/server"
)

func gracefulShutdown(ctx context.Context, apiServer *http.Server, dbService database.Service) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	<-ctx.Done() // Listen for the interrupt signal

	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	// Set a timeout to allow ongoing requests to finish within 5 seconds
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Close the database connection
	if err := dbService.Close(timeoutCtx); err != nil {
		log.Printf("Error closing database connection: %v", err)
	} else {
		log.Println("Database connection closed successfully")
	}

	if err := apiServer.Shutdown(timeoutCtx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	} else {
		log.Println("Server exited cleanly.")
	}
}

func main() {
	ctx := context.Background()
	config.Get().PrintBannerFromFile()
	httpServer, serverImpl := server.NewServer(ctx)

	// Run graceful shutdown in a separate goroutine
	go func() {
		gracefulShutdown(ctx, httpServer, serverImpl.GetDB())

	}()

	log.Printf("Server is configured to listen on %s", httpServer.Addr)

	// Start the server
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %s", err)
	}

	log.Println("Graceful shutdown complete.")
}
