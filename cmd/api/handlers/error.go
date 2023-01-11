package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// errors
var (
	ErrNotFAContract = errors.New("contract is not FA1.2 or FA2")
	ErrInvalidAuth   = errors.New("invalid authentication")
)

func handleError(c *gin.Context, repo models.GeneralRepository, err error, code int) bool {
	if err == nil {
		return false
	}

	switch code {
	case http.StatusUnauthorized:
		err = ErrInvalidAuth
	case 0:
		code = getErrorCode(err, repo)
		if code == http.StatusInternalServerError {
			if hub := sentrygin.GetHubFromContext(c); hub != nil {
				hub.CaptureMessage(err.Error())
			}
			logger.Err(err)
		}
	}

	c.AbortWithStatusJSON(code, getErrorMessage(err, repo))
	return true
}

func getErrorCode(err error, repo models.GeneralRepository) int {
	if repo.IsRecordNotFound(err) {
		return http.StatusNotFound
	}
	if errors.Is(err, ast.ErrValidation) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func getErrorMessage(err error, repo models.GeneralRepository) Error {
	if repo.IsRecordNotFound(err) {
		return Error{Message: "not found"}
	}
	return Error{Message: err.Error()}
}
