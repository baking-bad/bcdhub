package helpers

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type SentryConfig struct {
	DSN   string
	Debug bool
	Env   string
}

// InitSentry -
func InitSentry(cfg SentryConfig) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Env,
		Debug:            cfg.Debug,
		AttachStacktrace: true,
		BeforeSend:       beforeSend,
	}); err != nil {
		log.Err(err).Msg("Sentry initialization failed")
	}
}

func beforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	log.Debug().Msgf("[Sentry message] %s", event.Message)
	return event
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
		if hub := sentry.CurrentHub(); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
			})
			if eventId := hub.Recover(err); eventId != nil {
				sentry.Flush(time.Second * 5)
			}
		}
	}
}

// CatchErrorSentry -
func CatchErrorSentry(err error) {
	sentry.CaptureEvent(&sentry.Event{
		Message: err.Error(),
		Level:   sentry.LevelError,
		Exception: []sentry.Exception{
			{
				Value:      err.Error(),
				Stacktrace: sentry.ExtractStacktrace(errors.WithStack(err)),
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
