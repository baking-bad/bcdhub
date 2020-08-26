package pack

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
)

// Address -
func Address(address string) (string, error) {
	var buf bytes.Buffer

	prefix := address[0:3]

	decodedAddress, _, err := base58.CheckDecode(address)
	if err != nil {
		return "", err
	}
	decodedAddress = decodedAddress[2:]

	switch prefix {
	case "tz1":
		buf.Write([]byte{0, 0})
		buf.Write(decodedAddress)
	case "tz2":
		buf.Write([]byte{0, 1})
		buf.Write(decodedAddress)
	case "tz3":
		buf.Write([]byte{0, 2})
		buf.Write(decodedAddress)
	case "KT1":
		buf.WriteByte(1)
		buf.Write(decodedAddress)
		buf.WriteByte(0)
	default:
		return "", errors.Errorf("[pack.Address] Unknown address prefix: %s", prefix)
	}

	return hex.EncodeToString(buf.Bytes()), nil
}
