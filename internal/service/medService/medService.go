package medService

import (
	"context"
	"fmt"
	"kode/internal/reception"
	"kode/internal/storage"
	medicineProto "kode/proto/gen"
	"log/slog"
	"time"
)

type medStorage interface {
	AddMedicine(ctx context.Context, schedule storage.Medicine) (int64, error)
	GetMedicines(ctx context.Context, medId int64) ([]int64, error)
	GetMedicine(ctx context.Context, id int64) (*storage.Medicine, error)
	GetMedicinesByUserID(ctx context.Context, userID int64) ([]*storage.Medicine, error)
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

func (m *MedService) AddSchedule(ctx context.Context, name string, userId int64, takingDuration, treatmentDuration int32) (int64, error) {
	const fun = "medService.AddSchedule"
	log := m.log.With(slog.String("fun", fun))

	med := storage.Medicine{Name: name,
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
	log := m.log.With(slog.String("fun", fun))

	ids, err := m.storage.GetMedicines(ctx, userId)
	if err != nil {
		log.Error("Error getting medicines", "error", err)
		return nil, err
	}

	return ids, err
}

func (m *MedService) Schedule(ctx context.Context, userId, scheduleId int64) (*storage.Medicine, error) {
	const fun = "medService.Schedule"
	log := m.log.With(slog.String("fun", fun))
	med, err := m.storage.GetMedicine(ctx, scheduleId)
	if err != nil {
		log.Error("Error getting medicine", "error", err)
		return nil, err
	}
	if med == nil {
		return nil, storage.ErrNotFound
	}
	if med.UserId != userId {
		return nil, fmt.Errorf("schedule does not belong to the user")
	}
	med.Schedule = reception.GetReceptionIntake(med.TakingDuration)
	return med, err
}

func (m *MedService) NextTakings(ctx context.Context, userId int64) ([]*medicineProto.Medicines, error) {
	const fun = "medService.NextTakings"
	log := m.log.With(slog.String("fun", fun))

	med, err := m.storage.GetMedicinesByUserID(ctx, userId)
	if err != nil {
		log.Error("Error getting medicines", "error", err)
		return nil, err
	}

	var res []*medicineProto.Medicines
	now := time.Now()
	period := now.Add(m.period)
	for _, medId := range med {
		for _, t := range reception.GetReceptionIntake(medId.TakingDuration) {
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
