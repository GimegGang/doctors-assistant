package getNextTakings

import (
	"KODE_test/internal/reception"
	"KODE_test/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
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
		if err != nil || id < 1 {
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

		now := time.Now()
		nextPeriod := now.Add(period)

		var res getTakingResponse

		for _, m := range medicines {
			periods := reception.GetReceptionIntake(m)

			for _, p := range periods {
				intakeTime, err := time.Parse("15:04", p)

				if err != nil {
					log.Error("error parsing duration")
					w.WriteHeader(http.StatusInternalServerError)
					render.JSON(w, r, "internal server error")
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
					res.Medicines = append(res.Medicines, medicine{Name: m.Name, Time: p})
				}
			}
		}

		render.JSON(w, r, res)
	}
}
