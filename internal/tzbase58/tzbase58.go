package tzbase58

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

// EncodeFromHex - encodes hex string to base58 with prefix
func EncodeFromHex(input string, prefix []byte) (string, error) {
	bs, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	return EncodeFromBytes(bs, prefix), nil
}

// EncodeFromBytes - encodes bytes slice to base58 with prefix
func EncodeFromBytes(input, prefix []byte) string {
	payload := append(prefix, input...)
	cksum := checksum(payload)
	payload = append(payload, cksum...)

	return base58.Encode(payload)
}

func checksum(input []byte) []byte {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	return h2[:4]
}

// DecodeFromHex - decodes hex string from base58 with prefix
func DecodeFromHex(input string, prefixLen int) (string, error) {
	decoded, _, err := base58.CheckDecode(input)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decoded[prefixLen-1:]), nil
}
