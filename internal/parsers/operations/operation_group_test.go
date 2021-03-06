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

	generalRepo.
		EXPECT().
		GetByID(gomock.AssignableToTypeOf(&modelContract.Contract{})).
		DoAndReturn(readTestContractModel).
		AnyTimes()

	generalRepo.
		EXPECT().
		BulkInsert(gomock.AssignableToTypeOf([]models.Model{})).
		Return(nil).
		AnyTimes()

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2416))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:          2416,
				Key:          []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
				Value:        []byte(`{"prim":"Pair","args":[[],{"int":"6000"}]}`),
				ValueStrings: []string{},
				Level:        386026,
				Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
				Network:      "carthagenet",
				Timestamp:    timestamp,
				Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2417))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:          2417,
				Key:          []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
				KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
				Value:        nil,
				ValueStrings: []string{},
				Level:        386026,
				Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
				Network:      "carthagenet",
				Timestamp:    timestamp,
				Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	tests := []struct {
		name        string
		ParseParams *ParseParams
		filename    string
		address     string
		level       int64
		want        []models.Model
		wantErr     bool
	}{
		{
			name:        "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ParseParams: NewParseParams(rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, tbRepo),
			filename:    "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:        []models.Model{},
		}, {
			name: "opJXaAMkBrAbd1XFd23kS8vXiw63tU4rLUcLrZgqUCpCbhT1Pn9",
			ParseParams: NewParseParams(rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
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
					Parameters:      "{\"entrypoint\":\"default\",\"value\":{\"prim\":\"Right\",\"args\":[{\"prim\":\"Left\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Right\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1aSPEN4RTZbn4aXEsxDiix38dDmacGQ8sq\"},{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1invbJv3AEm55ct7QF2dVbWZuaDekssYkV\"},{\"int\":\"8010000\"}]}]}]}]}]}]}}",
					DeffatedStorage: "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}],{\"int\":\"62\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"63\"},{\"string\":\"Aspen Digital Token\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada700\"},{\"string\":\"ASPD\"}]},{\"int\":\"18000000\"}]}]}",
					Tags:            []string{"fa1-2"},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:       63,
					KeyHash:   "exprum2qtFLPHdeLWVasKCDw7YD5MrdiD4ra52PY2AUazaNGKyv6tx",
					Key:       []byte(`{"bytes":"0000a2560a416161def96031630886abe950c4baf036"}`),
					Value:     []byte(`{"int":"6141000"}`),
					Level:     1068669,
					Network:   "mainnet",
					Address:   "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
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
					Address:   "KT1S5iPRQ612wcNm6mXDqDhTNegGFcvTV7vM",
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
					Parameters:      "{\"entrypoint\":\"validateAccounts\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"True\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]},{\"bytes\":\"01796ad78734892d5ae4186e84a30290040732ada70076616c696461746552756c6573\"}]}}",
					DeffatedStorage: "{\"int\":\"61\"}",
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
					Parameters:      "{\"entrypoint\":\"validateRules\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"}]}]},{\"int\":\"2\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"string\":\"US\"}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"False\"},{\"bytes\":\"0000c644b537bdb0dac40fe742010106546effd69395\"}]}]},{\"int\":\"6\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000a2560a416161def96031630886abe950c4baf036\"},{\"bytes\":\"0000fdf98b65d53a9661e07f41093dcb6f3d931736ba\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"int\":\"14151000\"},{\"int\":\"0\"}]},{\"prim\":\"True\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"01bff38c4e363eacef338f7b2e15f00ca42fafa1ce00\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"8010000\"},{\"int\":\"18000000\"}]}]}]}}",
					DeffatedStorage: "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000056d8b91b541c9d20d51f929dcccca2f14928f1dc\"},{\"bytes\":\"010d25f77b84dc2164a5d1ce5e8a5d3ca2b1d0cbf900\"}]},[]]}",
					Tags:            []string{},
				},
			},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ParseParams: NewParseParams(
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
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
			address:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			level:    1151495,
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:     0,
					Network:          consts.Mainnet,
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         false,
					Nonce:            nil,
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            1151495,
					Kind:             "transaction",
					Initiator:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Source:           "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Fee:              43074,
					Counter:          6909186,
					GasLimit:         427673,
					StorageLimit:     47,
					Destination:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Parameters:       "{\"entrypoint\":\"redeem\",\"value\":{\"bytes\":\"a874aac22777351417c9bde0920cc7ed33e54453e1dd149a1f3a60521358d19a\"}}",
					Entrypoint:       "redeem",
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"int\":\"32\"},{\"prim\":\"Unit\"}]}",
					ParameterStrings: []string{},
					StorageStrings:   []string{},
					Tags:             []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          32,
					Key:          []byte(`{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"}`),
					KeyHash:      "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
					KeyStrings:   nil,
					ValueStrings: nil,
					OperationID:  "f79b897e69e64aa9b6d7f0199fed08f9",
					Level:        1151495,
					Address:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979843131,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&operation.Operation{
					ContentIndex:     0,
					Network:          consts.Mainnet,
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         true,
					Nonce:            setInt64(0),
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            1151495,
					Kind:             "transaction",
					Initiator:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Source:           "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Counter:          6909186,
					Destination:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Parameters:       "{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}",
					Entrypoint:       "transfer",
					Burned:           47000,
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"int\":\"31\"},{\"prim\":\"Pair\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DUP\"},{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DIP\",\"args\":[{\"int\":\"2\"},[{\"prim\":\"DUP\"}]]},{\"prim\":\"DIG\",\"args\":[{\"int\":\"2\"}]}]]},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"code\"}]},{\"prim\":\"PAIR\"},{\"prim\":\"PACK\"},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"NONE\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]}],[{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack code\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SOME\"}]]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"DROP\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bytes\"},{\"bytes\":\"05010000000866616c6c6261636b\"}]},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: no field fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SWAP\"}]]},{\"prim\":\"PAIR\"},{\"prim\":\"EXEC\"}],[{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"SWAP\"},{\"prim\":\"DROP\"},{\"prim\":\"PAIR\"}]]},{\"prim\":\"SWAP\"},{\"prim\":\"EXEC\"}]]}],{\"prim\":\"Pair\",\"args\":[{\"int\":\"1\"},{\"prim\":\"False\"}]}]}]}",
					ParameterStrings: nil,
					StorageStrings:   nil,
					Tags:             []string{"fa1-2"},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					Key:          []byte(`{"bytes": "05010000000b746f74616c537570706c79"}`),
					KeyHash:      "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
					KeyStrings:   nil,
					Value:        []byte(`{"bytes":"050098e1e8d78a02"}`),
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979845825,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					Key:          []byte(`{"bytes": "05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"}`),
					KeyHash:      "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
					KeyStrings:   nil,
					Value:        []byte(`{"bytes":"0507070080a5c1070200000000"}`),
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495, Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:     consts.Mainnet,
					IndexedTime: 1602764979845832,
					Timestamp:   timestamp,
					Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					Key:          []byte(`{"bytes": "05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"}`),
					KeyHash:      "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
					KeyStrings:   nil,
					Value:        []byte(`{"bytes":"05070700ba81bb090200000000"}`),
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979845839,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
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
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
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
			address:  "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
			level:    86142,
			filename: "./data/rpc/opg/onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt.json",
			want: []models.Model{
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
					DeffatedStorage:                    "{\"int\":\"0\"}\n",
					ParameterStrings:                   nil,
					StorageStrings:                     nil,
					Tags:                               nil,
				},
				&modelContract.Contract{
					Network:     "delphinet",
					Level:       86142,
					Timestamp:   timestamp,
					Language:    "unknown",
					Hash:        "e4b88b53b9227b3fc4fc0dbe148f249a7a1c755cf4cbc9c8fb5b5b78395a139d3f8e0fde5c27117df30553e98ecb4e3e8ddc9740292af18fbf36326cb55cebad",
					Tags:        []string{},
					Hardcoded:   []string{},
					FailStrings: []string{},
					Annotations: []string{"%decrement", "%increment"},
					Entrypoints: []string{"decrement", "increment"},
					Address:     "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
					Manager:     "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
				},
			},
		}, {
			name: "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
			ParseParams: NewParseParams(
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     386026,
					ChainID:   "test",
				}),
				WithNetwork("carthagenet"),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
			),
			address:  "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
			level:    386026,
			filename: "./data/rpc/opg/opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:     0,
					Network:          "carthagenet",
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:         false,
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            386026,
					Kind:             "transaction",
					Initiator:        "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:           "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Fee:              62628,
					Counter:          554732,
					GasLimit:         622830,
					StorageLimit:     154,
					Amount:           0,
					Destination:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Parameters:       "{\"entrypoint\":\"mint\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1XvMBRHwmXtXS2K6XYZdmcc5kdwB9STFJu\"},{\"int\":\"8500\"}]}}",
					Entrypoint:       "mint",
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c67479d5c0961a0fcac5c13a1a94b56a37236e98\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2417\"},{\"int\":\"2\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"015e691d7d78f7738e78e92379b54979738846e2ea00\"}]},{\"bytes\":\"01440ed695906addcb6aa4681c816824db9562475700\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"TezosTkNext\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2415\"},{\"prim\":\"False\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"string\":\"euroTz\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"6000\"},[]]}]}]}]}",
					ParameterStrings: []string{},
					StorageStrings:   []string{},
					Tags:             []string{},
				},
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(0),
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              386026,
					Kind:                               "transaction",
					Initiator:                          "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:                             "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Fee:                                0,
					Counter:                            554732,
					GasLimit:                           0,
					StorageLimit:                       0,
					Amount:                             0,
					Destination:                        "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Parameters:                         "{\"entrypoint\":\"mint\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000086b7990605548cb13db091c7a68a46a7aef3d0a2\"},{\"int\":\"8500\"}]}}",
					Entrypoint:                         "mint",
					Burned:                             77000,
					AllocatedDestinationContractBurned: 0,
					DeffatedStorage:                    "{\"prim\":\"Pair\",\"args\":[{\"int\":\"2416\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"013718908e90796befd5f7e1fa7312e6acc12314e500\"}]},{\"int\":\"14500\"}]}]}",
					ParameterStrings:                   []string{},
					StorageStrings:                     []string{},
					Tags:                               []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2416,
					Key:          []byte(`{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"}`),
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        []byte(`"{"prim":"Pair","args":[[],{"int":"8500"}]}"`),
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(1),
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              386026,
					Kind:                               "transaction",
					Initiator:                          "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:                             "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Fee:                                0,
					Counter:                            554732,
					GasLimit:                           0,
					StorageLimit:                       0,
					Amount:                             0,
					Destination:                        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Parameters:                         "{\"entrypoint\":\"receiveDataFromStandardSC\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"int\":\"-1\"},{\"int\":\"14500\"}]}}",
					Entrypoint:                         "receiveDataFromStandardSC",
					Burned:                             77000,
					AllocatedDestinationContractBurned: 0,
					DeffatedStorage:                    "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c67479d5c0961a0fcac5c13a1a94b56a37236e98\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2418\"},{\"int\":\"2\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"015e691d7d78f7738e78e92379b54979738846e2ea00\"}]},{\"bytes\":\"01440ed695906addcb6aa4681c816824db9562475700\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"TezosTkNext\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2415\"},{\"prim\":\"False\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"string\":\"euroTz\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14500\"},[]]}]}]}]}",
					ParameterStrings:                   []string{},
					StorageStrings:                     []string{},
					Tags:                               []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2417,
					Key:          []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        nil,
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapaction.BigMapAction{
					Action:         "remove",
					SourcePtr:      setInt64(2417),
					DestinationPtr: nil,
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(2416),
					DestinationPtr: setInt64(2418),
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2418,
					Key:          []byte(`{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"}`),
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        []byte(`{"prim":"Pair","args":[[],{"int":"6000"}]}`),
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2418,
					Key:          []byte(`{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"}`),
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        []byte(`"{"prim":"Pair","args":[[],{"int":"8500"}]}"`),
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc.
				EXPECT().
				GetScriptStorageRaw(tt.address, tt.level).
				DoAndReturn(
					func(address string, level int64) ([]byte, error) {
						storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
						return ioutil.ReadFile(storageFile)
					},
				).
				AnyTimes()

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
