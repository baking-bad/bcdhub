package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

func (s *StorageTestSuite) TestContractGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contract, err := s.contracts.Get(ctx, "KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br")
	s.Require().NoError(err)

	s.Require().EqualValues(4, contract.ID)
	s.Require().EqualValues(33, contract.Level)
	s.Require().EqualValues(37, contract.Account.ID)
	s.Require().EqualValues(1, contract.Account.Type)
	s.Require().EqualValues("KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br", contract.Account.Address)
}

func (s *StorageTestSuite) TestContractByHash() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	script, err := s.contracts.ByHash(ctx, "95d118c83ad81586ba4a39c07a47ff7804f5e6d1ebe1a943016d0c61b4940fb6")
	s.Require().NoError(err)

	s.Require().EqualValues(18, script.ID)
	s.Require().NotEmpty(script.Parameter)
	s.Require().NotEmpty(script.Code)
	s.Require().NotEmpty(script.Storage)
	s.Require().NotEmpty(script.Entrypoints)
}

func (s *StorageTestSuite) TestContractScript() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	script, err := s.contracts.Script(ctx, "KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br", bcd.SymLinkBabylon)
	s.Require().NoError(err)

	s.Require().EqualValues(4, script.ID)
	s.Require().NotEmpty(script.Parameter)
	s.Require().NotEmpty(script.Code)
	s.Require().NotEmpty(script.Storage)
	s.Require().NotEmpty(script.Entrypoints)
}

func (s *StorageTestSuite) TestContractParameter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.Parameter(ctx, 4)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

func (s *StorageTestSuite) TestContractStorage() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.Storage(ctx, 4)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

func (s *StorageTestSuite) TestContractCode() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.Code(ctx, 4)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

func (s *StorageTestSuite) TestContractViews() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.Views(ctx, 19)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

func (s *StorageTestSuite) TestContractScriptPart() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.ScriptPart(ctx, "KT1CMJQmuwwJopNnLhSDHXT3zQVUrNPLA8br", bcd.SymLinkBabylon, consts.STORAGE)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

func (s *StorageTestSuite) TestContractRecentlyCalled() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := s.contracts.RecentlyCalled(ctx, 0, 3)
	s.Require().NoError(err)
	s.Require().Len(data, 3)
}

func (s *StorageTestSuite) TestContractCount() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	count, err := s.contracts.Count(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(56, count)
}

func (s *StorageTestSuite) TestContractFindOne() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contract, err := s.contracts.FindOne(ctx, types.FA2Tag)
	s.Require().NoError(err)
	s.Require().Positive(contract.ID)
}
