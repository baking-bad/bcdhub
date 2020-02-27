package unpack

import (
	"testing"
)

func TestBytes(t *testing.T) {
	validTestCases := []struct {
		name    string
		input   string
		jsonstr string
		result  string
	}{
		{
			name:    "Small int",
			input:   "050006",
			jsonstr: `{ "int": "6" }`,
			result:  `6`,
		},
		{
			name:    "Negative small int",
			input:   "050046",
			jsonstr: `{ "int": "-6" }`,
			result:  `-6`,
		},
		{
			name:    "Medium int",
			input:   "0500840e",
			jsonstr: `{ "int": "900" }`,
			result:  `900`,
		},
		{
			name:    "Negative medium int",
			input:   "0500c40e",
			jsonstr: `{ "int": "-900" }`,
			result:  `-900`,
		},
		{
			name:    "Large int",
			input:   "0500ba9af7ea06",
			jsonstr: `{ "int": "917431994" }`,
			result:  `917431994`,
		},
		{
			name:    "Negative large int",
			input:   "0500c0f9b9d4c723",
			jsonstr: `{ "int": "-610913435200" }`,
			result:  `-610913435200`,
		},
		{
			name:    "String",
			input:   "0501000000096d696368656c696e65",
			jsonstr: `{ "string": "micheline" }`,
			result:  `"micheline"`,
		},
		{
			name:    "Empty string",
			input:   "050100000000",
			jsonstr: `{ "string": "" }`,
			result:  `""`,
		},
		{
			name:    "Bytes",
			input:   "050a000000080123456789abcdef",
			jsonstr: `{ "bytes": "0123456789abcdef" }`,
			result:  "0x0123456789abcdef",
		},
		{
			name:    "Mixed literal array",
			input:   "0502000000210061010000000574657a6f730100000000010000000b63727970746f6e6f6d6963",
			jsonstr: `[ { "int": "-33" }, { "string": "tezos" }, { "string": "" }, { "string": "cryptonomic" } ]`,
			result:  `{ -33 ; "tezos" ; "" ; "cryptonomic" }`,
		},
		{
			name:    "Single primitive",
			input:   "050343",
			jsonstr: `{ "prim": "PUSH" }`,
			result:  "PUSH",
		},
		{
			name:    "Single primitive with a single annotation",
			input:   "0504430000000440636261",
			jsonstr: `{ "prim": "PUSH", "annots": [ "@cba" ] }`,
			result:  "PUSH @cba",
		},
		{
			name:    "Single primitive with a single argument",
			input:   "05053d036d",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" } ] }`,
			result:  "NIL operation",
		},
		{
			name:    "Single primitive with a single argument and annotation",
			input:   "05063d036d0000000440636261",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" } ], "annots": [ "@cba" ] }`,
			result:  "NIL @cba operation",
		},
		{
			name:    "Single primitive with two arguments",
			input:   "05073d036d036d",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ] }`,
			result:  "NIL operation operation",
		},
		{
			name:    "Single primitive with two arguments and annotation",
			input:   "05083d036d036d0000000440636261",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@cba" ] }`,
			result:  "NIL @cba operation operation",
		},
		{
			name:    "Single primitive with more than two arguments and no annotations",
			input:   "05093d00000006036d036d036d00000000",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ] }`,
			result:  "NIL operation operation operation",
		},
		{
			name:    "Single primitive with more than two arguments and multiple annotations",
			input:   "05093d00000006036d036d036d00000011407265642040677265656e2040626c7565",
			jsonstr: `{ "prim": "NIL", "args": [ { "prim": "operation" }, { "prim": "operation" }, { "prim": "operation" } ], "annots": [ "@red", "@green", "@blue" ] }`,
			result:  "NIL @red @green @blue operation operation operation",
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000561646d696e",
			jsonstr: `{ "string": "admin" }`,
			result:  `"admin"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			jsonstr: `{ "bytes": "000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9" }`,
			result:  "0x000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050100000006706175736564",
			jsonstr: `{ "string": "paused" }`,
			result:  `"paused"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050303",
			jsonstr: `{ "prim": "False" }`,
			result:  "False",
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000866616c6c6261636b",
			jsonstr: `{ "string": "fallback" }`,
			result:  `"fallback"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "0502000000270316031607430368010000001655706172616d4e6f53756368456e747279506f696e7403420327",
			jsonstr: `[ { "prim": "CAR" }, { "prim": "CAR" }, { "prim": "PUSH", "args": [ { "prim": "string" }, { "string": "UparamNoSuchEntryPoint" } ] }, { "prim": "PAIR" }, { "prim": "FAILWITH" } ]`,
			result:  `{ CAR ; CAR ; PUSH string "UparamNoSuchEntryPoint" ; PAIR ; FAILWITH }`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "0501000000086e65774f776e6572",
			jsonstr: `{ "string": "newOwner" }`,
			result:  `"newOwner"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050306",
			jsonstr: `{ "prim": "None" }`,
			result:  "None",
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "0501000000096f70657261746f7273",
			jsonstr: `{ "string": "operators" }`,
			result:  `"operators"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050200000000",
			jsonstr: `[]`,
			result:  `{}`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050100000009746f6b656e636f6465",
			jsonstr: `{ "string": "tokencode" }`,
			result:  `"tokencode"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050100000005545a425443",
			jsonstr: `{ "string": "TZBTC" }`,
			result:  `"TZBTC"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050100000009746f6b656e6e616d65",
			jsonstr: `{ "string": "tokenname" }`,
			result:  `"tokenname"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000b746f74616c4275726e6564",
			jsonstr: `{ "string": "totalBurned" }`,
			result:  `"totalBurned"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050000",
			jsonstr: `{ "int": "0" }`,
			result:  "0",
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000b746f74616c4d696e746564",
			jsonstr: `{ "string": "totalMinted" }`,
			result:  `"totalMinted"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000b746f74616c537570706c79",
			jsonstr: `{ "string": "totalSupply" }`,
			result:  `"totalSupply"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "05010000000d72656465656d41646472657373",
			jsonstr: `{ "string": "redeemAddress" }`,
			result:  `"redeemAddress"`,
		},
		{
			name:    "KT1FgscaMyhxoVLbVirJVVKpRXgiSGtDG9Z4",
			input:   "050a00000016000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			jsonstr: `{ "bytes": "000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9" }`,
			result:  "0x000006a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			res := Bytes(tc.input)
			if res != tc.result {
				t.Errorf("\nInput: %v. \nGot: %v, \nexpected: %v.", tc.input, res, tc.result)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	validTestCases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "KT address",
			input:  "011fb03e3ff9fedaf3a2200ffc64d27812da734bba00",
			result: `KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A`,
		},
		{
			name:   "tz1 address",
			input:  "00009e6ac2e529a49aedbcdd0ac9542d5c0f4ce76f77",
			result: `tz1a5fMLLY5WCarCzH7RKTJHX9mJFN8eaaWG`,
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Address(tc.input)
			if err != nil {
				t.Errorf("error: %v", err)
				return
			}
			if res != tc.result {
				t.Errorf("\nInput: %v. \nGot: %v, \nexpected: %v.", tc.input, res, tc.result)
			}
		})
	}
}
