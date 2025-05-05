package reception

import (
	"fmt"
	"time"
)

func GetReceptionIntake(takingDuration int32) ([]string, error) {
	if takingDuration <= 0 {
		return nil, fmt.Errorf("receptions time must be greate 0")
	}

	start := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	end := start.Add(14 * time.Hour)

	schedule := make([]string, takingDuration)

	if takingDuration == 1 {
		schedule[0] = start.Format("15:04")
		return schedule, nil
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

	return schedule, nil
}
