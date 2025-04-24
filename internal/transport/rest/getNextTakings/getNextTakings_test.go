package getNextTakings

import (
	"context"
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

func (m *MockDB) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*storage.Medicine, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}

	return []*storage.Medicine{
		{
			Id:                0,
			Name:              "test",
			TakingDuration:    2,
			TreatmentDuration: 2,
			UserId:            1,
		},
	}, nil
}

func (m *MockDB) GetMedicines(ctx context.Context, medId int64) ([]int64, error) {
	return []int64{}, nil
}
func (m *MockDB) GetMedicine(ctx context.Context, id int64) (*storage.Medicine, error) {
	return &storage.Medicine{}, nil
}
func (m *MockDB) AddMedicine(ctx context.Context, schedule storage.Medicine) (int64, error) {
	return 0, nil
}

func setupRouter(log *slog.Logger, db *MockDB, period time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service := medService2.New(log, db, time.Hour*24)
	r.GET("/schedules", GetNextTakingsHandler(log, service))
	return r
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
			out:          `[{"name":"test","times":"08:00"},{"name":"test","times":"22:00"}]`,
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

			mockDB := &MockDB{shouldError: tc.shouldError}
			router := setupRouter(logger.MustLoad("local"), mockDB, time.Hour*20)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)

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
