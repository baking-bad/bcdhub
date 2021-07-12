package services

import (
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/service"
)

// StorageBased -
type StorageBased struct {
	name         string
	handler      Handler
	updatePeriod time.Duration
	repo         service.Repository
	state        service.State
	bulkSize     int64

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewStorageBased -
func NewStorageBased(
	name string,
	repo service.Repository,
	handler Handler,
	updatePeriod time.Duration,
	bulkSize int64,
) *StorageBased {
	if bulkSize < 10 {
		bulkSize = 10
	}
	return &StorageBased{
		name:         name,
		repo:         repo,
		handler:      handler,
		updatePeriod: updatePeriod,
		bulkSize:     bulkSize,
		stop:         make(chan struct{}, 1),
	}
}

// Init -
func (s *StorageBased) Init() error {
	logger.Info().Str("name", s.name).Msg("starting service...")
	state, err := s.repo.Get(s.name)
	if err != nil {
		return err
	}
	s.state = state
	return nil
}

// Start -
func (s *StorageBased) Start() {
	s.wg.Add(1)
	go s.work()
}

// Close -
func (s *StorageBased) Close() error {
	s.stop <- struct{}{}
	s.wg.Wait()

	close(s.stop)
	return nil
}

func (s *StorageBased) work() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.updatePeriod)
	defer ticker.Stop()

	isFull, err := s.do()
	if err != nil {
		logger.Err(err)
	}

	for {
		select {

		case <-s.stop:
			return

		case <-ticker.C:
			isFull, err = s.do()
			if err != nil {
				logger.Err(err)
				continue
			}

		default:
			if isFull {
				isFull, err = s.do()
				if err != nil {
					logger.Err(err)
					continue
				}
				ticker.Reset(s.updatePeriod)
			}
		}
	}
}

func (s *StorageBased) do() (bool, error) {
	items, err := s.handler.Chunk(s.state.LastID, s.bulkSize)
	if err != nil {
		return false, err
	}

	if err := s.handler.Handle(items); err != nil {
		return false, err
	}

	if len(items) > 0 && s.state.LastID < items[len(items)-1].GetID() {
		s.state.LastID = items[len(items)-1].GetID()
		if err := s.repo.Save(s.state); err != nil {
			return false, err
		}
	}

	return len(items) == int(s.bulkSize), nil
}
