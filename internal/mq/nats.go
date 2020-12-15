package mq

import (
	"strings"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	nats "github.com/nats-io/nats.go"
)

// NatsMessage -
type NatsMessage struct {
	msg *nats.Msg
}

// GetBody -
func (nm *NatsMessage) GetBody() []byte {
	return nm.msg.Data
}

// GetKey -
func (nm *NatsMessage) GetKey() string {
	parts := strings.Split(nm.msg.Subject, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// Ack -
func (nm *NatsMessage) Ack(flag bool) error {
	return nil
}

// Nats -
type Nats struct {
	service string
	queues  []Queue

	conn *nats.Conn
	data chan Data

	wg            sync.WaitGroup
	stop          chan struct{}
	subscriptions []*nats.Subscription
}

// NewNats -
func NewNats(service, url string, queues ...Queue) (*Nats, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &Nats{
		service:       service,
		queues:        queues,
		conn:          conn,
		data:          make(chan Data),
		stop:          make(chan struct{}),
		subscriptions: make([]*nats.Subscription, 0),
	}, nil
}

// WaitNewNats -
func WaitNewNats(service, url string, timeout int, queues ...Queue) *Nats {
	var n *Nats
	var err error

	for n == nil {
		n, err = NewNats(service, url, queues...)
		if err != nil {
			logger.Warning("Waiting mq up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return n
}

// SendRaw -
func (n *Nats) SendRaw(queue string, body []byte) error {
	if n.conn == nil {
		return ErrInvalidConnection
	}
	if n.conn.IsClosed() {
		return ErrConnectionIsClosed
	}
	return n.conn.Publish(queue, body)
}

// Send -
func (n *Nats) Send(msg IMessage) error {
	queues := msg.GetQueues()
	if len(queues) == 0 {
		return nil
	}
	message, err := msg.MarshalToQueue()
	if err != nil {
		return err
	}
	for _, queue := range queues {
		if err := n.SendRaw(queue, message); err != nil {
			return err
		}
	}
	return nil
}

// Consume -
func (n *Nats) Consume(queue string) (<-chan Data, error) {
	ch := make(chan *nats.Msg)
	sub, err := n.conn.ChanSubscribe(queue, ch)
	if err != nil {
		return nil, err
	}
	n.subscriptions = append(n.subscriptions, sub)

	n.wg.Add(1)
	go func(c <-chan *nats.Msg) {
		defer n.wg.Done()

		for {
			select {
			case <-n.stop:
				return
			case msg := <-c:
				n.data <- &NatsMessage{msg}
			}
		}
	}(ch)

	return n.data, nil
}

// GetQueues -
func (n *Nats) GetQueues() []string {
	queues := make([]string, len(n.queues))
	for i := range n.queues {
		queues[i] = n.queues[i].Name
	}
	return queues
}

// Close -
func (n *Nats) Close() error {
	for i := range n.subscriptions {
		if err := n.subscriptions[i].Unsubscribe(); err != nil {
			return err
		}
		n.stop <- struct{}{}
	}
	n.wg.Wait()
	n.conn.Close()
	return nil
}
