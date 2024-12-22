package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jorge-dev/centsible/internal/config"
	"github.com/jorge-dev/centsible/internal/database"
	"github.com/jorge-dev/centsible/internal/logger"
	"github.com/jorge-dev/centsible/server"
)

func gracefulShutdown(ctx context.Context, apiServer *http.Server, dbService database.Service) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	slog.Info("Shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := dbService.Close(timeoutCtx); err != nil {
		slog.Error("Error closing database connection", "error", err)
	} else {
		slog.Info("Database connection closed successfully")
	}

	if err := apiServer.Shutdown(timeoutCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server exited cleanly")
	}
}

func main() {
	cfg := logger.LogConfig{
		Level:      config.ParseLogLevel(config.Get().Logging.Level),
		JSONOutput: config.Get().AppEnv == "prod",
	}
	logger.InitLogger(cfg)

	ctx := context.Background()

	config.Get().PrintBannerFromFile()

	httpServer, serverImpl := server.NewServer(ctx)

	if httpServer == nil || serverImpl == nil {
		slog.Error("Server configuration is invalid")
	}

	go func() {
		gracefulShutdown(ctx, httpServer, serverImpl.GetDB())

	}()

	slog.Info("Server is configured to listen on", slog.String("address", httpServer.Addr))

	// Start the server
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error:", slog.String("errorMsg", err.Error()))

	}

	slog.Info("Graceful shutdown complete.")
}
