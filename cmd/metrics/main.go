package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var ctx *config.Context

const metricStoppedError = "METRICS_STOPPED"

func parseData(data amqp.Delivery) error {
	switch data.RoutingKey {
	case mq.QueueContracts:
		return getContract(data)
	case mq.QueueOperations:
		return getOperation(data)
	case mq.QueueMigrations:
		return getMigrations(data)
	case mq.QueueRecalc:
		return recalculateAll(data)
	case mq.QueueTransfers:
		return getTransfer(data)
	case mq.QueueBigMapDiffs:
		return getBigMapDiff(data)
	default:
		if data.RoutingKey == "" {
			return errors.Errorf(metricStoppedError)
		}
		return errors.Errorf("Unknown data routing key %s", data.RoutingKey)
	}
}

func handler(data amqp.Delivery) error {
	if err := parseData(data); err != nil {
		if err.Error() == metricStoppedError {
			return err
		}
		if !elastic.IsRecordNotFound(err) {
			return err
		}
	}

	if err := data.Ack(false); err != nil {
		return errors.Errorf("Error acknowledging message: %s", err)
	}
	return nil
}

func listenChannel(messageQueue mq.IMessageReceiver, queue string, closeChan chan struct{}) {
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
			logger.Info("Stopped %s queue", queue)
			return
		case msg := <-msgs:
			if err := handler(msg); err != nil {
				if err.Error() == metricStoppedError {
					logger.Warning("[%s] Rabbit MQ server stopped! Metrics service need to be restarted. Closing connection...", queue)
					return
				}
				logger.Errorf("[listenChannel] %s", err.Error())
				helpers.LocalCatchErrorSentry(localSentry, errors.Errorf("[listenChannel] %s", err.Error()))
			}
		}
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Metrics.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Metrics.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctx = config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithRPC(cfg.RPC),
		config.WithDatabase(cfg.DB),
		config.WithRabbit(cfg.RabbitMQ, cfg.Metrics.ProjectName, cfg.Metrics.MQ),
		config.WithAliases(consts.Mainnet),
		config.WithShare(cfg.SharePath),
	)
	defer ctx.Close()

	closeChan := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for _, queue := range ctx.MQ.GetQueues() {
		go listenChannel(ctx.MQ, queue, closeChan)
	}

	<-signals
	close(closeChan)
}
