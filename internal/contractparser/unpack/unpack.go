package unpack

import (
	"encoding/hex"
	"fmt"
	"unicode/utf8"

	"github.com/baking-bad/bcdhub/internal/contractparser/forging"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/domaintypes"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	signatureHexLength   = 128
	chainIDHexLength     = 8
	addressHexLength     = 44
	keyHashHexLength     = 42
	pKeyEd25519HexLength = 66
	pKey256HexLength     = 68
	ktPrefix             = "01"
	ktSuffix             = "00"

	// MainPrefix -
	MainPrefix = "05"
)

// PublicKey -
func PublicKey(input string) (string, error) {
	if len(input) != pKeyEd25519HexLength && len(input) != pKey256HexLength {
		return "", errors.Errorf("[PublicKey] Wrong length of %v. Expected %v or %v, Got: %v", input, pKeyEd25519HexLength, pKey256HexLength, len(input))
	}

	return domaintypes.DecodePublicKey(input)
}

// KeyHash -
func KeyHash(input string) (string, error) {
	if len(input) != keyHashHexLength {
		return "", errors.Errorf("[KeyHash] Wrong length of %v. Expected %v, Got: %v", input, keyHashHexLength, len(input))
	}

	return domaintypes.DecodeKeyHash(input)
}

// Address - unpack KT, tz1, tz2, tz3 addresses
func Address(input string) (string, error) {
	if len(input) != addressHexLength {
		return "", errors.Errorf("[Address] Wrong length of %v. Expected %v, Got: %v", input, addressHexLength, len(input))
	}

	if input[:2] == ktPrefix && input[len(input)-2:] == ktSuffix {
		return domaintypes.DecodeKT(input)
	}

	return domaintypes.DecodeTz(input)
}

// Signature -
func Signature(input string) (string, error) {
	if len(input) != signatureHexLength {
		return "", errors.Errorf("[Signature] Wrong length of %v. Expected %v, Got: %v", input, signatureHexLength, len(input))
	}

	return domaintypes.DecodeSignature(input)
}

// ChainID -
func ChainID(input string) (string, error) {
	if len(input) != chainIDHexLength {
		return "", errors.Errorf("[ChainID] Wrong length of %v. Expected %v, Got: %v", input, chainIDHexLength, len(input))
	}

	return domaintypes.DecodeChainID(input)
}

// Contract - unpack contract
func Contract(input string) (string, error) {
	if len(input) < addressHexLength {
		return "", errors.Errorf("[Contract] Wrong length of %v. Expected %v, Got: %v", input, addressHexLength, len(input))
	}

	address, err := Address(input[:addressHexLength])
	if err != nil {
		return "", errors.Errorf("[Contract] Cant decode Address %v error: %v", input, err)
	}

	if len(input) == addressHexLength {
		return address, nil
	}

	tail, err := hex.DecodeString(input[addressHexLength:])
	if err != nil {
		return "", errors.Errorf("[Contract] %v hex.DecodeString error: %v", input, err)
	}

	return fmt.Sprintf("%s%%%s", address, tail), nil
}

// Bytes - unpack bytes
func Bytes(input string) string {
	if len(input) >= 1 && input[:2] == MainPrefix {
		str, err := forging.Unforge(input[2:])
		if err == nil {
			data := gjson.Parse(str)
			res, err := formatter.MichelineToMichelson(data, false, formatter.DefLineSize)

			if err == nil {
				return res
			}
		}
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}

	return input
}
