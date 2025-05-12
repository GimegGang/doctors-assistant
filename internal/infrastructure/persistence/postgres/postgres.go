package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"kode/internal/entity"
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
	// TODO добавить миграции
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
	defer func() {
		_ = rows.Close()
	}()

	var res []int64

	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%s: %w", fun, err)
		}
		res = append(res, id)
	}

	if len(res) == 0 {
		return res, entity.ErrNotFound
	}

	return res, nil
}

func (s *StoragePostgres) AddMedicine(ctx context.Context, schedule entity.Medicine) (int64, error) {
	const fun = "internal/storage/postgres.AddMedicine"
	stmt, err := s.Prepare(`INSERT INTO medicine (name, taking_duration, treatment_duration, user_id, date) VALUES ($1, $2, $3, $4, $5) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	var lastID int64
	err = stmt.QueryRowContext(ctx, schedule.Name, schedule.TakingDuration, schedule.TreatmentDuration, schedule.UserId, time.Now()).Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}

	return lastID, nil
}

func (s *StoragePostgres) GetMedicine(ctx context.Context, id int64) (*entity.Medicine, error) {
	const fun = "internal/storage/postgres.GetMedicine"
	stmt, err := s.Prepare("SELECT * FROM medicine WHERE id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	var res entity.Medicine
	if err = stmt.QueryRowContext(ctx, id).Scan(&res.Id, &res.Name, &res.TakingDuration, &res.TreatmentDuration, &res.UserId, &res.Date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", fun, err)
	}

	if time.Now().After(res.Date.Add((time.Hour * 24) * time.Duration(res.TreatmentDuration))) {
		return nil, entity.ErrNotFound
	}

	return &res, nil
}

func (s *StoragePostgres) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*entity.Medicine, error) {
	const fun = "internal/storage/postgres.GetMedicinesByUserID"
	rows, err := s.QueryContext(ctx, `
        SELECT id, name, taking_duration, treatment_duration, user_id, date 
        FROM medicine 
        WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var medicines []*entity.Medicine
	for rows.Next() {
		var med entity.Medicine
		if err = rows.Scan(&med.Id, &med.Name, &med.TakingDuration, &med.TreatmentDuration, &med.UserId, &med.Date); err != nil {
			return nil, fmt.Errorf("%s: %w", fun, err)
		}
		if time.Now().Before(med.Date.Add((time.Hour * 24) * time.Duration(med.TreatmentDuration))) {
			medicines = append(medicines, &med)
		}
	}

	if len(medicines) == 0 {
		return nil, entity.ErrNotFound
	}

	return medicines, nil
}
