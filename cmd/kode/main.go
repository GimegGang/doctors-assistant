package main

import (
	"errors"
	"github.com/gin-gonic/gin"
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

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Middleware (аналоги Chi)
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(func(c *gin.Context) {
		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
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

	log.Info("Start Server", slog.String("address", srv.Addr))
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "error", err)
	}
}
