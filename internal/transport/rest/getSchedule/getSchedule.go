package getSchedule

import (
	"github.com/gin-gonic/gin"
	"kode/internal/service"
	"log/slog"
	"net/http"
	"strconv"
)

func GetScheduleHandler(log *slog.Logger, service service.MedServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handler.GetScheduleHandler"
		logger := log.With(
			slog.String("fun", fun),
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

		med, err := service.Schedule(c, userId, id)
		if err != nil {
			logger.Error("get schedule error", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, med)
	}
}
