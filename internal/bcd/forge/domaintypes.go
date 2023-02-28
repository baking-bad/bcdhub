package forge

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/pkg/errors"
)

// Address -
func Address(val string, tzOnly bool) ([]byte, error) {
	prefix := val[:3]
	address, err := encoding.DecodeBase58(val)
	if err != nil {
		return nil, err
	}
	switch prefix {
	case encoding.PrefixPublicKeyTZ1:
		address = append([]byte{0, 0}, address...)
	case encoding.PrefixPublicKeyTZ2:
		address = append([]byte{0, 1}, address...)
	case encoding.PrefixPublicKeyTZ3:
		address = append([]byte{0, 2}, address...)
	case encoding.PrefixPublicKeyTZ4:
		address = append([]byte{0, 3}, address...)
	case encoding.PrefixPublicKeyKT1:
		address = append([]byte{1}, address...)
		address = append(address, byte(0))
	case encoding.PrefixPublicKeyTxr1:
		address = append([]byte{2}, address...)
		address = append(address, byte(0))
	case encoding.PrefixOriginatedSmartRollup:
		address = append([]byte{3}, address...)
		address = append(address, byte(0))
	default:
		return nil, errors.Errorf("Invalid address prefix: %s", prefix)
	}
	if tzOnly {
		return address[1:], nil
	}
	return address, nil
}

// UnforgeAddress -
func UnforgeAddress(str string) (string, error) {
	if len(str) != 44 && len(str) != 42 {
		return str, errors.Wrapf(consts.ErrInvalidAddress, "UnforgeAddress: %s", str)
	}
	switch {
	case len(str) == 42:
		return UnforgeAddress("00" + str)
	case strings.HasPrefix(str, "0000"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ1))
	case strings.HasPrefix(str, "0001"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ2))
	case strings.HasPrefix(str, "0002"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ3))
	case strings.HasPrefix(str, "0003"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ3))
	case strings.HasPrefix(str, "01") && strings.HasSuffix(str, "00"):
		return encoding.EncodeBase58String(str[2:len(str)-2], []byte(encoding.PrefixPublicKeyKT1))
	case strings.HasPrefix(str, "02") && strings.HasSuffix(str, "00"):
		return encoding.EncodeBase58String(str[2:len(str)-2], []byte(encoding.PrefixPublicKeyTxr1))
	case strings.HasPrefix(str, "03") && strings.HasSuffix(str, "00"):
		return encoding.EncodeBase58String(str[2:len(str)-2], []byte(encoding.PrefixOriginatedSmartRollup))
	default:
		return str, errors.Wrapf(consts.ErrInvalidAddress, "UnforgeAddress: %s", str)
	}
}

// Contract -
func Contract(val string) (string, error) {
	parts := strings.Split(val, "%")
	if len(parts) == 1 {
		parts = append(parts, consts.DefaultEntrypoint)
	}
	res, err := Address(parts[0], false)
	if err != nil {
		return "", err
	}
	if parts[1] != consts.DefaultEntrypoint {
		res = append(res, []byte(parts[1])...)
	}
	return hex.EncodeToString(res), nil
}

// UnforgeContract -
func UnforgeContract(str string) (string, error) {
	if len(str) < 44 {
		return "", errors.Wrapf(consts.ErrInvalidAddress, "UnforgeContract: %s", str)
	}
	res, err := UnforgeAddress(str[:44])
	if err != nil {
		return "", err
	}
	if len(str) > 44 {
		decoded, err := hex.DecodeString(str[44:])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s%%%s", res, string(decoded)), nil
	}
	return res, nil
}

// PublicKey -
func PublicKey(val string) ([]byte, error) {
	prefix := val[4:]
	decoded, err := encoding.DecodeBase58String(val)
	if err != nil {
		return nil, err
	}
	switch prefix {
	case encoding.PrefixED25519PublicKey:
		return append([]byte{0}, decoded...), nil
	case encoding.PrefixSecp256k1PublicKey:
		return append([]byte{1}, decoded...), nil
	case encoding.PrefixP256PublicKey:
		return append([]byte{2}, decoded...), nil
	default:
		return nil, errors.Errorf("Invalid public key prefix: %s", prefix)
	}
}

// UnforgePublicKey -
func UnforgePublicKey(str string) (string, error) {
	if len(str) != 68 && len(str) != 66 {
		return "", errors.Wrapf(consts.ErrInvalidAddress, "UnforgePublicKey: %s", str)
	}
	switch {
	case strings.HasPrefix(str, "00"):
		return encoding.EncodeBase58String(str[2:], []byte(encoding.PrefixED25519PublicKey))
	case strings.HasPrefix(str, "01"):
		return encoding.EncodeBase58String(str[2:], []byte(encoding.PrefixSecp256k1PublicKey))
	case strings.HasPrefix(str, "02"):
		return encoding.EncodeBase58String(str[2:], []byte(encoding.PrefixP256PublicKey))
	default:
		return str, nil
	}
}

// UnforgeChainID -
func UnforgeChainID(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixChainID))
}

// UnforgeSignature -
func UnforgeSignature(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixGenericSignature))
}

// UnforgeBakerHash -
func UnforgeBakerHash(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixBakerHash))
}

// UnforgeOpgHash -
func UnforgeOpgHash(input string) (string, error) {
	if len(input) != 51 {
		return "", errors.Wrapf(consts.ErrInvalidOPGHash, "UnforgeOpgHash: %s", input)
	}

	return encoding.EncodeBase58String(input, []byte(encoding.PrefixOperationHash))
}
