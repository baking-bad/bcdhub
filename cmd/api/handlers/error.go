package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/logger"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func (ctx *Context) handleError(c *gin.Context, err error, code int) bool {
	if err == nil {
		return false
	}

	switch code {
	case http.StatusUnauthorized:
		err = errors.New("invalid authentication")
	case 0:
		code = ctx.getErrorCode(err)

		if code == http.StatusInternalServerError {
			if hub := sentrygin.GetHubFromContext(c); hub != nil {
				hub.CaptureMessage(err.Error())
			}
			logger.Err(err)
		}
	}

	c.AbortWithStatusJSON(code, Error{Message: ctx.getErrorText(err)})
	return true
}

func (ctx *Context) getErrorCode(err error) int {
	if ctx.Storage.IsRecordNotFound(err) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func (ctx *Context) getErrorText(err error) string {
	if ctx.Storage.IsRecordNotFound(err) {
		return "not found"
	}
	return err.Error()
}
