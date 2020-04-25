package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/baking-bad/bcdhub/cmd/indexer/indexer"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/tidwall/gjson"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Indexer.Sentry.Enabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Indexer.Sentry.Project)
		defer helpers.CatchPanicSentry()
	}

	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	indexers, err := indexer.CreateIndexers(cfg)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

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
	logger.Info("Stopped")
}
