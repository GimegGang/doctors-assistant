package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"kode/internal/app"
	"kode/internal/config"
	"kode/internal/logger"
	"kode/internal/service/medService"
	"kode/internal/storage/sqlite"
	"kode/internal/transport/rest/addHandler"
	"kode/internal/transport/rest/getNextTakings"
	"kode/internal/transport/rest/getSchedule"
	"kode/internal/transport/rest/getSchedules"
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

	service := medService.New(log, db, cfg.TimePeriod)

	router.POST("/schedule", addHandler.AddScheduleHandler(log, service))
	router.GET("/schedules", getSchedules.GetSchedulesHandler(log, service))
	router.GET("/schedule", getSchedule.GetScheduleHandler(log, service))
	router.GET("/next_takings", getNextTakings.GetNextTakingsHandler(log, service))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.RestAddress),
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("Start HTTP Server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	grpc := app.New(log, cfg.TimePeriod, cfg.GrpcAddress, service)
	go func() {
		if err := grpc.Start(); err != nil {
			log.Error("gRPC server failed", "error", err)
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
