package operations

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/golang/mock/gomock"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlES := gomock.NewController(t)
	defer ctrlES.Finish()
	es := elastic.NewMockIElastic(ctrlES)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	tests := []struct {
		name        string
		ParseParams *ParseParams
		filename    string
		want        []elastic.Model
		wantErr     bool
	}{
		{
			name: "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ParseParams: &ParseParams{
				es:  es,
				rpc: rpc,
			},
			filename: "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:     []elastic.Model{},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ParseParams: &ParseParams{
				es:  es,
				rpc: rpc,
				head: noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     1151495,
				},
				network: "mainnet",
			},
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
					Nonce:            setInt64(1),
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			metadata := &models.Metadata{
				ID: "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			}
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

			var filters map[string]interface{}
			es.
				EXPECT().
				GetContract(gomock.AssignableToTypeOf(filters)).
				DoAndReturn(
					func(args map[string]interface{}) (models.Contract, error) {
						return readTestContractModel(args["address"].(string))
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
				t.Errorf("Group.Parse() = %##v, want %##v", got, tt.want)
			}
		})
	}
}
