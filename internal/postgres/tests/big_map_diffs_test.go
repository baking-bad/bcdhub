package tests

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/testsuite"
)

func (s *StorageTestSuite) TestBigMapDiffsCurrent() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := s.bigMapDiffs.Current(ctx, "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo", 4)
	s.Require().NoError(err)

	s.Require().EqualValues(1, diff.ID)
	s.Require().EqualValues(4, diff.Ptr)
	s.Require().EqualValues(0, diff.Count)
	s.Require().EqualValues(33, diff.LastUpdateLevel)
	s.Require().EqualValues("KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i", diff.Contract)
	s.Require().False(diff.Removed)
	s.Require().Equal(testsuite.MustHexDecode("7b22737472696e67223a22227d"), []byte(diff.Key))
}

func (s *StorageTestSuite) TestBigMapDiffsGetForAddress() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diffs, err := s.bigMapDiffs.GetForAddress(ctx, "KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i")
	s.Require().NoError(err)
	s.Require().Len(diffs, 1)

	diff := diffs[0]
	s.Require().EqualValues(1, diff.ID)
	s.Require().EqualValues(4, diff.Ptr)
	s.Require().EqualValues(0, diff.Count)
	s.Require().EqualValues(33, diff.LastUpdateLevel)
	s.Require().EqualValues("KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i", diff.Contract)
	s.Require().False(diff.Removed)
	s.Require().Equal(testsuite.MustHexDecode("7b22737472696e67223a22227d"), []byte(diff.Key))
}

func (s *StorageTestSuite) TestBigMapDiffsGetByAddress() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diffs, err := s.bigMapDiffs.GetByAddress(ctx, "KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i")
	s.Require().NoError(err)
	s.Require().Len(diffs, 1)

	diff := diffs[0]
	s.Require().EqualValues(1, diff.ID)
	s.Require().EqualValues(4, diff.Ptr)
	s.Require().EqualValues(33, diff.Level)
	s.Require().EqualValues(2, diff.ProtocolID)
	s.Require().EqualValues(34, diff.OperationID)
	s.Require().EqualValues("KT1W3fGSo8XfRSESPAg3Jngzt3D8xpPqW64i", diff.Contract)
	s.Require().Equal("expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo", diff.KeyHash)
	s.Require().Equal(testsuite.MustHexDecode("7b22737472696e67223a22227d"), []byte(diff.Key))
}

func (s *StorageTestSuite) TestBigMapDiffsCount() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	count, err := s.bigMapDiffs.Count(ctx, 4)
	s.Require().NoError(err)
	s.Require().EqualValues(1, count)
}

func (s *StorageTestSuite) TestBigMapDiffsPrevious() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	previous, err := s.bigMapDiffs.Previous(ctx, []bigmapdiff.BigMapDiff{
		{
			ID:       55,
			Ptr:      41,
			KeyHash:  "exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX",
			Contract: "KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1",
		},
	})
	s.Require().NoError(err)
	s.Require().Len(previous, 1)

	diff := previous[0]
	s.Require().EqualValues(54, diff.ID)
	s.Require().EqualValues(41, diff.Ptr)
	s.Require().EqualValues(40, diff.Level)
	s.Require().EqualValues(2, diff.ProtocolID)
	s.Require().EqualValues(109, diff.OperationID)
	s.Require().EqualValues("KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1", diff.Contract)
	s.Require().Equal("exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", diff.KeyHash)
	s.Require().Equal(testsuite.MustHexDecode("11223344556677889900"), []byte(diff.Value))
}

func (s *StorageTestSuite) TestBigMapDiffsGetForOperation() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diffs, err := s.bigMapDiffs.GetForOperation(ctx, 109)
	s.Require().NoError(err)
	s.Require().Len(diffs, 5)
}

func (s *StorageTestSuite) TestBigMapDiffsGetByPtrAndKeyHash() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diffs, count, err := s.bigMapDiffs.GetByPtrAndKeyHash(ctx, 41, "exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", 10, 0)
	s.Require().NoError(err)
	s.Require().Len(diffs, 2)
	s.Require().EqualValues(2, count)
}

func (s *StorageTestSuite) TestBigMapDiffsGetByPtr() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	states, err := s.bigMapDiffs.GetByPtr(ctx, "KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1", 41)
	s.Require().NoError(err)
	s.Require().Len(states, 2)
}

func (s *StorageTestSuite) TestBigMapDiffsGet() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diffs, err := s.bigMapDiffs.Get(ctx, bigmapdiff.GetContext{
		Ptr:  testsuite.Ptr[int64](41),
		Size: 10,
	})
	s.Require().NoError(err)
	s.Require().Len(diffs, 2)
}

func (s *StorageTestSuite) TestBigMapDiffsGetStats() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stats, err := s.bigMapDiffs.GetStats(ctx, 41)
	s.Require().NoError(err)
	s.Require().EqualValues(2, stats.Total)
	s.Require().EqualValues(2, stats.Active)
	s.Require().EqualValues("KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1", stats.Contract)
}

func (s *StorageTestSuite) TestBigMapDiffsCurrentByContract() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := s.bigMapDiffs.CurrentByContract(ctx, "KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1")
	s.Require().NoError(err)
	s.Require().Len(diff, 4)
}

func (s *StorageTestSuite) TestBigMapDiffsStatesChangedAtLevel() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := s.bigMapDiffs.StatesChangedAtLevel(ctx, 40)
	s.Require().NoError(err)
	s.Require().Len(diff, 6)
}

func (s *StorageTestSuite) TestBigMapDiffsLastDiff() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	diff, err := s.bigMapDiffs.LastDiff(ctx, 41, "exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", true)
	s.Require().NoError(err)

	s.Require().EqualValues(55, diff.ID)
	s.Require().EqualValues(41, diff.Ptr)
	s.Require().EqualValues(40, diff.Level)
	s.Require().EqualValues(2, diff.ProtocolID)
	s.Require().EqualValues(109, diff.OperationID)
	s.Require().EqualValues("KT1NSpRTVR4MUwx64XCADXDUmpMGQw5yVNK1", diff.Contract)
	s.Require().Equal("exprurUjYU5axnk1qjE6F2t7uDtqR64bnsxGu3AHfWiVREftRDcRPX", diff.KeyHash)
}

func (s *StorageTestSuite) TestBigMapDiffsKeys() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	states, err := s.bigMapDiffs.Keys(ctx, bigmapdiff.GetContext{
		Ptr: testsuite.Ptr[int64](41),
	})
	s.Require().NoError(err)
	s.Require().Len(states, 2)
}
