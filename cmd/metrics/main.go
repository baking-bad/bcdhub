package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/streadway/amqp"
)

var ctx *config.Context

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
	default:
		return fmt.Errorf("Unknown data routing key %s", data.RoutingKey)
	}
}

func handler(data amqp.Delivery) error {
	if err := parseData(data); err != nil {
		if !strings.Contains(err.Error(), elastic.RecordNotFound) {
			return err
		}
	}

	if err := data.Ack(false); err != nil {
		return fmt.Errorf("Error acknowledging message: %s", err)
	}
	return nil
}

func listenChannel(messageQueue *mq.MQ, queue string, closeChan chan struct{}) {
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
				logger.Errorf("[listenChannel] %s", err.Error())
				helpers.LocalCatchErrorSentry(localSentry, fmt.Errorf("[listenChannel] %s", err.Error()))
			}
		}
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Metrics.Sentry.Enabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Metrics.Sentry.Project)
		defer helpers.CatchPanicSentry()
	}

	ctx = config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithRPC(cfg.RPC),
		config.WithDatabase(cfg.DB),
		config.WithRabbit(cfg.RabbitMQ),
	)
	defer ctx.Close()
	if err := ctx.LoadAliases(); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

	closeChan := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for i := range cfg.RabbitMQ.Queues {
		go listenChannel(ctx.MQ, cfg.RabbitMQ.Queues[i], closeChan)
	}

	<-signals
	close(closeChan)
}
