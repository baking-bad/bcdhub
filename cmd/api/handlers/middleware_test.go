package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBodyLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const limit = 16
	ctrl := gomock.NewController(t)
	repo := mock.NewMockGeneralRepository(ctrl)
	repo.EXPECT().IsRecordNotFound(gomock.Any()).Return(false).AnyTimes()

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(BodyLimitMiddleware(limit))
		r.POST("/", func(c *gin.Context) {
			var body map[string]any
			if err := c.ShouldBindJSON(&body); handleError(c, repo, err, http.StatusBadRequest) {
				return
			}
			c.Status(http.StatusOK)
		})
		return r
	}

	tests := []struct {
		name          string
		body          string
		contentLength int64
		want          int
	}{
		{name: "within limit", body: `{"a":"12345678"}`, contentLength: 16, want: http.StatusOK},
		{name: "declared too large", body: `{"a":"123456789"}`, contentLength: 17, want: http.StatusRequestEntityTooLarge},
		{name: "chunked too large", body: `{"a":"123456789"}`, contentLength: -1, want: http.StatusRequestEntityTooLarge},
		{name: "chunked within limit", body: `{"a":"12345678"}`, contentLength: -1, want: http.StatusOK},
		{name: "invalid JSON is still bad request", body: `{"a":`, contentLength: -1, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(tt.body))
			req.ContentLength = tt.contentLength
			w := httptest.NewRecorder()
			newRouter().ServeHTTP(w, req)
			require.Equal(t, tt.want, w.Code)
		})
	}
}
