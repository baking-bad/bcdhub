package decode

import (
	"testing"
)

func TestHexToMicheline(t *testing.T) {
	type testCase struct {
		name   string
		input  string
		result string
	}

	validTestCases := []testCase{
		testCase{
			name:   "Small int",
			input:  "0006",
			result: `{ "int": "6" }`,
		},
		testCase{
			name:   "Negative small int",
			input:  "0046",
			result: `{ "int": "-6" }`,
		},
		testCase{
			name:   "Medium int",
			input:  "00840e",
			result: `{ "int": "900" }`,
		},
		testCase{
			name:   "Negative medium int",
			input:  "00c40e",
			result: `{ "int": "-900" }`,
		},
		testCase{
			name:   "Large int",
			input:  "00ba9af7ea06",
			result: `{ "int": "917431994" }`,
		},
		testCase{
			name:   "Negative large int",
			input:  "00c0f9b9d4c723",
			result: `{ "int": "-610913435200" }`,
		},
		testCase{
			name:   "Negative large int from contract",
			input:  "00f7bda18081d73103c6da76151a15e6ab4b0bca788006481c2062b0df028ed8ac",
			result: `{ "int": "-109246922633079" }`,
		},
		testCase{
			name:   "String",
			input:  "01000000096d696368656c696e65",
			result: `{ "string": "micheline" }`,
		},
		testCase{
			name:   "Empty string",
			input:  "0100000000",
			result: `{ "string": "" }`,
		},
		testCase{
			name:   "Bytes",
			input:  "0a000000080123456789abcdef",
			result: `{ "bytes": "0123456789abcdef" }`,
		},
		testCase{
			name:   "Mixed literal array",
			input:  "02000000210061010000000574657a6f730100000000010000000b63727970746f6e6f6d6963",
			result: `[ { "int": "-33" }, { "string": "tezos" }, { "string": "" }, { "string": "cryptonomic" } ]`,
		},
		testCase{
			name:   "Single primitive",
			input:  "0343",
			result: `{ "prim": "PUSH" }`,
		},
		testCase{
			name:   "Single primitive with a single annotation",
			input:  "04430000000440636261",
			result: `{ "prim": "PUSH", "annots": [ "@cba" ] }`,
		},
		testCase{
			name:   "Single primitive with a single argument",
			input:  "053d036d",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" } ] }`,
		},
		testCase{
			name:   "Single primitive with a single argument and annotation",
			input:  "063d036d0000000440636261",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" } ], "annots": [ "@cba" ] }`,
		},
		testCase{
			name:   "Single primitive with two arguments",
			input:  "073d036d036d",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ] }`,
		},
		testCase{
			name:   "Single primitive with two arguments and annotation",
			input:  "083d036d036d0000000440636261",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@cba" ] }`,
		},
		testCase{
			name:   "Single primitive with more than two arguments and no annotations",
			input:  "093d00000006036d036d036d00000000",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ] }`,
		},
		testCase{
			name:   "Single primitive with more than two arguments and multiple annotations",
			input:  "093d00000006036d036d036d00000011407265642040677265656e2040626c7565",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@red", "@green", "@blue" ] }`,
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, err := HexToMicheline(tc.input)
			if err != nil {
				t.Errorf("HexToMicheline %v error. Input: %v", tc.name, tc.input)
			}

			if res != tc.result {
				t.Errorf("Input: %v. Got: %v, expected: %v.", tc.input, res, tc.result)
			}
		})
	}
}
