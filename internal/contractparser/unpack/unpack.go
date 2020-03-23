package unpack

import (
	"encoding/hex"
	"fmt"
	"unicode"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/rawbytes"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/tzbase58"
	"github.com/tidwall/gjson"
)

const (
	signatureHexLength = 128
	chainIDHexLength   = 8
	addressHexLength   = 44
)

const minPrintableASCII = 32

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
	input = input[:addressHexLength]
	if input[:2] == "01" && input[len(input)-2:] == "00" {
		return tzbase58.DecodeKT(input)
	}

	return tzbase58.DecodeTz(input)
}

// Signature -
func Signature(input string) (string, error) {
	if len(input) != signatureHexLength {
		return "", fmt.Errorf("[Signature] Wrong length of %v. Expected %v, Got: %v", input, signatureHexLength, len(input))
	}

	return tzbase58.DecodeSignature(input)
}

// ChainID -
func ChainID(input string) (string, error) {
	if len(input) != chainIDHexLength {
		return "", fmt.Errorf("[ChainID] Wrong length of %v. Expected %v, Got: %v", input, chainIDHexLength, len(input))
	}

	return tzbase58.DecodeChainID(input)
}

// Contract -
func Contract(input string) (string, error) {
	if len(input) < addressHexLength {
		return "", fmt.Errorf("[Contract] Wrong length of %v. Expected %v, Got: %v", input, addressHexLength, len(input))
	}

	address, err := Address(input[:addressHexLength])
	if err != nil {
		return "", fmt.Errorf("[Contract] Cant decode Address %v error: %v", input, err)
	}

	if len(input) == addressHexLength {
		return address, nil
	}

	tail, err := hex.DecodeString(input[addressHexLength:])
	if err != nil {
		return "", fmt.Errorf("[Contract] %v hex.DecodeString error: %v", input, err)
	}

	return fmt.Sprintf("%s%%%s", address, tail), nil
}

// Bytes -
func Bytes(input string) string {
	if len(input) > 1 && input[:2] == "05" {
		str, err := rawbytes.ToMicheline(input[2:])
		if err == nil {
			data := gjson.Parse(str)
			res, err := formatter.MichelineToMichelson(data, false, formatter.DefLineSize)

			if err == nil {
				return res
			}
		}
	}

	decoded, err := hex.DecodeString(input)
	if err == nil && isASCII(decoded) {
		return string(decoded)
	}

	return input
}

func isASCII(input []byte) bool {
	for _, c := range input {
		if c < minPrintableASCII || c > unicode.MaxASCII {
			return false
		}
	}
	return true
}
