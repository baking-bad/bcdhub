package tzbase58

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
)

// Length consts
const (
	KeyHashLength = 21
	AddressLength = 22
)

// DecodePublicKey -
func DecodePublicKey(input string) (string, error) {
	prefixes := map[string][]byte{
		"00": []byte{13, 15, 37, 217},
		"01": []byte{3, 254, 226, 86},
		"02": []byte{3, 178, 139, 127},
	}

	if _, ok := prefixes[input[:2]]; !ok {
		return "", fmt.Errorf("[DecodePublicKey] Unknown prefix %v", input[:2])
	}

	return encodeBase58(input[2:], prefixes[input[:2]])
}

// DecodeKeyHash -
func DecodeKeyHash(input string) (string, error) {
	prefixes := map[string][]byte{
		"00": []byte{6, 161, 159},
		"01": []byte{6, 161, 161},
		"02": []byte{6, 161, 164},
	}

	if _, ok := prefixes[input[:2]]; !ok {
		return "", fmt.Errorf("[DecodeKeyHash] Unknown prefix %v", input[:2])
	}

	return encodeBase58(input[2:], prefixes[input[:2]])
}

// DecodeSignature -
func DecodeSignature(input string) (string, error) {
	prefix := []byte{4, 130, 43}

	return encodeBase58(input, prefix)
}

// DecodeChainID -
func DecodeChainID(input string) (string, error) {
	prefix := []byte{87, 82, 0}

	return encodeBase58(input, prefix)
}

// DecodeKT -
func DecodeKT(input string) (string, error) {
	prefix := []byte{2, 90, 121}

	return encodeBase58(input[2:len(input)-2], prefix)
}

// DecodeTz -
func DecodeTz(input string) (string, error) {
	prefixes := map[string][]byte{
		"0000": []byte{6, 161, 159},
		"0001": []byte{6, 161, 161},
		"0002": []byte{6, 161, 164},
	}

	if _, ok := prefixes[input[:4]]; !ok {
		return "", fmt.Errorf("[DecodeTz] Unknown prefix %v %v", input[:4], input)
	}

	return encodeBase58(input[4:], prefixes[input[:4]])
}

// HasKT1Affixes -
func HasKT1Affixes(data []byte) bool {
	return data[0] == 0x01 && data[len(data)-1] == 0x00
}

func encodeBase58(input string, prefix []byte) (string, error) {
	bs, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	payload := append(prefix, bs...)
	cksum := checksum(payload)
	payload = append(payload, cksum...)
	res := base58.Encode(payload)

	return res, nil
}

func checksum(input []byte) []byte {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	return h2[:4]
}
