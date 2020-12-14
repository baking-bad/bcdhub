package mq

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func getQueueName(service, queue string) string {
	return fmt.Sprintf("%s.%s", queue, service)
}

// New -
func New(url, service string, needPublisher bool, timeout int, queues ...Queue) Mediator {
	switch url {
	case SandboxURL:
		n, err := NewNats(service, queues...)
		if err != nil {
			logger.Error(err)
			return nil
		}
		return n
	default:
		return WaitNew(url, service, needPublisher, timeout, queues...)
	}
}
