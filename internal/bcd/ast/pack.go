package ast

import (
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

const (
	packByte = 0x05
)

func pack(data []byte) []byte {
	return append([]byte{packByte}, data...)
}

func unpack(data []byte) []byte {
	if len(data) == 0 || data[0] != packByte {
		return nil
	}
	return data[1:]
}

func forgeAddress(val string, tzOnly bool) ([]byte, error) {
	prefix := val[:3]
	address, err := encoding.DecodeBase58(val)
	if err != nil {
		return nil, err
	}
	address = address[3:]
	switch prefix {
	case "tz1":
		address = append([]byte{0, 0}, address...)
	case "tz2":
		address = append([]byte{0, 1}, address...)
	case "tz3":
		address = append([]byte{0, 2}, address...)
	case "KT1":
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

func forgeContract(val string) (string, error) {
	parts := strings.Split(val, "%")
	if len(parts) == 1 {
		parts = append(parts, consts.DefaultEntrypoint)
	}
	res, err := forgeAddress(parts[0], false)
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

func forgeString(val string) ([]byte, error) {
	forger := new(forge.String)
	forger.StringValue = &val
	return forger.Forge()
}

func forgeBytes(val string) ([]byte, error) {
	forger := new(forge.Bytes)
	forger.BytesValue = &val
	return forger.Forge()
}

func forgeInt(val *base.BigInt) ([]byte, error) {
	forger := forge.NewInt()
	forger.IntValue = val
	return forger.Forge()
}

func forgePublicKey(val string) ([]byte, error) {
	prefix := val[4:]
	decoded, err := encoding.DecodeBase58String(val)
	if err != nil {
		return nil, err
	}
	switch prefix {
	case "edpk":
		return append([]byte{0}, decoded...), nil
	case "sppk":
		return append([]byte{1}, decoded...), nil
	case "p2pk":
		return append([]byte{2}, decoded...), nil
	default:
		return nil, errors.Errorf("Invalid public key prefix: %s", prefix)
	}
}
