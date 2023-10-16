package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestBlocksGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	block, err := s.blocks.Get(ctx, 47)
	s.Require().NoError(err)

	s.Require().EqualValues(47, block.ID)
	s.Require().EqualValues(47, block.Level)
	s.Require().EqualValues(3, block.ProtocolID)
	s.Require().Equal("BLwSEbi7iNcW8Cu6wMzN93aHasWudPFL3An62k52SzfH4gHaXf4", block.Hash)
	s.Require().Equal("2022-01-25T17:17:47Z", block.Timestamp.Format(time.RFC3339))
	s.Require().Equal("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", block.Protocol.Hash)
}

func (s *StorageTestSuite) TestBlocksLast() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	block, err := s.blocks.Last(ctx)
	s.Require().NoError(err)

	s.Require().EqualValues(47, block.ID)
	s.Require().EqualValues(47, block.Level)
	s.Require().EqualValues(3, block.ProtocolID)
	s.Require().Equal("BLwSEbi7iNcW8Cu6wMzN93aHasWudPFL3An62k52SzfH4gHaXf4", block.Hash)
	s.Require().Equal("2022-01-25T17:17:47Z", block.Timestamp.Format(time.RFC3339))
	s.Require().Equal("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", block.Protocol.Hash)
}
