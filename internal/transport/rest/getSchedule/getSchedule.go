package getSchedule

import (
	"github.com/gin-gonic/gin"
	"kode/internal/service"
	"kode/internal/transport/rest/restMiddleware"
	"log/slog"
	"net/http"
	"strconv"
)

func GetScheduleHandler(log *slog.Logger, service service.MedServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handler.GetScheduleHandler"
		log = log.With(slog.String("fun", fun), slog.String("trace-id", restMiddleware.GetTraceID(c.Request.Context())))

		strId := c.Query("schedule_id")
		if strId == "" {
			log.Error("missing parameter id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter schedule_id"})
			return
		}

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id <= 0 {
			log.Error("invalid parameter id", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter schedule_id"})
			return
		}

		userIdStr := c.Query("user_id")
		if userIdStr == "" {
			log.Error("missing parameter user_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter user_id"})
			return
		}

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil || userId < 1 {
			log.Error("invalid parameter user_id", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter user_id"})
			return
		}

		med, err := service.Schedule(c.Request.Context(), userId, id)
		if err != nil {
			log.Error("get schedule error", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Info("successful", slog.Any("request", struct {
			UserID     int64 `json:"userID"`
			ScheduleID int64 `json:"scheduleID"`
		}{UserID: userId, ScheduleID: id}), slog.Any("response", med))

		c.JSON(http.StatusOK, med)
	}
}
