package rollback

import (
	"database/sql"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/mock"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_stats "github.com/baking-bad/bcdhub/internal/models/mock/stats"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/testsuite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManager_Rollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	level := int64(11)
	storage := mock.NewMockGeneralRepository(ctrl)
	rb := mock.NewMockRollback(ctrl)
	blockRepo := mock_block.NewMockRepository(ctrl)
	statsRepo := mock_stats.NewMockRepository(ctrl)

	blockRepo.EXPECT().
		Get(gomock.Any(), level).
		Return(block.Block{
			Level: 11,
		}, nil).
		Times(1)

	storage.EXPECT().IsRecordNotFound(sql.ErrNoRows).Return(true).AnyTimes()

	statsRepo.EXPECT().
		Get(gomock.Any()).
		Return(stats.Stats{
			ID:                   1,
			ContractsCount:       100,
			OperationsCount:      100,
			EventsCount:          10,
			TransactionsCount:    70,
			OriginationsCount:    10,
			SrOriginationsCount:  10,
			TransferTicketsCount: 11,
		}, nil).
		Times(1)

	rb.EXPECT().
		GetOperations(gomock.Any(), level).
		Return([]operation.Operation{
			{
				Destination: account.Account{
					ID:      1,
					Address: "address_1",
					Type:    types.AccountTypeContract,
				},
				DestinationID: 1,
				Source: account.Account{
					ID:      3,
					Address: "address_3",
					Type:    types.AccountTypeTz,
				},
				SourceID: 3,
				Kind:     types.OperationKindOrigination,
			}, {
				Destination: account.Account{
					ID:      1,
					Address: "address_1",
					Type:    types.AccountTypeContract,
				},
				DestinationID: 1,
				Source: account.Account{
					ID:      2,
					Address: "address_2",
					Type:    types.AccountTypeTz,
				},
				SourceID: 2,
				Kind:     types.OperationKindTransaction,
			}, {
				Destination: account.Account{
					ID:      3,
					Address: "address_3",
					Type:    types.AccountTypeTz,
				},
				DestinationID: 3,
				Source: account.Account{
					ID:      2,
					Address: "address_2",
					Type:    types.AccountTypeTz,
				},
				SourceID: 2,
				Kind:     types.OperationKindTransaction,
			}, {
				Destination: account.Account{
					ID:      4,
					Address: "address_4",
					Type:    types.AccountTypeContract,
				},
				DestinationID: 4,
				Source: account.Account{
					ID:      2,
					Address: "address_2",
					Type:    types.AccountTypeTz,
				},
				SourceID: 2,
				Kind:     types.OperationKindTransaction,
			}, {
				Destination: account.Account{
					ID:      1,
					Address: "address_1",
					Type:    types.AccountTypeContract,
				},
				DestinationID: 1,
				Source: account.Account{
					ID:      3,
					Address: "address_3",
					Type:    types.AccountTypeTz,
				},
				SourceID: 3,
				Kind:     types.OperationKindTransaction,
			}, {
				Source: account.Account{
					ID:      1,
					Address: "address_1",
					Type:    types.AccountTypeContract,
				},
				SourceID: 1,
				Kind:     types.OperationKindEvent,
			}, {
				Destination: account.Account{
					ID:      1,
					Address: "address_1",
					Type:    types.AccountTypeContract,
				},
				DestinationID: 1,
				Source: account.Account{
					ID:      3,
					Address: "address_3",
					Type:    types.AccountTypeTz,
				},
				SourceID:           3,
				Kind:               types.OperationKindTransferTicket,
				TicketUpdatesCount: 2,
			},
		}, nil).
		Times(1)

	rb.EXPECT().
		DeleteAll(gomock.Any(), (*operation.Operation)(nil), level).
		Return(5, nil).
		Times(1)

	rb.EXPECT().
		DeleteAll(gomock.Any(), (*ticket.TicketUpdate)(nil), level).
		Return(0, nil).
		Times(1)

	rb.EXPECT().
		GetMigrations(gomock.Any(), level).
		Return([]migration.Migration{
			{
				ContractID: 1,
				Contract: contract.Contract{
					AccountID: 1,
				},
			},
		}, nil).
		Times(1)

	ts := time.Now().UTC()
	rb.EXPECT().
		GetLastAction(gomock.Any(), gomock.Any()).
		Return([]models.LastAction{
			{
				AccountId: 4,
				Time:      ts,
			}, {
				AccountId: 3,
				Time:      ts,
			}, {
				AccountId: 2,
				Time:      ts,
			}, {
				AccountId: 1,
				Time:      ts,
			},
		}, nil).
		Times(1)

	rb.EXPECT().
		GetTicketUpdates(gomock.Any(), gomock.Any()).
		Return([]ticket.TicketUpdate{
			{
				AccountId: 4,
				TicketId:  1,
				Amount:    decimal.RequireFromString("100"),
			}, {
				AccountId: 1,
				TicketId:  1,
				Amount:    decimal.RequireFromString("-100"),
			},
		}, nil).
		Times(1)

	rb.EXPECT().
		TicketBalances(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	rb.EXPECT().
		DeleteTickets(gomock.Any(), level).
		Return([]int64{1}, nil).
		Times(1)

	rb.EXPECT().
		DeleteTicketBalances(gomock.Any(), []int64{1}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		UpdateAccountStats(gomock.Any(), account.Account{
			ID:              4,
			LastAction:      ts,
			OperationsCount: 1,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		UpdateAccountStats(gomock.Any(), account.Account{
			ID:              3,
			LastAction:      ts,
			OperationsCount: 4,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		UpdateAccountStats(gomock.Any(), account.Account{
			ID:              2,
			LastAction:      ts,
			OperationsCount: 3,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		UpdateAccountStats(gomock.Any(), account.Account{
			ID:                 1,
			LastAction:         ts,
			OperationsCount:    5,
			MigrationsCount:    1,
			EventsCount:        1,
			TicketUpdatesCount: 2,
		}).
		Times(1)

	rb.EXPECT().
		StatesChangedAtLevel(gomock.Any(), level).
		Return([]bigmapdiff.BigMapState{
			{
				ID:              1,
				Ptr:             10,
				LastUpdateLevel: 11,
				Count:           10,
				LastUpdateTime:  ts,
				KeyHash:         "key_hash",
				Contract:        "address_1",
				Key:             types.MustNewBytes("deadbeaf"),
				Value:           types.MustNewBytes("00112233"),
				Removed:         false,
			}, {
				ID:              2,
				Ptr:             10,
				LastUpdateLevel: 11,
				Count:           10,
				LastUpdateTime:  ts,
				KeyHash:         "key_hash_2",
				Contract:        "address_1",
				Key:             types.MustNewBytes("deadbeaf0011"),
				Value:           types.MustNewBytes("001122334455"),
				Removed:         false,
			},
		}, nil).
		Times(1)

	ptr := int64(10)
	rb.EXPECT().
		LastDiff(gomock.Any(), ptr, "key_hash", false).
		Return(bigmapdiff.BigMapDiff{
			ID:          1,
			Ptr:         ptr,
			KeyHash:     "key_hash",
			Contract:    "address_1",
			Key:         types.MustNewBytes("deadbeaf"),
			Value:       types.MustNewBytes("deadbeaf"),
			Level:       9,
			Timestamp:   ts,
			ProtocolID:  2,
			OperationID: 10,
		}, nil).
		Times(1)

	rb.EXPECT().
		LastDiff(gomock.Any(), ptr, "key_hash_2", false).
		Return(bigmapdiff.BigMapDiff{}, sql.ErrNoRows).
		Times(1)

	rb.EXPECT().
		DeleteBigMapState(gomock.Any(), bigmapdiff.BigMapState{
			ID:              2,
			Ptr:             ptr,
			LastUpdateLevel: 11,
			Count:           10,
			LastUpdateTime:  ts,
			KeyHash:         "key_hash_2",
			Contract:        "address_1",
			Key:             types.MustNewBytes("deadbeaf0011"),
			Value:           types.MustNewBytes("001122334455"),
			Removed:         false,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		SaveBigMapState(gomock.Any(), bigmapdiff.BigMapState{
			ID:              1,
			Ptr:             10,
			LastUpdateLevel: 9,
			Count:           10,
			LastUpdateTime:  ts,
			KeyHash:         "key_hash",
			Contract:        "address_1",
			Key:             types.MustNewBytes("deadbeaf"),
			Value:           types.MustNewBytes("deadbeaf"),
			Removed:         false,
			IsRollback:      true,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		GlobalConstants(gomock.Any(), level).
		Return([]contract.GlobalConstant{
			{
				ID:        1,
				Timestamp: ts,
				Level:     11,
				Address:   "address_1",
				Value:     testsuite.MustHexDecode("deadbeaf"),
			},
		}, nil).
		Times(1)

	rb.EXPECT().
		Scripts(gomock.Any(), level).
		Return([]contract.Script{}, nil).
		Times(1)

	rb.EXPECT().
		DeleteScriptsConstants(gomock.Any(), []int64{}, []int64{1}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		DeleteAll(gomock.Any(), nil, level).
		Return(0, nil).
		Times(8)

	rb.EXPECT().
		Protocols(gomock.Any(), level).
		Return(nil).
		Times(1)

	rb.EXPECT().
		UpdateStats(gomock.Any(), stats.Stats{
			ID:                   1,
			ContractsCount:       100,
			OperationsCount:      95,
			EventsCount:          9,
			TransactionsCount:    66,
			OriginationsCount:    9,
			SrOriginationsCount:  10,
			TransferTicketsCount: 10,
		}).
		Return(nil).
		Times(1)

	rb.EXPECT().
		Commit().
		Return(nil).
		Times(1)

	t.Run("Rollback", func(t *testing.T) {
		state := block.Block{
			Level: 11,
		}
		err := NewManager(storage, blockRepo, rb, statsRepo).
			Rollback(
				t.Context(),
				types.Mainnet,
				state,
				10,
			)
		require.NoError(t, err)
	})
}
