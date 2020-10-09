package datasources

import "sync"

// Data source types
const (
	RabbitType = "rabbit"
)

// DataSource -
type DataSource interface {
	Run()
	Stop()
	GetType() string
	Subscribe() chan Data
	Unsubscribe(chan Data)
}

// DefaultSource -
type DefaultSource struct {
	subscribers sync.Map
}

// NewDefaultSource -
func NewDefaultSource() *DefaultSource {
	return &DefaultSource{}
}

// Subscribe -
func (s *DefaultSource) Subscribe() chan Data {
	ch := make(chan Data)
	s.subscribers.Store(ch, struct{}{})
	return ch
}

// Unsubscribe -
func (s *DefaultSource) Unsubscribe(ch chan Data) {
	if _, ok := s.subscribers.Load(ch); ok {
		close(ch)
		s.subscribers.Delete(ch)
	}
}

// Data -
type Data struct {
	Type string
	Body interface{}
}
