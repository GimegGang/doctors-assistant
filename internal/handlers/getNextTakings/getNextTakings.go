package getNextTakings

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"kode/internal/reception"
	"kode/internal/storage"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

type getTakings interface {
	GetMedicinesByUserID(userID int64) ([]*storage.Medicine, error)
}

type getTakingResponse struct {
	Medicines []medicine `json:"medicines"`
}

type medicine struct {
	Name string `json:"name"`
	Time string `json:"times"`
}

type ByTime []medicine

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time < a[j].Time }

var timeNow = time.Now // переменная для подмены в тестах

func GetNextTakingsHandler(log *slog.Logger, db getTakings, period time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		const fun = "handlers.NextTakingsHandler"
		log = log.With(
			slog.String("fun", fun),
			slog.String("request_id", c.GetHeader("X-Request-ID")),
		)

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

		medicines, err := db.GetMedicinesByUserID(id)
		if err != nil {
			if errors.Is(err, storage.ErrNoRows) {
				log.Warn("Medicine not found", slog.Any("error", err))
				c.JSON(http.StatusNotFound, gin.H{"error": "Medicine not found"})
				return
			}
			log.Error("error getting medicines")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting medicines"})
			return
		}

		now := timeNow()
		nextPeriod := now.Add(period)

		var res getTakingResponse

		resChan := make(chan medicine)
		errChan := make(chan error)

		var wg sync.WaitGroup

		for _, m := range medicines {
			periods := reception.GetReceptionIntake(m)

			for _, p := range periods {
				wg.Add(1)
				go func(m *storage.Medicine, p string) {
					defer wg.Done()
					intakeTime, err := time.Parse("15:04", p)

					if err != nil {
						errChan <- fmt.Errorf("error parsing medicine time: %w", err)
						return
					}

					intakeToday := time.Date(
						now.Year(), now.Month(), now.Day(),
						intakeTime.Hour(), intakeTime.Minute(), intakeTime.Second(), 0, now.Location(),
					)
					if intakeToday.Before(now) {
						intakeToday = intakeToday.Add(24 * time.Hour)
					}
					if intakeToday.After(now) && intakeToday.Before(nextPeriod) {
						resChan <- medicine{Name: m.Name, Time: p}
					}
				}(m, p)
			}
		}

		go func() {
			wg.Wait()
			close(resChan)
			close(errChan)
		}()

		for {
			select {
			case med, ok := <-resChan:
				if !ok {
					sort.Sort(ByTime(res.Medicines))
					c.JSON(http.StatusOK, res)
					return
				}
				res.Medicines = append(res.Medicines, med)

			case err, ok := <-errChan:
				if !ok {
					continue
				}
				log.Error("error in goroutine", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
	}
}
