package operations

import (
	"context"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	mock_contract "github.com/baking-bad/bcdhub/internal/models/mock/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMigration_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlContractRepo := gomock.NewController(t)
	defer ctrlContractRepo.Finish()
	contractRepo := mock_contract.NewMockRepository(ctrlContractRepo)

	tests := []struct {
		name      string
		operation *operation.Operation
		fileName  string
		want      *migration.Migration
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
				Level:      123,
				ProtocolID: 2,
				Destination: account.Account{
					Address: "destination",
				},
				Timestamp: timestamp,
				Hash:      []byte("hash"),
			},
			fileName: "./data/migration/test1.json",
			want:     nil,
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Level:      123,
				ProtocolID: 2,
				Destination: account.Account{
					Address: "destination",
				},
				Timestamp: timestamp,
				Hash:      []byte("hash"),
			},
			fileName: "./data/migration/test2.json",
			want: &migration.Migration{
				Level:      123,
				ProtocolID: 2,
				Timestamp:  timestamp,
				Hash:       []byte("hash"),
				Kind:       types.MigrationKindLambda,
				Contract:   contract.Contract{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var op noderpc.Operation
			err := readJSONFile(tt.fileName, &op)
			require.NoError(t, err)

			contractRepo.
				EXPECT().
				Get(gomock.Any(), tt.operation.Destination.Address).
				Return(contract.Contract{}, nil).
				AnyTimes()

			store := parsers.NewTestStore()
			err = NewMigration(contractRepo).Parse(context.Background(), op, tt.operation, "PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq", store)
			require.NoError(t, err)

			if tt.want != nil {
				tt.want.ID = store.Migrations[0].ID
				require.Equal(t, tt.want, store.Migrations[0])
			} else {
				require.Len(t, store.Migrations, 0)
			}
		})
	}
}
