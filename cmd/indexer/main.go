package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
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

	internalCtx := config.NewContext(
		config.WithConfigCopy(cfg),
		config.WithStorage(cfg.Storage, "indexer", 10, cfg.Indexer.Connections.Open, cfg.Indexer.Connections.Idle),
		config.WithSearch(cfg.Storage),
	)
	defer internalCtx.Close()

	indexers, err := indexer.CreateIndexers(ctx, internalCtx, cfg)
	if err != nil {
		cancel()
		logger.Err(err)
		helpers.CatchErrorSentry(err)
		return
	}

	countCPU := runtime.NumCPU()
	if countCPU > len(indexers)+1 {
		countCPU = len(indexers) + 1
	}
	logger.Warning().Msgf("Indexer started on %d CPU cores", countCPU)
	runtime.GOMAXPROCS(countCPU)

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
	logger.Info().Msg("Stopped")
}
