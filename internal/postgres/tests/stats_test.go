package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestRollbackStats() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stats, err := s.stats.Get(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(120, stats.ContractsCount)
	s.Require().EqualValues(192, stats.OperationsCount)
	s.Require().EqualValues(72, stats.TransactionsCount)
	s.Require().EqualValues(118, stats.OriginationsCount)
	s.Require().EqualValues(2, stats.EventsCount)
	s.Require().EqualValues(0, stats.SrOriginationsCount)
}
