package medService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"kode/internal/component/reception"
	"kode/internal/entity"
	medicineProto "kode/internal/transport/grpc/generated"
	"kode/internal/transport/rest/restMiddleware"
)

type medStorage interface {
	AddMedicine(ctx context.Context, schedule entity.Medicine) (int64, error)
	GetMedicines(ctx context.Context, medId int64) ([]int64, error)
	GetMedicine(ctx context.Context, id int64) (*entity.Medicine, error)
	GetMedicinesByUserID(ctx context.Context, userID int64) ([]*entity.Medicine, error)
}

type MedService struct {
	log     *slog.Logger
	storage medStorage
	period  time.Duration
}

func New(log *slog.Logger, storage medStorage, period time.Duration) *MedService {
	return &MedService{
		log:     log,
		storage: storage,
		period:  period,
	}
}

var timeNow = time.Now

func (m *MedService) AddSchedule(ctx context.Context, name string, userId int64, takingDuration, treatmentDuration int32) (int64, error) {
	const fun = "medService.AddSchedule"
	log := m.serviceLogger(ctx, fun)

	if name == "" || userId <= 0 || takingDuration <= 0 || treatmentDuration <= 0 {
		log.Error("invalid input data")
		return 0, errors.New("invalid input data")
	}

	med := entity.Medicine{Name: name,
		UserId:            userId,
		TakingDuration:    takingDuration,
		TreatmentDuration: treatmentDuration,
	}

	id, err := m.storage.AddMedicine(ctx, med)
	if err != nil {
		log.Error("Error adding medicine", "error", err)
		return 0, err
	}

	return id, err
}

func (m *MedService) Schedules(ctx context.Context, userId int64) ([]int64, error) {
	const fun = "medService.Schedules"
	log := m.serviceLogger(ctx, fun)

	if userId <= 0 {
		log.Error("ID must be greater 0")
		return nil, errors.New("ID must be greater 0")
	}

	ids, err := m.storage.GetMedicines(ctx, userId)
	if err != nil {
		log.Error("Error getting medicines", "error", err)
		return nil, err
	}

	return ids, err
}

func (m *MedService) Schedule(ctx context.Context, userId, scheduleId int64) (*entity.Medicine, error) {
	const fun = "medService.Schedule"
	log := m.serviceLogger(ctx, fun)

	med, err := m.storage.GetMedicine(ctx, scheduleId)
	if err != nil {
		log.Error("Error getting medicine", "error", err)
		return nil, err
	}
	if med == nil {
		return nil, entity.ErrNotFound
	}
	if med.UserId != userId {
		return nil, fmt.Errorf("schedule does not belong to the user")
	}
	med.Schedule, err = reception.GetReceptionIntake(med.TakingDuration)
	if err != nil {
		return nil, err
	}
	return med, nil
}

func (m *MedService) NextTakings(ctx context.Context, userId int64) ([]*medicineProto.Medicines, error) {
	const fun = "medService.NextTakings"
	log := m.serviceLogger(ctx, fun)

	med, err := m.storage.GetMedicinesByUserID(ctx, userId)
	if err != nil {
		log.Error("Error getting medicines", "error", err)
		return nil, err
	}
	// TODO подумать над переводом логики ниже в отдельный компонен для упрощения чтения
	var res []*medicineProto.Medicines
	now := timeNow()
	period := now.Add(m.period)
	for _, medId := range med {
		rec, err := reception.GetReceptionIntake(medId.TakingDuration)
		if err != nil {
			return nil, err
		}
		for _, t := range rec {
			intakeTime, err := time.Parse("15:04", t)
			if err != nil {
				log.Error("Error parsing time", "error", err)
				return nil, err
			}
			intakeToday := time.Date(
				now.Year(), now.Month(), now.Day(),
				intakeTime.Hour(), intakeTime.Minute(), intakeTime.Second(), 0, now.Location(),
			)
			if intakeToday.Before(now) {
				intakeToday = intakeToday.Add(24 * time.Hour)
			}
			if intakeToday.After(now) && intakeToday.Before(period) {
				res = append(res, &medicineProto.Medicines{Name: medId.Name, Times: t})
			}
		}
	}
	return res, nil
}

func (m *MedService) serviceLogger(ctx context.Context, fun string) *slog.Logger {
	log := m.log.With(slog.String("fun", fun))

	if traceID := restMiddleware.GetTraceID(ctx); traceID != "" {
		log = log.With(slog.String("trace-id", traceID))
	}

	return log
}
