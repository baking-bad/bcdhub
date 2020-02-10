package unpack

import (
	"encoding/hex"
	"unicode/utf8"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/unpack/rawbytes"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/unpack/tzbase58"
)

// PublicKey -
func PublicKey(input string) (string, error) {
	return tzbase58.DecodePublicKey(input)
}

// KeyHash -
func KeyHash(input string) (string, error) {
	return tzbase58.DecodeKeyHash(input)
}

// Address -
func Address(input string) (string, error) {
	if input[:2] == "01" && input[len(input)-2:] == "00" {
		return tzbase58.DecodeKT(input)
	}

	return tzbase58.DecodeTz(input)
}

// Bytes -
func Bytes(input string) string {
	if len(input) > 1 && input[:2] == "05" {
		str, err := rawbytes.ToMicheline(input[2:])
		if err == nil {
			return str
		}
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}

	return input
}
