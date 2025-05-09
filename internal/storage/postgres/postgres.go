package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"kode/internal/storage"
	"time"
)

type StoragePostgres struct {
	*sql.DB
}

func New(url string) (*StoragePostgres, error) {
	const fun = "internal/storage/postgres.New"

	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	//TODO добавить миграции
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS medicine (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			taking_duration BIGINT NOT NULL,
			treatment_duration BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			date TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create medicine table: %w", fun, err)
	}

	return &StoragePostgres{db}, nil
}

func (s *StoragePostgres) GetMedicines(ctx context.Context, medId int64) ([]int64, error) {
	const fun = "internal/storage/postgres.GetMedicines"
	rows, err := s.QueryContext(ctx, "SELECT id FROM medicine WHERE user_id = $1", medId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer rows.Close()

	var res []int64

	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%s: %w", fun, err)
		}
		res = append(res, id)
	}

	if len(res) == 0 {
		return res, storage.ErrNotFound
	}

	return res, nil
}

func (s *StoragePostgres) AddMedicine(ctx context.Context, schedule storage.Medicine) (int64, error) {
	const fun = "internal/storage/postgres.AddMedicine"
	stmt, err := s.Prepare(`INSERT INTO medicine (name, taking_duration, treatment_duration, user_id, date) VALUES ($1, $2, $3, $4, $5) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	var lastID int64
	err = stmt.QueryRowContext(ctx, schedule.Name, schedule.TakingDuration, schedule.TreatmentDuration, schedule.UserId, time.Now()).Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}

	return lastID, nil
}

func (s *StoragePostgres) GetMedicine(ctx context.Context, id int64) (*storage.Medicine, error) {
	const fun = "internal/storage/postgres.GetMedicine"
	stmt, err := s.Prepare("SELECT * FROM medicine WHERE id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	var res storage.Medicine
	if err = stmt.QueryRowContext(ctx, id).Scan(&res.Id, &res.Name, &res.TakingDuration, &res.TreatmentDuration, &res.UserId, &res.Date); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", fun, err)
	}

	if time.Now().After(res.Date.Add((time.Hour * 24) * time.Duration(res.TreatmentDuration))) {
		return nil, storage.ErrNotFound
	}

	return &res, nil
}

func (s *StoragePostgres) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*storage.Medicine, error) {
	const fun = "internal/storage/postgres.GetMedicinesByUserID"
	rows, err := s.QueryContext(ctx, `
        SELECT id, name, taking_duration, treatment_duration, user_id, date 
        FROM medicine 
        WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer rows.Close()

	var medicines []*storage.Medicine
	for rows.Next() {
		var med storage.Medicine
		if err = rows.Scan(&med.Id, &med.Name, &med.TakingDuration, &med.TreatmentDuration, &med.UserId, &med.Date); err != nil {
			return nil, fmt.Errorf("%s: %w", fun, err)
		}
		if time.Now().Before(med.Date.Add((time.Hour * 24) * time.Duration(med.TreatmentDuration))) {
			medicines = append(medicines, &med)
		}
	}

	if len(medicines) == 0 {
		return nil, storage.ErrNotFound
	}

	return medicines, nil
}
