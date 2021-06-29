package operations

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	modelContract "github.com/baking-bad/bcdhub/internal/models/contract"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_bm "github.com/baking-bad/bcdhub/internal/models/mock/bigmap"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_contract "github.com/baking-bad/bcdhub/internal/models/mock/contract"
	mock_proto "github.com/baking-bad/bcdhub/internal/models/mock/protocol"
	mock_token_balance "github.com/baking-bad/bcdhub/internal/models/mock/tokenbalance"
	mock_tzip "github.com/baking-bad/bcdhub/internal/models/mock/tzip"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlBmRepo := gomock.NewController(t)
	defer ctrlBmRepo.Finish()
	bmRepo := mock_bm.NewMockRepository(ctrlBmRepo)
	bmStateRepo := mock_bm.NewMockStateRepository(ctrlBmRepo)
	bmdRepo := mock_bm.NewMockDiffRepository(ctrlBmRepo)

	ctrlBlockRepo := gomock.NewController(t)
	defer ctrlBlockRepo.Finish()
	blockRepo := mock_block.NewMockRepository(ctrlBlockRepo)

	ctrlProtoRepo := gomock.NewController(t)
	defer ctrlProtoRepo.Finish()
	protoRepo := mock_proto.NewMockRepository(ctrlProtoRepo)

	ctrlTzipRepo := gomock.NewController(t)
	defer ctrlTzipRepo.Finish()
	tzipRepo := mock_tzip.NewMockRepository(ctrlTzipRepo)

	ctrlContractRepo := gomock.NewController(t)
	defer ctrlContractRepo.Finish()
	contractRepo := mock_contract.NewMockRepository(ctrlContractRepo)

	ctrlTokenBalanceRepo := gomock.NewController(t)
	defer ctrlTokenBalanceRepo.Finish()
	tbRepo := mock_token_balance.NewMockRepository(ctrlTokenBalanceRepo)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	ctrlScriptSaver := gomock.NewController(t)
	defer ctrlScriptSaver.Finish()
	scriptSaver := contract.NewMockScriptSaver(ctrlScriptSaver)

	scriptSaver.
		EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	tzipRepo.
		EXPECT().
		GetWithEvents(gomock.Any()).
		Return(make([]tzip.TZIP, 0), nil).
		AnyTimes()

	tzipRepo.
		EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	contractRepo.
		EXPECT().
		Get(gomock.Any(), gomock.Any()).
		DoAndReturn(readTestContractModel).
		AnyTimes()

	contractRepo.
		EXPECT().
		GetProjectIDByHash(gomock.Any()).
		Return("", nil).
		AnyTimes()

	generalRepo.
		EXPECT().
		Save(gomock.AssignableToTypeOf([]models.Model{})).
		Return(nil).
		AnyTimes()

	generalRepo.
		EXPECT().
		IsRecordNotFound(gomock.Any()).
		Return(true).
		AnyTimes()

	bmStateRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq(types.Carthagenet),
			gomock.Eq("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU"),
			gomock.Eq(int64(2416))).
		Return([]bigmap.State{
			{
				BigMap: bigmap.BigMap{
					Network:  types.Carthagenet,
					Contract: "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Ptr:      2416,
				},
				Key:             []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:         "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:           []byte(`{"prim":"Pair","args":[[],{"int":"6000"}]}`),
				LastUpdateLevel: 386026,
				LastUpdateTime:  timestamp,
			},
		}, nil).
		AnyTimes()

	var bmID int64
	for _, ptr := range []int{25167, 25166, 25165, 25164} {
		bmID++
		bmRepo.
			EXPECT().
			Get(gomock.Eq(types.Edo2net), gomock.Eq(int64(ptr)), gomock.Any()).
			Return(&bigmap.BigMap{
				ID:       bmID,
				Ptr:      int64(ptr),
				Network:  types.Edo2net,
				Contract: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
			}, nil).
			AnyTimes()

		bmStateRepo.
			EXPECT().
			GetByPtr(
				gomock.Eq(types.Edo2net),
				gomock.Eq("KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264"),
				gomock.Eq(int64(ptr))).
			Return([]bigmap.State{}, nil).
			AnyTimes()
	}

	for _, ptr := range []int{25171, 25170, 25169, 25168} {
		bmID++
		bmRepo.
			EXPECT().
			Get(gomock.Eq(types.Edo2net), gomock.Eq(int64(ptr)), gomock.Any()).
			Return(nil, gorm.ErrRecordNotFound).
			AnyTimes()
	}

	bmID++
	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Mainnet), gomock.Eq(int64(63)), gomock.Any()).
		Return(&bigmap.BigMap{
			ID:       bmID,
			Ptr:      63,
			Network:  types.Mainnet,
			Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
		}, nil).
		AnyTimes()

	bmID++
	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Mainnet), gomock.Eq(int64(32)), gomock.Any()).
		Return(&bigmap.BigMap{
			ID:       bmID,
			Ptr:      32,
			Network:  types.Mainnet,
			Contract: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
		}, nil).
		AnyTimes()

	bmID++
	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Mainnet), gomock.Eq(int64(31)), gomock.Any()).
		Return(&bigmap.BigMap{
			ID:       bmID,
			Ptr:      31,
			Network:  types.Mainnet,
			Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
		}, nil).
		AnyTimes()

	bmID++
	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Mainnet), gomock.Eq(int64(746)), gomock.Any()).
		Return(&bigmap.BigMap{
			ID:       bmID,
			Ptr:      746,
			Network:  types.Mainnet,
			Contract: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
		}, nil).
		AnyTimes()

	bmID++
	bmRepo.
		EXPECT().
		Get(gomock.Eq(types.Mainnet), gomock.Eq(int64(1264)), gomock.Any()).
		Return(&bigmap.BigMap{
			ID:       bmID,
			Ptr:      1264,
			Network:  types.Mainnet,
			Contract: "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
		}, nil).
		AnyTimes()

	bmStateRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq(types.Carthagenet),
			gomock.Eq("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ"),
			gomock.Eq(int64(2417))).
		Return([]bigmap.State{
			{
				BigMap: bigmap.BigMap{
					Network:  types.Carthagenet,
					Contract: "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Ptr:      2417,
				},
				Key:             []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:         "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:           nil,
				LastUpdateLevel: 386026,
				LastUpdateTime:  timestamp,
			},
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Delphinet),
			gomock.Eq("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"),
			gomock.Eq(int64(-1))).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			Network: types.Delphinet,
			SymLink: bcd.SymLinkBabylon,
			ID:      0,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Mainnet),
			gomock.Eq("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"),
			gomock.Eq(int64(-1))).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      1,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Mainnet),
			gomock.Eq("PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP"),
			gomock.Eq(int64(-1))).
		Return(protocol.Protocol{
			Hash:    "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      2,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Edo2net),
			gomock.Eq("PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA"),
			gomock.Eq(int64(-1))).
		Return(protocol.Protocol{
			Hash:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
			Network: types.Edo2net,
			SymLink: bcd.SymLinkBabylon,
			ID:      3,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Mainnet),
			gomock.Eq("PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i"),
			gomock.Eq(int64(-1))).
		Return(protocol.Protocol{
			Hash:    "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      4,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(gomock.Eq(int64(0))).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			Network: types.Delphinet,
			SymLink: bcd.SymLinkBabylon,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(gomock.Eq(int64(1))).
		Return(protocol.Protocol{
			Hash:    "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      1,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(gomock.Eq(int64(2))).
		Return(protocol.Protocol{
			Hash:    "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      2,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(gomock.Eq(int64(3))).
		Return(protocol.Protocol{
			Hash:    "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
			Network: types.Edo2net,
			SymLink: bcd.SymLinkBabylon,
			ID:      3,
		}, nil).
		AnyTimes()

	protoRepo.
		EXPECT().
		GetByID(gomock.Eq(int64(4))).
		Return(protocol.Protocol{
			Hash:    "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
			Network: types.Mainnet,
			SymLink: bcd.SymLinkBabylon,
			ID:      4,
		}, nil).
		AnyTimes()

	tests := []struct {
		name       string
		rpc        noderpc.INode
		ctx        *config.Context
		paramsOpts []ParseParamsOption
		filename   string
		storage    map[string]int64
		want       *parsers.Result
		wantErr    bool
	}{
		{
			name: "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1068669,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithNetwork(types.Mainnet),
			},
			filename: "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:     parsers.NewResult(),
		}, {
			name: "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1068669,
					ChainID:   "test",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
				WithNetwork(types.Mainnet),
			},
			storage: map[string]int64{
				"KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM": 1068668,
				"KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo": 1068668,
				"KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS": 1068668,
			},
			filename: "./data/rpc/opg/opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						Kind:            types.OperationKindTransaction,
						Source:          "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
						Fee:             37300,
						Counter:         5791164,
						GasLimit:        369423,
						StorageLimit:    90,
						Destination:     "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						Status:          types.OperationStatusApplied,
						Level:           1068669,
						Network:         types.Mainnet,
						Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Entrypoint:      "transfer",
						Timestamp:       timestamp,
						Burned:          70000,
						Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
						ProtocolID:      1,
						Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq\"},{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV\"},{\"int\":\"8010000\"}]}]}]}]}]}]}}"),
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}],{\"int\":\"62\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"63\"},{\"string\":\"Aspen Digital Token\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada700\"},{\"string\":\"ASPD\"}]},{\"int\":\"18000000\"}]}]}"),
						Tags:            types.FA12Tag,
						Transfers: []*transfer.Transfer{
							{
								Network:   types.Mainnet,
								Contract:  "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
								Initiator: "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
								Status:    types.OperationStatusApplied,
								Timestamp: timestamp,
								Level:     1068669,
								From:      "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
								To:        "tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV",
								TokenID:   0,
								Amount:    newDecimal("8010000"),
							},
						},
						BigMapDiffs: []*bigmap.Diff{
							{
								KeyHash:    "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
								Key:        []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
								Value:      []byte(`{"int":"6141000"}`),
								Level:      1068669,
								ProtocolID: 1,
								Timestamp:  timestamp,
								KeyStrings: []string{"tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq"},
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
									Ptr:      63,
								},
							}, {
								KeyHash:    "exprv2snyFbF6EDZd2YAHnnmNBoFt7bbaXhGSWGXHv4a4wnxS359ob",
								Key:        []byte(`{"bytes":"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba"}`),
								Value:      []byte(`{"int":"8010000"}`),
								Level:      1068669,
								ProtocolID: 1,
								Timestamp:  timestamp,
								KeyStrings: []string{"tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV"},
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
									Ptr:      63,
								},
							},
						},
					}, {
						Kind:            types.OperationKindTransaction,
						Source:          "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						Destination:     "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
						Status:          types.OperationStatusApplied,
						Level:           1068669,
						Counter:         5791164,
						Network:         types.Mainnet,
						Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Nonce:           setInt64(0),
						Entrypoint:      "validateAccounts",
						Internal:        true,
						Timestamp:       timestamp,
						ProtocolID:      1,
						Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
						Parameters:      []byte("{\"entrypoint\":\"validateAccounts\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"True\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]},{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada70076616c696461746552756c6573\"}]}}"),
						DeffatedStorage: []byte("{\"int\":\"61\"}"),
					}, {
						Kind:            types.OperationKindTransaction,
						Source:          "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
						Destination:     "KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS",
						Status:          types.OperationStatusApplied,
						Level:           1068669,
						Counter:         5791164,
						Network:         types.Mainnet,
						Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
						Nonce:           setInt64(1),
						Entrypoint:      "validateRules",
						Internal:        true,
						Timestamp:       timestamp,
						ProtocolID:      1,
						Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
						Parameters:      []byte("{\"entrypoint\":\"validateRules\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}]}]},{\"int\":\"2\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000c644b537bdb0dac40fe742010106546effd69395\"}]}]},{\"int\":\"6\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]},{\"prim\":\"True\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01bff38c4e363eacef338f7b2e15f00ca42fafa1ce00\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]}}"),
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]},[]]}"),
					},
				},
				BigMapState: []*bigmap.State{
					{
						KeyHash:         "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
						Key:             []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
						Value:           []byte(`{"int":"6141000"}`),
						LastUpdateLevel: 1068669,
						LastUpdateTime:  timestamp,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
							Ptr:      63,
						},
						BigMapID: 9,
					}, {
						KeyHash:         "exprv2snyFbF6EDZd2YAHnnmNBoFt7bbaXhGSWGXHv4a4wnxS359ob",
						Key:             []byte(`{"bytes":"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba"}`),
						Value:           []byte(`{"int":"8010000"}`),
						LastUpdateLevel: 1068669,
						LastUpdateTime:  timestamp,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
							Ptr:      63,
						},
						BigMapID: 9,
					},
				},
				TokenBalances: []*tokenbalance.TokenBalance{
					{
						Network:  types.Mainnet,
						Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						Address:  "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
						TokenID:  0,
						Balance:  newDecimal("-8010000"),
					}, {
						Network:  types.Mainnet,
						Contract: "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
						Address:  "tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV",
						TokenID:  0,
						Balance:  newDecimal("8010000"),
					},
				},
			},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     1151495,
					ChainID:   "test",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
				WithNetwork(types.Mainnet),
			},
			storage: map[string]int64{
				"KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr": 1151494,
				"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn": 1151494,
			},
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						ContentIndex:    0,
						Network:         types.Mainnet,
						ProtocolID:      1,
						Hash:            "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
						Internal:        false,
						Nonce:           nil,
						Status:          types.OperationStatusApplied,
						Timestamp:       timestamp,
						Level:           1151495,
						Kind:            types.OperationKindTransaction,
						Initiator:       "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
						Source:          "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
						Fee:             43074,
						Counter:         6909186,
						GasLimit:        427673,
						StorageLimit:    47,
						Destination:     "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
						Parameters:      []byte("{\"entrypoint\":\"redeem\",\"value\":{\"bytes\":\"a874aac22777351417c9bde0920cc7ed33e54453e1dd149a1f3a60521358d19a\"}}"),
						Entrypoint:      "redeem",
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"int\":\"32\"},{\"prim\":\"Unit\"}]}"),
						BigMapDiffs: []*bigmap.Diff{
							{
								BigMapID: 10,
								BigMap: bigmap.BigMap{
									Network:   types.Mainnet,
									Contract:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
									Ptr:       32,
									KeyType:   []byte(``),
									ValueType: []byte(``),
								},
								Key:        []byte(`{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
								KeyHash:    "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
								Level:      1151495,
								Timestamp:  timestamp,
								ProtocolID: 1,
							},
						},
					}, {
						ContentIndex:    0,
						Network:         types.Mainnet,
						ProtocolID:      1,
						Hash:            "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
						Internal:        true,
						Nonce:           setInt64(0),
						Status:          types.OperationStatusApplied,
						Timestamp:       timestamp,
						Level:           1151495,
						Kind:            types.OperationKindTransaction,
						Initiator:       "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
						Source:          "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
						Counter:         6909186,
						Destination:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Parameters:      []byte("{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}"),
						Entrypoint:      "transfer",
						Burned:          47000,
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"int\":\"31\"},{\"prim\":\"Pair\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DUP\"},{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DIP\",\"args\":[{\"int\":\"2\"},[{\"prim\":\"DUP\"}]]},{\"prim\":\"DIG\",\"args\":[{\"int\":\"2\"}]}]]},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"code\"}]},{\"prim\":\"PAIR\"},{\"prim\":\"PACK\"},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"NONE\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]}],[{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack code\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SOME\"}]]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"DROP\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bytes\"},{\"bytes\":\"05010000000866616c6c6261636b\"}]},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: no field fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SWAP\"}]]},{\"prim\":\"PAIR\"},{\"prim\":\"EXEC\"}],[{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"SWAP\"},{\"prim\":\"DROP\"},{\"prim\":\"PAIR\"}]]},{\"prim\":\"SWAP\"},{\"prim\":\"EXEC\"}]]}],{\"prim\":\"Pair\",\"args\":[{\"int\":\"1\"},{\"prim\":\"False\"}]}]}]}"),
						Tags:            types.FA12Tag,
						Transfers: []*transfer.Transfer{
							{
								Network:   types.Mainnet,
								Contract:  "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
								Initiator: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
								Status:    types.OperationStatusApplied,
								Timestamp: timestamp,
								Level:     1151495,
								From:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
								To:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
								TokenID:   0,
								Amount:    newDecimal("7874880"),
							},
						},
						BigMapDiffs: []*bigmap.Diff{
							{
								BigMapID: 11,
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
									Ptr:      31,
								},
								Key:        []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
								KeyHash:    "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
								Value:      []byte(`{"bytes":"050098e1e8d78a02"}`),
								Level:      1151495,
								Timestamp:  timestamp,
								ProtocolID: 1,
								KeyStrings: []string{"totalSupply"},
							}, {
								BigMapID: 11,
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
									Ptr:      31,
								},
								Key:        []byte(`{"bytes":"05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
								KeyHash:    "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
								Value:      []byte(`{"bytes":"0507070080a5c1070200000000"}`),
								Level:      1151495,
								Timestamp:  timestamp,
								ProtocolID: 1,
								KeyStrings: []string{"ledger", "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE"},
							}, {
								BigMapID: 11,
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
									Ptr:      31,
								},
								Key:        []byte(`{"bytes":"05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
								KeyHash:    "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
								Value:      []byte(`{"bytes":"05070700ba81bb090200000000"}`),
								Level:      1151495,
								Timestamp:  timestamp,
								ProtocolID: 1,
								KeyStrings: []string{"ledger", "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr"},
							},
						},
					},
				},
				BigMapState: []*bigmap.State{
					{
						BigMapID: 10,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
							Ptr:      32,
						},
						Key:             []byte(`{"bytes":"80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
						KeyHash:         "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
						Removed:         true,
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						BigMapID: 11,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						Key:             []byte(`{"bytes":"05010000000b746f74616c537570706c79"}`),
						KeyHash:         "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
						Value:           []byte(`{"bytes":"050098e1e8d78a02"}`),
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						BigMapID: 11,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
						KeyHash:         "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
						Value:           []byte(`{"bytes":"0507070080a5c1070200000000"}`),
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					}, {
						BigMapID: 11,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
							Ptr:      31,
						},
						Key:             []byte(`{"bytes":"05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
						KeyHash:         "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
						Value:           []byte(`{"bytes":"05070700ba81bb090200000000"}`),
						LastUpdateLevel: 1151495,
						LastUpdateTime:  timestamp,
					},
				},
				TokenBalances: []*tokenbalance.TokenBalance{
					{
						Network:  types.Mainnet,
						Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Address:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
						TokenID:  0,
						Balance:  newDecimal("-7874880"),
					}, {
						Network:  types.Mainnet,
						Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
						Address:  "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
						TokenID:  0,
						Balance:  newDecimal("7874880"),
					},
				},
			},
		}, {
			name: "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     86142,
					ChainID:   "test",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  250,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
				WithNetwork(types.Delphinet),
			},
			storage: map[string]int64{
				"KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR": 86142,
			},
			filename: "./data/rpc/opg/onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						ContentIndex:                       0,
						Network:                            types.Delphinet,
						ProtocolID:                         0,
						Hash:                               "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
						Internal:                           false,
						Status:                             types.OperationStatusApplied,
						Timestamp:                          timestamp,
						Level:                              86142,
						Kind:                               types.OperationKindOrigination,
						Initiator:                          "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
						Source:                             "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
						Fee:                                510,
						Counter:                            654594,
						GasLimit:                           1870,
						StorageLimit:                       371,
						Amount:                             0,
						Destination:                        "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
						Burned:                             87750,
						AllocatedDestinationContractBurned: 64250,
						DeffatedStorage:                    []byte("{\"int\":\"0\"}\n"),
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Network:     types.Delphinet,
						Level:       86142,
						Timestamp:   timestamp,
						Language:    "unknown",
						Hash:        "e4b88b53b9227b3fc4fc0dbe148f249a7a1c755cf4cbc9c8fb5b5b78395a139d3f8e0fde5c27117df30553e98ecb4e3e8ddc9740292af18fbf36326cb55cebad",
						Entrypoints: []string{"decrement", "increment"},
						Address:     "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
						Manager:     "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
					},
				},
			},
		}, {
			name: "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
					Level:     301436,
					ChainID:   "test",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     400000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
				WithNetwork(types.Mainnet),
			},
			storage: map[string]int64{
				"KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4": 301436,
			},
			filename: "./data/rpc/opg/onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						Kind:                               types.OperationKindOrigination,
						Source:                             "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
						Fee:                                1555,
						Counter:                            983250,
						GasLimit:                           12251,
						StorageLimit:                       351,
						Destination:                        "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
						Status:                             types.OperationStatusApplied,
						Level:                              301436,
						Network:                            types.Mainnet,
						Hash:                               "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
						Timestamp:                          timestamp,
						Burned:                             331000,
						Initiator:                          "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
						ProtocolID:                         2,
						DeffatedStorage:                    []byte("[]"),
						AllocatedDestinationContractBurned: 257000,
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Network:     types.Mainnet,
						Level:       301436,
						Timestamp:   timestamp,
						Language:    "unknown",
						Hash:        "0569cf67a58ae603cbfa740c3181b588608f8967e8a7d1ea49e00c9325e9e1b67dc32cd1ec1f9cdc73699dd793ded16ac6f14511b61b63240e8f647b3aed17a3",
						Tags:        types.Tags(0),
						Entrypoints: []string{"default"},
						Address:     "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
						Manager:     "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
					},
				},
			},
		}, {
			name: "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapState:   bmStateRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					Level:     72207,
					ChainID:   "test",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     400000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
				WithNetwork(types.Edo2net),
			},
			storage: map[string]int64{
				"KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264": 72206,
				"KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU": 72207,
			},
			filename: "./data/rpc/opg/op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						Kind:            types.OperationKindTransaction,
						Source:          "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
						Fee:             4045,
						Counter:         155670,
						GasLimit:        37831,
						StorageLimit:    5265,
						Destination:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
						Status:          types.OperationStatusApplied,
						Level:           72207,
						Network:         types.Edo2net,
						Hash:            "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
						Timestamp:       timestamp,
						Entrypoint:      "@entrypoint_1",
						Initiator:       "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
						Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Unit\"}]}}"),
						ProtocolID:      3,
						DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000e527ed176ccf8f8297f674a9886a2ba8a55818d9\"},{\"prim\":\"Left\",\"args\":[{\"bytes\":\"016ebc941b2ae4e305470f392fa050e41ca1e52b4500\"}]}]}"),
						BigMapActions: []*bigmap.Action{
							{
								Action:    types.BigMapActionRemove,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(1),
								Source: bigmap.BigMap{
									Network:  types.Edo2net,
									Contract: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:      25167,
								},
							}, {
								Action:    types.BigMapActionRemove,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(2),
								Source: bigmap.BigMap{
									Network:  types.Edo2net,
									Contract: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:      25166,
								},
							}, {
								Action:    types.BigMapActionRemove,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(3),
								Source: bigmap.BigMap{
									Network:  types.Edo2net,
									Contract: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:      25165,
								},
							}, {
								Action:    types.BigMapActionRemove,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(4),
								Source: bigmap.BigMap{
									Network:  types.Edo2net,
									Contract: "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:      25164,
								},
							},
						},
					}, {
						Kind:                               types.OperationKindOrigination,
						Source:                             "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
						Nonce:                              setInt64(0),
						Destination:                        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Status:                             types.OperationStatusApplied,
						Level:                              72207,
						Network:                            types.Edo2net,
						Hash:                               "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
						Timestamp:                          timestamp,
						Burned:                             5245000,
						Counter:                            155670,
						Internal:                           true,
						Initiator:                          "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
						ProtocolID:                         3,
						AllocatedDestinationContractBurned: 257000,
						Tags:                               types.FA2Tag,
						DeffatedStorage:                    []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},[]]},{\"int\":\"25168\"},{\"int\":\"25169\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Unit\"}]},{\"int\":\"25170\"}]},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},{\"int\":\"0\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"25171\"}]},{\"int\":\"2\"},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"}]},{\"int\":\"11\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]}],{\"prim\":\"DROP\"},{\"prim\":\"NIL\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"PAIR\"}]]},{\"int\":\"500\"},{\"int\":\"1000\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"1000\"},{\"int\":\"2592000\"}]},{\"int\":\"1\"},{\"int\":\"1\"}]},[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bool\"},{\"prim\":\"True\"}]}],[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"nat\"},{\"int\":\"0\"}]}]]}"),
						BigMapActions: []*bigmap.Action{
							{
								Action:    types.BigMapActionCopy,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(1),
								Source: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:          25167,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
								Destination: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
									Ptr:          25171,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
							}, {
								Action:    types.BigMapActionCopy,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(2),
								Source: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:          25166,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
								Destination: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
									Ptr:          25170,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
							}, {
								Action:    types.BigMapActionCopy,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(3),
								Source: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:          25165,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
								Destination: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
									Ptr:          25169,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
							}, {
								Action:    types.BigMapActionCopy,
								Level:     72207,
								Timestamp: timestamp,
								SourceID:  newInt64Ptr(4),
								Source: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
									Ptr:          25164,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
								Destination: bigmap.BigMap{
									Network:      types.Edo2net,
									Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
									Ptr:          25168,
									CreatedLevel: 72207,
									CreatedAt:    timestamp,
								},
							},
						}},
				},
				BigMaps: []*bigmap.BigMap{
					{
						Network:      types.Edo2net,
						Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Ptr:          25171,
						Name:         "proposals",
						CreatedLevel: 72207,
						CreatedAt:    timestamp,
					}, {
						Network:      types.Edo2net,
						Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Ptr:          25170,
						Name:         "operators",
						CreatedLevel: 72207,
						CreatedAt:    timestamp,
					}, {
						Network:      types.Edo2net,
						Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Ptr:          25169,
						Name:         "metadata",
						CreatedLevel: 72207,
						CreatedAt:    timestamp,
					}, {
						Network:      types.Edo2net,
						Contract:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Ptr:          25168,
						Name:         "ledger",
						CreatedLevel: 72207,
						CreatedAt:    timestamp,
					},
				},
				Contracts: []*modelContract.Contract{
					{
						Network:     types.Edo2net,
						Level:       72207,
						Timestamp:   timestamp,
						Language:    "unknown",
						Hash:        "d3bfdacb039f6e8added88c45046b7a8f6a2b91744859ace29f4c19294c9a394857598e2b331394cac91a7a2c543cadaa60282c5eb2c87f83f001f5e563cea36",
						Tags:        types.FA2Tag,
						FailStrings: []string{"FA2_INSUFFICIENT_BALANCE"},
						Annotations: []string{"%token_address", "%drop_proposal", "%transfer_contract_tokens", "%permits_counter", "%remove_operator", "%mint", "%ledger", "%voters", "%owner", "%balance", "%transfer", "%from_", "%max_voting_period", "%not_in_migration", "%start_date", "%custom_entrypoints", "%proposal_check", "%accept_ownership", "%migrate", "%set_quorum_threshold", "%amount", "%proposals", "%min_voting_period", "%rejected_proposal_return_value", "%burn", "%flush", "%max_quorum_threshold", "%migratingTo", "%operators", "%proposer", "%call_FA2", "%argument", "%params", "%transfer_ownership", "%voting_period", "%request", "%confirm_migration", "%frozen_token", "%param", "%admin", "%migration_status", "%proposal_key_list_sort_by_date", "%requests", "%update_operators", "%add_operator", "%getVotePermitCounter", "%propose", "%vote", "%vote_amount", "%proposer_frozen_token", "%callCustom", "%txs", "%operator", "%quorum_threshold", "%to_", "%set_voting_period", "%callback", "%contract_address", "%downvotes", "%max_votes", "%balance_of", "%proposal_key", "%vote_type", "%signature", "%decision_lambda", "%token_id", "%permit", "%key", "%extra", "%pending_owner", "%upvotes", "%max_proposals", "%min_quorum_threshold", "%proposal_metadata", "%metadata", "%migratedTo"},
						Entrypoints: []string{"callCustom", "accept_ownership", "burn", "balance_of", "transfer", "update_operators", "confirm_migration", "drop_proposal", "flush", "getVotePermitCounter", "migrate", "mint", "propose", "set_quorum_threshold", "set_voting_period", "transfer_ownership", "vote", "transfer_contract_tokens"},
						Address:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
						Manager:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					},
				},
			},
		}, {
			name: "ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
					Level:     1516349,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     400000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
				WithNetwork(types.Mainnet),
			},
			storage: map[string]int64{
				"KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy": 1516349,
			},
			filename: "./data/rpc/opg/ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c.json",
			want: &parsers.Result{
				Operations: []*operation.Operation{
					{
						Kind:            types.OperationKindTransaction,
						Source:          "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
						Fee:             2235,
						Counter:         9432992,
						GasLimit:        18553,
						Destination:     "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
						Status:          types.OperationStatusApplied,
						Level:           1516349,
						Network:         types.Mainnet,
						Hash:            "ooz1bkCQeYsZYP7vb4Dx7pYPRpWN11Z3G3yP1v4HAfdNXuHRv9c",
						Timestamp:       timestamp,
						Entrypoint:      "transfer",
						Tags:            types.FA2Tag,
						Initiator:       "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
						Parameters:      []byte(`{"entrypoint":"transfer","value":[{"prim":"Pair","args":[{"string":"tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb"},[{"prim":"Pair","args":[{"string":"tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi"},{"prim":"Pair","args":[{"int":"12"},{"int":"1"}]}]}]]}]}`),
						ProtocolID:      4,
						DeffatedStorage: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"746"},{"int":"4992269"}]},{"prim":"Pair","args":[{"int":"747"},{"int":"748"}]}]},{"int":"749"}]}`),
						Transfers: []*transfer.Transfer{
							{
								TokenID:   12,
								From:      "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
								To:        "tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi",
								Amount:    decimal.NewFromInt(1),
								Network:   types.Mainnet,
								Contract:  "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
								Initiator: "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
								Status:    types.OperationStatusApplied,
								Timestamp: timestamp,
								Level:     1516349,
							},
						},
						BigMapDiffs: []*bigmap.Diff{
							{
								BigMapID: 12,
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
									Ptr:      746,
								},
								KeyHash:      "expruSKSLw7MS3ou3pPd7MUXy5QDPtVvkUNF4yWS2g6n8mXGzDJCG7",
								Key:          []byte(`{"int":"12" }`),
								Value:        []byte(`{"bytes":"00009e96262b1bfc9a709603668843d52994358be677"}`),
								ValueStrings: []string{"tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi"},
								Level:        1516349,
								Timestamp:    timestamp,
								ProtocolID:   4,
							},
						},
					},
				},
				BigMapState: []*bigmap.State{
					{
						BigMapID: 12,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
							Ptr:      746,
						},
						KeyHash:         "expruSKSLw7MS3ou3pPd7MUXy5QDPtVvkUNF4yWS2g6n8mXGzDJCG7",
						Key:             []byte(`{"int":"12"}`),
						Value:           []byte(`{"bytes":"00009e96262b1bfc9a709603668843d52994358be677"}`),
						LastUpdateLevel: 1516349,
						LastUpdateTime:  timestamp,
					},
				},
				TokenBalances: []*tokenbalance.TokenBalance{
					{
						Network:  types.Mainnet,
						Contract: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
						Address:  "tz1aCzsYRUgDZBV7zb7Si6q2AobrocFW5qwb",
						TokenID:  12,
						Balance:  decimal.NewFromInt(-1),
					}, {
						Network:  types.Mainnet,
						Contract: "KT1QcxwB4QyPKfmSwjH1VRxa6kquUjeDWeEy",
						Address:  "tz1a6ZKyEoCmfpsY74jEq6uKBK8RQXdj1aVi",
						TokenID:  12,
						Balance:  decimal.NewFromInt(1),
					},
				},
			},
		}, {
			name: "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
			rpc:  rpc,
			ctx: &config.Context{
				Storage:       generalRepo,
				Contracts:     contractRepo,
				BigMaps:       bmRepo,
				BigMapDiffs:   bmdRepo,
				Blocks:        blockRepo,
				Protocols:     protoRepo,
				TZIP:          tzipRepo,
				TokenBalances: tbRepo,
				Cache:         cache.NewCache(),
				SharePath:     "./test",
			},
			paramsOpts: []ParseParamsOption{
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
					Level:     1520888,
					ChainID:   "NetXdQprcVkpaWU",
				}),
				WithNetwork(types.Mainnet),
			},
			filename: "./data/rpc/opg/oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8.json",
			storage: map[string]int64{
				"KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS": 1520888,
				"KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc": 1520888,
			},
			want: &parsers.Result{
				BigMapState: []*bigmap.State{
					{
						BigMapID: 13,
						BigMap: bigmap.BigMap{
							Network:  types.Mainnet,
							Contract: "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
							Ptr:      1264,
						},
						LastUpdateLevel: 1520888,
						LastUpdateTime:  timestamp,
						KeyHash:         "exprvKwnhi4q3tSmdvgqXACxfN6zARGkoikHv7rqohvQKg4cWdgsii",
						Key:             []byte(`{"bytes":"62616c6c732e74657a"}`),
						Value:           []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
					},
				},
				Operations: []*operation.Operation{
					{
						Kind:            types.OperationKindTransaction,
						Hash:            "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
						Source:          "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
						Initiator:       "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
						Status:          types.OperationStatusApplied,
						Fee:             5043,
						Counter:         10671622,
						GasLimit:        46511,
						StorageLimit:    400,
						Amount:          0,
						Timestamp:       timestamp,
						Level:           1520888,
						Network:         types.Mainnet,
						Entrypoint:      "update_record",
						ProtocolID:      4,
						Destination:     "KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc",
						Parameters:      []byte(`{"entrypoint":"update_record","value":{"prim":"Pair","args":[{"bytes":"62616c6c732e74657a"},{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"tz1dDQc4KsTHEFe3USc66Wti2pBatZ3UDbD4"}]},{"prim":"Pair","args":[{"string":"tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT"},[]]}]}]}}`),
						DeffatedStorage: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"01535d971759846a1f2be8610e36f2db40fe8ce40800"},{"int":"1268"}]},{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"}]}`),
					},
					{
						Kind:            types.OperationKindTransaction,
						Hash:            "oocFt4vkkgQGfoRH54328cJUbDdWvj3x6KEs5Arm4XhqwwJmnJ8",
						Internal:        true,
						Timestamp:       timestamp,
						Status:          types.OperationStatusApplied,
						Level:           1520888,
						Nonce:           newInt64Ptr(0),
						Counter:         10671622,
						Network:         types.Mainnet,
						ProtocolID:      4,
						Entrypoint:      "execute",
						Initiator:       "tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
						Source:          "KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc",
						Destination:     "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
						Parameters:      []byte(`{"entrypoint":"execute","value":{"prim":"Pair","args":[{"string":"UpdateRecord"},{"prim":"Pair","args":[{"bytes":"0507070a0000000962616c6c732e74657a070705090a000000160000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c07070a000000160000753f63893674b6d523f925f0d787bf9270b95c330200000000"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]}]}}`),
						DeffatedStorage: []byte(`{"prim":"Pair","args":[[{"int":"1260"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1261"},{"int":"1262"}]},{"prim":"Pair","args":[{"int":"1263"},{"int":"9824"}]}]},{"prim":"Pair","args":[{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"},{"int":"1264"}]},{"int":"1265"},{"int":"1266"}],[{"bytes":"014796e76af90e6327adfab057bbbe0375cd2c8c1000"},{"bytes":"015c6799f783b8d118b704267f634c5d24d19e9a9f00"},{"bytes":"0168e9b7d86646e312c76dfbedcbcdb24320875a3600"},{"bytes":"019178a76f3c41a9541d2291cad37dd5fb96a6850500"},{"bytes":"01ac3638385caa4ad8126ea84e061f4f49baa44d3c00"},{"bytes":"01d2a0974172cf6fc8b1eefdebd5bea681616f7c6f00"}]]}`),
						BigMapDiffs: []*bigmap.Diff{
							{
								BigMapID: 13,
								BigMap: bigmap.BigMap{
									Network:  types.Mainnet,
									Contract: "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
									Ptr:      1264,
								},
								ProtocolID: 4,
								Level:      1520888,
								Timestamp:  timestamp,
								KeyHash:    "exprvKwnhi4q3tSmdvgqXACxfN6zARGkoikHv7rqohvQKg4cWdgsii",
								Key:        []byte(`{"bytes":"62616c6c732e74657a"}`),
								Value:      []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
								KeyStrings: []string{"balls.tez"},
								ValueStrings: []string{
									"tz1dDQc4KsTHEFe3USc66Wti2pBatZ3UDbD4",
									"balls.tez",
									"tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for address, level := range tt.storage {
				rpc.
					EXPECT().
					GetScriptStorageRaw(address, level).
					DoAndReturn(
						func(address string, level int64) ([]byte, error) {
							storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
							return ioutil.ReadFile(storageFile)
						},
					).
					AnyTimes()
			}

			var op noderpc.OperationGroup
			if err := readJSONFile(tt.filename, &op); err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}

			parseParams, err := NewParseParams(tt.rpc, tt.ctx, tt.paramsOpts...)
			if err != nil {
				t.Errorf(`NewParseParams = error %v`, err)
				return
			}

			opg := NewGroup(parseParams)
			got, err := opg.Parse(op)
			if (err != nil) != tt.wantErr {
				t.Errorf("Group.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareParserResponse(t, got, tt.want) {
				t.Errorf("Group.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
