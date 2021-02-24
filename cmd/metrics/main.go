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
	Cache               *ccache.Cache
	AliasesCacheSeconds time.Duration
	*config.Context
}

var ctx Context

var handlers = map[string]BulkHandler{
	mq.QueueContracts:   getContract,
	mq.QueueOperations:  getOperation,
	mq.QueueMigrations:  getMigrations,
	mq.QueueBigMapDiffs: getBigMapDiff,
	mq.QueueRecalc:      recalculateAll,
	mq.QueueProjects:    getProject,
}

var managers = map[string]*BulkManager{}

func listenChannel(messageQueue mq.IMessageReceiver, queue string, closeChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "queue", queue)

	msgs, err := messageQueue.Consume(queue)
	if err != nil {
		panic(err)
	}

	logger.Info("Connected to %s queue", queue)
	for {
		select {
		case <-closeChan:
			if manager, ok := managers[queue]; ok {
				manager.Stop()
				wg.Done()
			}
			logger.Info("Stopped %s queue", queue)
			return
		case msg := <-msgs:
			if manager, ok := managers[msg.GetKey()]; ok {
				manager.Add(msg)
				continue
			}

			if msg.GetKey() == "" {
				logger.Warning("[%s] Rabbit MQ server stopped! Metrics service need to be restarted. Closing connection...", queue)
				return
			}
			logger.Errorf("Unknown data routing key %s", msg.GetKey())
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
		config.WithStorage(cfg.Storage),
		config.WithRPC(cfg.RPC),
		config.WithDatabase(cfg.DB),
		config.WithRabbit(cfg.RabbitMQ, cfg.Metrics.ProjectName, cfg.Metrics.MQ),
		config.WithShare(cfg.SharePath),
		config.WithDomains(cfg.Domains),
		config.WithConfigCopy(cfg),
	)
	defer configCtx.Close()

	ctx = Context{
		Cache:               ccache.New(ccache.Configure().MaxSize(10)),
		AliasesCacheSeconds: time.Second * time.Duration(configCtx.Config.Metrics.CacheAliasesSeconds),
		Context:             configCtx,
	}

	var wg sync.WaitGroup

	closeChan := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-signals
		for range ctx.MQ.GetQueues() {
			closeChan <- struct{}{}
		}
	}()

	for _, queue := range ctx.MQ.GetQueues() {
		if handler, ok := handlers[queue]; ok {
			managers[queue] = NewBulkManager(30, 10, handler)
			wg.Add(1)
			go managers[queue].Run()
		}
		wg.Add(1)
		go listenChannel(ctx.MQ, queue, closeChan, &wg)
	}

	wg.Wait()

	close(closeChan)
	close(signals)
}
