package tzbase58

import "testing"

func TestEncodeFromHex(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix []byte
		res    string
	}{
		{
			name:   "tz1",
			input:  "06a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
			prefix: []byte{6, 161, 159},
			res:    "tz1LFEVYR7YRCxT6Nm3Zfjdnfj77xZqhbR5U",
		},
		{
			name:   "tz1",
			input:  "44c6f8bc6088cd3b64f0bca87f812634c3f0ed30",
			prefix: []byte{6, 161, 159},
			res:    "tz1RugzxKA8NwuymbGcy2wkSTvfRJpckfmDF",
		},
		{
			name:   "tz1",
			input:  "9e6ac2e529a49aedbcdd0ac9542d5c0f4ce76f77",
			prefix: []byte{6, 161, 159},
			res:    "tz1a5fMLLY5WCarCzH7RKTJHX9mJFN8eaaWG",
		},
		{
			name:   "tz3",
			input:  "358cbffa97149631cfb999fa47f0035fb1ea8636",
			prefix: []byte{6, 161, 164},
			res:    "tz3RDC3Jdn4j15J7bBHZd29EUee9gVB1CxD9",
		},
		{
			name:   "KT",
			input:  "1fb03e3ff9fedaf3a2200ffc64d27812da734bba",
			prefix: []byte{2, 90, 121},
			res:    "KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A",
		},
		{
			name:   "KT",
			input:  "6f516588d2ee560385e386708a13bd63da907cf3",
			prefix: []byte{2, 90, 121},
			res:    "KT1JjN5bTE9yayzYHiBm6ruktwEWSHRF8aDm",
		},
		{
			name:   "KT",
			input:  "e5bae183211979a662665319a0900df3542e65ba",
			prefix: []byte{2, 90, 121},
			res:    "KT1VXUBQbYMt58yoKhNo73Zf8HTMfAd8Fqge",
		},
		{
			name:   "sig",
			input:  "bdc36db614aaa6084549020d376bb2469b5ea888dca2f7afbe5a0095bcc45ca0d8b5f00a051969437fe092debbcfe19d66378fbb74104de7eb1ecd895a64a80a",
			prefix: []byte{4, 130, 43},
			res:    "signpEFVQ1rW3TnVhc3PXf6SHRj7PvxwfJhBukWfB5X9rDhzpEk3ms5gRh763e922n52uQcjeqhqPdYi7WbFs2ERrNAPmCZJ",
		},
		{
			name:   "sig",
			input:  "a04991b4e938cc42d6c01c42be3649a81a9f80d244d9b90e7ec4edf8e0a7b68b6c212da2fef076e48fed66802fa83442b960a36afdb3e60c3cf14d4010f41f03",
			prefix: []byte{4, 130, 43},
			res:    "sigixZejtj1GfDpyiWAQAmvbtnNmCXKyADqVvCaXJH9xHyhSnYYV8696Z3kkns5DNV7oMnMPfNzo3qm84DfEx1XG6saZmHiA",
		},
		{
			name:   "chainID/main",
			input:  "7a06a770",
			prefix: []byte{87, 82, 0},
			res:    "NetXdQprcVkpaWU",
		},
		{
			name:   "chainID/babylon",
			input:  "458aa837",
			prefix: []byte{87, 82, 0},
			res:    "NetXUdfLh6Gm88t",
		},
		{
			name:   "chainID/carthage",
			input:  "9caecab9",
			prefix: []byte{87, 82, 0},
			res:    "NetXjD3HPJJjmcd",
		},
		{
			name:   "chainID/zeronet",
			input:  "0f6f0310",
			prefix: []byte{87, 82, 0},
			res:    "NetXKakFj1A7ouL",
		},
		{
			name:   "ed25519_public_key",
			input:  "4e4ca2abb4baeed702a0ac5b0de9b5607dd1fedb399c0ce25e15b3868f67269e",
			prefix: []byte{13, 15, 37, 217},
			res:    "edpkuEhzJqdFBCWMw6TU3deADRK2fq3GuwWFUphwyH7ero1Na4oGFP",
		},
		{
			name:   "secp256k1_public_key",
			input:  "030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
			prefix: []byte{3, 254, 226, 86},
			res:    "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
		},
		{
			name:   "p256_public_key",
			input:  "031a3ad5ea94de6912f9bc83fd31de49816e90602c5252d77b5b233bfe711b0dd2",
			prefix: []byte{3, 178, 139, 127},
			res:    "p2pk66iTZwLmRPshQgUr2HE3RUzSFwAN5MNaBQ5rfduT1dGKXd25pNN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeFromHex(tt.input, tt.prefix)
			if err != nil {
				t.Errorf("Error in encodeFromHex: %v", err)
				return
			}

			if result != tt.res {
				t.Errorf("Invalid base58 encoding. Got: %v, Expected: %v", result, tt.res)
			}
		})
	}
}

func TestEncodeFromBytes(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		prefix []byte
		res    string
	}{
		{
			name:   "expr",
			input:  []byte{67, 91, 208, 213, 143, 94, 239, 63, 51, 236, 101, 133, 225, 61, 89, 133, 12, 217, 196, 85, 255, 147, 117, 80, 141, 224, 126, 19, 72, 147, 180, 24},
			prefix: []byte{13, 44, 64, 27},
			res:    "expru2YV8AanTTUSV4K21P7X4DzbuWQFVk7NewDuP1A5uamffiiFA3",
		},
		{
			name:   "expr",
			input:  []byte{26, 54, 49, 4, 200, 100, 110, 132, 202, 175, 188, 34, 28, 19, 50, 17, 137, 249, 114, 244, 43, 241, 183, 143, 187, 115, 39, 157, 233, 111, 130, 84},
			prefix: []byte{13, 44, 64, 27},
			res:    "exprtiRSZkLKYRess9GZ3ryb4cVQD36WLo2oysZBFxKTZ2jXqcHWGj",
		},
		{
			name:   "expr",
			input:  []byte{152, 104, 71, 141, 151, 129, 55, 21, 140, 230, 76, 215, 115, 57, 114, 50, 133, 215, 94, 29, 166, 199, 223, 41, 238, 72, 34, 186, 36, 195, 235, 94},
			prefix: []byte{13, 44, 64, 27},
			res:    "exprufzwVGdAX7zG91UpiAkR2yVxEDE75tHD5YgSBmYMUx22teZTCM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := EncodeFromBytes(tt.input, tt.prefix); result != tt.res {
				t.Errorf("Invalid base58 encoding. Got: %v, Expected: %v", result, tt.res)
			}
		})
	}
}

func TestDecodeToHex(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix []byte
		res    string
	}{
		{
			name:   "tz1",
			input:  "tz1LFEVYR7YRCxT6Nm3Zfjdnfj77xZqhbR5U",
			prefix: []byte{6, 161, 159},
			res:    "06a868bd80219eb1f6a25108d1bdaa98ae27b2d9",
		},
		{
			name:   "tz1",
			input:  "tz1RugzxKA8NwuymbGcy2wkSTvfRJpckfmDF",
			prefix: []byte{6, 161, 159},
			res:    "44c6f8bc6088cd3b64f0bca87f812634c3f0ed30",
		},
		{
			name:   "tz1",
			input:  "tz1a5fMLLY5WCarCzH7RKTJHX9mJFN8eaaWG",
			prefix: []byte{6, 161, 159},
			res:    "9e6ac2e529a49aedbcdd0ac9542d5c0f4ce76f77",
		},
		{
			name:   "tz3",
			input:  "tz3RDC3Jdn4j15J7bBHZd29EUee9gVB1CxD9",
			prefix: []byte{6, 161, 164},
			res:    "358cbffa97149631cfb999fa47f0035fb1ea8636",
		},
		{
			name:   "KT",
			input:  "KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A",
			prefix: []byte{2, 90, 121},
			res:    "1fb03e3ff9fedaf3a2200ffc64d27812da734bba",
		},
		{
			name:   "secp256k1_public_key",
			input:  "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
			prefix: []byte{3, 254, 226, 86},
			res:    "030ed412d33412ab4b71df0aaba07df7ddd2a44eb55c87bf81868ba09a358bc0e0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeToHex(tt.input, tt.prefix)
			if err != nil {
				t.Errorf("Error in DecodeFromHex: %v", err)
				return
			}

			if result != tt.res {
				t.Errorf("Invalid base58 decoding. Got: %v, Expected: %v", result, tt.res)
			}
		})
	}
}
