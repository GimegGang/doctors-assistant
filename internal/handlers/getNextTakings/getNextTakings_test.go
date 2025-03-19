package getNextTakings

import (
	"encoding/json"
	"errors"
	"kode/internal/logger"
	"kode/internal/storage"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		out          string
	}{
		{
			name:         "normal case",
			input:        "user_id=1",
			shouldError:  false,
			expectedCode: http.StatusOK,
			out:          `{"medicines":[{"name":"test","times":"08:00"},{"name":"test","times":"22:00"}]}`,
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
			now := time.Now()
			timeNow = func() time.Time {
				return time.Date(now.Year(), now.Month(), now.Day(), 7, 30, 0, 0, time.Local)
			}
			defer func() { timeNow = time.Now }()

			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)
			rr := httptest.NewRecorder()
			mockDB := &MockDB{shouldError: tc.shouldError}
			handler := GetNextTakingsHandler(logger.MustLoad("local"), mockDB, time.Hour*15)
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
