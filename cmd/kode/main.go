package main

import (
	"KODE_test/internal/config"
	"KODE_test/internal/handlers"
	"KODE_test/internal/logger"
	"KODE_test/internal/storage/sqlite"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		log.Error("Error opening database", err)
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)

	router.Post("/schedule", handlers.AddScheduleHandler(log, db))
	router.Get("/schedules", handlers.GetSchedulesHandler(log, db))
	router.Get("/schedule", handlers.GetScheduleHandler(log, db)) //в моей реализации параметр user_id не требуется
	router.Get("/next_takings", handlers.NextTakingsHandler(log, db, cfg.TimePeriod))

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
