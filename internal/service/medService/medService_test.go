package medService

import (
	"context"
	"errors"
	"kode/internal/entity"
	"kode/pkg/logger"
	"slices"
	"testing"
	"time"
)

type mockStorage struct {
	shouldError bool
	shouldEmpty bool
}

func (m *mockStorage) GetMedicines(ctx context.Context, medId int64) ([]int64, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	return []int64{1, 2, 3}, nil
}

func (m *mockStorage) AddMedicine(ctx context.Context, schedule entity.Medicine) (int64, error) {
	if m.shouldError {
		return 0, errors.New("error")
	}
	return 5, nil
}

func (m *mockStorage) GetMedicine(ctx context.Context, id int64) (*entity.Medicine, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	if m.shouldEmpty {
		return nil, nil
	}
	return &entity.Medicine{
		Id:                id,
		Name:              "test",
		UserId:            1,
		TakingDuration:    1,
		TreatmentDuration: 1,
		Date:              time.Date(2025, 5, 11, 9, 0, 0, 0, time.UTC),
	}, nil
}

func (m *mockStorage) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*entity.Medicine, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	if m.shouldEmpty {
		return []*entity.Medicine{}, nil
	}
	return []*entity.Medicine{
		{
			Id:                1,
			Name:              "medicine1",
			UserId:            userID,
			TakingDuration:    3,
			TreatmentDuration: 7,
		},
		{
			Id:                2,
			Name:              "medicine2",
			UserId:            userID,
			TakingDuration:    2,
			TreatmentDuration: 5,
		},
	}, nil
}

func TestSchedules(t *testing.T) {
	tests := []struct {
		name        string
		input       int64
		shouldError bool
		out         []int64
		outError    bool
	}{
		{
			name:        "Normal Case",
			input:       int64(5),
			shouldError: false,
			out:         []int64{1, 2, 3},
			outError:    false,
		},
		{
			name:        "Zero Test",
			input:       int64(0),
			shouldError: false,
			out:         nil,
			outError:    true,
		},
		{
			name:        "Minus Test",
			input:       int64(-5),
			shouldError: false,
			out:         nil,
			outError:    true,
		},
		{
			name:        "DB error Test",
			input:       int64(5),
			shouldError: true,
			out:         nil,
			outError:    true,
		},
	}

	log := logger.New("local")
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := New(log, &mockStorage{shouldError: tc.shouldError}, time.Hour)
			res, err := service.Schedules(ctx, tc.input)
			if err != nil && !tc.outError {
				t.Errorf("Unexpected error\nin: %v\nerr: %v\n", tc.input, err)
			}
			if !slices.Equal(res, tc.out) {
				t.Errorf("GetMedicines() returned wrong res: \ngot %v \nwant %v", res, tc.out)
			}
		})
	}
}

func TestAddSchedule(t *testing.T) {
	type testInput struct {
		name              string
		userId            int64
		takingDuration    int32
		treatmentDuration int32
	}

	tests := []struct {
		name        string
		input       *testInput
		shouldError bool
		out         int64
		outError    bool
	}{
		{
			name: "Normal Test",
			input: &testInput{
				name:              "test",
				userId:            1,
				takingDuration:    1,
				treatmentDuration: 1,
			},
			shouldError: false,
			out:         5,
			outError:    false,
		},
		{
			name: "DB error Test",
			input: &testInput{
				name:              "test",
				userId:            1,
				takingDuration:    1,
				treatmentDuration: 1,
			},
			shouldError: true,
			out:         0,
			outError:    true,
		},
		{
			name: "Invalid input Test",
			input: &testInput{
				name:              "",
				userId:            0,
				takingDuration:    1,
				treatmentDuration: -5,
			},
			shouldError: false,
			out:         0,
			outError:    true,
		},
	}

	log := logger.New("local")
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := New(log, &mockStorage{shouldError: tc.shouldError}, time.Hour)
			res, err := service.AddSchedule(ctx, tc.input.name, tc.input.userId, tc.input.takingDuration, tc.input.treatmentDuration)
			if err != nil && !tc.outError {
				t.Errorf("Unexpected error\nin: %v\nerr: %v\n", tc.input, err)
			}
			if res != tc.out {
				t.Errorf("GetMedicines() returned wrong res: \ngot %d \nwant %d", res, tc.out)
			}
		})
	}
}

func TestSchedule(t *testing.T) {
	tests := []struct {
		name         string
		userId       int64
		scheduleId   int64
		shouldError  bool
		shouldEmpty  bool
		expectedName string
		outError     bool
	}{
		{
			name:         "Normal Case",
			userId:       1,
			scheduleId:   1,
			shouldError:  false,
			shouldEmpty:  false,
			expectedName: "test",
			outError:     false,
		},
		{
			name:        "Not Found Case",
			userId:      1,
			scheduleId:  1,
			shouldError: false,
			shouldEmpty: true,
			outError:    true,
		},
		{
			name:        "DB Error Case",
			userId:      1,
			scheduleId:  1,
			shouldError: true,
			shouldEmpty: false,
			outError:    true,
		},
		{
			name:        "Wrong User Case",
			userId:      2,
			scheduleId:  1,
			shouldError: false,
			shouldEmpty: false,
			outError:    true,
		},
	}

	log := logger.New("local")
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := New(log, &mockStorage{shouldError: tc.shouldError, shouldEmpty: tc.shouldEmpty}, time.Hour)
			res, err := service.Schedule(ctx, tc.userId, tc.scheduleId)

			if err != nil && !tc.outError {
				t.Errorf("Unexpected error\nuserId: %v, scheduleId: %v\nerr: %v\n", tc.userId, tc.scheduleId, err)
			}

			if !tc.outError {
				if res == nil {
					t.Errorf("Expected medicine, got nil")
				} else if res.Name != tc.expectedName {
					t.Errorf("Expected medicine name '%s', got '%s'", tc.expectedName, res.Name)
				}
			}
		})
	}
}

func TestNextTakings(t *testing.T) {
	oldTimeNow := timeNow
	defer func() {
		timeNow = oldTimeNow
	}()

	fixedTime := time.Date(2025, 5, 11, 10, 0, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return fixedTime
	}

	tests := []struct {
		name        string
		userId      int64
		shouldError bool
		shouldEmpty bool
		expectedLen int
		outError    bool
	}{
		{
			name:        "Normal Case",
			userId:      1,
			shouldError: false,
			shouldEmpty: false,
			expectedLen: 1,
			outError:    false,
		},
		{
			name:        "Empty Case",
			userId:      1,
			shouldError: false,
			shouldEmpty: true,
			expectedLen: 0,
			outError:    false,
		},
		{
			name:        "DB Error Case",
			userId:      1,
			shouldError: true,
			shouldEmpty: false,
			outError:    true,
		},
		{
			name:        "Invalid User Case",
			userId:      0,
			shouldError: false,
			shouldEmpty: false,
			outError:    true,
		},
	}

	log := logger.New("local")
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := New(log, &mockStorage{shouldError: tc.shouldError, shouldEmpty: tc.shouldEmpty}, 6*time.Hour)
			res, err := service.NextTakings(ctx, tc.userId)

			if err != nil && !tc.outError {
				t.Errorf("Unexpected error\nuserId: %v\nerr: %v\n", tc.userId, err)
			}

			if !tc.outError && len(res) != tc.expectedLen {
				t.Errorf("Expected %d medicines, got %d", tc.expectedLen, len(res))
			}
		})
	}
}
