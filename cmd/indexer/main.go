package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/baking-bad/bcdhub/cmd/indexer/indexer"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/dipdup-io/workerpool"
	"github.com/pyroscope-io/client/pyroscope"
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

	var profiler *pyroscope.Profiler
	if cfg.Profiler != nil && cfg.Profiler.Server != "" {
		profiler, err = pyroscope.Start(pyroscope.Config{
			ApplicationName: "bcdhub.indexer",
			ServerAddress:   cfg.Profiler.Server,
			Tags: map[string]string{
				"hostname": os.Getenv("BCDHUB_SERVICE"),
				"project":  "bcdhub",
				"service":  "indexer",
			},

			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
				pyroscope.ProfileGoroutines,
				pyroscope.ProfileMutexCount,
				pyroscope.ProfileMutexDuration,
				pyroscope.ProfileBlockCount,
				pyroscope.ProfileBlockDuration,
			},
		})
		if err != nil {
			panic(err)
		}
	}

	g := workerpool.NewGroup()
	ctx, cancel := context.WithCancel(context.Background())

	indexers, err := indexer.CreateIndexers(ctx, cfg, g)
	if err != nil {
		cancel()
		logger.Err(err)
		helpers.CatchErrorSentry(err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for i := range indexers {
		g.GoCtx(ctx, indexers[i].Start)
	}

	<-sigChan
	cancel()

	g.Wait()
	for i := range indexers {
		if err := indexers[i].Close(); err != nil {
			panic(err)
		}
	}

	if profiler != nil {
		if err := profiler.Stop(); err != nil {
			panic(err)
		}
	}
	logger.Info().Msg("Stopped")
}
