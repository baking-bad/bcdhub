package ast

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/pkg/errors"
)

// errors
var (
	ErrValidation = errors.New("validation error")
)

// ValidatorConstraint -
type ValidatorConstraint interface {
	~string
}

// Validator -
type Validator[T ValidatorConstraint] func(T) error

var (
	hexRegex           = regexp.MustCompile("^[0-9a-fA-F]+$")
	hexWithPrefixRegex = regexp.MustCompile("^(0x)?[0-9a-fA-F]+$")
)

// AddressValidator -
func AddressValidator(value string) error {
	switch len(value) {
	case 44, 42:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "address '%s' should be hexademical without prefixes", value)
		}
	case 36:
		if !bcd.IsAddressLazy(value) {
			return errors.Wrapf(ErrValidation, "invalid address '%s'", value)
		}
		if !bcd.IsAddress(value) {
			return errors.Wrapf(ErrValidation, "invalid address '%s'", value)
		}
	default:
		return errors.Wrapf(ErrValidation, "invalid address '%s'", value)
	}

	return nil
}

// ContractValidator -
func ContractValidator(value string) error {
	switch len(value) {
	case 44, 42:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "contract '%s' should be hexademical without prefixes", value)
		}
	case 36:
		if !bcd.IsContractLazy(value) {
			return errors.Wrapf(ErrValidation, "invalid contract address '%s'", value)
		}
		if !bcd.IsContract(value) {
			return errors.Wrapf(ErrValidation, "invalid contract address '%s'", value)
		}
	default:
		if len(value) < 38 {
			return errors.Wrap(ErrValidation, "invalid contract address length")
		}
		if value[36] != '%' {
			return errors.Wrap(ErrValidation, "invalid contract address format address%%entrypoint")
		}
		address := value[:36]
		if !bcd.IsContractLazy(address) {
			return errors.Wrapf(ErrValidation, "invalid contract address '%s'", address)
		}
		if !bcd.IsContract(address) {
			return errors.Wrapf(ErrValidation, "invalid contract address '%s'", address)
		}
	}

	return nil
}

// BakerHashValidator -
func BakerHashValidator(value string) error {
	switch len(value) {
	case 40:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "baker hash '%s' should be hexademical without prefixes", value)
		}
	case 36:
		if !bcd.IsBakerHash(value) {
			return errors.Wrapf(ErrValidation, "invalid baker hash '%s'", value)
		}
	default:
		return errors.Wrap(ErrValidation, "invalid baker hash length")
	}

	return nil
}

// PublicKeyValidator -
func PublicKeyValidator(value string) error {
	switch len(value) {
	case 68, 66:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "public key '%s' should be hexademical without prefixes", value)
		}
	case 55, 54:
		if strings.HasPrefix(value, encoding.PrefixED25519PublicKey) ||
			strings.HasPrefix(value, encoding.PrefixP256PublicKey) ||
			strings.HasPrefix(value, encoding.PrefixSecp256k1PublicKey) {
			return nil
		}
		return errors.Wrapf(ErrValidation, "invalid public key '%s'", value)
	default:
		return errors.Wrap(ErrValidation, "invalid public key length")
	}

	return nil
}

// BytesValidator -
func BytesValidator(value string) error {
	if len(value)%2 > 0 {
		return errors.Wrapf(ErrValidation, "invalid bytes in hex length '%s'", value)
	}
	if value != "" && !hexWithPrefixRegex.MatchString(value) {
		return errors.Wrapf(ErrValidation, "bytes '%s' should be hexademical without prefixes", value)
	}
	return nil
}

// ChainIDValidator -
func ChainIDValidator(value string) error {
	switch len(value) {
	case 8:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "chain id '%s' should be hexademical without prefixes", value)
		}
	case 15:
		if strings.HasPrefix(value, encoding.PrefixChainID) {
			return nil
		}
		return errors.Wrapf(ErrValidation, "invalid chain id '%s'", value)
	default:
		return errors.Wrap(ErrValidation, "invalid chain id length")
	}

	return nil
}

// SignatureValidator -
func SignatureValidator(value string) error {
	switch len(value) {
	case 128:
		if !hexRegex.MatchString(value) {
			return errors.Wrapf(ErrValidation, "signature '%s' should be hexademical without prefixes", value)
		}
	case 96:
		if strings.HasPrefix(value, encoding.PrefixGenericSignature) {
			return nil
		}
		return errors.Wrapf(ErrValidation, "invalid signature '%s'", value)
	case 98:
		if strings.HasPrefix(value, encoding.PrefixP256Signature) {
			return nil
		}
		return errors.Wrapf(ErrValidation, "invalid signature '%s'", value)
	case 99:
		if strings.HasPrefix(value, encoding.PrefixED25519Signature) ||
			strings.HasPrefix(value, encoding.PrefixSecp256k1Signature) {
			return nil
		}
		return errors.Wrapf(ErrValidation, "invalid signature '%s'", value)
	default:
		return errors.Wrap(ErrValidation, "invalid signature length")
	}

	return nil
}
