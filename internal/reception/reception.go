package reception

import (
	"KODE_test/internal/storage"
	"time"
)

func GetReceptionIntake(medicine *storage.Medicine) []string {
	if medicine == nil || medicine.TakingDuration <= 0 {
		return nil
	}

	start := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	end := start.Add(14 * time.Hour)

	schedule := make([]string, medicine.TakingDuration)

	if medicine.TakingDuration == 1 {
		schedule[0] = start.Format("15:04")
		return schedule
	}

	step := 14 * time.Hour / time.Duration(medicine.TakingDuration-1)

	for i := 0; i < medicine.TakingDuration; i++ {
		t := start.Add(time.Duration(i) * step)
		t = t.Round(15 * time.Minute)
		if t.After(end) {
			t = end
		}
		schedule[i] = t.Format("15:04")
	}

	return schedule
}
