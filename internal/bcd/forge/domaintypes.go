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
	case encoding.PrefixPublicKeyTZ5:
		address = append([]byte{0, 4}, address...)
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
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ4))
	case strings.HasPrefix(str, "0004"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ5))
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
	if len(val) < 4 {
		return nil, errors.Errorf("Invalid public key: %s", val)
	}
	prefix := val[:4]
	decoded, err := encoding.DecodeBase58(val)
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
	case encoding.PrefixBLS12_381PublicKey:
		return append([]byte{3}, decoded...), nil
	case encoding.PrefixMLDSA44PublicKey:
		return append([]byte{4}, decoded...), nil
	default:
		return nil, errors.Errorf("Invalid public key prefix: %s", prefix)
	}
}

// publicKeyTags - forged public key tag -> base58 prefix and length of hex-encoded key without tag
var publicKeyTags = map[string]struct {
	prefix string
	length int
}{
	"00": {encoding.PrefixED25519PublicKey, 64},
	"01": {encoding.PrefixSecp256k1PublicKey, 66},
	"02": {encoding.PrefixP256PublicKey, 66},
	"03": {encoding.PrefixBLS12_381PublicKey, 96},
	"04": {encoding.PrefixMLDSA44PublicKey, 2624},
}

// UnforgePublicKey - converts forged public key to base58 form. Keys of unknown type are returned as is (hex).
func UnforgePublicKey(str string) (string, error) {
	if len(str) < 2 {
		return str, nil
	}
	tag, ok := publicKeyTags[str[:2]]
	if !ok || len(str)-2 != tag.length {
		return str, nil
	}
	return encoding.EncodeBase58String(str[2:], []byte(tag.prefix))
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

// UnforgeTxRollupL2Address - 20-byte BLS public key hash to `tz4` address
func UnforgeTxRollupL2Address(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixPublicKeyTZ4))
}

// UnforgeOpgHash -
func UnforgeOpgHash(input string) (string, error) {
	if len(input) != 51 {
		return "", errors.Wrapf(consts.ErrInvalidOPGHash, "UnforgeOpgHash: %s", input)
	}

	return encoding.EncodeBase58String(input, []byte(encoding.PrefixOperationHash))
}
