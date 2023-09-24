package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestBlocksGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	block, err := s.blocks.Get(ctx, 40)
	s.Require().NoError(err)

	s.Require().EqualValues(40, block.ID)
	s.Require().EqualValues(40, block.Level)
	s.Require().EqualValues(2, block.ProtocolID)
	s.Require().Equal("BL68PgM93vRHeZe9dcKJuNixLQSLq13ZguthVu8fXDJT33bcQqf", block.Hash)
	s.Require().Equal("2022-01-25T17:01:51Z", block.Timestamp.Format(time.RFC3339))
	s.Require().Equal("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", block.Protocol.Hash)
}

func (s *StorageTestSuite) TestBlocksLast() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	block, err := s.blocks.Last(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(40, block.ID)
	s.Require().EqualValues(40, block.Level)
	s.Require().EqualValues(2, block.ProtocolID)
	s.Require().Equal("BL68PgM93vRHeZe9dcKJuNixLQSLq13ZguthVu8fXDJT33bcQqf", block.Hash)
	s.Require().Equal("2022-01-25T17:01:51Z", block.Timestamp.Format(time.RFC3339))
	s.Require().Equal("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", block.Protocol.Hash)
}
