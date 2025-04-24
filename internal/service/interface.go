package service

import (
	"context"
	"kode/internal/storage"
	medicineProto "kode/proto/gen"
)

type MedServiceInterface interface {
	AddSchedule(ctx context.Context, name string, userId int64, takingDuration, treatmentDuration int32) (int64, error)
	Schedules(ctx context.Context, userId int64) ([]int64, error)
	Schedule(ctx context.Context, userId, scheduleId int64) (*storage.Medicine, error)
	NextTakings(ctx context.Context, userId int64) ([]*medicineProto.Medicines, error)
}
