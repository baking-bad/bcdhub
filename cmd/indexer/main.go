package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/baking-bad/bcdhub/cmd/indexer/indexer"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/tidwall/gjson"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Indexer.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Indexer.ProjectName)
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

	countCPU := runtime.NumCPU()
	if countCPU > len(indexers) {
		countCPU = len(indexers)
	}
	logger.Warning("Indexer started on %d CPU cores", countCPU)
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
	logger.Info("Stopped")
}
