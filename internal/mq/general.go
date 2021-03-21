package mq

import (
	"fmt"
)

func getQueueName(service, queue string) string {
	return fmt.Sprintf("%s.%s", queue, service)
}

// New -
func New(url, service string, needPublisher bool, timeout int, queues ...Queue) Mediator {
	// switch {
	// case strings.HasPrefix(url, NatsURLPrefix):
	// 	return WaitNewNats(service, url, timeout, queues...)
	// case strings.HasPrefix(url, RabbitURLPrefix):
	return WaitNewRabbit(url, service, needPublisher, timeout, queues...)
	// default:
	// 	logger.Errorf("Unknown message queue URL: %s", url)
	// 	return nil
	// }
}
