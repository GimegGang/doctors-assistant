package handlers

import (
	"KODE_test/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type getSchedules interface {
	GetMedicines(medId int64) ([]*int64, error)
}

type getSchedulesResponse struct {
	Schedules []*int64 `json:"schedules_id"`
}

func GetSchedulesHandler(log *slog.Logger, db getSchedules) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fun = "handler.GetSchedulesHandler"
		log.With(slog.String("fun", fun), slog.String("request_id", middleware.GetReqID(r.Context())))

		userId := r.URL.Query().Get("user_id")
		if userId == "" {
			log.Error("missing parameter user_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "missing parameter user_id")
			return
		}

		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			log.Error("invalid parameter user_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "invalid parameter user_id")
			return
		}

		schedules, err := db.GetMedicines(id)
		if err != nil {
			if errors.Is(err, storage.ErrNoRows) {
				log.Warn("Medicine not found", slog.Any("error", err))
				render.JSON(w, r, "Medicine not found")
				return
			}
			log.Error("error getting schedules", "id", id, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "internal server error")
			return
		}

		render.JSON(w, r, getSchedulesResponse{Schedules: schedules})
	}
}
