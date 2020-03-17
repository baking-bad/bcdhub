package hash

import (
	"crypto/sha256"

	"github.com/baking-bad/bcdhub/internal/contractparser/pack"
	"github.com/btcsuite/btcutil/base58"

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
	payload := encodeBase58(blakeHash[:], prefix)

	return payload, nil
}

func encodeBase58(input, prefix []byte) string {
	payload := append(prefix, input...)
	cksum := checksum(payload)
	payload = append(payload, cksum...)

	return base58.Encode(payload)
}

func checksum(input []byte) []byte {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	return h2[:4]
}
