package forging

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"
)

type intForger struct{}

// Decode -
func (d intForger) Unforge(dec *decoder, code *strings.Builder) (int, error) {
	var buffer bytes.Buffer

	for {
		buf := make([]byte, 1)
		if _, err := dec.Read(buf); err != nil {
			break
		}
		buffer.Write(buf)
		if buf[0] < 127 {
			break
		}
	}

	ret, err := d.DecodeSigned(buffer.Bytes())
	if err != nil {
		return buffer.Len(), err
	}
	// log.Printf("[int Decode] data: %x, value: %v\n", buffer.Bytes(), ret)

	fmt.Fprintf(code, `{ "int": "%s" }`, ret)
	return buffer.Len(), nil
}

// DecodeSigned -
func (d intForger) DecodeSigned(source []byte) (string, error) {
	if len(source) == 0 {
		return "", errors.Errorf("expected non-empty byte array")
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

	// Convert from base 2 to base 10
	ret := new(big.Int)
	_, success := ret.SetString(bitString, 2)
	if !success {
		return "", errors.Errorf("failed to parse bit string %s to big.Int", bitString)
	}

	return fmt.Sprintf("%v", ret), nil
}
