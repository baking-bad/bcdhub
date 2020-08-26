package tzbase58

import (
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
)

// EncodeFromHex - encodes hex string to base58 with prefix
func EncodeFromHex(input string, prefix []byte) (string, error) {
	if len(prefix) < 1 {
		return "", errors.Errorf("Invalid prefix %v. Should be at least 1 symbol length", prefix)
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

// DecodeToHex - decodes hex string from base58 with prefix
func DecodeToHex(input string, prefix []byte) (string, error) {
	decoded, version, err := base58.CheckDecode(input)
	if err != nil {
		return "", err
	}
	if len(prefix) > 0 {
		if version != prefix[0] {
			return "", errors.Errorf("[DecodeToHex] Unknown version %v %v", version, prefix[0])
		}

		for i := range prefix[1:] {
			if decoded[i] != prefix[i+1] {
				return "", errors.Errorf("[DecodeToHex] Unknown prefix %v %v", decoded[:2], prefix)
			}
		}
	}
	return hex.EncodeToString(decoded[len(prefix)-1:]), nil
}
