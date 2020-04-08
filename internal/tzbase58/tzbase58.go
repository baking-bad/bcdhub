package tzbase58

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
)

// EncodeFromHex - encodes hex string to base58 with prefix
func EncodeFromHex(input string, prefix []byte) (string, error) {
	if len(prefix) < 1 {
		return "", fmt.Errorf("Invalid prefix %v. Should be at least 1 symbol length", prefix)
	}

	bs, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	return EncodeFromBytes(bs, prefix), nil
}

// EncodeFromBytes - encodes bytes slice to base58 with prefix
func EncodeFromBytes(input, prefix []byte) string {
	payload := append(prefix[1:], input...)
	return base58.CheckEncode(payload, prefix[0])
}

// DecodeFromHex - decodes hex string from base58 with prefix
func DecodeFromHex(input string, prefixLen int) (string, error) {
	decoded, _, err := base58.CheckDecode(input)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decoded[prefixLen-1:]), nil
}
