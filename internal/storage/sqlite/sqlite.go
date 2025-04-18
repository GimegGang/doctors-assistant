package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"kode/internal/storage"
	"time"
)

type Storage struct {
	*sql.DB
}

func New(storagePath string) (*Storage, error) {
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

	return &Storage{db}, nil
}

func (s *Storage) GetMedicines(medId int64) ([]int64, error) {
	const fun = "internal/storage/mysql.GetSchedules"
	rows, err := s.Query("SELECT id FROM medicine WHERE user_id = ?", medId)
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
		return res, storage.ErrNoRows
	}

	return res, nil
}

func (s *Storage) AddMedicine(schedule storage.Medicine) (int64, error) {
	const fun = "internal/storage/mysql.AddSchedule"

	stmt, err := s.Prepare(`INSERT INTO medicine (name, taking_duration, treatment_duration, user_id, date) values (?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(schedule.Name, schedule.TakingDuration, schedule.TreatmentDuration, schedule.UserId, time.Now())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fun, err)
	}

	return lastID, nil
}

func (s *Storage) GetMedicine(id int64) (*storage.Medicine, error) {
	const fun = "internal/storage/mysql.GetSchedule"

	stmt, err := s.Prepare("SELECT * FROM medicine WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fun, err)
	}
	defer stmt.Close()

	var res storage.Medicine

	if err = stmt.QueryRow(id).Scan(&res.Id, &res.Name, &res.TakingDuration, &res.TreatmentDuration, &res.UserId, &res.Date); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, storage.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", fun, err)
	}

	if time.Now().After(res.Date.Add((time.Hour * 24) * time.Duration(res.TreatmentDuration))) {
		return nil, nil
	}

	return &res, nil
}

func (s *Storage) GetMedicinesByUserID(userID int64) ([]*storage.Medicine, error) {
	const fun = "internal/storage/sqlite.GetMedicinesByUserID"

	rows, err := s.Query(`
        SELECT id, name, taking_duration, treatment_duration, user_id, date 
        FROM medicine 
        WHERE user_id = ?
    `, userID)
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
		return nil, storage.ErrNoRows
	}

	return medicines, nil
}
