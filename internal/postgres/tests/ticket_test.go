package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestTicketGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates, err := s.ticketUpdates.Get(ctx, "KT1SM849krq9FFxGWCZyc7X5GvAz8XnRmXnf", 10, 0)
	s.Require().NoError(err)
	s.Require().Len(updates, 2)

	update := updates[0]
	s.Require().EqualValues(2, update.ID)
	s.Require().EqualValues(104, update.OperationID)
	s.Require().EqualValues(40, update.Level)
	s.Require().EqualValues(133, update.TicketerID)
	s.Require().EqualValues(131, update.AccountID)
	s.Require().EqualValues("43", update.Amount.String())
	s.Require().NotEmpty(update.Content)
	s.Require().NotEmpty(update.ContentType)
}

func (s *StorageTestSuite) TestTicketHas() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ok, err := s.ticketUpdates.Has(ctx, 133)
	s.Require().NoError(err)
	s.Require().True(ok)

	ok, err = s.ticketUpdates.Has(ctx, 1)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *StorageTestSuite) TestTicketForOperation() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates, err := s.ticketUpdates.ForOperation(ctx, 104)
	s.Require().NoError(err)
	s.Require().Len(updates, 2)

	update := updates[1]
	s.Require().EqualValues(2, update.ID)
	s.Require().EqualValues(104, update.OperationID)
	s.Require().EqualValues(40, update.Level)
	s.Require().EqualValues(133, update.TicketerID)
	s.Require().EqualValues(131, update.AccountID)
	s.Require().EqualValues("43", update.Amount.String())
	s.Require().NotEmpty(update.Content)
	s.Require().NotEmpty(update.ContentType)
}
