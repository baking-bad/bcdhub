package tests

import (
	"context"
	"sort"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/testsuite"
)

func (s *StorageTestSuite) TestOperationsLast() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operation, err := s.operations.Last(ctx, map[string]interface{}{
		"destination_id": 88,
	}, -1)
	s.Require().NoError(err)

	s.Require().EqualValues(67, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(37, operation.Level)
	s.Require().EqualValues(540, operation.Counter)
	s.Require().EqualValues(518, operation.Fee)
	s.Require().EqualValues(2155, operation.GasLimit)
	s.Require().EqualValues(0, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(2054659, operation.ConsumedGas)
	s.Require().EqualValues(87, operation.StorageSize)
	s.Require().EqualValues(92, operation.InitiatorID)
	s.Require().EqualValues(92, operation.SourceID)
	s.Require().EqualValues(88, operation.DestinationID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().EqualValues("@entrypoint_0", operation.Entrypoint.String())
	s.Require().NotEmpty(operation.Parameters)
	s.Require().NotEmpty(operation.DeffatedStorage)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationsGetByHash() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	hash := testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324")

	operations, err := s.operations.GetByHash(ctx, hash)
	s.Require().NoError(err)
	s.Require().Len(operations, 1)

	operation := operations[0]

	s.Require().EqualValues(67, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(37, operation.Level)
	s.Require().EqualValues(540, operation.Counter)
	s.Require().EqualValues(518, operation.Fee)
	s.Require().EqualValues(2155, operation.GasLimit)
	s.Require().EqualValues(0, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(2054659, operation.ConsumedGas)
	s.Require().EqualValues(87, operation.StorageSize)
	s.Require().EqualValues(92, operation.InitiatorID)
	s.Require().EqualValues(92, operation.SourceID)
	s.Require().EqualValues(88, operation.DestinationID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().EqualValues("@entrypoint_0", operation.Entrypoint.String())
	s.Require().NotEmpty(operation.Parameters)
	s.Require().NotEmpty(operation.DeffatedStorage)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationsGetById() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operation, err := s.operations.GetByID(ctx, 67)
	s.Require().NoError(err)

	s.Require().EqualValues(67, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(37, operation.Level)
	s.Require().EqualValues(540, operation.Counter)
	s.Require().EqualValues(518, operation.Fee)
	s.Require().EqualValues(2155, operation.GasLimit)
	s.Require().EqualValues(0, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(2054659, operation.ConsumedGas)
	s.Require().EqualValues(87, operation.StorageSize)
	s.Require().EqualValues(92, operation.InitiatorID)
	s.Require().EqualValues(92, operation.SourceID)
	s.Require().EqualValues(88, operation.DestinationID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().EqualValues("@entrypoint_0", operation.Entrypoint.String())
	s.Require().NotEmpty(operation.Parameters)
	s.Require().NotEmpty(operation.DeffatedStorage)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationGroups() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	opg, err := s.operations.OPG(ctx, "KT18sWFGnjvobw9BiHrm9bjpWr3AFAAVas9w", 10, -1)
	s.Require().NoError(err)
	s.Require().Len(opg, 2)

	operation := opg[0]
	s.Require().EqualValues(67, operation.LastID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(37, operation.Level)
	s.Require().EqualValues(540, operation.Counter)
	s.Require().EqualValues(0, operation.Flow)
	s.Require().EqualValues(518, operation.TotalCost)
	s.Require().EqualValues(0, operation.Internals)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationGetByHashAndCounter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	hash := testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324")
	operations, err := s.operations.GetByHashAndCounter(ctx, hash, 540)
	s.Require().NoError(err)
	s.Require().Len(operations, 1)

	operation := operations[0]
	s.Require().EqualValues(67, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(37, operation.Level)
	s.Require().EqualValues(540, operation.Counter)
	s.Require().EqualValues(518, operation.Fee)
	s.Require().EqualValues(2155, operation.GasLimit)
	s.Require().EqualValues(0, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(2054659, operation.ConsumedGas)
	s.Require().EqualValues(87, operation.StorageSize)
	s.Require().EqualValues(92, operation.InitiatorID)
	s.Require().EqualValues(92, operation.SourceID)
	s.Require().EqualValues(88, operation.DestinationID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().EqualValues("@entrypoint_0", operation.Entrypoint.String())
	s.Require().NotEmpty(operation.Parameters)
	s.Require().NotEmpty(operation.DeffatedStorage)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationGetImplicitOperation() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operations, err := s.operations.GetByHashAndCounter(ctx, nil, 2)
	s.Require().NoError(err)
	s.Require().Len(operations, 1)

	operation := operations[0]

	s.Require().EqualValues(1, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(2, operation.Level)
	s.Require().EqualValues(2, operation.Counter)
	s.Require().EqualValues(0, operation.Fee)
	s.Require().EqualValues(0, operation.GasLimit)
	s.Require().EqualValues(0, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(2468, operation.ConsumedGas)
	s.Require().EqualValues(4630, operation.StorageSize)
	s.Require().EqualValues(0, operation.InitiatorID)
	s.Require().EqualValues(0, operation.SourceID)
	s.Require().EqualValues(2, operation.DestinationID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindTransaction, operation.Kind)
	s.Require().Empty(operation.Parameters)
	s.Require().NotEmpty(operation.DeffatedStorage)
	s.Require().Empty(operation.Hash)
}

func (s *StorageTestSuite) TestOperationListEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operations, err := s.operations.ListEvents(ctx, 37, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(operations, 2)

	operation := operations[0]
	s.Require().EqualValues(192, operation.ID)
	s.Require().EqualValues(0, operation.ContentIndex)
	s.Require().EqualValues(41, operation.Level)
	s.Require().EqualValues(136723, operation.Counter)
	s.Require().EqualValues(0, operation.Fee)
	s.Require().EqualValues(1818, operation.GasLimit)
	s.Require().EqualValues(60000, operation.StorageLimit)
	s.Require().EqualValues(0, operation.Amount)
	s.Require().EqualValues(100000, operation.ConsumedGas)
	s.Require().EqualValues(0, operation.StorageSize)
	s.Require().EqualValues(42, operation.InitiatorID)
	s.Require().EqualValues(37, operation.SourceID)
	s.Require().EqualValues(types.OperationStatusApplied, operation.Status)
	s.Require().EqualValues(types.OperationKindEvent, operation.Kind)
	s.Require().EqualValues("add_account_event", operation.Tag.String())
	s.Require().NotEmpty(operation.Payload)
	s.Require().NotEmpty(operation.PayloadType)
	s.Require().Equal(testsuite.MustHexDecode("3006fe3748e23bee8499ddd4ef69c3f910b1de0aa04080cc5be242b5123c1207"), operation.Hash)
}

func (s *StorageTestSuite) TestOperationsGetByCursors() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operations, err := s.operations.GetByCursors(ctx,
		operation.Cursor{ID: 67, Timestamp: time.Date(2022, 1, 25, 16, 54, 35, 0, time.UTC)},
		operation.Cursor{ID: 104, Timestamp: time.Date(2022, 1, 25, 17, 1, 51, 0, time.UTC)},
	)
	s.Require().NoError(err)
	s.Require().Len(operations, 2)

	sort.Slice(operations, func(i, j int) bool { return operations[i].ID < operations[j].ID })

	s.Require().EqualValues(67, operations[0].ID)
	s.Require().Equal(testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324"), operations[0].Hash)
	s.Require().EqualValues(104, operations[1].ID)
	s.Require().Equal(testsuite.MustHexDecode("d5567a8596a08a9cc49e22d7f5d8741161a1146f8369e897242bb8d59a8f6a7a"), operations[1].Hash)
}

func (s *StorageTestSuite) TestOperationsGetByCursorsWrongTimestamp() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operations, err := s.operations.GetByCursors(ctx,
		operation.Cursor{ID: 67, Timestamp: time.Date(2022, 1, 25, 17, 1, 51, 0, time.UTC)},
	)
	s.Require().NoError(err)
	s.Require().Empty(operations)
}

func (s *StorageTestSuite) TestOperationsGetByCursorsEmpty() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	operations, err := s.operations.GetByCursors(ctx)
	s.Require().NoError(err)
	s.Require().Empty(operations)
}
