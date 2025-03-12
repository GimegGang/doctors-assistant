package getNextTakings

import (
	"KODE_test/internal/reception"
	"KODE_test/internal/storage"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
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

var timeNow = time.Now // переменная для подмены в тестах

func GetNextTakingsHandler(log *slog.Logger, db getTakings, period time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fun = "handlers.NextTakingsHandler"
		log = log.With(slog.String("fun", fun), slog.String("request_id", middleware.GetReqID(r.Context())))

		strId := r.URL.Query().Get("user_id")
		if strId == "" {
			log.Error("missing parameter id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "missing parameter id")
			return
		}

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id < 0 {
			log.Error("invalid parameter id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "invalid parameter id")
			return
		}

		medicines, err := db.GetMedicinesByUserID(id)
		if err != nil {
			if errors.Is(err, storage.ErrNoRows) {
				log.Warn("Medicine not found", slog.Any("error", err))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, "Medicine not found")
				return
			}
			log.Error("error getting medicines")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "error getting medicines")
			return
		}

		now := timeNow() // применяю переменную для подмены в тестах
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
					render.JSON(w, r, res)
					return
				}
				res.Medicines = append(res.Medicines, med)

			case err, ok := <-errChan:
				if !ok {
					continue
				}
				log.Error("error in goroutine", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, "internal server error")
				return
			}
		}
	}
}
