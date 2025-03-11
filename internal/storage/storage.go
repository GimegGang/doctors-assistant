package storage

import (
	"errors"
	"time"
)

type Medicine struct {
	Id                int64     `json:"id"`
	Name              string    `json:"name"`
	TakingDuration    int       `json:"taking_duration"`
	TreatmentDuration int       `json:"treatment_duration"`
	UserId            int64     `json:"user_id"`
	Schedule          []string  `json:"schedule"`
	Date              time.Time `json:"date"`
}

// Errors
var (
	ErrNoRows = errors.New("sql: no rows in result set")
)
