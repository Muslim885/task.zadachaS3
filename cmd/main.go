package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskS3/internal/config"
	"taskS3/internal/server"
	"taskS3/pkg/logger"
)

func main() {
	log, err := logger.NewSugared()
	if err != nil {
		os.Stderr.WriteString("CRITICAL: Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	srv, err := server.New(cfg, log.Desugar())
	if err != nil {
		log.Fatal("Failed to create server: ", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Infof("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed: ", err)
		}
	}()
	sig := <-quit
	log.Infof("Received signal: %v. Shutting down gracefully...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
