package base

import (
	"fmt"
	"math/big"
)

// BigInt -
type BigInt struct {
	*big.Int
}

// NewBigInt -
func NewBigInt(val int64) *BigInt {
	return &BigInt{
		Int: big.NewInt(val),
	}
}

// MarshalJSON -
func (b *BigInt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, b.String())), nil
}

// UnmarshalJSON -
func (b *BigInt) UnmarshalJSON(p []byte) error {
	if string(p) == `null` {
		return nil
	}
	z := big.NewInt(0)
	if len(p) > 2 && p[0] == '"' && p[len(p)-1] == '"' { // trim quotes
		p = p[1 : len(p)-1]
	}
	if _, ok := z.SetString(string(p), 10); !ok {
		return fmt.Errorf("not a valid big integer: %s", p)
	}
	b.Int = z
	return nil
}
