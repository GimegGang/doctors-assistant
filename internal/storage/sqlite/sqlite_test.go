package sqlite

import (
	"errors"
	"kode/internal/storage"
	"os"
	"testing"
)

func setup(t *testing.T) (*Storage, func()) {
	tmpFile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	return db, func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}
}

func TestAddMedicine(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	med := storage.Medicine{
		Name:              "Aspirin",
		TakingDuration:    3,
		TreatmentDuration: 7,
		UserId:            1,
	}

	id, err := db.AddMedicine(med)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected ID > 0, got %d", id)
	}

	retrievedMed, err := db.GetMedicine(id)
	if err != nil {
		t.Fatalf("GetMedicine failed: %v", err)
	}
	if retrievedMed.Name != med.Name {
		t.Errorf("Expected name %s, got %s", med.Name, retrievedMed.Name)
	}
}

func TestGetMedicine(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	med := storage.Medicine{
		Name:              "Ibuprofen",
		TakingDuration:    2,
		TreatmentDuration: 5,
		UserId:            2,
	}

	id, err := db.AddMedicine(med)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}

	retrievedMed, err := db.GetMedicine(id)
	if err != nil {
		t.Fatalf("GetMedicine failed: %v", err)
	}

	if retrievedMed.Name != med.Name {
		t.Errorf("Expected name %s, got %s", med.Name, retrievedMed.Name)
	}
	if retrievedMed.TakingDuration != med.TakingDuration {
		t.Errorf("Expected taking duration %d, got %d", med.TakingDuration, retrievedMed.TakingDuration)
	}
	if retrievedMed.UserId != med.UserId {
		t.Errorf("Expected user ID %d, got %d", med.UserId, retrievedMed.UserId)
	}
}

func TestGetMedicinesByUserID(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	med1 := storage.Medicine{
		Name:              "Aspirin",
		TakingDuration:    3,
		TreatmentDuration: 7,
		UserId:            1,
	}
	med2 := storage.Medicine{
		Name:              "Ibuprofen",
		TakingDuration:    2,
		TreatmentDuration: 5,
		UserId:            1,
	}

	_, err := db.AddMedicine(med1)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}
	_, err = db.AddMedicine(med2)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}

	medicines, err := db.GetMedicinesByUserID(1)
	if err != nil {
		t.Fatalf("GetMedicinesByUserID failed: %v", err)
	}

	if len(medicines) != 2 {
		t.Errorf("Expected 2 medicines, got %d", len(medicines))
	}

	for _, med := range medicines {
		if med.UserId != 1 {
			t.Errorf("Expected user ID 1, got %d", med.UserId)
		}
	}
}

func TestGetMedicines(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	med1 := storage.Medicine{
		Name:              "Aspirin",
		TakingDuration:    3,
		TreatmentDuration: 7,
		UserId:            1,
	}
	med2 := storage.Medicine{
		Name:              "Ibuprofen",
		TakingDuration:    2,
		TreatmentDuration: 5,
		UserId:            1,
	}

	id1, err := db.AddMedicine(med1)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}
	id2, err := db.AddMedicine(med2)
	if err != nil {
		t.Fatalf("AddMedicine failed: %v", err)
	}

	ids, err := db.GetMedicines(1)
	if err != nil {
		t.Fatalf("GetMedicines failed: %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 ids, got %d", len(ids))
	}

	found := false
	for _, id := range ids {
		if *id == id1 || *id == id2 {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected ids %d and %d, got %v", id1, id2, ids)
	}
}

func TestGetMedicine_NotFound(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.GetMedicine(-1)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, storage.ErrNoRows) {
		t.Errorf("Expected ErrNoRows, got %v", err)
	}
}

func TestGetMedicinesByUserID_NotFound(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.GetMedicinesByUserID(-1)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, storage.ErrNoRows) {
		t.Errorf("Expected ErrNoRows, got %v", err)
	}
}
