package decode

import (
	"encoding/hex"
	"unicode/utf8"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/decode/rawbase58"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/decode/rawbytes"
)

// PublicKey -
func PublicKey(input string) (string, error) {
	return rawbase58.DecodePublicKey(input)
}

// KeyHash -
func KeyHash(input string) (string, error) {
	return rawbase58.DecodeKeyHash(input)
}

// Address -
func Address(input string) (string, error) {
	if input[:2] == "01" && input[len(input)-2:] == "00" {
		return rawbase58.DecodeKT(input)
	}

	return rawbase58.DecodeTz(input)
}

// Bytes -
func Bytes(input string) string {
	str, _, err := rawbytes.HexToMicheline(input)
	if err == nil {
		return str
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}

	return input
}
