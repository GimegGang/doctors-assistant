package sqlite

import (
	"context"
	"testing"
	"time"

	"kode/internal/entity"
)

func TestStorageSqlite(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now()
	userID := int64(1)

	medicine := entity.Medicine{
		Name:              "Aspirin",
		TakingDuration:    7,
		TreatmentDuration: 30,
		UserId:            userID,
		Date:              now,
	}

	t.Run("AddMedicine", func(t *testing.T) {
		id, err := db.AddMedicine(ctx, medicine)
		if err != nil {
			t.Fatalf("AddMedicine failed: %v", err)
		}
		if id <= 0 {
			t.Error("Expected positive ID, got", id)
		}
		medicine.Id = id
	})

	t.Run("GetMedicine", func(t *testing.T) {
		got, err := db.GetMedicine(ctx, medicine.Id)
		if err != nil {
			t.Fatalf("GetMedicine failed: %v", err)
		}

		if got.Name != medicine.Name {
			t.Errorf("Expected name %q, got %q", medicine.Name, got.Name)
		}
		if got.TakingDuration != medicine.TakingDuration {
			t.Errorf("Expected taking duration %d, got %d", medicine.TakingDuration, got.TakingDuration)
		}
	})

	t.Run("GetMedicine not found", func(t *testing.T) {
		_, err := db.GetMedicine(ctx, 9999)
		if err != entity.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("GetMedicines", func(t *testing.T) {
		ids, err := db.GetMedicines(ctx, userID)
		if err != nil {
			t.Fatalf("GetMedicines failed: %v", err)
		}
		if len(ids) == 0 {
			t.Error("Expected at least one medicine ID, got none")
		}
		found := false
		for _, id := range ids {
			if id == medicine.Id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find medicine ID %d in results", medicine.Id)
		}
	})

	t.Run("GetMedicinesByUserID", func(t *testing.T) {
		meds, err := db.GetMedicinesByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("GetMedicinesByUserID failed: %v", err)
		}
		if len(meds) == 0 {
			t.Error("Expected at least one medicine, got none")
		}
		found := false
		for _, m := range meds {
			if m.Id == medicine.Id {
				found = true
				if m.Name != medicine.Name {
					t.Errorf("Expected name %q, got %q", medicine.Name, m.Name)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected to find medicine ID %d in results", medicine.Id)
		}
	})
}
