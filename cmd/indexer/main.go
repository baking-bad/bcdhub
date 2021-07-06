package main

import (
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

	indexers, err := indexer.CreateIndexers(cfg)
	if err != nil {
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
		go indexers[i].Sync(&wg)
	}

	<-sigChan

	for i := range indexers {
		go indexers[i].Stop()
	}
	wg.Wait()
	logger.Info().Msg("Stopped")
}
