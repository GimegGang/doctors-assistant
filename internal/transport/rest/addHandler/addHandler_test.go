package addHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/logger"
	medService2 "kode/internal/service/medService"
	"kode/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
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
func (m *MockDB) GetMedicines(medId int64) ([]int64, error)       { return []int64{}, nil }
func (m *MockDB) GetMedicine(id int64) (*storage.Medicine, error) { return &storage.Medicine{}, nil }
func (m *MockDB) GetMedicinesByUserID(userID int64) ([]*storage.Medicine, error) {
	return []*storage.Medicine{}, nil
}

func setupRouter(log *slog.Logger, db *MockDB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service := medService2.New(log, db, time.Hour)
	r.POST("/schedule", AddScheduleHandler(log, service))
	return r
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
			mockDB := &MockDB{shouldError: tc.shouldError}
			router := setupRouter(logger.MustLoad("local"), mockDB)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/schedule", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tc.expectedCode)
			}

			if tc.out != "" {
				var exp, actual interface{}

				err := json.Unmarshal([]byte(tc.out), &exp)
				if err != nil {
					t.Fatalf("Failed to unmarshal json: %v", err)
				}

				err = json.Unmarshal(w.Body.Bytes(), &actual)
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
