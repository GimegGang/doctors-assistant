package getSchedules

import (
	"errors"
	"kode/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockDB struct {
	shouldError bool
}

func (m *MockDB) GetMedicines(medId int64) ([]*int64, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	return []*int64{new(int64)}, nil
}

func TestGetSchedulesHandler(t *testing.T) {
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
			name:         "empty case",
			input:        "",
			shouldError:  false,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "database error",
			input:        "user_id=1",
			shouldError:  true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)
			rr := httptest.NewRecorder()
			mockDB := &MockDB{shouldError: tc.shouldError}
			handler := GetSchedulesHandler(logger.MustLoad("local"), mockDB)
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tc.expectedCode)
			}
		})
	}
}
