package reception

import (
	"KODE_test/internal/storage"
	"testing"
)

func TestGetReceptionIntake(t *testing.T) {
	tests := []struct {
		name        string
		input       *storage.Medicine
		expectedLen int
	}{
		{
			name: "normal case",
			input: &storage.Medicine{
				TakingDuration: 3,
			},
			expectedLen: 3,
		},
		{
			name:        "empty case",
			input:       nil,
			expectedLen: 0,
		},
		{
			name: "invalid case",
			input: &storage.Medicine{
				TakingDuration: -1,
			},
			expectedLen: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := GetReceptionIntake(tc.input)
			if len(res) != tc.expectedLen {
				t.Errorf("GetReceptionIntake() returned wrong length: got %v want %v", len(res), tc.expectedLen)
			}
		})
	}
}
