package addHandler

import (
	"bytes"
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

func (m *MockDB) AddMedicine(schedule storage.Medicine) (int64, error) {
	if m.shouldError {
		return 0, errors.New("error")
	}
	return 1, nil
}

func TestAddScheduleHandler(t *testing.T) {
	testCases := []struct {
		name         string
		input        storage.Medicine
		shouldError  bool
		expectedCode int
		out          string
	}{
		{
			name: "normal case",
			input: storage.Medicine{
				Name:              "TEST",
				TakingDuration:    1,
				TreatmentDuration: 1,
				UserId:            1,
			},
			shouldError:  false,
			expectedCode: http.StatusOK,
			out:          `{"id":1}`,
		},
		{
			name: "input error case",
			input: storage.Medicine{
				Name:              "TEST",
				TakingDuration:    -1,
				TreatmentDuration: 1,
				UserId:            1,
			},
			shouldError:  false,
			expectedCode: http.StatusBadRequest,
			out:          "",
		},
		{
			name: "database error case",
			input: storage.Medicine{
				Name:              "TEST",
				TakingDuration:    1,
				TreatmentDuration: 1,
				UserId:            1,
			},
			shouldError:  true,
			expectedCode: http.StatusInternalServerError,
			out:          "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.input)
			req, _ := http.NewRequest(http.MethodPost, "/schedule", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()
			mockDB := &MockDB{shouldError: tc.shouldError}
			handler := AddScheduleHandler(logger.MustLoad("local"), mockDB)
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
