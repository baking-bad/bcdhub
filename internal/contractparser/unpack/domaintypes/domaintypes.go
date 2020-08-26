package domaintypes

import (
	"github.com/baking-bad/bcdhub/internal/tzbase58"
	"github.com/pkg/errors"
)

// Length consts
const (
	KeyHashBytesLength = 21
	AddressBytesLength = 22
)

// DecodePublicKey -
func DecodePublicKey(input string) (string, error) {
	prefixes := map[string][]byte{
		"00": {13, 15, 37, 217},
		"01": {3, 254, 226, 86},
		"02": {3, 178, 139, 127},
	}

	if _, ok := prefixes[input[:2]]; !ok {
		return "", errors.Errorf("[DecodePublicKey] Unknown prefix %v", input[:2])
	}

	return tzbase58.EncodeFromHex(input[2:], prefixes[input[:2]])
}

// DecodeKeyHash -
func DecodeKeyHash(input string) (string, error) {
	prefixes := map[string][]byte{
		"00": {6, 161, 159},
		"01": {6, 161, 161},
		"02": {6, 161, 164},
	}

	if _, ok := prefixes[input[:2]]; !ok {
		return "", errors.Errorf("[DecodeKeyHash] Unknown prefix %v", input[:2])
	}

	return tzbase58.EncodeFromHex(input[2:], prefixes[input[:2]])
}

// DecodeSignature -
func DecodeSignature(input string) (string, error) {
	prefix := []byte{4, 130, 43}

	return tzbase58.EncodeFromHex(input, prefix)
}

// DecodeChainID -
func DecodeChainID(input string) (string, error) {
	prefix := []byte{87, 82, 0}

	return tzbase58.EncodeFromHex(input, prefix)
}

// DecodeKT -
func DecodeKT(input string) (string, error) {
	prefix := []byte{2, 90, 121}

	return tzbase58.EncodeFromHex(input[2:len(input)-2], prefix)
}

// DecodeTz -
func DecodeTz(input string) (string, error) {
	prefixes := map[string][]byte{
		"0000": {6, 161, 159},
		"0001": {6, 161, 161},
		"0002": {6, 161, 164},
	}

	if _, ok := prefixes[input[:4]]; !ok {
		return "", errors.Errorf("[DecodeTz] Unknown prefix %v %v", input[:4], input)
	}

	return tzbase58.EncodeFromHex(input[4:], prefixes[input[:4]])
}

// HasKT1Affixes -
func HasKT1Affixes(data []byte) bool {
	return data[0] == 0x01 && data[len(data)-1] == 0x00
}

// DecodeOpgHash -
func DecodeOpgHash(input string) (string, error) {
	if len(input) != 51 {
		return "", errors.Errorf("[DecodeOpgHash] invalid input length: %d != 51", len(input))
	}

	return tzbase58.DecodeToHex(input, []byte{5, 116})
}
