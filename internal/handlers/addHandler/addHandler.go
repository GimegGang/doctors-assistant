package addHandler

import (
	"github.com/gin-gonic/gin"
	"kode/internal/storage"
	"log/slog"
	"net/http"
)

type addSchedule interface {
	AddMedicine(schedule storage.Medicine) (int64, error)
}

type addScheduleResponse struct {
	Id int64 `json:"id"`
}

func AddScheduleHandler(log *slog.Logger, db addSchedule) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handler.AddScheduleHandler"
		log = log.With(
			slog.String("fun", fun),
			slog.String("request_id", c.GetHeader("X-Request-ID")),
		)

		var req storage.Medicine

		if err := c.BindJSON(&req); err != nil {
			log.Error("error decoding request body", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.Name == "" || req.TreatmentDuration <= 0 || req.TakingDuration <= 0 || req.UserId <= 0 {
			log.Error("invalid request", "req", req)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		id, err := db.AddMedicine(req)
		if err != nil {
			log.Error("error adding schedule", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, addScheduleResponse{Id: id})
	}
}
