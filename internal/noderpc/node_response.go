package noderpc

import (
	stdJSON "encoding/json"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

// Header is a header in a block returned by the Tezos RPC API.
type Header struct {
	Level       int64     `json:"level"`
	Protocol    string    `json:"protocol"`
	Timestamp   time.Time `json:"timestamp"`
	ChainID     string    `json:"chain_id"`
	Hash        string    `json:"hash"`
	Predecessor string    `json:"predecessor"`
}

// Int64StringSlice -
type Int64StringSlice []int64

// UnmarshalJSON -
func (slice *Int64StringSlice) UnmarshalJSON(data []byte) error {
	s := make([]string, 0)
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*slice = make([]int64, len(s))
	for i := range s {
		value, err := strconv.ParseInt(s[i], 10, 64)
		if err != nil {
			return err
		}
		(*slice)[i] = value
	}
	return nil
}

// Constants -
type Constants struct {
	CostPerByte                  int64            `json:"cost_per_byte,string"`
	HardGasLimitPerOperation     int64            `json:"hard_gas_limit_per_operation,string"`
	HardStorageLimitPerOperation int64            `json:"hard_storage_limit_per_operation,string"`
	TimeBetweenBlocks            Int64StringSlice `json:"time_between_blocks"`
}

// ContractData -
type ContractData struct {
	RawScript stdJSON.RawMessage `json:"script"`
	Script    gjson.Result       `json:"-"`
	Balance   int64              `json:"balance"`
	Counter   int64              `json:"counter"`
	Manager   string             `json:"manager"`
	Delegate  struct {
		Value string `json:"value"`
	} `json:"delegate"`
}
