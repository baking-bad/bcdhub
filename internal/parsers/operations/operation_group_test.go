package operations

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelContract "github.com/baking-bad/bcdhub/internal/models/contract"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_contract "github.com/baking-bad/bcdhub/internal/models/mock/contract"
	mock_token_balance "github.com/baking-bad/bcdhub/internal/models/mock/tokenbalance"
	mock_tzip "github.com/baking-bad/bcdhub/internal/models/mock/tzip"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/golang/mock/gomock"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlBmdRepo := gomock.NewController(t)
	defer ctrlBmdRepo.Finish()
	bmdRepo := mock_bmd.NewMockRepository(ctrlBmdRepo)

	ctrlBlockRepo := gomock.NewController(t)
	defer ctrlBlockRepo.Finish()
	blockRepo := mock_block.NewMockRepository(ctrlBlockRepo)

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
		GetWithEvents().
		Return(make([]tzip.TZIP, 0), nil).
		AnyTimes()

	tzipRepo.
		EXPECT().
		GetWithEventsCounts().
		Return(int64(0), nil).
		AnyTimes()

	tbRepo.
		EXPECT().
		Update(gomock.Any()).
		Return(nil).
		AnyTimes()

	contractRepo.
		EXPECT().
		Get(gomock.Any(), gomock.Any()).
		DoAndReturn(readTestContractModel).
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

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2416))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:       2416,
				Key:       []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:   "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:     []byte(`{"prim":"Pair","args":[[],{"int":"6000"}]}`),
				Level:     386026,
				Contract:  "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
				Network:   "carthagenet",
				Timestamp: timestamp,
				Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	for _, ptr := range []int{25167, 25166, 25165, 25164} {
		bmdRepo.
			EXPECT().
			GetByPtr(
				gomock.Eq("KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264"),
				gomock.Eq("edo2net"),
				gomock.Eq(int64(ptr))).
			Return([]bigmapdiff.BigMapDiff{}, nil).
			AnyTimes()
	}

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2417))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:       2417,
				Key:       []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:   "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				Value:     nil,
				Level:     386026,
				Contract:  "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
				Network:   "carthagenet",
				Timestamp: timestamp,
				Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	tests := []struct {
		name        string
		ParseParams *ParseParams
		filename    string
		storage     map[string]int64
		want        []models.Model
		wantErr     bool
	}{
		{
			name:        "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ParseParams: NewParseParams(rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo),
			filename:    "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:        []models.Model{},
		}, {
			name: "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
			ParseParams: NewParseParams(rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     1068669,
					ChainID:   "test",
				}),
				WithShareDirectory("./test"),
				WithNetwork(consts.Mainnet),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			storage: map[string]int64{
				"KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM": 1068668,
				"KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo": 1068668,
				"KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS": 1068668,
			},
			filename: "./data/rpc/opg/opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9.json",
			want: []models.Model{
				&operation.Operation{
					Kind:            "transaction",
					Source:          "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					Fee:             37300,
					Counter:         5791164,
					GasLimit:        369423,
					StorageLimit:    90,
					Destination:     "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
					Status:          "applied",
					Level:           1068669,
					Network:         "mainnet",
					Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
					Entrypoint:      "transfer",
					Timestamp:       timestamp,
					Burned:          70000,
					Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					Protocol:        "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq\"},{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV\"},{\"int\":\"8010000\"}]}]}]}]}]}]}}"),
					DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}],{\"int\":\"62\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"63\"},{\"string\":\"Aspen Digital Token\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada700\"},{\"string\":\"ASPD\"}]},{\"int\":\"18000000\"}]}]}"),
					Tags:            []string{"fa1-2"},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       63,
					KeyHash:   "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
					Key:       []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
					Value:     []byte(`{"int":"6141000"}`),
					Level:     1068669,
					Network:   "mainnet",
					Contract:  "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Timestamp: timestamp,
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       63,
					KeyHash:   "exprv2snyFbF6EDZd2YAHnnmNBoFt7bbaXhGSWGXHv4a4wnxS359ob",
					Key:       []byte(`{"bytes":"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba"}`),
					Value:     []byte(`{"int":"8010000"}`),
					Level:     1068669,
					Network:   "mainnet",
					Contract:  "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Timestamp: timestamp,
				},
				&transfer.Transfer{
					Network:      consts.Mainnet,
					Contract:     "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
					Initiator:    "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					Hash:         "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
					Status:       consts.Applied,
					Timestamp:    timestamp,
					Level:        1068669,
					From:         "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					To:           "tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV",
					TokenID:      0,
					AmountBigInt: big.NewInt(8010000),
					Counter:      5791164,
				},
				&operation.Operation{
					Kind:            "transaction",
					Source:          "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
					Destination:     "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
					Status:          "applied",
					Level:           1068669,
					Counter:         5791164,
					Network:         "mainnet",
					Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
					Nonce:           setInt64(0),
					Entrypoint:      "validateAccounts",
					Internal:        true,
					Timestamp:       timestamp,
					Protocol:        "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					Parameters:      []byte("{\"entrypoint\":\"validateAccounts\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"True\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]},{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada70076616c696461746552756c6573\"}]}}"),
					DeffatedStorage: []byte("{\"int\":\"61\"}"),
					Tags:            []string{},
				},
				&operation.Operation{
					Kind:            "transaction",
					Source:          "KT19nHqEWZxFFbbDL1b7Y86escgEN7qUShGo",
					Destination:     "KT1KemKUx79keZgFW756jQrqKcZJ21y4SPdS",
					Status:          "applied",
					Level:           1068669,
					Counter:         5791164,
					Network:         "mainnet",
					Hash:            "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
					Nonce:           setInt64(1),
					Entrypoint:      "validateRules",
					Internal:        true,
					Timestamp:       timestamp,
					Protocol:        "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Initiator:       "tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq",
					Parameters:      []byte("{\"entrypoint\":\"validateRules\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}]}]},{\"int\":\"2\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000c644b537bdb0dac40fe742010106546effd69395\"}]}]},{\"int\":\"6\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]},{\"prim\":\"True\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01bff38c4e363eacef338f7b2e15f00ca42fafa1ce00\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]}}"),
					DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]},[]]}"),
					Tags:            []string{},
				},
			},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ParseParams: NewParseParams(
				rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     1151495,
					ChainID:   "test",
				}),
				WithNetwork(consts.Mainnet),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			storage: map[string]int64{
				"KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr": 1151494,
				"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn": 1151494,
			},
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:    0,
					Network:         consts.Mainnet,
					Protocol:        "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:            "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:        false,
					Nonce:           nil,
					Status:          consts.Applied,
					Timestamp:       timestamp,
					Level:           1151495,
					Kind:            "transaction",
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
					Tags:            []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       32,
					Key:       []byte(`{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
					KeyHash:   "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
					Level:     1151495,
					Contract:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Network:   consts.Mainnet,
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&operation.Operation{
					ContentIndex:    0,
					Network:         consts.Mainnet,
					Protocol:        "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:            "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:        true,
					Nonce:           setInt64(0),
					Status:          consts.Applied,
					Timestamp:       timestamp,
					Level:           1151495,
					Kind:            "transaction",
					Initiator:       "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Source:          "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Counter:         6909186,
					Destination:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Parameters:      []byte("{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}"),
					Entrypoint:      "transfer",
					Burned:          47000,
					DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"int\":\"31\"},{\"prim\":\"Pair\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DUP\"},{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DIP\",\"args\":[{\"int\":\"2\"},[{\"prim\":\"DUP\"}]]},{\"prim\":\"DIG\",\"args\":[{\"int\":\"2\"}]}]]},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"code\"}]},{\"prim\":\"PAIR\"},{\"prim\":\"PACK\"},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"NONE\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]}],[{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack code\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SOME\"}]]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"DROP\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bytes\"},{\"bytes\":\"05010000000866616c6c6261636b\"}]},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: no field fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SWAP\"}]]},{\"prim\":\"PAIR\"},{\"prim\":\"EXEC\"}],[{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"SWAP\"},{\"prim\":\"DROP\"},{\"prim\":\"PAIR\"}]]},{\"prim\":\"SWAP\"},{\"prim\":\"EXEC\"}]]}],{\"prim\":\"Pair\",\"args\":[{\"int\":\"1\"},{\"prim\":\"False\"}]}]}]}"),
					Tags:            []string{"fa1-2"},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       31,
					Key:       []byte(`{"bytes": "05010000000b746f74616c537570706c79"}`),
					KeyHash:   "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
					Value:     []byte(`{"bytes":"050098e1e8d78a02"}`),
					Level:     1151495,
					Contract:  "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:   consts.Mainnet,
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:     31,
					Key:     []byte(`{"bytes": "05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
					KeyHash: "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
					Value:   []byte(`{"bytes":"0507070080a5c1070200000000"}`),
					Level:   1151495, Contract: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:   consts.Mainnet,
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       31,
					Key:       []byte(`{"bytes": "05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
					KeyHash:   "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
					Value:     []byte(`{"bytes":"05070700ba81bb090200000000"}`),
					Level:     1151495,
					Contract:  "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:   consts.Mainnet,
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&transfer.Transfer{
					Network:      consts.Mainnet,
					Contract:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Initiator:    "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Hash:         "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Status:       consts.Applied,
					Timestamp:    timestamp,
					Level:        1151495,
					From:         "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					To:           "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					TokenID:      0,
					AmountBigInt: big.NewInt(7.87488e+06),
					Counter:      6909186,
					Nonce:        setInt64(0),
				},
			},
		}, {
			name: "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
			ParseParams: NewParseParams(
				rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     86142,
					ChainID:   "test",
				}),
				WithNetwork("delphinet"),
				WithConstants(protocol.Constants{
					CostPerByte:                  250,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
				WithShareDirectory("test"),
			),
			storage: map[string]int64{
				"KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR": 86142,
			},
			filename: "./data/rpc/opg/onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt.json",
			want: []models.Model{
				&modelContract.Contract{
					Network:     "delphinet",
					Level:       86142,
					Timestamp:   timestamp,
					Language:    "unknown",
					Hash:        "e4b88b53b9227b3fc4fc0dbe148f249a7a1c755cf4cbc9c8fb5b5b78395a139d3f8e0fde5c27117df30553e98ecb4e3e8ddc9740292af18fbf36326cb55cebad",
					Tags:        []string{},
					Entrypoints: []string{"decrement", "increment"},
					Address:     "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
					Manager:     "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
				},
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "delphinet",
					Protocol:                           "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Hash:                               "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
					Internal:                           false,
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              86142,
					Kind:                               "origination",
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
					Tags:                               nil,
				},
			},
		}, {
			name: "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
			ParseParams: NewParseParams(
				rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
					Level:     301436,
					ChainID:   "test",
				}),
				WithNetwork("mainnet"),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     400000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			storage: map[string]int64{
				"KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4": 301436,
			},
			filename: "./data/rpc/opg/onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC.json",
			want: []models.Model{
				&modelContract.Contract{
					Network:     "mainnet",
					Level:       301436,
					Timestamp:   timestamp,
					Language:    "unknown",
					Hash:        "0569cf67a58ae603cbfa740c3181b588608f8967e8a7d1ea49e00c9325e9e1b67dc32cd1ec1f9cdc73699dd793ded16ac6f14511b61b63240e8f647b3aed17a3",
					Tags:        []string{},
					Entrypoints: []string{"default"},
					Address:     "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
					Manager:     "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
				},
				&operation.Operation{
					Kind:                               "origination",
					Source:                             "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
					Fee:                                1555,
					Counter:                            983250,
					GasLimit:                           12251,
					StorageLimit:                       351,
					Destination:                        "KT1AbjG7vtpV8osdoJXcMRck8eTwst8dWoz4",
					Status:                             "applied",
					Level:                              301436,
					Network:                            "mainnet",
					Hash:                               "onv6Q1dNejAGEJeQzwRannWsDSGw85FuFdhLnBrY18TBcC9p8kC",
					Timestamp:                          timestamp,
					Burned:                             331000,
					Initiator:                          "tz1MXrEgDNnR8PDryN8sq4B2m9Pqcf57wBqM",
					Protocol:                           "PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
					DeffatedStorage:                    []byte("[]"),
					AllocatedDestinationContractBurned: 257000,
				},
			},
		}, {
			name: "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
			ParseParams: NewParseParams(
				rpc, generalRepo, contractRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					Level:     72207,
					ChainID:   "test",
				}),
				WithNetwork("edo2net"),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     400000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			storage: map[string]int64{
				"KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264": 72206,
				"KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU": 72207,
			},
			filename: "./data/rpc/opg/op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg.json",
			want: []models.Model{
				&operation.Operation{
					Kind:            "transaction",
					Source:          "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
					Fee:             4045,
					Counter:         155670,
					GasLimit:        37831,
					StorageLimit:    5265,
					Destination:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Status:          "applied",
					Level:           72207,
					Network:         "edo2net",
					Hash:            "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
					Timestamp:       timestamp,
					Entrypoint:      "@entrypoint_1",
					Initiator:       "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
					Parameters:      []byte("{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Unit\"}]}}"),
					Protocol:        "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					DeffatedStorage: []byte("{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000e527ed176ccf8f8297f674a9886a2ba8a55818d9\"},{\"prim\":\"Left\",\"args\":[{\"bytes\":\"016ebc941b2ae4e305470f392fa050e41ca1e52b4500\"}]}]}"),
				},
				&bigmapaction.BigMapAction{
					Action:    "remove",
					SourcePtr: setInt64(25167),
					Level:     72207,
					Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Network:   "edo2net",
					Timestamp: timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:    "remove",
					SourcePtr: setInt64(25166),
					Level:     72207,
					Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Network:   "edo2net",
					Timestamp: timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:    "remove",
					SourcePtr: setInt64(25165),
					Level:     72207,
					Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Network:   "edo2net",
					Timestamp: timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:    "remove",
					SourcePtr: setInt64(25164),
					Level:     72207,
					Address:   "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Network:   "edo2net",
					Timestamp: timestamp,
				},
				&modelContract.Contract{
					Network:     "edo2net",
					Level:       72207,
					Timestamp:   timestamp,
					Language:    "unknown",
					Hash:        "d3bfdacb039f6e8added88c45046b7a8f6a2b91744859ace29f4c19294c9a394857598e2b331394cac91a7a2c543cadaa60282c5eb2c87f83f001f5e563cea36",
					Tags:        []string{"nft_ledger", "fa2"},
					FailStrings: []string{"FA2_INSUFFICIENT_BALANCE"},
					Annotations: []string{"%token_address", "%drop_proposal", "%transfer_contract_tokens", "%permits_counter", "%remove_operator", "%mint", "%ledger", "%voters", "%owner", "%balance", "%transfer", "%from_", "%max_voting_period", "%not_in_migration", "%start_date", "%custom_entrypoints", "%proposal_check", "%accept_ownership", "%migrate", "%set_quorum_threshold", "%amount", "%proposals", "%min_voting_period", "%rejected_proposal_return_value", "%burn", "%flush", "%max_quorum_threshold", "%migratingTo", "%operators", "%proposer", "%call_FA2", "%argument", "%params", "%transfer_ownership", "%voting_period", "%request", "%confirm_migration", "%frozen_token", "%param", "%admin", "%migration_status", "%proposal_key_list_sort_by_date", "%requests", "%update_operators", "%add_operator", "%getVotePermitCounter", "%propose", "%vote", "%vote_amount", "%proposer_frozen_token", "%callCustom", "%txs", "%operator", "%quorum_threshold", "%to_", "%set_voting_period", "%callback", "%contract_address", "%downvotes", "%max_votes", "%balance_of", "%proposal_key", "%vote_type", "%signature", "%decision_lambda", "%token_id", "%permit", "%key", "%extra", "%pending_owner", "%upvotes", "%max_proposals", "%min_quorum_threshold", "%proposal_metadata", "%metadata", "%migratedTo"},
					Entrypoints: []string{"callCustom", "accept_ownership", "burn", "balance_of", "transfer", "update_operators", "confirm_migration", "drop_proposal", "flush", "getVotePermitCounter", "migrate", "mint", "propose", "set_quorum_threshold", "set_voting_period", "transfer_ownership", "vote", "transfer_contract_tokens"},
					Address:     "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Manager:     "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(25167),
					DestinationPtr: setInt64(25171),
					Level:          72207,
					Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Network:        "edo2net",
					Timestamp:      timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(25166),
					DestinationPtr: setInt64(25170),
					Level:          72207,
					Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Network:        "edo2net",
					Timestamp:      timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(25165),
					DestinationPtr: setInt64(25169),
					Level:          72207,
					Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Network:        "edo2net",
					Timestamp:      timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(25164),
					DestinationPtr: setInt64(25168),
					Level:          72207,
					Address:        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Network:        "edo2net",
					Timestamp:      timestamp,
				},
				&operation.Operation{
					Kind:                               "origination",
					Source:                             "KT1C2MfcjWb5R1ZDDxVULCsGuxrf5fEn5264",
					Nonce:                              setInt64(0),
					Destination:                        "KT1JgHoXtZPjVfG82BY3FSys2VJhKVZo2EJU",
					Status:                             "applied",
					Level:                              72207,
					Network:                            "edo2net",
					Hash:                               "op4fFMvYsxvSUKZmLWC7aUf25VMYqigaDwTZCAoBBi8zACbHTNg",
					Timestamp:                          timestamp,
					Burned:                             5245000,
					Counter:                            155670,
					Internal:                           true,
					Initiator:                          "tz1gXhGAXgKvrXjn4t16rYUXocqbch1XXJFN",
					Protocol:                           "PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA",
					AllocatedDestinationContractBurned: 257000,
					DeffatedStorage:                    []byte("{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},[]]},{\"int\":\"25168\"},{\"int\":\"25169\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Unit\"}]},{\"int\":\"25170\"}]},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"},{\"int\":\"0\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"25171\"}]},{\"int\":\"2\"},{\"string\":\"tz1QozfhaUW4wLnohDo6yiBUmh7cPCSXE9Af\"}]},{\"int\":\"11\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]}],{\"prim\":\"DROP\"},{\"prim\":\"NIL\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"PAIR\"}]]},{\"int\":\"500\"},{\"int\":\"1000\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"1000\"},{\"int\":\"2592000\"}]},{\"int\":\"1\"},{\"int\":\"1\"}]},[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bool\"},{\"prim\":\"True\"}]}],[{\"prim\":\"DROP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"nat\"},{\"int\":\"0\"}]}]]}"),
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
			opg := NewGroup(tt.ParseParams)
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
