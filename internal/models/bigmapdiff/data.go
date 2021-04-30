package bigmapdiff

import "fmt"

// Bucket -
type Bucket struct {
	BigMapDiff

	KeysCount int64
}

// Stats -
type Stats struct {
	Total    int64
	Active   int64
	Contract string
}

// OPG -
type OPG struct {
	Hash    string
	Counter int64
	Nonce   *int64
}

// HashKey -
func (opg OPG) HashKey() string {
	if opg.Nonce == nil {
		return fmt.Sprintf("%s_%v", opg.Hash, opg.Counter)
	}
	return fmt.Sprintf("%s_%v_%d", opg.Hash, opg.Counter, *opg.Nonce)
}
