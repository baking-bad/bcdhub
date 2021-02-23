package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/stretchr/testify/assert"
)

func TestMigration_Parse(t *testing.T) {
	timestamp := time.Now()

	tests := []struct {
		name      string
		operation *operation.Operation
		fileName  string
		want      *migration.Migration
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
				Network:     "mainnet",
				Level:       123,
				Protocol:    "protocol",
				Destination: "destination",
				Timestamp:   timestamp,
				Hash:        "hash",
			},
			fileName: "./data/migration/test1.json",
			want:     nil,
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Network:     "mainnet",
				Level:       123,
				Protocol:    "protocol",
				Destination: "destination",
				Timestamp:   timestamp,
				Hash:        "hash",
			},
			fileName: "./data/migration/test2.json",
			want: &migration.Migration{
				Network:   "mainnet",
				Level:     123,
				Protocol:  "protocol",
				Address:   "destination",
				Timestamp: timestamp,
				Hash:      "hash",
				Kind:      consts.MigrationLambda,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := readJSONFile(tt.fileName)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.fileName, err)
				return
			}
			got, err := NewMigration().Parse(data, tt.operation)
			if err != nil {
				t.Errorf("Migration.Parse() = %s", err)
				return
			}
			if tt.want != nil {
				tt.want.ID = got.ID
				tt.want.IndexedTime = got.IndexedTime
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
