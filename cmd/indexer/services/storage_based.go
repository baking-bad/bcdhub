package services

import (
	"context"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/service"
)

// StorageBased -
type StorageBased[M models.Constraint] struct {
	name         string
	handler      Handler[M]
	updatePeriod time.Duration
	repo         service.Repository
	state        service.State
	bulkSize     int

	wg sync.WaitGroup
}

// NewStorageBased -
func NewStorageBased[M models.Constraint](
	name string,
	repo service.Repository,
	handler Handler[M],
	updatePeriod time.Duration,
	bulkSize int,
) *StorageBased[M] {
	if bulkSize < 10 {
		bulkSize = 10
	}
	return &StorageBased[M]{
		name:         name,
		repo:         repo,
		handler:      handler,
		updatePeriod: updatePeriod,
		bulkSize:     bulkSize,
	}
}

// Init -
func (s *StorageBased[M]) Init() error {
	logger.Info().Str("name", s.name).Msg("starting service...")
	state, err := s.repo.Get(s.name)
	if err != nil {
		return err
	}
	s.state = state
	return nil
}

// Start -
func (s *StorageBased[M]) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.work(ctx)
}

// Close -
func (s *StorageBased[M]) Close() error {
	s.wg.Wait()
	return nil
}

func (s *StorageBased[M]) work(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.updatePeriod)
	defer ticker.Stop()

	isFull, err := s.do(ctx, &s.wg)
	if err != nil {
		logger.Err(err)
	}

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			isFull, err = s.do(ctx, &s.wg)
			if err != nil {
				logger.Err(err)
				continue
			}

		default:
			if isFull {
				isFull, err = s.do(ctx, &s.wg)
				if err != nil {
					logger.Err(err)
					continue
				}
				ticker.Reset(s.updatePeriod)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (s *StorageBased[M]) do(ctx context.Context, wg *sync.WaitGroup) (bool, error) {
	items, err := s.handler.Chunk(s.state.LastID, s.bulkSize)
	if err != nil {
		return false, err
	}

	if err := s.handler.Handle(ctx, items, wg); err != nil {
		return false, err
	}

	if len(items) > 0 && s.state.LastID < items[len(items)-1].GetID() {
		s.state.LastID = items[len(items)-1].GetID()
		if err := s.repo.Save(s.state); err != nil {
			return false, err
		}
	}

	return len(items) == s.bulkSize, nil
}
