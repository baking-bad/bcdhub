package types

import "time"

// BigMapDiff -
type BigMapDiff struct {
	Ptr   int64
	Key   []byte
	Value []byte

	ID               int64
	KeyHash          string
	OperationHash    string
	OperationCounter int64
	OperationNonce   *int64
	Level            int64
	Address          string
	Network          string
	IndexedTime      int64
	Timestamp        time.Time
	Protocol         int64
}
