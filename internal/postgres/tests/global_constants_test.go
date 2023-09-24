package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestGCGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	constant, err := s.globalConstants.Get(ctx, "expruv45XuhGc4fdRzTwwXpmp2ZyqwmUYeMmnKbxkCn5Q8uCtwkhM6")
	s.Require().NoError(err)

	s.Require().EqualValues(1, constant.ID)
	s.Require().EqualValues("expruv45XuhGc4fdRzTwwXpmp2ZyqwmUYeMmnKbxkCn5Q8uCtwkhM6", constant.Address)
	s.Require().EqualValues("2022-01-23T17:10:55Z", constant.Timestamp.Format(time.RFC3339))
	s.Require().NotEmpty(constant.Value)
	s.Require().EqualValues(30, constant.Level)
}

func (s *StorageTestSuite) TestGCAll() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	constants, err := s.globalConstants.All(
		ctx,
		"expruv45XuhGc4fdRzTwwXpmp2ZyqwmUYeMmnKbxkCn5Q8uCtwkhM6",
		"expru5X5fvCer8tbRkSAtwyVCs9FUCq46JQG7QCAkhZSumjbZBUGzb",
	)
	s.Require().NoError(err)
	s.Require().Len(constants, 2)
}

func (s *StorageTestSuite) TestGCList() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	constants, err := s.globalConstants.List(ctx, 10, 0, "links_count", "desc")
	s.Require().NoError(err)
	s.Require().Len(constants, 3)
}

func (s *StorageTestSuite) TestGCForContract() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	constants, err := s.globalConstants.ForContract(ctx, "KT1AafHA1C1vk959wvHWBispY9Y2f3fxBUUo", 10, 0)
	s.Require().NoError(err)
	s.Require().Len(constants, 2)
}

func (s *StorageTestSuite) TestGCContractList() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contracts, err := s.globalConstants.ContractList(ctx, "expruv45XuhGc4fdRzTwwXpmp2ZyqwmUYeMmnKbxkCn5Q8uCtwkhM6", 10, 0)
	s.Require().NoError(err)
	s.Require().Len(contracts, 1)
}
