package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"kode/internal/entity"
	"time"
)

type StorageSqlite struct {
	*sql.DB
}

func New(storagePath string) (*StorageSqlite, error) {
	const fun = "internal/storage/sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS medicine (
    	id INTEGER PRIMARY KEY,
    	name TEXT NOT NULL,
    	taking_duration INTEGER NOT NULL,
    	treatment_duration INTEGER NOT NULL,
    	user_id INTEGER NOT NULL,
    	date DATE NOT NULL
	);`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create medicine table: %w", fun, err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_index ON medicine (user_id);`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to createй index on medicine: %w", fun, err)
	}

	return &StorageSqlite{db}, nil
}

func (s *StorageSqlite) GetMedicines(ctx context.Context, medId int64) ([]int64, error) {
	const fun = "internal/storage/mysql.GetSchedules"
	rows, err := s.QueryContext(ctx, "SELECT id FROM medicine WHERE user_id = ?", medId)
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
		return res, entity.ErrNotFound
	}

	return res, nil
}

func (s *StorageSqlite) AddMedicine(ctx context.Context, schedule entity.Medicine) (int64, error) {
	const fun = "internal/storage/mysql.AddSchedule"

	stmt, err := s.Prepare(`INSERT INTO medicine (name, taking_duration, treatment_duration, user_id, date) values (?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, schedule.Name, schedule.TakingDuration, schedule.TreatmentDuration, schedule.UserId, time.Now())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}

	return lastID, nil
}

func (s *StorageSqlite) GetMedicine(ctx context.Context, id int64) (*entity.Medicine, error) {
	const fun = "internal/storage/mysql.GetSchedule"

	stmt, err := s.Prepare("SELECT * FROM medicine WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	var res entity.Medicine

	if err = stmt.QueryRowContext(ctx, id).Scan(&res.Id, &res.Name, &res.TakingDuration, &res.TreatmentDuration, &res.UserId, &res.Date); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", fun, err)
	}

	if time.Now().After(res.Date.Add((time.Hour * 24) * time.Duration(res.TreatmentDuration))) {
		return nil, entity.ErrNotFound
	}

	return &res, nil
}

func (s *StorageSqlite) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*entity.Medicine, error) {
	const fun = "internal/storage/sqlite.GetMedicinesByUserID"

	rows, err := s.QueryContext(ctx, `
        SELECT id, name, taking_duration, treatment_duration, user_id, date 
        FROM medicine 
        WHERE user_id = ?
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer rows.Close()

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
