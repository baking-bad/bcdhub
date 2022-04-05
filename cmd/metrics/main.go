package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/baking-bad/bcdhub/cmd/metrics/services"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

var ctxs config.Contexts

const (
	bulkSize = 100
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
	}

	if cfg.Metrics.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Metrics.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctxs = config.NewContexts(
		cfg, cfg.Metrics.Networks,
		config.WithStorage(cfg.Storage, cfg.Metrics.ProjectName, 0, cfg.Metrics.Connections.Open, cfg.Metrics.Connections.Idle),
		config.WithRPC(cfg.RPC, false),
		config.WithSearch(cfg.Storage),
		config.WithConfigCopy(cfg),
	)
	defer ctxs.Close()

	if err := ctxs.Any().Searcher.CreateIndexes(); err != nil {
		logger.Err(err)
		return
	}

	workers := make([]services.Service, 0)

	for _, ctx := range ctxs {
		workers = append(workers, services.NewUnknown(ctx, time.Minute*30, time.Second*2, -time.Hour*24))
		workers = append(workers, services.NewStorageBased[domains.BigMapDiff](
			"contract_metadata",
			ctx.Services,
			services.NewContractMetadataHandler(ctx),
			time.Second*15,
			bulkSize,
		))
		workers = append(workers, services.NewStorageBased[domains.BigMapDiff](
			"token_metadata",
			ctx.Services,
			services.NewTokenMetadataHandler(ctx),
			time.Second*15,
			bulkSize,
		))
		workers = append(workers, services.NewStorageBased[*operation.Operation](
			"operations",
			ctx.Services,
			services.NewOperationsHandler(ctx),
			time.Second*15,
			bulkSize,
		))
		workers = append(workers, services.NewStorageBased[*contract.Contract](
			"contracts",
			ctx.Services,
			services.NewContractsHandler(ctx),
			time.Second*15,
			bulkSize,
		))
		workers = append(workers, services.NewStorageBased[*bigmapdiff.BigMapDiff](
			"big_map_diffs",
			ctx.Services,
			services.NewBigMapDiffHandler(ctx),
			time.Second*15,
			bulkSize,
		))
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	cancelledContext, cancel := context.WithCancel(context.Background())

	for i := range workers {
		if err := workers[i].Init(); err != nil {
			if err := stop(workers, i-1, signals); err != nil {
				logger.Err(err)
			}
			logger.Err(err)
			return
		}
		workers[i].Start(cancelledContext)
	}

	<-signals
	cancel()

	if err := stop(workers, len(workers), signals); err != nil {
		logger.Err(err)
	}
}

func stop(workers []services.Service, running int, signals chan os.Signal) error {
	if running > 0 {
		if running > len(workers) {
			running = len(workers)
		}
		for i := 0; i < running; i++ {
			if err := workers[i].Close(); err != nil {
				return err
			}
		}
	}

	close(signals)
	return nil
}
