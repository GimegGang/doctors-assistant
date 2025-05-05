package reception

import (
	"fmt"
	"slices"
	"testing"
)

func TestReception(t *testing.T) {
	tests := []struct {
		name   string
		input  int32
		output []string
		err    error
	}{
		{
			name:  "normal case",
			input: 15,
			output: []string{
				"08:00",
				"09:00",
				"10:00",
				"11:00",
				"12:00",
				"13:00",
				"14:00",
				"15:00",
				"16:00",
				"17:00",
				"18:00",
				"19:00",
				"20:00",
				"21:00",
				"22:00",
			},
			err: nil,
		},
		{
			name:  "first border",
			input: 1,
			output: []string{
				"08:00",
			},
			err: nil,
		},
		{
			name:  "end border",
			input: 57, // так как у нас 14 * 4 + 1 промежутков времени без повторений
			output: []string{
				"08:00", "08:15", "08:30", "08:45", "09:00", "09:15", "09:30", "09:45", "10:00", "10:15", "10:30", "10:45",
				"11:00", "11:15", "11:30", "11:45", "12:00", "12:15", "12:30", "12:45", "13:00", "13:15", "13:30", "13:45",
				"14:00", "14:15", "14:30", "14:45", "15:00", "15:15", "15:30", "15:45", "16:00", "16:15", "16:30", "16:45",
				"17:00", "17:15", "17:30", "17:45", "18:00", "18:15", "18:30", "18:45", "19:00", "19:15", "19:30", "19:45",
				"20:00", "20:15", "20:30", "20:45", "21:00", "21:15", "21:30", "21:45", "22:00",
			},
			err: nil,
		},
		{
			name:   "zero test",
			input:  0,
			output: nil,
			err:    fmt.Errorf("receptions time must be greate 0"),
		},
		{
			name:   "less zero test",
			input:  -5,
			output: nil,
			err:    fmt.Errorf("receptions time must be greate 0"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := GetReceptionIntake(tc.input)
			if !slices.Equal(res, tc.output) {
				t.Errorf("GetReceptionIntake() returned wrong res: \ngot %v \nwant %v", res, tc.output)
			}
			if (err == nil) != (tc.err == nil) {
				t.Errorf("GetReceptionIntake() error presence mismatch:\ngot  %v\nwant %v", err, tc.err)
			}
			if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Errorf("GetReceptionIntake() error message mismatch:\ngot  %q\nwant %q", err.Error(), tc.err.Error())
			}
		})
	}
}
