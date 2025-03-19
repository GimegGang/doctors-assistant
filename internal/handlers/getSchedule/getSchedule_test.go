package getSchedule

import (
	"encoding/json"
	"errors"
	"kode/internal/logger"
	"kode/internal/storage"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type MockDB struct {
	shouldError bool
}

func (m *MockDB) GetMedicine(id int64) (*storage.Medicine, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	return &storage.Medicine{
		Id:                0,
		Name:              "test",
		TakingDuration:    2,
		TreatmentDuration: 2,
		UserId:            1,
	}, nil
}

func TestGetScheduleHandler(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		shouldError  bool
		expectedCode int
		out          string
	}{
		{
			name:         "normal case",
			input:        "schedule_id=1&user_id=1",
			shouldError:  false,
			expectedCode: http.StatusOK,
			out: `{"id":0,"name":"test","taking_duration":2,"treatment_duration":2,"user_id":1,
							"schedule":["08:00","22:00"],"date":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:         "error schedule_id input case",
			input:        "schedule_id=-1&user_id=1",
			shouldError:  false,
			expectedCode: http.StatusBadRequest,
			out:          "",
		},
		{
			name:         "error user_id input case",
			input:        "schedule_id=1&user_id=5",
			shouldError:  false,
			expectedCode: http.StatusForbidden,
			out:          "",
		},
		{
			name:         "empty case",
			input:        "",
			shouldError:  true,
			expectedCode: http.StatusBadRequest,
			out:          "",
		},
		{
			name:         "database error case",
			input:        "schedule_id=1&user_id=1",
			shouldError:  true,
			expectedCode: http.StatusInternalServerError,
			out:          "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)
			rr := httptest.NewRecorder()
			mockDB := &MockDB{shouldError: tc.shouldError}
			handler := GetScheduleHandler(logger.MustLoad("local"), mockDB)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tc.expectedCode)
			}

			if tc.out != "" {
				var exp, actual interface{}

				err := json.Unmarshal([]byte(tc.out), &exp)
				if err != nil {
					t.Fatalf("Failed to unmarshal json: %v", err)
				}

				err = json.Unmarshal(rr.Body.Bytes(), &actual)
				if err != nil {
					t.Fatalf("Failed to unmarshal json: %v", err)
				}

				if !reflect.DeepEqual(exp, actual) {
					t.Fatalf("Expected: %v, got: %v", exp, actual)
				}
			}
		})
	}
}
