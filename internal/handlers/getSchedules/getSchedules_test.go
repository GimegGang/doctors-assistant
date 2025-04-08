package getSchedules

import (
	"errors"
	"github.com/gin-gonic/gin"
	"kode/internal/logger"
	"log/slog"
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

func setupRouter(log *slog.Logger, db *MockDB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/schedules", GetSchedulesHandler(log, db))
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
