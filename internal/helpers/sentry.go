package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// InitSentry -
func InitSentry(debug bool, environment, dsn string) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Debug:            debug,
		AttachStacktrace: true,
		BeforeSend:       beforeSend,
	}); err != nil {
		logger.Info("Sentry initialization failed: %v\n", err)
	}
}

func beforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	logger.Info("[Sentry message] %s", event.Message)
	return event
}

// SentryMiddleware -
func SentryMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{})
}

// SetTagSentry -
func SetTagSentry(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// SetUserIDSentry -
func SetUserIDSentry(id string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: id,
		})
	})
}

// CatchPanicSentry -
func CatchPanicSentry() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		sentry.Flush(time.Second * 5)
	}
}

// CatchErrorSentry -
func CatchErrorSentry(err error) {
	withStack := fmt.Sprintf("%+v", errors.WithStack(err))
	frames := make([]sentry.Frame, 0)
	lines := strings.Split(withStack, "\n")
	if len(lines) > 1 {
		lines = lines[1:]
	}

	var frame sentry.Frame
	for i, line := range lines {
		if i%2 == 0 {
			dotsSplits := strings.SplitAfterN(line, ".", 2)
			frame = sentry.Frame{
				Function: dotsSplits[1],
				Module:   dotsSplits[0],
			}
		} else {
			parts := strings.Split(line, ":")
			frame.AbsPath = parts[0]
			if len(parts) > 1 {
				lineNo, _ := strconv.Atoi(parts[1])
				frame.Lineno = lineNo
			}
			frames = append(frames, frame)
		}
	}

	for left, right := 0, len(frames)-1; left < right; left, right = left+1, right-1 {
		frames[left], frames[right] = frames[right], frames[left]
	}

	stackTrace := &sentry.Stacktrace{
		Frames: frames,
	}
	sentry.CaptureEvent(&sentry.Event{
		Message: err.Error(),
		Level:   sentry.LevelError,
		Exception: []sentry.Exception{
			{
				Value:      err.Error(),
				Stacktrace: stackTrace,
			},
		},
	})
	sentry.Flush(time.Second * 5)
}

// GetLocalSentry -
func GetLocalSentry() *sentry.Hub {
	return sentry.CurrentHub().Clone()
}

// SetLocalTagSentry -
func SetLocalTagSentry(hub *sentry.Hub, key, value string) {
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// LocalCatchErrorSentry -
func LocalCatchErrorSentry(hub *sentry.Hub, err error) {
	hub.CaptureException(err)
	hub.Flush(time.Second * 5)
}

// LocalCatchPanicSentry -
func LocalCatchPanicSentry(hub *sentry.Hub) {
	if err := recover(); err != nil {
		hub.CaptureMessage(err.(string))
		hub.Flush(time.Second * 5)
	}
}
