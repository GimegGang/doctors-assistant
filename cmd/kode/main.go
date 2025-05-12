package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"kode/internal/app"
	"kode/internal/config"
	"kode/internal/entity"
	"kode/internal/infrastructure/persistence/postgres"
	"kode/internal/infrastructure/persistence/sqlite"
	"kode/internal/service/medService"
	"kode/pkg/logger"
)

func main() {
	cfg := config.MustLoad("config/config.yaml")
	log := logger.New(cfg.Env)
	log.Info("config is loaded", slog.Any("config", cfg))

	var db entity.StorageInterface
	var err error

	switch cfg.Env {
	case "prod":
		db, err = postgres.New("host=localhost port=5432 user=postgres dbname=postgres sslmode=disable")
		log.Info("Using PostgreSQL database")
	default:
		db, err = sqlite.New(cfg.StoragePath)
		log.Info("Using SQLite database")
	}

	if err != nil {
		log.Error("Error opening database", "error", err)
		os.Exit(1)
	}

	service := medService.New(log, db, cfg.TimePeriod)
	application := app.New(log, cfg, service)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err = application.Start(); err != nil {
			log.Error("app failed")
			stop()
		}
	}()

	<-ctx.Done()
	application.Stop()
}
