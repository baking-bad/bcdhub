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
	subscribers map[chan Data]struct{}
	mux         sync.Mutex
}

// NewDefaultSource -
func NewDefaultSource() *DefaultSource {
	return &DefaultSource{
		subscribers: make(map[chan Data]struct{}),
	}
}

// Subscribe -
func (s *DefaultSource) Subscribe() chan Data {
	ch := make(chan Data)
	s.mux.Lock()
	defer s.mux.Unlock()

	s.subscribers[ch] = struct{}{}

	return ch
}

// Unsubscribe -
func (s *DefaultSource) Unsubscribe(ch chan Data) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribers[ch]; ok {
		close(ch)
		delete(s.subscribers, ch)
	}
}

// Data -
type Data struct {
	Type string
	Body interface{}
}
