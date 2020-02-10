package tzbase58

import "testing"

func TestDecodePublicKey(t *testing.T) {
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
			res, err := DecodePublicKey(tt.input)
			if err != nil {
				t.Errorf("Error in PublicKey. Input: %v. Error: %v", tt.input, err)
			}
			if res != tt.result {
				t.Errorf("Error in PublicKey. Input: %v. Got: %v. Expected: %v.", tt.input, res, tt.result)
			}
		})
	}
}

func TestDecodeKeyHash(t *testing.T) {
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
			res, err := DecodeKeyHash(tt.input)

			if err != nil {
				t.Errorf("Error in KeyHash. Error: %v", err)
			}

			if res != tt.result {
				t.Errorf("Error in Keyhash. Got: %v. Expected: %v", res, tt.result)
			}
		})
	}
}

func TestDecodeTz(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			res, err := DecodeTz(tt.address)
			if err != nil {
				t.Errorf("Error in Address. Error: %v", err)
			}
			if res != tt.result {
				t.Errorf("Error in Address. Got %v, Expected: %v", res, tt.result)
			}
		})
	}
}

func TestDecodeKT(t *testing.T) {
	tests := []struct {
		address string
		result  string
	}{
		{
			address: "0168b709e887ddc34c3c9e468b5819b2f012b60ef700",
			result:  "KT1J8T7U6J1BAo9fJAxvedHsNErnejwvPyUH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			res, err := DecodeKT(tt.address)
			if err != nil {
				t.Errorf("Error in Address. Error: %v", err)
			}
			if res != tt.result {
				t.Errorf("Error in Address. Got %v, Expected: %v", res, tt.result)
			}
		})
	}
}
