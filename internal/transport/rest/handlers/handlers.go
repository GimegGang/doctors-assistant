package handlers

import (
	"github.com/gin-gonic/gin"
	"kode/internal/service"
	"kode/internal/transport/rest/api"
	"kode/internal/transport/rest/restMiddleware"
	"log/slog"
	"net/http"
)

type RequestHandler struct {
	log     *slog.Logger
	service service.MedServiceInterface
}

func New(log *slog.Logger, service service.MedServiceInterface) *RequestHandler {
	return &RequestHandler{
		log:     log,
		service: service,
	}
}

func (h *RequestHandler) GetNextTakings(c *gin.Context, params api.GetNextTakingsParams) {
	const fun = "handlers.GetNextTakings"
	log := h.log.With(slog.String("fun", fun), slog.String("trace-id", restMiddleware.GetTraceID(c.Request.Context())))

	medicines, err := h.service.NextTakings(c, params.UserId)
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

	var response []*api.NextTaking
	for _, m := range medicines {
		response = append(response, &api.NextTaking{
			Name: &m.Name,
			Time: &m.Times,
		})
	}
	log.Info("successful", slog.Int64("userId", params.UserId))
	c.JSON(http.StatusOK, response)
}

func (h *RequestHandler) GetSchedule(c *gin.Context, params api.GetScheduleParams) {
	const fun = "handlers.GetSchedule"
	log := h.log.With(slog.String("fun", fun), slog.String("trace-id", restMiddleware.GetTraceID(c.Request.Context())))

	medicine, err := h.service.Schedule(c, params.UserId, params.ScheduleId)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := api.Medicine{
		Id:                &medicine.Id,
		Name:              medicine.Name,
		TakingDuration:    medicine.TakingDuration,
		TreatmentDuration: medicine.TakingDuration,
		Schedule:          &medicine.Schedule,
		UserId:            medicine.UserId,
		Date:              &medicine.Date,
	}

	log.Info("successful", slog.Int64("userId", params.UserId))
	c.JSON(http.StatusOK, response)
}

func (h *RequestHandler) PostSchedule(c *gin.Context) {
	const fun = "handlers.PostSchedule"
	log := h.log.With(slog.String("fun", fun), slog.String("trace-id", restMiddleware.GetTraceID(c.Request.Context())))

	var medicine api.Medicine
	if err := c.ShouldBindJSON(&medicine); err != nil {
		log.Error("error decoding request body", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	id, err := h.service.AddSchedule(c, medicine.Name, medicine.UserId, medicine.TakingDuration, medicine.TreatmentDuration)
	if err != nil {
		log.Error("error adding schedule", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	log.Info("schedule added", slog.Int64("id", id))
	c.JSON(http.StatusOK, api.AddScheduleResponse{Id: &id})
}

func (h *RequestHandler) GetSchedules(c *gin.Context, params api.GetSchedulesParams) {
	const fun = "handlers.GetSchedules"
	log := h.log.With(slog.String("fun", fun), slog.String("trace-id", restMiddleware.GetTraceID(c.Request.Context())))

	medicine, err := h.service.Schedules(c, params.UserId)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("successful", slog.Int64("userId", params.UserId))
	c.JSON(http.StatusOK, api.GetSchedulesResponse{SchedulesId: &medicine})
}
