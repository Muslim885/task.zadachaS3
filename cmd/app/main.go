package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"taskS3/internal/config"
	"taskS3/internal/server"
)

func main() {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	log := logger.Sugar()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Запуск сервера
	srv := server.New(cfg, logger)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запуск в горутине
	go func() {
		if err := srv.Run(); err != nil {
			log.Error("Server failed: ", err)
		}
	}()

	// Ожидание сигнала
	<-ctx.Done()

	log.Info("Shutting down gracefully...")

	// Graceful shutdown с таймаутом
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown: ", err)
	}

	log.Info("Server exited")
}
