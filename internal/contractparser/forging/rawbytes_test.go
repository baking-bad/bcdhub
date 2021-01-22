package forging

import (
	"testing"
)

func TestUnforge(t *testing.T) {
	validTestCases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "Small int",
			input:  "0006",
			result: `{ "int": "6" }`,
		},
		{
			name:   "Negative small int",
			input:  "0046",
			result: `{ "int": "-6" }`,
		},
		{
			name:   "Medium int",
			input:  "00840e",
			result: `{ "int": "900" }`,
		},
		{
			name:   "Negative medium int",
			input:  "00c40e",
			result: `{ "int": "-900" }`,
		},
		{
			name:   "Large int",
			input:  "00ba9af7ea06",
			result: `{ "int": "917431994" }`,
		},
		{
			name:   "Negative large int",
			input:  "00c0f9b9d4c723",
			result: `{ "int": "-610913435200" }`,
		},
		{
			name:   "String",
			input:  "01000000096d696368656c696e65",
			result: `{ "string": "micheline" }`,
		},
		{
			name:   "Empty string",
			input:  "0100000000",
			result: `{ "string": "" }`,
		},
		{
			name:   "Bytes",
			input:  "0a000000080123456789abcdef",
			result: `{ "bytes": "0123456789abcdef" }`,
		},
		{
			name:   "Mixed literal array",
			input:  "02000000210061010000000574657a6f730100000000010000000b63727970746f6e6f6d6963",
			result: `[ { "int": "-33" }, { "string": "tezos" }, { "string": "" }, { "string": "cryptonomic" } ]`,
		},
		{
			name:   "Single primitive",
			input:  "0343",
			result: `{ "prim": "PUSH" }`,
		},
		{
			name:   "Single primitive with a single annotation",
			input:  "04430000000440636261",
			result: `{ "prim": "PUSH", "annots": [ "@cba" ] }`,
		},
		{
			name:   "Single primitive with a single argument",
			input:  "053d036d",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" } ] }`,
		},
		{
			name:   "Single primitive with a single argument and annotation",
			input:  "063d036d0000000440636261",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" } ], "annots": [ "@cba" ] }`,
		},
		{
			name:   "Single primitive with two arguments",
			input:  "073d036d036d",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ] }`,
		},
		{
			name:   "Single primitive with two arguments and annotation",
			input:  "083d036d036d0000000440636261",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@cba" ] }`,
		},
		{
			name:   "Single primitive with more than two arguments and no annotations",
			input:  "093d00000006036d036d036d00000000",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ] }`,
		},
		{
			name:   "Single primitive with more than two arguments and multiple annotations",
			input:  "093d00000006036d036d036d00000011407265642040677265656e2040626c7565",
			result: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@red", "@green", "@blue" ] }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000561646d696e",
			result: `{ "string": "admin" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			result: `{ "string": "tz1LFEVYR7YRCxT6Nm3Zfjdnfj77xZqhbR5U" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0100000006706175736564",
			result: `{ "string": "paused" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0303",
			result: `{ "prim": "False" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000866616c6c6261636b",
			result: `{ "string": "fallback" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "02000000270316031607430368010000001655706172616d4e6f53756368456e747279506f696e7403420327",
			result: `[ { "prim": "CAR" }, { "prim": "CAR" }, { "prim": "PUSH", "args": [ { "prim": "string" }, { "string": "UparamNoSuchEntryPoint" } ] }, { "prim": "PAIR" }, { "prim": "FAILWITH" } ]`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "01000000086e65774f776e6572",
			result: `{ "string": "newOwner" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0306",
			result: `{ "prim": "None" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "01000000096f70657261746f7273",
			result: `{ "string": "operators" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0200000000",
			result: `[]`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0100000009746f6b656e636f6465",
			result: `{ "string": "tokencode" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0100000005545a425443",
			result: `{ "string": "TZBTC" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0100000009746f6b656e6e616d65",
			result: `{ "string": "tokenname" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000b746f74616c4275726e6564",
			result: `{ "string": "totalBurned" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0000",
			result: `{ "int": "0" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000b746f74616c4d696e746564",
			result: `{ "string": "totalMinted" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000b746f74616c537570706c79",
			result: `{ "string": "totalSupply" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "010000000d72656465656d41646472657373",
			result: `{ "string": "redeemAddress" }`,
		},
		{
			name:   "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:  "0a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			result: `{ "string": "tz1LFEVYR7YRCxT6Nm3Zfjdnfj77xZqhbR5U" }`,
		},
		{
			name:   "prim",
			input:  "070700000000",
			result: `{ "prim": "Pair", "args": [ { "int": "0" }, { "int": "0" } ] }`,
		},
		{
			name:   "edo prim",
			input:  "0707000007070000070700000000",
			result: `{ "prim": "Pair", "args": [ { "int": "0" }, { "prim": "Pair", "args": [ { "int": "0" }, { "prim": "Pair", "args": [ { "int": "0" }, { "int": "0" } ] } ] } ] }`,
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Unforge(tc.input)
			if err != nil {
				t.Errorf("Input: %v. Error: %v.", tc.input, err)
			}
			if res != tc.result {
				t.Errorf("Input: %v. Got: %v, expected: %v.", tc.input, res, tc.result)
			}
		})
	}
}

func TestToMichelineErrors(t *testing.T) {
	errorTestCases := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "offset error",
			input: "00f7bda18081d73103c6da76151a15e6ab4b0bca788006481c2062b0df028ed8ac",
			err:   "input is not empty",
		},
		{
			name:  "prim keyword error",
			input: "03fe636820fd6e933da2de8a131df78c56ba64d1574749db4a20c3eb03a126b3",
			err:   "Invalid data: [prim] invalid prim keyword fe",
		},
	}

	for _, tc := range errorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Unforge(tc.input)
			if err == nil {
				t.Errorf("Empty error. Input: %v.", tc.input)
				return
			}
			if err.Error() != tc.err {
				t.Errorf("Input: %v. Got: %v, expected: %v.", tc.input, err, tc.err)
			}
		})
	}
}
