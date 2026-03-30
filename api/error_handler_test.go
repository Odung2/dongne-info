package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"dongne-info/api"
	"dongne-info/model"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "NotFoundError → 404",
			err:            &model.NotFoundError{Resource: "공고", ID: "1"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ValidationError → 400",
			err:            &model.ValidationError{Field: "district", Message: "invalid"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ExternalAPIError → 502",
			err:            &model.ExternalAPIError{Service: "test", Err: fmt.Errorf("fail")},
			expectedStatus: http.StatusBadGateway,
		},
		{
			name:           "일반 에러 → 500",
			err:            fmt.Errorf("unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			api.HandleError(c, tt.err)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}
