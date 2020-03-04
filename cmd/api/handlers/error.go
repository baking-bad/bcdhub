package handlers

import (
	"errors"
	"net/http"
	"strings"

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
		c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
		return true
	}
	return false
}

func getErrorCode(err error) int {
	if strings.HasPrefix(err.Error(), "Unknown") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
