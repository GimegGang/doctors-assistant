package getSchedule

import (
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/reception"
	"kode/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
)

type getSchedule interface {
	GetMedicine(id int64) (*storage.Medicine, error)
}

func GetScheduleHandler(log *slog.Logger, db getSchedule) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handler.GetScheduleHandler"
		logger := log.With(
			slog.String("fun", fun),
			slog.String("request_id", c.GetHeader("X-Request-ID")),
		)

		strId := c.Query("schedule_id")
		if strId == "" {
			logger.Error("missing parameter id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter schedule_id"})
			return
		}

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id <= 0 {
			logger.Error("invalid parameter id", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter schedule_id"})
			return
		}

		userIdStr := c.Query("user_id")
		if userIdStr == "" {
			logger.Error("missing parameter user_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter user_id"})
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil || userId < 1 {
			logger.Error("invalid parameter user_id", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter user_id"})
			return
		}

		med, err := db.GetMedicine(id)
		if err != nil {
			if errors.Is(err, storage.ErrNoRows) {
				logger.Warn("Medicine not found", slog.Any("error", err))
				c.JSON(http.StatusNotFound, gin.H{"error": "Medicine not found"})
				return
			}
			logger.Error("error getting medicine", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if med == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Medicine not found"})
			return
		}

		if userId != med.UserId {
			logger.Error("invalid user id")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid user id"})
			return
		}

		schedule := reception.GetReceptionIntake(med)
		med.Schedule = schedule

		c.JSON(http.StatusOK, med)
	}
}
