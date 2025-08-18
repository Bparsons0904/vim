package main

import (
	"context"
	"os"
	"os/signal"
	"server/internal/app"
	"server/internal/logger"
	"server/internal/server"
	"syscall"
	"time"
)

func gracefulShutdown(
	app *app.App,
	appServer *server.AppServer,
	done chan bool,
	log logger.Logger,
) {
	log = log.Function("gracefulShutdown")

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := appServer.FiberApp.ShutdownWithContext(ctx); err != nil {
		log.Er("Server forced to shutdown", err)
	}

	if err := app.Database.Close(); err != nil {
		log.Er("failed to close database", err)
	}

	log.Info("Server exiting")
	done <- true
}

func main() {
	log := logger.New("main")

	app, err := app.New()
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Er("failed to close app", err)
		}
	}()

	server, err := server.New(app)
	if err != nil {
		os.Exit(1)
	}

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	go func() {
		err := server.Listen(app.Config.ServerPort)
		if err != nil {
			os.Exit(1)
		}
	}()

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(app, server, done, log)

	// Wait for the graceful shutdown to complete
	<-done
	log.Info("Graceful shutdown complete.")
}
