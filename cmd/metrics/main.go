package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/streadway/amqp"
)

var ctx *Context

func parseData(data amqp.Delivery) error {
	switch data.RoutingKey {
	case mq.QueueContracts:
		return getContract(data)
	case mq.QueueOperations:
		return getOperation(data)
	default:
		return fmt.Errorf("Unknown data routing key %s", data.RoutingKey)
	}
	return nil
}

func handler(data amqp.Delivery) error {
	if err := parseData(data); err != nil {
		return err
	}

	if err := data.Ack(false); err != nil {
		return fmt.Errorf("Error acknowledging message: %s", err)
	}
	return nil
}

func listenChannel(messageQueue *mq.MQ, queue string, closeChan chan struct{}) {
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
			}
		}
	}
}

func main() {
	var err error
	var cfg config
	if err = jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	ctx, err = newContext(cfg)
	if err != nil {
		panic(err)
	}
	defer ctx.Close()

	closeChan := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for i := range cfg.Mq.Queues {
		go listenChannel(ctx.MQ, cfg.Mq.Queues[i], closeChan)
	}

	<-signals
	close(closeChan)
}
