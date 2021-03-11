package forge

import (
	"bytes"
	"fmt"
	"math"
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
	var buffer bytes.Buffer
	for i := range data {
		buffer.WriteByte(data[i])
		if data[i] < 127 {
			break
		}
	}

	return buffer.Len(), val.decode(buffer.Bytes())
}

// Forge -
func (val *Int) Forge() ([]byte, error) {
	data, err := val.encode()
	if err != nil {
		return nil, err
	}
	return append([]byte{ByteInt}, data...), nil
}

func (val *Int) decode(source []byte) error {
	if len(source) == 0 {
		return errors.Errorf("expected non-empty byte array")
	}

	segments := make([]string, len(source))
	for i, curByte := range source {
		segments[i] = fmt.Sprintf("%08b", curByte)
	}

	for i, segment := range segments {
		segments[i] = segment[1:]
	}

	firstSegment := []rune(segments[0])
	isNegative := firstSegment[0] == '1'
	segments[0] = string(firstSegment[1:])

	segments = reverse(segments)

	bitStringBuf := new(bytes.Buffer)
	for _, segment := range segments {
		bitStringBuf.WriteString(segment)
	}
	bitString := bitStringBuf.String()

	if isNegative {
		bitString = "-" + bitString
	}

	if _, ok := val.IntValue.SetString(bitString, 2); !ok {
		return errors.Errorf("failed to parse bit string %s to big.Int", bitString)
	}

	return nil
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
