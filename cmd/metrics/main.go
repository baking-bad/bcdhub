package main

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
)

// Context -
type Context struct {
	*config.Context

	cache *ccache.Cache
}

var ctx Context

var handlers = map[string]BulkHandler{
	mq.QueueContracts:   getContract,
	mq.QueueOperations:  getOperation,
	mq.QueueBigMapDiffs: getBigMapDiff,
}

var managers = map[string]*BulkManager{}

func listenChannel(messageQueue mq.IMessageReceiver, queue string, closeChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "queue", queue)

	msgs, err := messageQueue.Consume(queue)
	if err != nil {
		logger.Error(err)
		return
	}

	ticker := time.NewTicker(time.Second * time.Duration(15))
	defer ticker.Stop()

	logger.Info("Connected to %s queue", queue)
	for {
		select {
		case <-ticker.C:
			if manager, ok := managers[queue]; ok {
				manager.Exec()
			}
		case <-closeChan:
			logger.Info("Stopped %s queue", queue)
			return
		case msg := <-msgs:
			if manager, ok := managers[msg.RoutingKey]; ok {
				manager.Add(msg)
				continue
			}

			if msg.RoutingKey == "" {
				logger.Warning("[%s] Rabbit MQ server stopped! Metrics service need to be restarted. Closing connection...", queue)
				return
			}
			logger.Errorf("Unknown data routing key %s", msg.RoutingKey)
			helpers.LocalCatchErrorSentry(localSentry, errors.Errorf("[listenChannel] %s", err.Error()))
		}
	}
}

func main() {
	logger.Warning("Metrics started on 5 CPU cores")
	runtime.GOMAXPROCS(5)

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Metrics.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Metrics.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	configCtx := config.NewContext(
		config.WithStorage(cfg.Storage, cfg.Metrics.ProjectName, 0),
		config.WithRPC(cfg.RPC),
		config.WithRabbit(cfg.RabbitMQ, cfg.Metrics.ProjectName, cfg.Metrics.MQ),
		config.WithSearch(cfg.Storage),
		config.WithShare(cfg.SharePath),
		config.WithDomains(cfg.Domains),
		config.WithConfigCopy(cfg),
	)
	defer configCtx.Close()

	ctx = Context{
		Context: configCtx,
		cache:   ccache.New(ccache.Configure().MaxSize(100)),
	}

	if err := ctx.Searcher.CreateIndexes(); err != nil {
		logger.Fatal(err)
	}

	var wg sync.WaitGroup

	threadsCount := len(ctx.MQ.GetQueues()) + 2

	closeChan := make(chan struct{}, threadsCount)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for _, queue := range ctx.MQ.GetQueues() {
		if handler, ok := handlers[queue]; ok {
			managers[queue] = NewBulkManager(50, 10, handler)
		}
		wg.Add(1)
		go listenChannel(ctx.MQ, queue, closeChan, &wg)
	}

	wg.Add(1)
	go timeBasedTask(time.Minute, ctx.updateMaterializedViews, closeChan, &wg)

	wg.Add(1)
	go timeBasedTask(time.Hour*6, ctx.updateSeriesMaterializedViews, closeChan, &wg)

	<-signals
	for i := 0; i < threadsCount; i++ {
		closeChan <- struct{}{}
	}

	wg.Wait()

	close(closeChan)
	close(signals)
}
