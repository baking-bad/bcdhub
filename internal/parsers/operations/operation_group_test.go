package operations

import (
	"fmt"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	mock_elastic "github.com/baking-bad/bcdhub/internal/elastic/mock"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/golang/mock/gomock"
	"github.com/tidwall/gjson"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlES := gomock.NewController(t)
	defer ctrlES.Finish()
	es := mock_elastic.NewMockIElastic(ctrlES)

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

	es.
		EXPECT().
		GetTZIPWithViews().
		Return(make([]models.TZIP, 0), nil).
		AnyTimes()
	es.
		EXPECT().
		UpdateTokenBalances(gomock.Any()).
		Return(nil).
		AnyTimes()

	tests := []struct {
		name        string
		ParseParams *ParseParams
		filename    string
		address     string
		level       int64
		want        []elastic.Model
		wantErr     bool
	}{
		{
			name:        "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ParseParams: NewParseParams(rpc, es),
			filename:    "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:        []elastic.Model{},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ParseParams: NewParseParams(
				rpc, es,
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     1151495,
				}),
				WithNetwork("mainnet"),
				WithConstants(models.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			address:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			level:    1151495,
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: []elastic.Model{
				&models.Operation{
					ContentIndex:     0,
					Network:          "mainnet",
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         false,
					Nonce:            nil,
					Status:           "applied",
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
				&models.BigMapDiff{
					Ptr:          32,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"},
					KeyHash:      "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
					KeyStrings:   nil,
					ValueStrings: nil,
					OperationID:  "f79b897e69e64aa9b6d7f0199fed08f9",
					Level:        1151495,
					Address:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Network:      "mainnet",
					IndexedTime:  1602764979843131,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.Operation{
					ContentIndex:     0,
					Network:          "mainnet",
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         true,
					Nonce:            setInt64(0),
					Status:           "applied",
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
					Tags:             []string{"fa12"},
				},
				&models.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05010000000b746f74616c537570706c79"},
					KeyHash:      "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
					KeyStrings:   nil,
					Value:        `{"bytes":"050098e1e8d78a02"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      "mainnet",
					IndexedTime:  1602764979845825,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"},
					KeyHash:      "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
					KeyStrings:   nil,
					Value:        `{"bytes":"0507070080a5c1070200000000"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495, Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:     "mainnet",
					IndexedTime: 1602764979845832,
					Timestamp:   timestamp,
					Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"},
					KeyHash:      "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
					KeyStrings:   nil,
					Value:        `{"bytes":"05070700ba81bb090200000000"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      "mainnet",
					IndexedTime:  1602764979845839,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.Transfer{
					Network:   "mainnet",
					Contract:  "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Initiator: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Hash:      "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Status:    "applied",
					Timestamp: timestamp,
					Level:     1151495,
					From:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					To:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					TokenID:   0,
					Amount:    7.87488e+06,
					Counter:   6909186,
					Nonce:     setInt64(0),
				},
			},
		}, {
			name: "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
			ParseParams: NewParseParams(
				rpc, es,
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     86142,
				}),
				WithNetwork("delphinet"),
				WithConstants(models.Constants{
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
			want: []elastic.Model{
				&models.Operation{
					ContentIndex:                       0,
					Network:                            "delphinet",
					Protocol:                           "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Hash:                               "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
					Internal:                           false,
					Status:                             "applied",
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
				&models.Metadata{
					ID:        "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
					Parameter: map[string]string{"babylon": "{\"0\":{\"prim\":\"or\",\"args\":[\"0/0\",\"0/1\"],\"type\":\"namedunion\"},\"0/0\":{\"fieldname\":\"decrement\",\"prim\":\"int\",\"type\":\"int\",\"name\":\"decrement\"},\"0/1\":{\"fieldname\":\"increment\",\"prim\":\"int\",\"type\":\"int\",\"name\":\"increment\"}}"},
					Storage:   map[string]string{"babylon": "{\"0\":{\"prim\":\"int\",\"type\":\"int\"}}"},
				},
				&models.Contract{
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
				rpc, es,
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     386026,
				}),
				WithNetwork("carthagenet"),
				WithConstants(models.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
			),
			address:  "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
			level:    386026,
			filename: "./data/rpc/opg/opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x.json",
			want: []elastic.Model{
				&models.Operation{
					ContentIndex:     0,
					Network:          "carthagenet",
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:         false,
					Status:           "applied",
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
				&models.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(0),
					Status:                             "applied",
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
				&models.BigMapDiff{
					Ptr:          2416,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"},
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"8500\"}]}",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(1),
					Status:                             "applied",
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
				&models.BigMapDiff{
					Ptr:          2417,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        "",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.BigMapAction{
					Action:         "remove",
					SourcePtr:      setInt64(2417),
					DestinationPtr: nil,
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&models.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(2416),
					DestinationPtr: setInt64(2418),
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&models.BigMapDiff{
					Ptr:          2418,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"6000\"}]}",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&models.BigMapDiff{
					Ptr:          2418,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"},
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"8500\"}]}",
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
			metadata := &models.Metadata{ID: tt.address}
			es.
				EXPECT().
				GetByID(gomock.AssignableToTypeOf(metadata)).
				DoAndReturn(
					func(val *models.Metadata) error {
						buf, err := readTestMetadataModel(val.GetID())
						if err != nil {
							return err
						}
						val.Parameter = buf.Parameter
						val.Storage = buf.Storage
						return nil
					},
				).
				AnyTimes()

			es.
				EXPECT().
				GetByID(gomock.AssignableToTypeOf(&models.Contract{})).
				DoAndReturn(readTestContractModel).
				AnyTimes()

			es.
				EXPECT().
				GetBigMapDiffsByPtr(
					gomock.Eq("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU"),
					gomock.Eq("carthagenet"),
					gomock.Eq(int64(2416))).
				Return([]models.BigMapDiff{
					{
						Ptr:          2416,
						BinPath:      "0/0/0/1/0",
						Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
						KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
						KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
						Value:        `{"prim":"Pair","args":[[],{"int":"6000"}]}`,
						ValueStrings: []string{},
						Level:        386026,
						Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
						Network:      "carthagenet",
						Timestamp:    timestamp,
						Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					},
				}, nil).
				AnyTimes()

			es.
				EXPECT().
				GetBigMapDiffsByPtr(
					gomock.Eq("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ"),
					gomock.Eq("carthagenet"),
					gomock.Eq(int64(2417))).
				Return([]models.BigMapDiff{
					{
						Ptr:          2417,
						BinPath:      "0/0/0/1/0",
						Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
						KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
						KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
						Value:        "",
						ValueStrings: []string{},
						Level:        386026,
						Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
						Network:      "carthagenet",
						Timestamp:    timestamp,
						Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					},
				}, nil).
				AnyTimes()

			rpc.
				EXPECT().
				GetScriptStorageJSON(tt.address, tt.level).
				DoAndReturn(
					func(address string, level int64) (gjson.Result, error) {
						storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
						return readJSONFile(storageFile)
					},
				).
				AnyTimes()

			data, err := readJSONFile(tt.filename)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			opg := NewGroup(tt.ParseParams)
			got, err := opg.Parse(data)
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
