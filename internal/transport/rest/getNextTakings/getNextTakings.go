package getNextTakings

import (
	"github.com/gin-gonic/gin"
	"kode/internal/service"
	"kode/internal/transport/rest/middleware"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type medicine struct {
	Name string `json:"name"`
	Time string `json:"times"`
}

type ByTime []medicine

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time < a[j].Time }

var timeNow = time.Now // переменная для подмены в тестах

func GetNextTakingsHandler(log *slog.Logger, service service.MedServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handlers.NextTakingsHandler"
		log = log.With(slog.String("fun", fun), slog.String("trace-id", middleware.GetTraceID(c.Request.Context())))

		strId := c.Query("user_id")
		if strId == "" {
			log.Error("missing parameter id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter id"})
			return
		}

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id < 0 {
			log.Error("invalid parameter id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter id"})
			return
		}

		medicines, err := service.NextTakings(c.Request.Context(), id)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if medicines == nil {
			log.Info("no next takings found")
			c.JSON(http.StatusNotFound, gin.H{"error": "no next takings found"})
			return
		}
		var response []*medicine
		for _, m := range medicines {
			response = append(response, &medicine{Name: m.Name, Time: m.Times})
		}
		c.JSON(http.StatusOK, response)
	}
}
