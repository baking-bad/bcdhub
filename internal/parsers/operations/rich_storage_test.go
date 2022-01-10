package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
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

	protocols := map[int64]string{
		2: "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		3: "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
	}

	tests := []struct {
		name          string
		operation     *operation.Operation
		filename      string
		sourcePtr     int64
		want          *parsers.Result
		wantErr       bool
		wantOperation operation.Operation
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
				Level: 1151463,
				Destination: account.Account{
					Network: types.Mainnet,
					Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Type:    types.AccountTypeContract,
				},
				Network:    types.Mainnet,
				Timestamp:  timestamp,
				ProtocolID: 2,
				Kind:       types.OperationKindTransaction,
			},
			filename: "./data/rich_storage/test1.json",

			wantOperation: operation.Operation{
				Level: 1151463,
				Destination: account.Account{
					Network: types.Mainnet,
					Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Type:    types.AccountTypeContract,
				},
				Network:    types.Mainnet,
				Timestamp:  timestamp,
				ProtocolID: 2,
				Kind:       types.OperationKindTransaction,
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:        31,
						KeyHash:    "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key:        []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						Value:      []byte(`{"bytes":"050098e1e8d78a02"}`),
						Level:      1151463,
						Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:    types.Mainnet,
						Timestamp:  timestamp,
						ProtocolID: 2,
					}, {
						Ptr:        31,
						KeyHash:    "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key:        []byte(`{"bytes":"05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3"}`),
						Value:      []byte(`{"bytes":"05070700bdf4160200000000"}`),
						Level:      1151463,
						Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:    types.Mainnet,
						Timestamp:  timestamp,
						ProtocolID: 2,
					}, {
						Ptr:        31,
						KeyHash:    "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key:        []byte(`{"bytes":"05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200"}`),
						Value:      []byte(`{"bytes":"0507070084a99c750200000000"}`),
						Level:      1151463,
						Contract:   "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:    types.Mainnet,
						Timestamp:  timestamp,
						ProtocolID: 2,
					},
				},
			},
			want: &parsers.Result{
				BigMapState: []*bigmapdiff.BigMapState{
					{
						Ptr:             31,
						KeyHash:         "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key:             []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						Value:           []byte(`{"bytes":"050098e1e8d78a02"}`),
						LastUpdateLevel: 1151463,
						Contract:        "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:         types.Mainnet,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:             31,
						KeyHash:         "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3"}`),
						Value:           []byte(`{"bytes":"05070700bdf4160200000000"}`),
						LastUpdateLevel: 1151463,
						Contract:        "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:         types.Mainnet,
						LastUpdateTime:  timestamp,
					}, {
						Ptr:             31,
						KeyHash:         "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200"}`),
						Value:           []byte(`{"bytes":"0507070084a99c750200000000"}`),
						LastUpdateLevel: 1151463,
						Contract:        "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Network:         types.Mainnet,
						LastUpdateTime:  timestamp,
					},
				},
			},
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Level: 359942,
				Destination: account.Account{
					Network: types.Carthagenet,
					Address: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
					Type:    types.AccountTypeContract,
				},
				Network:    types.Carthagenet,
				Timestamp:  timestamp,
				ProtocolID: 2,
				Kind:       types.OperationKindOrigination,
			},
			sourcePtr: 1055,
			filename:  "./data/rich_storage/test2.json",
			want: &parsers.Result{
				BigMapState: []*bigmapdiff.BigMapState{},
			},
			wantOperation: operation.Operation{
				Level: 359942,
				Destination: account.Account{
					Network: types.Carthagenet,
					Address: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
					Type:    types.AccountTypeContract,
				},
				Network:     types.Carthagenet,
				Timestamp:   timestamp,
				ProtocolID:  2,
				Kind:        types.OperationKindOrigination,
				BigMapDiffs: []*bigmapdiff.BigMapDiff{},
				BigMapActions: []*bigmapaction.BigMapAction{
					{
						Action:         types.BigMapActionCopy,
						SourcePtr:      setInt64(1055),
						DestinationPtr: setInt64(1509),
						Level:          359942,
						Address:        "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
						Network:        types.Carthagenet,
						Timestamp:      timestamp,
					},
				},
			},
		}, {
			name: "test 3",
			operation: &operation.Operation{
				Level: 220,
				Destination: account.Account{
					Network: types.Edo2net,
					Address: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
					Type:    types.AccountTypeContract,
				},
				Network:    types.Edo2net,
				Timestamp:  timestamp,
				ProtocolID: 3,
				Kind:       types.OperationKindOrigination,
			},
			sourcePtr: 17,
			filename:  "./data/rich_storage/test3.json",
			want: &parsers.Result{
				BigMapState: []*bigmapdiff.BigMapState{
					{
						Ptr:             17,
						KeyHash:         "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Key:             []byte(`{"string":""}`),
						Value:           []byte(`{"bytes":"68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f747a69702d31362f656d6f6a692d696e2d6d657461646174612e6a736f6e"}`),
						LastUpdateLevel: 220,
						Contract:        "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
						Network:         types.Edo2net,
						LastUpdateTime:  timestamp,
					},
				},
			},
			wantOperation: operation.Operation{
				Level: 220,
				Destination: account.Account{
					Network: types.Edo2net,
					Address: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
					Type:    types.AccountTypeContract,
				},
				Network:    types.Edo2net,
				Timestamp:  timestamp,
				ProtocolID: 3,
				Kind:       types.OperationKindOrigination,
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:        17,
						KeyHash:    "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Key:        []byte(`{"string":""}`),
						Value:      []byte(`{"bytes":"68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f747a69702d31362f656d6f6a692d696e2d6d657461646174612e6a736f6e"}`),
						Level:      220,
						Contract:   "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
						Network:    types.Edo2net,
						Timestamp:  timestamp,
						ProtocolID: 3,
					},
				},
				BigMapActions: []*bigmapaction.BigMapAction{
					{
						Action:    types.BigMapActionAlloc,
						SourcePtr: setInt64(17),
						Level:     220,
						Address:   "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
						Network:   types.Edo2net,
						Timestamp: timestamp,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc.
				EXPECT().
				GetScriptStorageRaw(gomock.Any(), gomock.Any()).
				DoAndReturn(readStorage).
				AnyTimes()

			bmdRepo.
				EXPECT().
				GetByPtr(tt.operation.Network, tt.operation.Destination.Address, tt.sourcePtr).
				Return([]bigmapdiff.BigMapState{}, nil).
				AnyTimes()

			var op noderpc.Operation
			if err := readJSONFile(tt.filename, &op); err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			proto, ok := protocols[tt.operation.ProtocolID]
			if !ok {
				t.Errorf(`unknown protocol ID: %d`, tt.operation.ProtocolID)
				return
			}

			symLink, err := bcd.GetProtoSymLink(proto)
			if err != nil {
				t.Error(err)
				return
			}

			script, err := readTestScript(tt.operation.Network, tt.operation.Destination.Address, symLink)
			if err != nil {
				t.Errorf(`readTestScript= error %v`, err)
				return
			}
			tt.operation.Script = script

			tt.operation.AST, err = ast.NewScriptWithoutCode(script)
			if err != nil {
				t.Errorf("NewScriptWithoutCode() error = %v", err)
				return
			}

			parser, err := NewRichStorage(bmdRepo, rpc, proto)
			if err != nil {
				t.Errorf(`NewRichStorage = error %v`, err)
				return
			}

			got, err := parser.Parse(op, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("RichStorage.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if compareOperations(t, tt.operation, &tt.wantOperation) {
				return
			}
			compareRichStorage(t, got, tt.want)
		})
	}
}

func compareRichStorage(t *testing.T, expected, got *parsers.Result) {
	assert.Len(t, got.BigMapState, len(expected.BigMapState))

	for i := range expected.BigMapState {
		expected.BigMapState[i].ID = got.BigMapState[i].GetID()
	}

	assert.Equal(t, expected.BigMapState, got.BigMapState)
}

func setInt64(x int64) *int64 {
	return &x
}
