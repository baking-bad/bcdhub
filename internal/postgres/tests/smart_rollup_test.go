package tests

import (
	"context"
	"time"
)

func (s *StorageTestSuite) TestSrList() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	smartRollups, err := s.smartRollups.List(ctx, 3, 0, "ASC")
	s.Require().NoError(err)
	s.Require().Len(smartRollups, 3)

	sr := smartRollups[0]
	s.Require().EqualValues(1, sr.ID)
	s.Require().NotEmpty(sr.Type)
	s.Require().NotEmpty(sr.Kernel)
	s.Require().Equal("wasm_2_0_0", sr.PvmKind)
	s.Require().EqualValues(6552, sr.Size)
	s.Require().EqualValues("sr1BP9kkXc1T4sRpX4kZuQoWKauLkMjijEDv", sr.Address.Address)
}

func (s *StorageTestSuite) TestSrGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sr, err := s.smartRollups.Get(ctx, "sr1BP9kkXc1T4sRpX4kZuQoWKauLkMjijEDv")
	s.Require().NoError(err)

	s.Require().EqualValues(1, sr.ID)
	s.Require().NotEmpty(sr.Type)
	s.Require().NotEmpty(sr.Kernel)
	s.Require().Equal("wasm_2_0_0", sr.PvmKind)
	s.Require().EqualValues(6552, sr.Size)
	s.Require().EqualValues("sr1BP9kkXc1T4sRpX4kZuQoWKauLkMjijEDv", sr.Address.Address)
}
