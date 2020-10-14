package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/models"
)

func TestMigration_Parse(t *testing.T) {
	timestamp := time.Now()

	tests := []struct {
		name      string
		operation *models.Operation
		fileName  string
		want      *models.Migration
	}{
		{
			name: "test 1",
			operation: &models.Operation{
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
			operation: &models.Operation{
				Network:     "mainnet",
				Level:       123,
				Protocol:    "protocol",
				Destination: "destination",
				Timestamp:   timestamp,
				Hash:        "hash",
			},
			fileName: "./data/migration/test2.json",
			want: &models.Migration{
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
			got := NewMigration(tt.operation).Parse(data)
			if got == nil {
				if got != tt.want {
					t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				}
				return
			}
			if got.Network != tt.want.Network {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Level != tt.want.Level {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Protocol != tt.want.Protocol {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Address != tt.want.Address {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Timestamp != tt.want.Timestamp {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Hash != tt.want.Hash {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
			if got.Kind != tt.want.Kind {
				t.Errorf("Migration.Parse() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
