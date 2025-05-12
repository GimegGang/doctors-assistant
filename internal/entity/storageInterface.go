package entity

import "context"

type StorageInterface interface {
	AddMedicine(ctx context.Context, schedule Medicine) (int64, error)
	GetMedicines(ctx context.Context, medId int64) ([]int64, error)
	GetMedicine(ctx context.Context, id int64) (*Medicine, error)
	GetMedicinesByUserID(ctx context.Context, userID int64) ([]*Medicine, error)
}
