package main

import (
	"context"
	"kode/internal/app"
	"kode/internal/config"
	"kode/internal/logger"
	"kode/internal/service/medService"
	"kode/internal/storage/postgres"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad("config/config.yaml")
	log := logger.New(cfg.Env)
	log.Info("config is loaded", slog.Any("config", cfg))

	//db, err := sqlite.New(cfg.StoragePath) //sqlite
	db, err := postgres.New("host=localhost port=5432 user=gimeg dbname=postgres sslmode=disable") //postgres

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
