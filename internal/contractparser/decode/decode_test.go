package decode

import (
	"testing"
)

func TestPublicKey(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "ed25519 public key",
			input:  "004e4ca2abb4baeed702a0ac5b0de9b5607dd1fedb399c0ce25e15b3868f67269e",
			result: "edpkuEhzJqdFBCWMw6TU3deADRK2fq3GuwWFUphwyH7ero1Na4oGFP",
		},
		{
			name:   "secp256k1 public key",
			input:  "01030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			result: "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
		},
		{
			name:   "p256 public key",
			input:  "02031a3ad5ea94de6912f9bc83fd31de49816e90602c5252d77b5b233bfe711b0dd2",
			result: "p2pk66iTZwLmRPshQgUr2HE3RUzSFwAN5MNaBQ5rfduT1dGKXd25pNN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := PublicKey(tt.input)
			if err != nil {
				t.Errorf("Error in PublicKey. Input: %v. Error: %v", tt.input, err)
			}
			if res != tt.result {
				t.Errorf("Error in PublicKey. Input: %v. Got: %v. Expected: %v.", tt.input, res, tt.result)
			}
		})
	}
}

func TestKeyHash(t *testing.T) {
	tests := []struct {
		input  string
		result string
	}{
		{
			input:  "0010fc2282886d9cf8a1eebdc2733e302c7b110f38",
			result: "tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS",
		},
		{
			input:  "003c8c2fe0f75ce212558df94c7a7306c2eeadd979",
			result: "tz1RABAzdLWVvxAFf1wpeUALAkp32mVhSGXX",
		},
		{
			input:  "004bf0acca4cc9e034b1d5f0f783c78e5ed44d866e",
			result: "tz1SZZgtvMVXaBKPcez4gfjKUsDz1gs6vg6X",
		},
		{
			input:  "0079e68d8f0a8d64ec856e193efc0a347ef4adf8ee",
			result: "tz1WkaeRycRr999GrVFepJd9Nqi1FWqGyGqq",
		},
		{
			input:  "01028562fb176188114cf437a757cdc75bc4aa8cae",
			result: "tz28YZoayJjVz2bRgGeVjxE8NonMiJ3r2Wdu",
		},
		{
			input:  "029d6a61cd3510193e257128da8f09a0b173bff695",
			result: "tz3agP9LGe2cXmKQyYn6T68BHKjjktDbbSWX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := KeyHash(tt.input)

			if err != nil {
				t.Errorf("Error in KeyHash. Error: %v", err)
			}

			if res != tt.result {
				t.Errorf("Error in Keyhash. Got: %v. Expected: %v", res, tt.result)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	tests := []struct {
		address string
		result  string
	}{
		{
			address: "000010fc2282886d9cf8a1eebdc2733e302c7b110f38",
			result:  "tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS",
		},
		{
			address: "00003c8c2fe0f75ce212558df94c7a7306c2eeadd979",
			result:  "tz1RABAzdLWVvxAFf1wpeUALAkp32mVhSGXX",
		},
		{
			address: "00004bf0acca4cc9e034b1d5f0f783c78e5ed44d866e",
			result:  "tz1SZZgtvMVXaBKPcez4gfjKUsDz1gs6vg6X",
		},
		{
			address: "000079e68d8f0a8d64ec856e193efc0a347ef4adf8ee",
			result:  "tz1WkaeRycRr999GrVFepJd9Nqi1FWqGyGqq",
		},
		{
			address: "0001028562fb176188114cf437a757cdc75bc4aa8cae",
			result:  "tz28YZoayJjVz2bRgGeVjxE8NonMiJ3r2Wdu",
		},
		{
			address: "00029d6a61cd3510193e257128da8f09a0b173bff695",
			result:  "tz3agP9LGe2cXmKQyYn6T68BHKjjktDbbSWX",
		},
		{
			address: "0168b709e887ddc34c3c9e468b5819b2f012b60ef700",
			result:  "KT1J8T7U6J1BAo9fJAxvedHsNErnejwvPyUH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			res, err := Address(tt.address)
			if err != nil {
				t.Errorf("Error in Address. Error: %v", err)
			}
			if res != tt.result {
				t.Errorf("Error in Address. Got %v, Expected: %v", res, tt.result)
			}
		})
	}
}

func TestBytes(t *testing.T) {
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
			res := Bytes(tc.input)
			if res != tc.result {
				t.Errorf("Input: %v. Got: %v, expected: %v.", tc.input, res, tc.result)
			}
		})
	}
}
