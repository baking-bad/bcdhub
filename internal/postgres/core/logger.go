package core

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/go-pg/pg/v10"
)

type logQueryHook struct{}

// BeforeQuery -
func (h *logQueryHook) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	event.StartTime = time.Now()
	return ctx, nil
}

func (h *logQueryHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	query, err := event.FormattedQuery()
	if err != nil {
		return err
	}

	// logger.Info().Interface("params", event.Params).Msg("")
	if event.Err != nil {
		logger.Info().Msgf("[%d ms] %s : %s", time.Since(event.StartTime).Milliseconds(), event.Err.Error(), string(query))
	} else {
		logger.Info().Msgf("[%d ms] %d rows | %s", time.Since(event.StartTime).Milliseconds(), event.Result.RowsReturned(), string(query))
	}
	return nil
}
