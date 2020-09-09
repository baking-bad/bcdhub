package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/streadway/amqp"
)

// Context -
type Context struct {
	*config.Context
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Compiler.Sentry.Enabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Compiler.Sentry.Project)
		defer helpers.CatchPanicSentry()
	}

	context := &Context{
		config.NewContext(
			config.WithRPC(cfg.RPC),
			config.WithDatabase(cfg.DB),
			config.WithRabbitReceiver(cfg.RabbitMQ, "compiler"),
			config.WithRabbitPublisher(cfg.RabbitMQ),
			config.WithElasticSearch(cfg.Elastic),
		),
	}

	msgs, err := context.MQReceiver.Consume(mq.QueueCompilations)
	if err != nil {
		logger.Fatal(err)
	}

	defer context.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	logger.Info("Connected to %s queue", mq.QueueCompilations)

	for {
		select {
		case <-signals:
			logger.Info("Stopped compiler")
			return
		case msg := <-msgs:
			if err := context.handleMessage(msg); err != nil {
				logger.Error(err)
			}
		}
	}

}

func (ctx *Context) handleMessage(data amqp.Delivery) error {
	defer func(d amqp.Delivery) {
		if err := data.Ack(false); err != nil {
			logger.Errorf("Error acknowledging message: %s", err)
		}
	}(data)

	return ctx.parseData(data)
}

func (ctx *Context) parseData(data amqp.Delivery) error {
	if data.RoutingKey != mq.QueueCompilations {
		return fmt.Errorf("[parseData] Unknown data routing key %s", data.RoutingKey)
	}

	var ct compilation.Task
	if err := json.Unmarshal(data.Body, &ct); err != nil {
		return fmt.Errorf("[parseData] Unmarshal message body error: %s", err)
	}

	defer os.RemoveAll(ct.Dir) // clean up

	switch ct.Kind {
	case compilation.KindVerification:
		return ctx.verification(ct)
	case compilation.KindCompilation:
		log.Fatal("not implemented", compilation.KindCompilation)
	case compilation.KindDeployment:
		log.Fatal("not implemented", compilation.KindDeployment)
	}

	return fmt.Errorf("[parseData] Unknown compilation task kind %s", ct.Kind)
}
