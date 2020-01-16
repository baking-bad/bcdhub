package models

// BigMapDiff -
type BigMapDiff struct {
	Network string      `json:"network"`
	Address string      `json:"address"`
	Level   int64       `json:"level"`
	Ptr     int64       `json:"ptr"`
	BinPath string      `json:"bin_path"`
	Key     interface{} `json:"key"`
	KeyHash string      `json:"key_hash"`
	Value   string      `json:"value"`
}
