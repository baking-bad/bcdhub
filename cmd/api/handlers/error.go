package handlers

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// pgQueryCanceled is raised by postgres when statement_timeout expires
const pgQueryCanceled = "57014"

// statusClientClosedRequest is the nginx convention for a client-aborted request
const statusClientClosedRequest = 499

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func skipError(err error) bool {
	return errors.Is(err, noderpc.ErrNodeRPCError)
}

func handleError(c *gin.Context, repo models.GeneralRepository, err error, code int) bool {
	if err == nil {
		return false
	}

	switch code {
	case http.StatusUnauthorized:
		err = errors.New("invalid authentication")
	case 0:
		code = getErrorCode(err, repo)
		if code == http.StatusInternalServerError && !skipError(err) {
			if hub := sentrygin.GetHubFromContext(c); hub != nil {
				hub.CaptureMessage(err.Error())
			}
		}

		log.Err(err).Msg("unexpected error")
	}

	c.AbortWithStatusJSON(code, getErrorMessage(err, code, repo))
	return true
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgQueryCanceled
}

func getErrorCode(err error, repo models.GeneralRepository) int {
	if repo.IsRecordNotFound(err) {
		return http.StatusNotFound
	}
	if errors.Is(err, consts.ErrValidation) ||
		errors.Is(err, consts.ErrJSONDataIsAbsent) ||
		errors.Is(err, consts.ErrInvalidType) {
		return http.StatusBadRequest
	}
	if isTimeoutError(err) {
		return http.StatusGatewayTimeout
	}
	if errors.Is(err, context.Canceled) {
		return statusClientClosedRequest
	}
	return http.StatusInternalServerError
}

func getErrorMessage(err error, code int, repo models.GeneralRepository) Error {
	if repo.IsRecordNotFound(err) {
		return Error{Message: "not found"}
	}
	switch code {
	case http.StatusGatewayTimeout:
		return Error{Message: "request timed out"}
	case statusClientClosedRequest:
		return Error{Message: "request canceled"}
	case http.StatusInternalServerError:
		return Error{Message: "internal server error"}
	}
	return Error{Message: err.Error()}
}
