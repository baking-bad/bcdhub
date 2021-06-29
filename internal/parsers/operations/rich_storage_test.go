package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_bm "github.com/baking-bad/bcdhub/internal/models/mock/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRichStorage_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlBmRepo := gomock.NewController(t)
	defer ctrlBmRepo.Finish()
	bmRepo := mock_bm.NewMockRepository(ctrlBmRepo)
	bmStateRepo := mock_bm.NewMockStateRepository(ctrlBmRepo)

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	protocols := map[int64]string{
		2: "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		3: "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
	}

	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Carthagenet), gomock.Eq(int64(1055)), gomock.Any()).
		Return(&bigmap.BigMap{
			Network:  types.Carthagenet,
			Contract: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
			Ptr:      1055,
		}, nil).
		AnyTimes()

	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Carthagenet), gomock.Eq(int64(1509)), gomock.Any()).
		Return(&bigmap.BigMap{
			Network:  types.Carthagenet,
			Contract: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
			Ptr:      1509,
		}, nil).
		AnyTimes()

	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Edo2net), gomock.Eq(int64(17)), gomock.Any()).
		Return(&bigmap.BigMap{
			Network:  types.Edo2net,
			Contract: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
			Ptr:      17,
		}, nil).
		AnyTimes()

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
				Level:       1151463,
				Destination: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
				Network:     types.Mainnet,
				Timestamp:   timestamp,
				ProtocolID:  2,
				Kind:        types.OperationKindTransaction,
			},
			filename: "./data/rich_storage/test1.json",

			wantOperation: operation.Operation{
				Level:       1151463,
				Destination: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
				Network:     types.Mainnet,
				Timestamp:   timestamp,
				ProtocolID:  2,
				Kind:        types.OperationKindTransaction,
				BigMapDiffs: []*bigmap.Diff{
					{
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:    "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key:        []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						Value:      []byte(`{"bytes":"050098e1e8d78a02"}`),
						Level:      1151463,
						Timestamp:  timestamp,
						ProtocolID: 2,
					}, {
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:    "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key:        []byte(`{"bytes":"05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3"}`),
						Value:      []byte(`{"bytes":"05070700bdf4160200000000"}`),
						Level:      1151463,
						Timestamp:  timestamp,
						ProtocolID: 2,
					}, {
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:    "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key:        []byte(`{"bytes":"05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200"}`),
						Value:      []byte(`{"bytes":"0507070084a99c750200000000"}`),
						Level:      1151463,
						Timestamp:  timestamp,
						ProtocolID: 2,
					},
				},
				BigMapActions: []*bigmap.Action{},
			},
			want: &parsers.Result{
				BigMapState: []*bigmap.State{
					{
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:         "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Key:             []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						Value:           []byte(`{"bytes":"050098e1e8d78a02"}`),
						LastUpdateLevel: 1151463,
						LastUpdateTime:  timestamp,
					}, {
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:         "exprtzVE8dHF7nePZxF6PSRf3yhfecTEKavyCZpndJGN2hz6PzQkFi",
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a00000016000093e93e23e5d157a80852297eccc7a42d7080ddd3"}`),
						Value:           []byte(`{"bytes":"05070700bdf4160200000000"}`),
						LastUpdateLevel: 1151463,
						LastUpdateTime:  timestamp,
					}, {
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						KeyHash:         "expruyvqmgBYpF54i1c4p6r3oVV7FmW7ZH8EyjSjahKoQEfWPmcjGg",
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a000000160139c8ade2617663981fa2b87592c9ad92714d14c200"}`),
						Value:           []byte(`{"bytes":"0507070084a99c750200000000"}`),
						LastUpdateLevel: 1151463,
						LastUpdateTime:  timestamp,
					},
				},
			},
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Level:       359942,
				Destination: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
				Network:     types.Carthagenet,
				Timestamp:   timestamp,
				ProtocolID:  2,
				Kind:        types.OperationKindOrigination,
			},
			sourcePtr: 1055,
			filename:  "./data/rich_storage/test2.json",
			want: &parsers.Result{
				BigMapState: []*bigmap.State{},
				BigMaps: []*bigmap.BigMap{
					{
						CreatedLevel: 359942,
						CreatedAt:    timestamp,
						Network:      types.Carthagenet,
						Contract:     "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
						Ptr:          1509,
					},
				},
			},
			wantOperation: operation.Operation{
				Level:       359942,
				Destination: "KT1Xk1XJD2M8GYFUXRN12oMvDAysECDWwGdS",
				Network:     types.Carthagenet,
				Timestamp:   timestamp,
				ProtocolID:  2,
				Kind:        types.OperationKindOrigination,
				BigMapDiffs: []*bigmap.Diff{},
				BigMapActions: []*bigmap.Action{
					{
						Action:    types.BigMapActionCopy,
						Level:     359942,
						Timestamp: timestamp,
					},
				},
			},
		}, {
			name: "test 3",
			operation: &operation.Operation{
				Level:       220,
				Destination: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
				Network:     types.Edo2net,
				Timestamp:   timestamp,
				ProtocolID:  3,
				Kind:        types.OperationKindOrigination,
			},
			sourcePtr: 17,
			filename:  "./data/rich_storage/test3.json",
			want: &parsers.Result{
				BigMapState: []*bigmap.State{
					{
						BigMap: bigmap.BigMap{
							Network:  types.Edo2net,
							Contract: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
							Ptr:      17,
						},
						KeyHash:         "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Key:             []byte(`{"string":""}`),
						Value:           []byte(`{"bytes":"68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f747a69702d31362f656d6f6a692d696e2d6d657461646174612e6a736f6e"}`),
						LastUpdateLevel: 220,
						LastUpdateTime:  timestamp,
					},
				},
			},
			wantOperation: operation.Operation{
				Level:       220,
				Destination: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
				Network:     types.Edo2net,
				Timestamp:   timestamp,
				ProtocolID:  3,
				Kind:        types.OperationKindOrigination,
				BigMapDiffs: []*bigmap.Diff{
					{
						BigMap: bigmap.BigMap{
							Network:  types.Edo2net,
							Contract: "KT1C2Nh1VUjUt64JY44rx8bQPpjy3eSYoAu2",
							Ptr:      17,
						},
						KeyHash:    "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo",
						Key:        []byte(`{"string":""}`),
						Value:      []byte(`{"bytes":"68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f747a69702d31362f656d6f6a692d696e2d6d657461646174612e6a736f6e"}`),
						Level:      220,
						Timestamp:  timestamp,
						ProtocolID: 3,
					},
				},
				BigMapActions: []*bigmap.Action{
					{
						Action:    types.BigMapActionAlloc,
						Level:     220,
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

			bmStateRepo.
				EXPECT().
				GetByPtr(tt.operation.Network, tt.operation.Destination, tt.sourcePtr).
				Return([]bigmap.State{}, nil).
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
			script, err := fetch.Contract(tt.operation.Network, tt.operation.Destination, proto, "./test")
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			tt.operation.Script = script

			tt.operation.AST, err = ast.NewScriptWithoutCode(script)
			if err != nil {
				t.Errorf("NewScriptWithoutCode() error = %v", err)
				return
			}

			parser, err := NewRichStorage(bmRepo, bmStateRepo, generalRepo, rpc, proto)
			if err != nil {
				t.Errorf(`NewRichStorage = error %v`, err)
				return
			}

			got, err := parser.Parse(op, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("RichStorage.Parse() error = %v, wantErr %v", err, tt.wantErr)
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
