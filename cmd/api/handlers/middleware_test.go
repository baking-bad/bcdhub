package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBodyLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const limit = 16
	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(BodyLimitMiddleware(limit))
		r.POST("/", func(c *gin.Context) {
			if _, err := io.ReadAll(c.Request.Body); err != nil {
				c.Status(http.StatusBadRequest)
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
		{name: "within limit", body: strings.Repeat("a", limit), contentLength: limit, want: http.StatusOK},
		{name: "declared too large", body: strings.Repeat("a", limit+1), contentLength: limit + 1, want: http.StatusRequestEntityTooLarge},
		{name: "chunked too large", body: strings.Repeat("a", limit+1), contentLength: -1, want: http.StatusBadRequest},
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
