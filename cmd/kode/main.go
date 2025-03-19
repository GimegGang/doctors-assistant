package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"kode/internal/config"
	"kode/internal/handlers/addHandler"
	"kode/internal/handlers/getNextTakings"
	"kode/internal/handlers/getSchedule"
	"kode/internal/handlers/getSchedules"
	"kode/internal/logger"
	"kode/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoad("config/config.yaml")
	log := logger.MustLoad(cfg.Env)

	log.Info("config is loaded", slog.Any("config", cfg))

	db, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error opening database", "error", err)
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)

	router.Post("/schedule", addHandler.AddScheduleHandler(log, db))
	router.Get("/schedules", getSchedules.GetSchedulesHandler(log, db))
	router.Get("/schedule", getSchedule.GetScheduleHandler(log, db))
	router.Get("/", getNextTakings.GetNextTakingsHandler(log, db, cfg.TimePeriod))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("Start Server", slog.String("address", srv.Addr))
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "error", err)
	}
}
