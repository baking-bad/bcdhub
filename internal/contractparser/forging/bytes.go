package forging

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/domaintypes"
	"github.com/pkg/errors"
)

type bytesForger struct{}

// Decode -
func (d bytesForger) Unforge(dec *decoder, code *strings.Builder) (int, error) {
	length, err := unforgeLength(dec)
	if err != nil {
		return 4, err
	}
	if dec.Len() < length {
		return 4, &invalidDataError{
			typ:     "bytes",
			message: fmt.Sprintf("Not enough data in reader: %d < %d", dec.Len(), length),
		}
	}

	data := make([]byte, length)
	if _, err := dec.Read(data); err != nil && err != io.EOF {
		return 4 + length, err
	}

	// log.Printf("[bytes Decode] data: %x\n", data)

	if length == domaintypes.KeyHashBytesLength {
		if res, err := decodeKeyHash(data); err == nil {
			fmt.Fprintf(code, `{ "string": "%s" }`, res)
			return 4 + length, nil
		}
	}

	if length == domaintypes.AddressBytesLength {
		if res, err := decodeAddress(data); err == nil {
			fmt.Fprintf(code, `{ "string": "%s" }`, res)
			return 4 + length, nil
		}
	}

	s := hex.EncodeToString(data)
	intDec := newDecoder(strings.NewReader(s))

	var bufBuilder strings.Builder
	l, err := hexToMicheline(intDec, &bufBuilder)
	if err != nil || intDec.Len() > 0 {
		if _, ok := err.(*invalidDataError); ok || intDec.Len() > 0 {
			fmt.Fprintf(code, `{ "bytes": "%x" }`, data)
			return 4 + length, nil
		}
		return l + 4, err
	}
	fmt.Fprintf(code, `{ "bytes": "%s" }`, bufBuilder.String())
	return 4 + length, nil
}

func decodeKeyHash(data []byte) (string, error) {
	return domaintypes.DecodeKeyHash(hex.EncodeToString(data))
}

func decodeAddress(data []byte) (string, error) {
	if domaintypes.HasKT1Affixes(data) {
		if res, err := domaintypes.DecodeKT(hex.EncodeToString(data)); err == nil {
			return res, nil
		}
	}

	if res, err := domaintypes.DecodeTz(hex.EncodeToString(data)); err == nil {
		return res, nil
	}

	return "", errors.Errorf("decodeAddress: can't decode address from bytes: %v", data)
}
