package operations

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRichStorage_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlES := gomock.NewController(t)
	defer ctrlES.Finish()
	es := elastic.NewMockIElastic(ctrlES)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	tests := []struct {
		name      string
		operation *models.Operation
		filename  string
		sourcePtr int64
		want      storage.RichStorage
		wantErr   bool
	}{
		{
			name: "test 1",
			operation: &models.Operation{
				ID:          "operation_id",
				Level:       1151463,
				Destination: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
				Network:     "mainnet",
				Timestamp:   timestamp,
				Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				Kind:        "transaction",
			},
			filename: "./data/rich_storage/test1.json",
			want: storage.RichStorage{
				Models: []elastic.Model{
					&models.BigMapDiff{
						Ptr:     31,
						KeyHash: "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key: map[string]interface{}{
							"bytes": "05010000000b746f74616c537570706c79",
						},
						Value:       `{"bytes": "050098e1e8d78a02"}`,
						BinPath:     "0/0",
						OperationID: "operation_id",
						Level:       1151463,
						Address:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:     "mainnet",
						Timestamp:   timestamp,
						Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					}, &models.BigMapDiff{
						Ptr:     31,
						KeyHash: "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key: map[string]interface{}{
							"bytes": "05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3",
						},
						Value:       `{"bytes": "05070700bdf4160200000000"}`,
						BinPath:     "0/0",
						OperationID: "operation_id",
						Level:       1151463,
						Address:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:     "mainnet",
						Timestamp:   timestamp,
						Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					}, &models.BigMapDiff{
						Ptr:     31,
						KeyHash: "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key: map[string]interface{}{
							"bytes": "05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200",
						},
						Value:       `{"bytes": "0507070084a99c750200000000"}`,
						BinPath:     "0/0",
						OperationID: "operation_id",
						Level:       1151463,
						Address:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:     "mainnet",
						Timestamp:   timestamp,
						Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					},
				},
			},
		}, {
			name: "test 2",
			operation: &models.Operation{
				ID:          "operation_id",
				Level:       359942,
				Destination: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
				Network:     "carthagenet",
				Timestamp:   timestamp,
				Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				Kind:        "origination",
			},
			sourcePtr: 1055,
			filename:  "./data/rich_storage/test2.json",
			want: storage.RichStorage{
				Models: []elastic.Model{
					&models.BigMapAction{
						Action:         "copy",
						SourcePtr:      setInt64(1055),
						DestinationPtr: setInt64(1509),
						OperationID:    "operation_id",
						Level:          359942,
						Address:        "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
						Network:        "carthagenet",
						Timestamp:      timestamp,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", tt.operation.Destination, tt.operation.Level)
			storageJSON, err := readJSONFile(storageFile)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, storageFile, err)
				return
			}

			tt.want.DeffatedStorage = storageJSON.String()

			rpc.
				EXPECT().
				GetScriptStorageJSON(tt.operation.Destination, tt.operation.Level).
				Return(storageJSON, nil).
				AnyTimes()

			es.
				EXPECT().
				GetBigMapDiffsByPtr(tt.operation.Destination, tt.operation.Network, tt.sourcePtr).
				Return([]models.BigMapDiff{}, nil).
				AnyTimes()

			metadata, err := readTestMetadata(tt.operation.Destination)
			if err != nil {
				t.Errorf(`readTestMetadata("%s") = error %v`, tt.operation.Destination, err)
				return
			}
			data, err := readJSONFile(tt.filename)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}

			got, err := NewRichStorage(es, rpc, tt.operation, metadata).Parse(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RichStorage.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// log.Print(cmp.Diff(got, tt.want))

			if !compareRichStorage(t, got, tt.want) {
				t.Errorf("RichStorage.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareRichStorage(t *testing.T, one, two storage.RichStorage) bool {
	if one.Empty != two.Empty {
		return false
	}
	if !assert.JSONEq(t, one.DeffatedStorage, two.DeffatedStorage) {
		return false
	}
	if len(one.Models) != len(two.Models) {
		return false
	}

	for i := range one.Models {
		switch oneItem := one.Models[i].(type) {
		case *models.BigMapDiff:
			twoItem, ok := two.Models[i].(*models.BigMapDiff)
			if !ok {
				return false
			}
			if !compareBigMapDiff(t, oneItem, twoItem) {
				return false
			}
		case *models.BigMapAction:
			twoItem, ok := two.Models[i].(*models.BigMapAction)
			if !ok {
				return false
			}
			if !compareBigMapAction(oneItem, twoItem) {
				return false
			}
		}
	}

	return true
}

func compareBigMapDiff(t *testing.T, one, two *models.BigMapDiff) bool {
	if one.Address != two.Address {
		log.Printf("Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.KeyHash != two.KeyHash {
		log.Printf("KeyHash: %s != %s", one.KeyHash, two.KeyHash)
		return false
	}
	if !assert.JSONEq(t, one.Value, two.Value) {
		log.Printf("Value: %s != %s", one.Value, two.Value)
		return false
	}
	if one.BinPath != two.BinPath {
		log.Printf("BinPath: %s != %s", one.BinPath, two.BinPath)
		return false
	}
	if one.OperationID != two.OperationID {
		log.Printf("OperationID: %s != %s", one.OperationID, two.OperationID)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	if one.Protocol != two.Protocol {
		log.Printf("Protocol: %s != %s", one.Protocol, two.Protocol)
		return false
	}
	if !reflect.DeepEqual(one.Key, two.Key) {
		log.Printf("Key: %s != %s", one.Key, two.Key)
		return false
	}
	return true
}

func compareBigMapAction(one, two *models.BigMapAction) bool {
	if one.Action != two.Action {
		log.Printf("Action: %s != %s", one.Action, two.Action)
		return false
	}
	if one.SourcePtr != nil && two.SourcePtr != nil && *one.SourcePtr != *two.SourcePtr {
		log.Printf("SourcePtr: %d != %d", one.SourcePtr, two.SourcePtr)
		return false
	}
	if one.DestinationPtr != nil && two.DestinationPtr != nil && *one.DestinationPtr != *two.DestinationPtr {
		log.Printf("DestinationPtr: %d != %d", one.DestinationPtr, two.DestinationPtr)
		return false
	}
	if one.OperationID != two.OperationID {
		log.Printf("OperationID: %s != %s", one.OperationID, two.OperationID)
		return false
	}
	if one.Level != two.Level {
		log.Printf("Level: %d != %d", one.Level, two.Level)
		return false
	}
	if one.Address != two.Address {
		log.Printf("Address: %s != %s", one.Address, two.Address)
		return false
	}
	if one.Network != two.Network {
		log.Printf("Network: %s != %s", one.Network, two.Network)
		return false
	}
	if one.Timestamp != two.Timestamp {
		log.Printf("Timestamp: %s != %s", one.Timestamp, two.Timestamp)
		return false
	}
	return true
}

func setInt64(x int64) *int64 {
	return &x
}
