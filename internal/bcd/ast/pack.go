package ast

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/pkg/errors"
)

const (
	packByte = 0x05
)

// Pack -
func Pack(node Base) (string, error) {
	data, err := Forge(node, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("05%s", data), nil
}

// Unpack -
func Unpack(data []byte) (UntypedAST, error) {
	trimmed, err := unpack(data)
	if err != nil {
		return nil, err
	}
	unforger := forge.NewMichelson()
	if _, err := unforger.Unforge(trimmed); err != nil {
		return nil, err
	}
	return unforger.Nodes, nil
}

func unpack(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != packByte {
		return nil, errors.Errorf("Invalid unpack data: %v", data)
	}
	return data[1:], nil
}

func getOptimizedAddress(val string, tzOnly bool) ([]byte, error) {
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
	case encoding.PrefixPublicKeyKT1:
		address = append([]byte{1}, address...)
		address = append(address, byte(0))
	default:
		return nil, errors.Errorf("Invalid address prefix: %s", prefix)
	}
	if tzOnly {
		return address[1:], nil
	}
	return address, nil
}

func fromOptimizedAddress(str string) (string, error) {
	if len(str) != 44 && len(str) != 42 {
		return "", errors.Wrapf(consts.ErrInvalidAddress, "fromOptimizedAddress: %s", str)
	}
	switch {
	case strings.HasPrefix(str, "0000"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ1))
	case strings.HasPrefix(str, "0001"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ2))
	case strings.HasPrefix(str, "0002"):
		return encoding.EncodeBase58String(str[4:], []byte(encoding.PrefixPublicKeyTZ3))
	case strings.HasPrefix(str, "01") && strings.HasSuffix(str, "00"):
		return encoding.EncodeBase58String(str[2:len(str)-2], []byte(encoding.PrefixPublicKeyKT1))
	case len(str) == 42:
		return fromOptimizedAddress("00" + str)
	default:
		return str, nil
	}
}

func getOptimizedContract(val string) (string, error) {
	parts := strings.Split(val, "%")
	if len(parts) == 1 {
		parts = append(parts, consts.DefaultEntrypoint)
	}
	res, err := getOptimizedAddress(parts[0], false)
	if err != nil {
		return "", err
	}
	if parts[1] != consts.DefaultEntrypoint {
		decoded, err := hex.DecodeString(parts[1])
		if err != nil {
			return "", err
		}
		res = append(res, decoded...)
	}
	return hex.EncodeToString(res), nil
}

func fromOptimizedContract(str string) (string, error) {
	if len(str) != 44 && len(str) != 42 {
		return "", errors.Wrapf(consts.ErrInvalidAddress, "fromOptimizedContract: %s", str)
	}
	res, err := fromOptimizedAddress(str[:44])
	if err != nil {
		return "", err
	}
	if len(str) > 44 {
		return fmt.Sprintf("%s%%%s", res, str[44:]), nil
	}
	return res, nil
}

func getOptimizedPublicKey(val string) ([]byte, error) {
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

func fromOptimizedPublicKey(str string) (string, error) {
	if len(str) != 66 {
		return "", errors.Wrapf(consts.ErrInvalidAddress, "fromOptimizedPublicKey: %s", str)
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

func fromOptimizedChainID(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixChainID))
}

func fromOptimizedSignature(str string) (string, error) {
	return encoding.EncodeBase58String(str, []byte(encoding.PrefixGenericSignature))
}
