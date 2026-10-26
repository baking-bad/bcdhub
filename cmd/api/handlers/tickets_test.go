package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/account"
	mock_operation "github.com/baking-bad/bcdhub/internal/models/mock/operation"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/testsuite"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testTicketTimestamp1 = time.Date(2022, 1, 25, 16, 54, 35, 0, time.UTC)
	testTicketTimestamp2 = time.Date(2022, 1, 25, 17, 1, 51, 0, time.UTC)
	testOperationHash1   = testsuite.MustHexDecode("a47ce950e17c7f06af8e9e992ae10ad2b7aca0cf8a72c0ce451684f673f20324")
	testOperationHash2   = testsuite.MustHexDecode("d5567a8596a08a9cc49e22d7f5d8741161a1146f8369e897242bb8d59a8f6a7a")
)

func testTicketUpdate(id, operationID int64, ts time.Time) ticket.TicketUpdate {
	return ticket.TicketUpdate{
		ID:          id,
		Timestamp:   ts,
		OperationId: operationID,
		Level:       40,
		TicketId:    1,
		AccountId:   105,
		Amount:      decimal.NewFromInt(42),
		Account:     account.Account{Address: "tz1aSkwEot3L2kmUvcoxzjMomb9mvBNuzFK6"},
		Ticket: ticket.Ticket{
			ID:          1,
			Ticketer:    account.Account{Address: "KT1B74ywpG3nX2F3ZgpYi4MnsyuoYyQtQTAn"},
			ContentType: []byte(`{"prim":"string"}`),
			Content:     []byte(`{"string":"abc"}`),
		},
	}
}

func TestPrepareTicketUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operations := mock_operation.NewMockRepository(ctrl)
	ctx := &config.Context{Operations: operations}

	updates := []ticket.TicketUpdate{
		testTicketUpdate(1, 67, testTicketTimestamp1),
		testTicketUpdate(2, 104, testTicketTimestamp2),
		testTicketUpdate(3, 104, testTicketTimestamp2),
	}

	operations.EXPECT().
		GetByCursors(gomock.Any(),
			operation.Cursor{ID: 67, Timestamp: testTicketTimestamp1},
			operation.Cursor{ID: 104, Timestamp: testTicketTimestamp2},
			operation.Cursor{ID: 104, Timestamp: testTicketTimestamp2},
		).
		Return([]operation.Operation{
			{ID: 104, Hash: testOperationHash2},
			{ID: 67, Hash: testOperationHash1},
		}, nil).
		Times(1)

	result, err := prepareTicketUpdates(context.Background(), ctx, updates, nil)
	require.NoError(t, err)
	require.Len(t, result, 3)

	wantHashes := []string{
		encoding.MustEncodeOperationHash(testOperationHash1),
		encoding.MustEncodeOperationHash(testOperationHash2),
		encoding.MustEncodeOperationHash(testOperationHash2),
	}
	for i := range result {
		require.EqualValues(t, updates[i].ID, result[i].ID)
		require.Equal(t, wantHashes[i], result[i].OperationHash)
		require.Equal(t, "42", result[i].Amount)
		require.Equal(t, "KT1B74ywpG3nX2F3ZgpYi4MnsyuoYyQtQTAn", result[i].Ticketer)
		require.Equal(t, "tz1aSkwEot3L2kmUvcoxzjMomb9mvBNuzFK6", result[i].Address)
		require.NotEmpty(t, result[i].ContentType)
		require.NotNil(t, result[i].Content)
	}
}

func TestPrepareTicketUpdatesWithHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operations := mock_operation.NewMockRepository(ctrl)
	ctx := &config.Context{Operations: operations}

	updates := []ticket.TicketUpdate{
		testTicketUpdate(1, 67, testTicketTimestamp1),
		testTicketUpdate(2, 67, testTicketTimestamp1),
	}

	result, err := prepareTicketUpdates(context.Background(), ctx, updates, testOperationHash1)
	require.NoError(t, err)
	require.Len(t, result, 2)

	want := encoding.MustEncodeOperationHash(testOperationHash1)
	for i := range result {
		require.Equal(t, want, result[i].OperationHash)
	}
}

func TestPrepareTicketUpdatesEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// no expectations: any repository call fails the test
	operations := mock_operation.NewMockRepository(ctrl)
	ctx := &config.Context{Operations: operations}

	result, err := prepareTicketUpdates(context.Background(), ctx, nil, nil)
	require.NoError(t, err)
	require.Empty(t, result)
	require.NotNil(t, result)
}

func TestPrepareTicketUpdatesMissingOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operations := mock_operation.NewMockRepository(ctrl)
	ctx := &config.Context{Operations: operations}

	updates := []ticket.TicketUpdate{
		testTicketUpdate(1, 67, testTicketTimestamp1),
		testTicketUpdate(2, 104, testTicketTimestamp2),
	}

	operations.EXPECT().
		GetByCursors(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]operation.Operation{
			{ID: 67, Hash: testOperationHash1},
		}, nil).
		Times(1)

	_, err := prepareTicketUpdates(context.Background(), ctx, updates, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "operation_id=104")
}

func TestPrepareTicketUpdatesStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operations := mock_operation.NewMockRepository(ctrl)
	ctx := &config.Context{Operations: operations}

	errStorage := errors.New("storage error")
	operations.EXPECT().
		GetByCursors(gomock.Any(), gomock.Any()).
		Return(nil, errStorage).
		Times(1)

	_, err := prepareTicketUpdates(context.Background(), ctx, []ticket.TicketUpdate{
		testTicketUpdate(1, 67, testTicketTimestamp1),
	}, nil)
	require.ErrorIs(t, err, errStorage)
}
