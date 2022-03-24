package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/baking-bad/bcdhub/cmd/indexer/indexer"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
		return
	}

	if cfg.Indexer.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Indexer.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctx, cancel := context.WithCancel(context.Background())

	indexers, err := indexer.CreateIndexers(ctx, cfg)
	if err != nil {
		cancel()
		logger.Err(err)
		helpers.CatchErrorSentry(err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	var wg sync.WaitGroup
	for i := range indexers {
		wg.Add(1)
		go indexers[i].Sync(ctx, &wg)
	}

	<-sigChan
	cancel()

	wg.Wait()
	for i := range indexers {
		if err := indexers[i].Close(); err != nil {
			panic(err)
		}
	}
	logger.Info().Msg("Stopped")
}
