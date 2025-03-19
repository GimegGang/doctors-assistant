package addHandler

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
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

func AddScheduleHandler(log *slog.Logger, db addSchedule) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fun = "handler.AddScheduleHandler"
		log.With(slog.String("fun", fun), slog.String("request_id", middleware.GetReqID(r.Context())))

		var req storage.Medicine

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("error decoding request body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "invalid request")
			return
		}

		if req.Name == "" || req.TreatmentDuration <= 0 || req.TakingDuration <= 0 || req.UserId <= 0 {
			log.Error("invalid request", "req", req)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "invalid request")
		}

		id, err := db.AddMedicine(req)

		if err != nil {
			log.Error("error adding schedule", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "internal server error")
			return
		}

		render.JSON(w, r, addScheduleResponse{Id: id})
	}
}
