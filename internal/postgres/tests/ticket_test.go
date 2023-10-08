package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/ticket"
)

func (s *StorageTestSuite) TestTicketGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates, err := s.ticketUpdates.Updates(ctx, "KT1SM849krq9FFxGWCZyc7X5GvAz8XnRmXnf", 10, 0)
	s.Require().NoError(err)
	s.Require().Len(updates, 2)

	update := updates[0]
	s.Require().EqualValues(2, update.ID)
	s.Require().EqualValues(104, update.OperationId)
	s.Require().EqualValues(40, update.Level)
	s.Require().EqualValues(1, update.TicketId)
	s.Require().EqualValues(131, update.AccountId)
	s.Require().EqualValues("43", update.Amount.String())
	s.Require().EqualValues("KT1SM849krq9FFxGWCZyc7X5GvAz8XnRmXnf", update.Ticket.Ticketer.Address)
}

func (s *StorageTestSuite) TestTicketForOperation() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates, err := s.ticketUpdates.UpdatesForOperation(ctx, 104)
	s.Require().NoError(err)
	s.Require().Len(updates, 2)

	update := updates[1]
	s.Require().EqualValues(2, update.ID)
	s.Require().EqualValues(104, update.OperationId)
	s.Require().EqualValues(40, update.Level)
	s.Require().EqualValues(1, update.TicketId)
	s.Require().EqualValues(131, update.AccountId)
	s.Require().EqualValues("43", update.Amount.String())
	s.Require().EqualValues("KT1SM849krq9FFxGWCZyc7X5GvAz8XnRmXnf", update.Ticket.Ticketer.Address)
}

func (s *StorageTestSuite) TestBalancesForAccount() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	balances, err := s.ticketUpdates.BalancesForAccount(ctx, 131, ticket.BalanceRequest{
		Limit:               10,
		WithoutZeroBalances: true,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 2)

	balance := balances[0]
	s.Require().EqualValues(131, balance.AccountId)
	s.Require().EqualValues(1, balance.TicketId)
	s.Require().EqualValues("43", balance.Amount.String())
	s.Require().NotEmpty(balance.Ticket.Content)
	s.Require().NotEmpty(balance.Ticket.ContentType)
	s.Require().EqualValues("KT1SM849krq9FFxGWCZyc7X5GvAz8XnRmXnf", balance.Ticket.Ticketer.Address)
}

func (s *StorageTestSuite) TestBalancesForAccountEmpty() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	balances, err := s.ticketUpdates.BalancesForAccount(ctx, 12, ticket.BalanceRequest{
		Limit: 10,
	})
	s.Require().NoError(err)
	s.Require().Len(balances, 0)
}
