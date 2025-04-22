package getSchedules

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
)

type medService interface {
	Schedules(ctx context.Context, userId int64) ([]int64, error)
}

type getSchedulesResponse struct {
	Schedules []int64 `json:"schedules_id"`
}

func GetSchedulesHandler(log *slog.Logger, service medService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handler.GetSchedulesHandler"
		log.With(slog.String("fun", fun), slog.String("request_id", c.GetHeader("X-Request-ID")))

		userId := c.Query("user_id")
		if userId == "" {
			log.Error("missing parameter user_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter user_id"})
			return
		}

		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			log.Error("invalid parameter user_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter user_id"})
			return
		}

		schedules, err := service.Schedules(c, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Warn("Medicine not found", slog.Any("error", err))
				c.JSON(http.StatusOK, gin.H{"error": "Medicine not found"})
				return
			}
			log.Error("error getting schedules", "id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting schedules"})
			return
		}

		c.JSON(http.StatusOK, getSchedulesResponse{schedules})
	}
}
