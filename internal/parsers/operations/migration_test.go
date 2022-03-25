package operations

import (
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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
				Hash:      "hash",
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
				Hash:      "hash",
			},
			fileName: "./data/migration/test2.json",
			want: &migration.Migration{
				Level:      123,
				ProtocolID: 2,
				Timestamp:  timestamp,
				Hash:       "hash",
				Kind:       types.MigrationKindLambda,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var op noderpc.Operation
			if err := readJSONFile(tt.fileName, &op); err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.fileName, err)
				return
			}

			contractRepo.
				EXPECT().
				Get(gomock.Eq(tt.operation.Destination.Address)).
				Return(contract.Contract{}, nil).
				AnyTimes()

			store := parsers.NewTestStore()
			if err := NewMigration(contractRepo).Parse(op, tt.operation, store); err != nil {
				t.Errorf("Migration.Parse() = %s", err)
				return
			}
			if tt.want != nil {
				tt.want.ID = store.Migrations[0].ID
				assert.Equal(t, tt.want, store.Migrations[0])
			} else {
				assert.Len(t, store.Migrations, 0)
			}
		})
	}
}
