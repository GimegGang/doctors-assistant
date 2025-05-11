package entity

import (
	"context"
	medicineProto "kode/internal/transport/grpc/generated"
)

type MedServiceInterface interface {
	AddSchedule(ctx context.Context, name string, userId int64, takingDuration, treatmentDuration int32) (int64, error)
	Schedules(ctx context.Context, userId int64) ([]int64, error)
	Schedule(ctx context.Context, userId, scheduleId int64) (*Medicine, error)
	NextTakings(ctx context.Context, userId int64) ([]*medicineProto.Medicines, error)
}
