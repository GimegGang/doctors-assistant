package getSchedules

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/logger"
	medService2 "kode/internal/service/medService"
	"kode/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockDB struct {
	shouldError bool
}

func (m *MockDB) GetMedicines(ctx context.Context, medId int64) ([]int64, error) {
	if m.shouldError {
		return nil, errors.New("error")
	}
	return []int64{3}, nil
}
func (m *MockDB) GetMedicine(ctx context.Context, id int64) (*storage.Medicine, error) {
	return &storage.Medicine{}, nil
}
func (m *MockDB) GetMedicinesByUserID(ctx context.Context, userID int64) ([]*storage.Medicine, error) {
	return []*storage.Medicine{}, nil
}
func (m *MockDB) AddMedicine(ctx context.Context, schedule storage.Medicine) (int64, error) {
	return 0, nil
}

func setupRouter(log *slog.Logger, db *MockDB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service := medService2.New(log, db, time.Hour)
	r.GET("/schedules", GetSchedulesHandler(log, service))
	return r
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
			mockDB := &MockDB{shouldError: tc.shouldError}
			router := setupRouter(logger.MustLoad("local"), mockDB)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/schedules?"+tc.input, nil)

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tc.expectedCode)
			}
		})
	}
}
