package getNextTakings

import (
	"KODE_test/internal/logger"
	"KODE_test/internal/storage"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockDB struct {
	shouldError bool
}

func (m *MockDB) GetMedicinesByUserID(userID int64) ([]*storage.Medicine, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}

	var medicines []*storage.Medicine
	medicines = append(medicines, &storage.Medicine{
		Id:                0,
		Name:              "test",
		TakingDuration:    2,
		TreatmentDuration: 2,
		UserId:            1,
	})

	return medicines, nil
}

func TestGetNextTakingsHandler(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		shouldError  bool
		expectedCode int
	}{
		{
			name:         "normal case",
			input:        "user_id=1",
			shouldError:  false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty time case",
			input:        "user_id=1",
			shouldError:  false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty case",
			input:        "",
			shouldError:  false,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "error database case",
			input:        "user_id=1",
			shouldError:  true,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "error input case",
			input:        "user_id=-1",
			shouldError:  true,
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)
			rr := httptest.NewRecorder()
			mockDB := &MockDB{shouldError: tc.shouldError}
			handler := GetNextTakingsHandler(logger.MustLoad("local"), mockDB, time.Hour)
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tc.expectedCode)
			}
		})
	}
}
