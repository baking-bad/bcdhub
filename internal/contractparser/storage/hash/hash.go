package hash

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/pack"
	"github.com/baking-bad/bcdhub/internal/tzbase58"

	"github.com/tidwall/gjson"
	"golang.org/x/crypto/blake2b"
)

// Key -
func Key(input gjson.Result) (string, error) {
	packed, err := pack.Micheline(input)
	if err != nil {
		return "", err
	}

	blakeHash := blake2b.Sum256(packed)
	prefix := []byte{0x0D, 0x2C, 0x40, 0x1B}

	return tzbase58.EncodeFromBytes(blakeHash[:], prefix), nil
}
