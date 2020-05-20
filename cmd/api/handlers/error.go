package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error, code int) bool {
	if err != nil {
		if code == http.StatusUnauthorized {
			err = errors.New("Invalid authentication")
		}
		if code == 0 {
			code = getErrorCode(err)
		}
		c.AbortWithStatusJSON(code, Error{Message: err.Error()})
		return true
	}
	return false
}

func getErrorCode(err error) int {
	if elastic.IsRecordNotFound(err) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
