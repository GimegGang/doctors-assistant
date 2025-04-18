package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/app"
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
	"os/signal"
	"syscall"
	"time"
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

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-Request-ID"); id != "" {
			c.Header("X-Request-ID", id)
		}
		c.Next()
	})

	router.POST("/schedule", addHandler.AddScheduleHandler(log, db))
	router.GET("/schedules", getSchedules.GetSchedulesHandler(log, db))
	router.GET("/schedule", getSchedule.GetScheduleHandler(log, db))
	router.GET("/next_takings", getNextTakings.GetNextTakingsHandler(log, db, cfg.TimePeriod))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	grpc := app.New(log, cfg.TimePeriod, 1234, db)
	go func() {
		if err := grpc.Start(); err != nil {
			log.Error("gRPC server failed", "error", err)
		}
	}()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("Start HTTP Server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErr:
		log.Error("server error", "error", err)
		grpc.Stop()
		os.Exit(1)
	case sign := <-stop:
		log.Info("Shutting down", slog.Any("signal", sign))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("HTTP server shutdown error", "error", err)
		}

		grpc.Stop()
		log.Info("Server stopped")
	}
}
