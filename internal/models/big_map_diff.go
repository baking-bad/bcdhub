package models

// BigMapDiff -
type BigMapDiff struct {
	Ptr         int64       `json:"ptr,omitempty"`
	BinPath     string      `json:"bin_path"`
	Key         interface{} `json:"key"`
	KeyHash     string      `json:"key_hash"`
	Value       string      `json:"value"`
	OperationID string      `json:"operation_id"`
	Level       int64       `json:"level"`
}
