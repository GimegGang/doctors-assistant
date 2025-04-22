package reception

import (
	"time"
)

func GetReceptionIntake(takingDuration int32) []string {
	if takingDuration <= 0 {
		return nil
	}

	start := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	end := start.Add(14 * time.Hour)

	schedule := make([]string, takingDuration)

	if takingDuration == 1 {
		schedule[0] = start.Format("15:04")
		return schedule
	}

	step := 14 * time.Hour / time.Duration(takingDuration-1)

	for i := int32(0); i < takingDuration; i++ {
		t := start.Add(time.Duration(i) * step)
		t = t.Round(15 * time.Minute)
		if t.After(end) {
			t = end
		}
		schedule[i] = t.Format("15:04")
	}

	return schedule
}
