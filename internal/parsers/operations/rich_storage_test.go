package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRichStorage_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlBmdRepo := gomock.NewController(t)
	defer ctrlBmdRepo.Finish()
	bmdRepo := mock_bmd.NewMockRepository(ctrlBmdRepo)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	tests := []struct {
		name      string
		operation *operation.Operation
		filename  string
		sourcePtr int64
		want      storage.RichStorage
		wantErr   bool
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
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
				Models: []models.Model{
					&bigmapdiff.BigMapDiff{
						Ptr:         31,
						KeyHash:     "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key:         []byte(`{"bytes": "05010000000b746f74616c537570706c79"}`),
						Value:       []byte(`{"bytes": "050098e1e8d78a02"}`),
						OperationID: "operation_id",
						Level:       1151463,
						Address:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:     "mainnet",
						Timestamp:   timestamp,
						Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					}, &bigmapdiff.BigMapDiff{
						Ptr:         31,
						KeyHash:     "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key:         []byte(`{"bytes": "05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3"}`),
						Value:       []byte(`{"bytes": "05070700bdf4160200000000"}`),
						OperationID: "operation_id",
						Level:       1151463,
						Address:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:     "mainnet",
						Timestamp:   timestamp,
						Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					}, &bigmapdiff.BigMapDiff{
						Ptr:         31,
						KeyHash:     "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key:         []byte(`{"bytes": "05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200"}`),
						Value:       []byte(`{"bytes": "0507070084a99c750200000000"}`),
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
			operation: &operation.Operation{
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
				Models: []models.Model{
					&bigmapaction.BigMapAction{
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
		}, {
			name: "test 3",
			operation: &operation.Operation{
				ID:          "operation_id",
				Level:       220,
				Destination: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
				Network:     "edo2net",
				Timestamp:   timestamp,
				Protocol:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
				Kind:        "origination",
			},
			sourcePtr: 17,
			filename:  "./data/rich_storage/test3.json",
			want: storage.RichStorage{
				Models: []models.Model{
					&bigmapaction.BigMapAction{
						Action:      "alloc",
						SourcePtr:   setInt64(17),
						OperationID: "operation_id",
						Level:       220,
						Address:     "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
						Network:     "edo2net",
						Timestamp:   timestamp,
					},
					&bigmapdiff.BigMapDiff{
						Ptr:         17,
						KeyHash:     "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Key:         []byte(`{"string": ""}`),
						Value:       []byte(`{"bytes": "68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f747a69702d31362f656d6f6a692d696e2d6d657461646174612e6a736f6e"}`),
						OperationID: "operation_id",
						Level:       220,
						Address:     "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
						Network:     "edo2net",
						Timestamp:   timestamp,
						Protocol:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageJSON, err := readStorage(tt.operation.Destination, tt.operation.Level)
			if err != nil {
				t.Errorf(`readStorage("%s", %d) = error %v`, tt.operation.Destination, tt.operation.Level, err)
				return
			}
			tt.want.DeffatedStorage = string(storageJSON)

			rpc.
				EXPECT().
				GetScriptStorageRaw(tt.operation.Destination, tt.operation.Level).
				DoAndReturn(readStorage).
				AnyTimes()

			bmdRepo.
				EXPECT().
				GetByPtr(tt.operation.Destination, tt.operation.Network, tt.sourcePtr).
				Return([]bigmapdiff.BigMapDiff{}, nil).
				AnyTimes()

			var op noderpc.Operation
			if err := readJSONFile(tt.filename, &op); err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			script, err := fetch.Contract(tt.operation.Destination, tt.operation.Network, tt.operation.Protocol, "./test")
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			tt.operation.Script = script

			parser, err := NewRichStorage(bmdRepo, rpc, tt.operation.Protocol)
			if err != nil {
				t.Errorf(`NewRichStorage = error %v`, err)
				return
			}

			got, err := parser.Parse(op, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("RichStorage.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

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

	return compareParserResponse(t, one.Models, two.Models)
}

func setInt64(x int64) *int64 {
	return &x
}
