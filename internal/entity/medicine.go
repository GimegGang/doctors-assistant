package entity

import (
	"time"
)

type Medicine struct {
	Id                int64     `json:"id"`
	Name              string    `json:"name"`
	TakingDuration    int32     `json:"taking_duration"`
	TreatmentDuration int32     `json:"treatment_duration"`
	UserId            int64     `json:"user_id"`
	Schedule          []string  `json:"schedule"`
	Date              time.Time `json:"date"`
}
