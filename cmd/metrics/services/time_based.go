package services

import (
	"context"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
)

// TimeBased -
type TimeBased struct {
	handler func(ctx context.Context) error
	period  time.Duration

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewTimeBased -
func NewTimeBased(handler func(ctx context.Context) error, period time.Duration) *TimeBased {
	return &TimeBased{
		period:  period,
		handler: handler,
		stop:    make(chan struct{}, 1),
	}
}

// Init -
func (s *TimeBased) Init() error {
	return nil
}

// Start -
func (s *TimeBased) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.work(ctx)
}

// Close -
func (s *TimeBased) Close() error {
	s.stop <- struct{}{}
	s.wg.Wait()

	close(s.stop)
	return nil
}

func (s *TimeBased) work(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	// init event
	if err := s.handler(ctx); err != nil {
		logger.Err(err)
	}

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			if err := s.handler(ctx); err != nil {
				logger.Err(err)
			}
		}
	}
}
