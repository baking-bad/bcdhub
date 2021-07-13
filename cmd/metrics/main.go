package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baking-bad/bcdhub/cmd/metrics/services"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
)

var ctx *config.Context

const (
	bulkSize = 100
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
	}

	if cfg.Metrics.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Metrics.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctx = config.NewContext(
		config.WithStorage(cfg.Storage, cfg.Metrics.ProjectName, 0),
		config.WithRPC(cfg.RPC),
		config.WithSearch(cfg.Storage),
		config.WithShare(cfg.SharePath),
		config.WithDomains(cfg.Domains),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	if err := ctx.Searcher.CreateIndexes(); err != nil {
		logger.Err(err)
		return
	}

	workers := []services.Service{
		services.NewView(ctx.StorageDB.DB, "head_stats", time.Minute),
		services.NewStorageBased(
			"projects",
			ctx.Services,
			services.NewProjectsHandler(ctx),
			time.Second*15,
			bulkSize,
		),
		services.NewStorageBased(
			"contract_metadata",
			ctx.Services,
			services.NewContractMetadataHandler(ctx),
			time.Second*15,
			bulkSize,
		),
		services.NewStorageBased(
			"token_metadata",
			ctx.Services,
			services.NewTokenMetadataHandler(ctx),
			time.Second*15,
			bulkSize,
		),
		services.NewStorageBased(
			"tezos_domains",
			ctx.Services,
			services.NewTezosDomainHandler(ctx),
			time.Second*15,
			bulkSize,
		),
		services.NewStorageBased(
			"operations",
			ctx.Services,
			services.NewOperationsHandler(ctx),
			time.Second*15,
			bulkSize,
		),
		services.NewStorageBased(
			"big_map_diffs",
			ctx.Services,
			services.NewBigMapDiffHandler(ctx),
			time.Second*15,
			bulkSize,
		),
	}

	for network := range ctx.Config.Indexer.Networks {
		for _, view := range []string{
			"series_contract_by_month_",
			"series_operation_by_month_",
			"series_paid_storage_size_diff_by_month_",
			"series_consumed_gas_by_month_",
		} {
			name := fmt.Sprintf("%s%s", view, network)
			workers = append(workers, services.NewView(ctx.StorageDB.DB, name, time.Minute))
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for i := range workers {
		if err := workers[i].Init(); err != nil {
			if err := stop(workers, i-1, signals); err != nil {
				logger.Err(err)
			}
			logger.Err(err)
			return
		}
		workers[i].Start()
	}

	<-signals

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
