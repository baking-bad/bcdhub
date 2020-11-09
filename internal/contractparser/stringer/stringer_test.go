package stringer

import (
	"reflect"
	"sort"
	"testing"
)

func TestGet(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		res   []string
	}{
		{
			name:  "opAh17oT2tV9bxguwpXzG5Mhvm33ZYkk2Jhmq4bGXeUqNMfE6V8/deffated_storage",
			input: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[[],[]]},{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"250"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1000000"},{"int":"1"}]},{"prim":"Pair","args":[[],{"int":"1614805920"}]}]}]}],[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1000000"},{"int":"0"}]},{"prim":"Pair","args":[[{"prim":"Elt","args":[{"bytes":"01d50e3f6f059dc86f5591455549313ce42d0c50f100"},{"prim":"False"}]}],{"string":"TRIANON ROYAL DE MUSIQUE DE VERSAILLES"}]}]}]},{"prim":"Pair","args":[{"bytes":"0000d27fcbd31910d2226ba4c8f646d3d4c7b2f3a756"},[{"prim":"Elt","args":[{"bytes":"01f8f6c6a0af7c20251bc7df108f2a6e2879a06c9a00"},{"prim":"False"}]}]]}]}`,
			res:   []string{"KT1U1JZaXoG4u1EPnhHL4R4otzkWc1L34q3c", "KT1XHAmdRKugP1Q38CxDmpcRSxq143KpEiYx", "TRIANON ROYAL DE MUSIQUE DE VERSAILLES", "tz1eq3gqb2iZHjHVHoPJqV84gZdBF2TMQiH4"},
		},
		{
			name:  "onugZkopb9EgucdiASTSCNiPcjrWuQVhB551ZSmyzyiFh7WVp8N/storage",
			input: `{"prim": "Pair","args": [{"int": "9"},{"prim": "Pair","args": [{"prim": "Pair","args": [{"bytes": "0000107c4009f2bcfcc248d6952998af5b7203b8ff59"},[{"prim": "Elt","args": [{"string": "By"},{"string": "https://SmartPy.io"}]},{"prim": "Elt","args": [{"string": "Help"},{"string": "Use Build to define a new game board and Play to make moves"}]},{"prim": "Elt","args": [{"string": "Play at"},{"string": "https://smartpy.io/demo/explore.html?address=KT1UvfyLytrt71jh63YV4Yex5SmbNXpWHxtg"}]},{"prim": "Elt","args": [{"string": "SmartPy Template"},{"string": "https://smartpy.io/demo/index.html?template=tictactoeFactory.py"}]}]]},{"prim": "False"}]}]}`,
			res:   []string{"By", "Help", "Play at", "SmartPy Template", "Use Build to define a new game board and Play to make moves", "https://SmartPy.io", "https://smartpy.io/demo/explore.html?address=KT1UvfyLytrt71jh63YV4Yex5SmbNXpWHxtg", "https://smartpy.io/demo/index.html?template=tictactoeFactory.py", "tz1M9CMEtsXm3QxA7FmMU2Qh7xzsuGXVbcDr"},
		},
		{
			name:  "opLoRjVhZ5mW3BjEbkLErzumtuwX5GqjRM7yRKFxLQF8dbUfqZJ/big_map_diff",
			input: `{"bytes": "0507070100000004636f6465010000000963616c6c5061757365"}`,
			res:   []string{"code", "callPause"},
		},
		{
			name:  "ooqmgMXibYMtdzcNZ4mfV15hyWwxY1MHjj44U8nLMvYvsAywAAx/big_map_diff/0",
			input: `{"bytes": "05010000000b746f74616c4d696e746564"}`,
			res:   []string{"totalMinted"},
		},
		{
			name:  "ooqmgMXibYMtdzcNZ4mfV15hyWwxY1MHjj44U8nLMvYvsAywAAx/big_map_diff/2",
			input: `{"bytes": "05070701000000066c65646765720a0000001600009472982d7f6b096bc57d6da95e0b8ec8ee37e72f"}`,
			res:   []string{"ledger", "tz1ZAwyfujwED4yUhQAtc1eqm4gW5u2Xiw77"},
		},
		{
			name:  "ooFJJPfNp2D4Tb2zT1JtqnFamtirAPMUPVfwYmHdVZtoScDUBii/storage",
			input: `{"prim": "Pair","args": [{"int": "31"},{"prim": "Pair","args": [[{"prim": "DUP"},{"prim": "CAR"},{"prim": "DIP","args": [[{"prim": "CDR"}]]},{"prim": "DUP"},{"prim": "DUP"},{"prim": "CAR"},{"prim": "DIP","args": [[{"prim": "CDR"}]]},{"prim": "DIP","args": [[{"prim": "DIP","args": [{"int": "2"},[{"prim": "DUP"}]]},{"prim": "DIG","args": [{"int": "2"}]}]]},{"prim": "PUSH","args": [{"prim": "string"},{"string": "code"}]},{"prim": "PAIR"},{"prim": "PACK"},{"prim": "GET"},{"prim": "IF_NONE","args": [[{"prim": "NONE","args": [{"prim": "lambda","args": [{"prim": "pair","args": [{"prim": "bytes"},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]},{"prim": "pair","args": [{"prim": "list","args": [{"prim": "operation"}]},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]}]}]}],[{"prim": "UNPACK","args": [{"prim": "lambda","args": [{"prim": "pair","args": [{"prim": "bytes"},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]},{"prim": "pair","args": [{"prim": "list","args": [{"prim": "operation"}]},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]}]}]},{"prim": "IF_NONE","args": [[{"prim": "PUSH","args": [{"prim": "string"},{"string": "UStore: failed to unpack code"}]},{"prim": "FAILWITH"}],[]]},{"prim": "SOME"}]]},{"prim": "IF_NONE","args": [[{"prim": "DROP"},{"prim": "DIP","args": [[{"prim": "DUP"},{"prim": "PUSH","args": [{"prim": "bytes"},{"bytes": "05010000000866616c6c6261636b"}]},{"prim": "GET"},{"prim": "IF_NONE","args": [[{"prim": "PUSH","args": [{"prim": "string"},{"string": "UStore: no field fallback"}]},{"prim": "FAILWITH"}],[]]},{"prim": "UNPACK","args": [{"prim": "lambda","args": [{"prim": "pair","args": [{"prim": "pair","args": [{"prim": "string"},{"prim": "bytes"}]},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]},{"prim": "pair","args": [{"prim": "list","args": [{"prim": "operation"}]},{"prim": "big_map","args": [{"prim": "bytes"},{"prim": "bytes"}]}]}]}]},{"prim": "IF_NONE","args": [[{"prim": "PUSH","args": [{"prim": "string"},{"string": "UStore: failed to unpack fallback"}]},{"prim": "FAILWITH"}],[]]},{"prim": "SWAP"}]]},{"prim": "PAIR"},{"prim": "EXEC"}],[{"prim": "DIP","args": [[{"prim": "SWAP"},{"prim": "DROP"},{"prim": "PAIR"}]]},{"prim": "SWAP"},{"prim": "EXEC"}]]}],{"prim": "Pair","args": [{"int": "1"},{"prim": "False"}]}]}]}`,
			res:   []string{"UStore: failed to unpack code", "UStore: failed to unpack fallback", "UStore: no field fallback", "code", "fallback"},
		},
		{
			name:  "onuCcYunPpSkbkNQVyknaCs6KQxMew1L1EAFTRAuQFZm5vViF7g/big_map_diff",
			input: `{"bytes": "05020000013f03210316051f02000000020317050d036e072f0200000029034f07430368010000001a55706172616d417267756d656e74556e7061636b4661696c6564034203270200000000034203210316051f02000000020317051f02000000af0321074303690a0000000b0501000000056f776e65720329072f02000000210743036801000000165553746f72653a206e6f206669656c64206f776e657203270200000000050d036e072f020000002907430368010000001e5553746f72653a206661696c656420746f20756e7061636b206f776e657203270200000000034803190325072c0200000000020000001f034f07430368010000001053656e64657249734e6f744f776e6572034203270346030c0346074303690a0000000e0501000000086e65774f776e65720350053d036d034203210316051f020000000203170342"}`,
			res:   []string{"SenderIsNotOwner", "UStore: failed to unpack owner", "UStore: no field owner", "UparamArgumentUnpackFailed", "newOwner", "owner"},
		},
		{
			name:  "onuCcYunPpSkbkNQVyknaCs6KQxMew1L1EAFTRAuQFZm5vViF7g/big_map_diff",
			input: `{"bytes": "05020000014203210316051f02000000020317050d036e072f0200000029034f07430368010000001a55706172616d417267756d656e74556e7061636b4661696c6564034203270200000000034203210316051f02000000020317051f02000000af0321074303690a0000000b0501000000056f776e65720329072f02000000210743036801000000165553746f72653a206e6f206669656c64206f776e657203270200000000050d036e072f020000002907430368010000001e5553746f72653a206661696c656420746f20756e7061636b206f776e657203270200000000034803190325072c0200000000020000001f034f07430368010000001053656e64657249734e6f744f776e657203420327030c0346074303690a0000001305010000000d72656465656d416464726573730350053d036d034203210316051f020000000203170342"}`,
			res:   []string{"SenderIsNotOwner", "UStore: failed to unpack owner", "UStore: no field owner", "UparamArgumentUnpackFailed", "owner", "redeemAddress"},
		},
		{
			name:  "onuCcYunPpSkbkNQVyknaCs6KQxMew1L1EAFTRAuQFZm5vViF7g/big_map_diff",
			input: `{"bytes": "05020000020103210316051f02000000020317050d0765055f0362036e072f0200000029034f07430368010000001a55706172616d417267756d656e74556e7061636b4661696c6564034203270200000000034203210316051f0200000002031703210316051f02000000020317051f020000004b0555055f07650362076503680765036807650362076003680368072f0200000025034f074303680100000016556e6578706563746564436f6e747261637454797065034203270200000000034203210316051f02000000020317051f020000000b051f02000000020321034c034203210316051f02000000020317051f02000000a8074303690a0000001305010000000d746f6b656e4d657461646174610329072f020000002907430368010000001e5553746f72653a206e6f206669656c6420746f6b656e4d6574616461746103270200000000050d07650362076503680765036807650362076003680368072f02000000310743036801000000265553746f72653a206661696c656420746f20756e7061636b20746f6b656e4d65746164617461032702000000000538020000003903300325072c02000000000200000027074307650368036c07070100000014496e76616c696453696e676c65546f6b656e4964030b03270321034c0320051f02000000020313034d053d036d034c031b034203210316051f020000000203170342"}`,
			res:   []string{"InvalidSingleTokenId", "UStore: failed to unpack tokenMetadata", "UStore: no field tokenMetadata", "UnexpectedContractType", "UparamArgumentUnpackFailed", "tokenMetadata"},
		},
		{
			name:  "onuCcYunPpSkbkNQVyknaCs6KQxMew1L1EAFTRAuQFZm5vViF7g/big_map_diff",
			input: `{"bytes": "05020000017f03210316051f02000000020317050d0765036c036e072f0200000029034f07430368010000001a55706172616d417267756d656e74556e7061636b4661696c6564034203270200000000034203210316051f0200000002031703210316051f02000000020317051f02000000350555036e072f0200000025034f074303680100000016556e6578706563746564436f6e747261637454797065034203270200000000034203210316051f02000000020317051f020000000b051f02000000020321034c03420317074303690a0000001305010000000d72656465656d416464726573730329072f020000002907430368010000001e5553746f72653a206e6f206669656c642072656465656d4164647265737303270200000000050d036e072f02000000310743036801000000265553746f72653a206661696c656420746f20756e7061636b2072656465656d4164647265737303270200000000051f02000000020313034d053d036d034c031b034203210316051f020000000203170342"}`,
			res:   []string{"UStore: failed to unpack redeemAddress", "UStore: no field redeemAddress", "UnexpectedContractType", "UparamArgumentUnpackFailed", "redeemAddress"},
		},
		{
			name:  "onuCcYunPpSkbkNQVyknaCs6KQxMew1L1EAFTRAuQFZm5vViF7g/big_map_diff",
			input: `{"bytes": "05020000025003210316051f02000000020317050d036c072f0200000029034f07430368010000001a55706172616d417267756d656e74556e7061636b4661696c6564034203270200000000034203210316032003170321074303690a0000000e0501000000086e65774f776e65720329072f02000000240743036801000000195553746f72653a206e6f206669656c64206e65774f776e657203270200000000050d0563036e072f020000002c0743036801000000215553746f72653a206661696c656420746f20756e7061636b206e65774f776e657203270200000000072f0200000029034f07430368010000001a4e6f74496e5472616e736665724f776e6572736869704d6f6465034203270200000034034803190325072c02000000000200000022034f07430368010000001353656e64657249734e6f744e65774f776e6572034203270321074303690a0000000e0501000000086e65774f776e65720329072f02000000240743036801000000195553746f72653a206e6f206669656c64206e65774f776e657203270200000000050d0563036e072f020000002c0743036801000000215553746f72653a206661696c656420746f20756e7061636b206e65774f776e657203270200000000072f0200000029034f07430368010000001a4e6f74496e5472616e736665724f776e6572736869704d6f646503420327020000003b030c0346074303690a0000000b0501000000056f776e65720350053e036e030c0346074303690a0000000e0501000000086e65774f776e65720350053d036d034203210316051f020000000203170342"}`,
			res:   []string{"NotInTransferOwnershipMode", "SenderIsNotNewOwner", "UStore: failed to unpack newOwner", "UStore: no field newOwner", "UparamArgumentUnpackFailed", "newOwner", "owner"},
		},
		{
			name:  "KT1ChNsEFxwyCbJyWGSL3KdjeXE28AY1Kaog/",
			input: `{"prim": "Pair","args": [{"prim": "Pair","args": [{"prim": "Some","args": [{"prim": "Pair","args": [{"prim": "Pair","args": [{"prim": "Pair","args": [{"bytes": "54657a6f7320566f7465"},{"prim": "True"}]},{"bytes": "68747470733a2f2f74657a6f732e766f74652f62616b65722e6a736f6e"}]},{"prim": "Pair","args": [{"prim": "Pair","args": [{"int": "9400"},[{"string": "tz1ggJg924JdWdww6oLGFksRgKctfrDMUNfE"}]]},{"prim": "Pair","args": [{"prim": "Pair","args": [{"prim": "Pair","args": [{"int": "500000"},{"prim": "True"}]},{"prim": "Pair","args": [{"int": "6"},{"prim": "Pair","args": [{"int": "1"},{"int": "200"}]}]}]},{"prim": "Pair","args": [{"prim": "Pair","args": [{"prim": "False"},{"int": "15503"}]},{"prim": "Pair","args": [{"int": "100"},{"prim": "True"}]}]}]}]}]}]},{"prim": "Some","args": [{"string": "tz1ggJg924JdWdww6oLGFksRgKctfrDMUNfE"}]}]},{"string": "2020-04-03T18:17:21Z"}]}`,
			res:   []string{"Tezos Vote", "https://tezos.vote/baker.json", "tz1ggJg924JdWdww6oLGFksRgKctfrDMUNfE"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.input)

			sort.Strings(result)
			sort.Strings(tt.res)

			if !reflect.DeepEqual(result, tt.res) {
				t.Errorf("Res didnt match.\nGot %#v\nExp %#v", result, tt.res)
			}
		})
	}
}
