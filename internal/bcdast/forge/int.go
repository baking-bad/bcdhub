package forge

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
)

// Int -
type Int Node

// NewInt -
func NewInt() *Int {
	return &Int{
		IntValue: new(big.Int),
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

	return buffer.Len(), val.decodeSigned(buffer.Bytes())
}

func (val *Int) decodeSigned(source []byte) error {
	if len(source) == 0 {
		return errors.Errorf("expected non-empty byte array")
	}

	// Split input into 8-bit bitstrings
	segments := make([]string, len(source))
	for i, curByte := range source {
		segments[i] = fmt.Sprintf("%08b", curByte)
	}

	// Trim off leading continuation bit from each segment
	for i, segment := range segments {
		segments[i] = segment[1:]
	}

	// Trim off the sign flag from the first segment
	firstSegment := []rune(segments[0])
	isNegative := firstSegment[0] == '1'
	segments[0] = string(firstSegment[1:])

	// Reverse the order of the segments.
	// Source: https://github.com/golang/go/wiki/SliceTricks#reversing
	for i := len(segments)/2 - 1; i >= 0; i-- {
		opp := len(segments) - 1 - i
		segments[i], segments[opp] = segments[opp], segments[i]
	}

	// Concat all the bits
	bitStringBuf := bytes.Buffer{}
	for _, segment := range segments {
		bitStringBuf.WriteString(segment)
	}
	bitString := bitStringBuf.String()

	// Add sign flag
	if isNegative {
		bitString = "-" + bitString
	}

	if _, ok := val.IntValue.SetString(bitString, 2); !ok {
		return errors.Errorf("failed to parse bit string %s to big.Int", bitString)
	}

	return nil
}
