package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestAccountsGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	account, err := s.accounts.Get(ctx, "KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br")
	s.Require().NoError(err)

	s.Require().EqualValues(45, account.ID)
	s.Require().EqualValues(1, account.Type)
	s.Require().EqualValues("KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br", account.Address)
}
