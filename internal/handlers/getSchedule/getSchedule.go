package getSchedule

import (
	"KODE_test/internal/reception"
	"KODE_test/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type getSchedule interface {
	GetMedicine(id int64) (*storage.Medicine, error)
}

func GetScheduleHandler(log *slog.Logger, db getSchedule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fun = "handler.GetScheduleHandler"
		logger := log.With(
			slog.String("fun", fun),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		strId := r.URL.Query().Get("schedule_id")
		if strId == "" {
			logger.Error("missing parameter id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "missing parameter id")
			return
		}

		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil || id < 1 {
			logger.Error("invalid parameter id", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "invalid parameter id")
			return
		}

		med, err := db.GetMedicine(id)
		if err != nil {
			if errors.Is(err, storage.ErrNoRows) {
				log.Warn("Medicine not found", slog.Any("error", err))
				render.JSON(w, r, "Medicine not found")
				return
			}
			logger.Error("error getting medicine", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "internal server error")
			return
		}

		if med == nil {
			render.JSON(w, r, "Medicine not found")
			return
		}

		schedule := reception.GetReceptionIntake(med)
		med.Schedule = schedule

		render.JSON(w, r, med)
	}
}
