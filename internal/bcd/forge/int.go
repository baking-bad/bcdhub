package forge

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// Int -
type Int base.Node

// NewInt -
func NewInt() *Int {
	return &Int{
		IntValue: types.NewBigInt(0),
	}
}

// Unforge -
func (val *Int) Unforge(data []byte) (int, error) {
	buffer := new(bytes.Buffer)
	for i := range data {
		buffer.WriteByte(data[i])
		if data[i] < 128 {
			break
		}
	}

	parts := buffer.Bytes()
	for i := len(parts) - 1; i > 0; i-- {
		num := int64(parts[i] & 0x7f)
		val.IntValue.Int = val.IntValue.Lsh(val.IntValue.Int, 7)
		val.IntValue.Int = val.IntValue.Or(val.IntValue.Int, big.NewInt(num))
	}

	if len(parts) > 0 {
		num := int64(parts[0] & 0x3f)
		val.IntValue.Int = val.IntValue.Lsh(val.IntValue.Int, 6)
		val.IntValue.Int = val.IntValue.Or(val.IntValue.Int, big.NewInt(num))

		if parts[0]&0x40 > 0 {
			val.IntValue.Int = val.IntValue.Neg(val.IntValue.Int)
		}
	}

	return buffer.Len(), nil
}

// Forge -
func (val *Int) Forge() ([]byte, error) {
	data, err := val.encode()
	if err != nil {
		return nil, err
	}
	return append([]byte{ByteInt}, data...), nil
}

func (val *Int) encode() ([]byte, error) {
	if val.IntValue == nil {
		return nil, errors.New("Invalid int value")
	}

	isNegative := val.IntValue.Sign() == -1
	bits := val.IntValue.Text(2)
	if isNegative {
		bits = bits[1:]
	}
	bitsCount := len(bits)

	var pad int
	switch {
	case (bitsCount-6)%7 == 0:
		pad = bitsCount
	case bitsCount > 6:
		pad = bitsCount + 7 - (bitsCount-6)%7
	default:
		pad = 6
	}
	bits = fmt.Sprintf("%0*s", pad, bits)

	segments := make([]string, 0)
	for i := 0; i <= pad/7; i++ {
		idx := 7 * i
		length := int(math.Min(7, float64(pad-7*i)))
		segments = append(segments, bits[idx:(idx+length)])
	}

	segments = reverse(segments)
	if isNegative {
		segments[0] = fmt.Sprintf("1%s", segments[0])
	} else {
		segments[0] = fmt.Sprintf("0%s", segments[0])
	}

	data := make([]byte, 0)

	for i := 0; i < len(segments); i++ {
		prefix := "1"
		if i == len(segments)-1 {
			prefix = "0"
		}
		val, err := strconv.ParseUint(prefix+segments[i], 2, 8)
		if err != nil {
			return nil, err
		}
		data = append(data, byte(val))
	}

	return data, nil
}

func reverse(arr []string) []string {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
	return arr
}
