package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestProtocolGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	protocol, err := s.protocols.Get(ctx, "PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", -1)
	s.Require().NoError(err)

	s.Require().EqualValues(3, protocol.ID)
	s.Require().EqualValues(2, protocol.StartLevel)
	s.Require().EqualValues("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", protocol.Hash)
	s.Require().EqualValues("NetXnHfVqm9iesp", protocol.ChainID)
}

func (s *StorageTestSuite) TestProtocolGetByLevel() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	protocol, err := s.protocols.Get(ctx, "", 3)
	s.Require().NoError(err)

	s.Require().EqualValues(3, protocol.ID)
	s.Require().EqualValues(2, protocol.StartLevel)
	s.Require().EqualValues("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", protocol.Hash)
	s.Require().EqualValues("NetXnHfVqm9iesp", protocol.ChainID)
}

func (s *StorageTestSuite) TestProtocolGetById() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	protocol, err := s.protocols.GetByID(ctx, 3)
	s.Require().NoError(err)

	s.Require().EqualValues(3, protocol.ID)
	s.Require().EqualValues(2, protocol.StartLevel)
	s.Require().EqualValues("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx", protocol.Hash)
	s.Require().EqualValues("NetXnHfVqm9iesp", protocol.ChainID)
}
