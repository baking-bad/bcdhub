package mq

import (
	"sync"

	"github.com/pkg/errors"
)

// Sandbox -
type Sandbox struct {
	service string
	queues  []Queue

	channels sync.Map
}

// SandboxMessage -
type SandboxMessage struct {
	body []byte
	key  string
}

// GetBody -
func (sm *SandboxMessage) GetBody() []byte {
	return sm.body
}

// GetKey -
func (sm *SandboxMessage) GetKey() string {
	return sm.key
}

// Ack -
func (sm *SandboxMessage) Ack(flag bool) error {
	return nil
}

// NewSandbox -
func NewSandbox(service string, queues ...Queue) *Sandbox {
	s := Sandbox{
		service: service,
		queues:  queues,
	}

	for i := range queues {
		s.channels.Store(getQueueName(service, queues[i].Name), make(chan Data))
	}

	return &s
}

// SendRaw -
func (s *Sandbox) SendRaw(queue string, body []byte) error {
	name := getQueueName(s.service, queue)
	value, _ := s.channels.LoadOrStore(name, make(chan Data))
	channel := value.(chan Data)
	channel <- &SandboxMessage{
		body, queue,
	}
	return nil
}

// Send -
func (s *Sandbox) Send(msg IMessage) error {
	queues := msg.GetQueues()
	if len(queues) == 0 {
		return nil
	}
	message, err := msg.MarshalToQueue()
	if err != nil {
		return err
	}
	for _, queue := range queues {
		if err := s.SendRaw(queue, message); err != nil {
			return err
		}
	}
	return nil
}

// Consume -
func (s *Sandbox) Consume(queue string) (<-chan Data, error) {
	name := getQueueName(s.service, queue)
	value, ok := s.channels.Load(name)
	if !ok {
		return nil, errors.Wrap(ErrUnknownQueue, name)
	}
	return value.(chan Data), nil
}

// GetQueues -
func (s *Sandbox) GetQueues() []string {
	queues := make([]string, len(s.queues))
	for i := range s.queues {
		queues[i] = s.queues[i].Name
	}
	return queues
}

// Close -
func (s *Sandbox) Close() error {
	s.channels.Range(func(key, value interface{}) bool {
		channel := value.(chan Data)
		close(channel)
		s.channels.Delete(key)
		return true
	})
	return nil
}
