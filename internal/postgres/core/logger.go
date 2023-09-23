package core

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type logQueryHook struct{}

// BeforeQuery -
func (h *logQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	event.StartTime = time.Now()
	return ctx
}

// AfterQuery -
func (h *logQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if event.Err != nil {
		log.Trace().Msgf("[%d mcs] %s : %s", time.Since(event.StartTime).Microseconds(), event.Err.Error(), event.Query)
	} else {
		log.Trace().Msgf("[%d mcs] %s", time.Since(event.StartTime).Microseconds(), event.Query)
	}
}
