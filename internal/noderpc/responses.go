package noderpc

import (
	stdJSON "encoding/json"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
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
	Balance   int64              `json:"balance,string"`
	Counter   int64              `json:"counter,string"`
	Manager   string             `json:"manager"`
	Delegate  struct {
		Value string `json:"value"`
	} `json:"delegate"`
}

// OperationGroup -
type OperationGroup struct {
	Protocol  string      `json:"protocol"`
	ChainID   string      `json:"chain_id"`
	Hash      string      `json:"hash"`
	Branch    string      `json:"branch"`
	Signature string      `json:"signature"`
	Contents  []Operation `json:"contents"`
}

// Operation -
type Operation struct {
	Kind          string             `json:"kind"`
	Source        string             `json:"source"`
	Destination   *string            `json:"destination,omitempty"`
	PublicKey     string             `json:"public_key,omitempty"`
	ManagerPubKey string             `json:"manager_pubkey,omitempty"`
	Delegate      string             `json:"delegate,omitempty"`
	Fee           int64              `json:"fee,string"`
	Counter       int64              `json:"counter,string"`
	Balance       *int64             `json:"balance,omitempty,string"`
	GasLimit      int64              `json:"gas_limit,string"`
	StorageLimit  int64              `json:"storage_limit,string"`
	Amount        *int64             `json:"amount,omitempty,string"`
	Nonce         *int64             `json:"nonce,omitempty"`
	Parameters    stdJSON.RawMessage `json:"parameters,omitempty"`
	Metadata      *OperationMetadata `json:"metadata,omitempty"`
	Result        *OperationResult   `json:"result,omitempty"`
	Script        stdJSON.RawMessage `json:"script,omitempty"`
}

// GetResult -
func (op Operation) GetResult() *OperationResult {
	switch {
	case op.Metadata != nil && op.Metadata.OperationResult != nil:
		return op.Metadata.OperationResult
	case op.Result != nil:
		return op.Result
	default:
		return nil
	}
}

// Script -
type Script struct {
	Code    *ast.Script        `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// OperationMetadata -
type OperationMetadata struct {
	OperationResult    *OperationResult `json:"operation_result,omitempty"`
	Internal           []Operation      `json:"internal_operation_results,omitempty"`
	InternalOperations []Operation      `json:"internal_operations,omitempty"`
}

// OperationResult -
type OperationResult struct {
	Status                       string             `json:"status"`
	Storage                      stdJSON.RawMessage `json:"storage,omitempty"`
	ConsumedGas                  int64              `json:"consumed_gas,string"`
	ConsumedMilligas             *int64             `json:"consumed_milligas,omitempty,string"`
	StorageSize                  *int64             `json:"storage_size,omitempty,string"`
	Originated                   []string           `json:"originated_contracts,omitempty"`
	PaidStorageSizeDiff          *int64             `json:"paid_storage_size_diff,omitempty,string"`
	AllocatedDestinationContract *bool              `json:"allocated_destination_contract,omitempty"`
	BigMapDiffs                  []BigMapDiff       `json:"big_map_diff,omitempty"`
	Errors                       stdJSON.RawMessage `json:"errors,omitempty"`
}

// BigMapDiff -
type BigMapDiff struct {
	Action       string             `json:"action"`
	KeyHash      string             `json:"key_hash"`
	BigMap       *int64             `json:"big_map,omitempty,string"`
	SourceBigMap *int64             `json:"source_big_map,omitempty,string"`
	DestBigMap   *int64             `json:"destination_big_map,omitempty,string"`
	Key          stdJSON.RawMessage `json:"key"`
	Value        stdJSON.RawMessage `json:"value,omitempty"`
	KeyType      stdJSON.RawMessage `json:"key_type,omitempty"`
	ValueType    stdJSON.RawMessage `json:"value_type,omitempty"`
}

// RunCodeResponse -
type RunCodeResponse struct {
	Operations  []Operation        `json:"operations"`
	Storage     stdJSON.RawMessage `json:"storage"`
	BigMapDiffs []BigMapDiff       `json:"big_map_diff,omitempty"`
}

// RunCodeError -
type RunCodeError struct {
	ID string `json:"id"`
}

// OperationError -
type OperationError struct {
	ID              string             `json:"id"`
	Kind            string             `json:"kind"`
	Contract        string             `json:"contract,omitempty"`
	Location        *int64             `json:"location,omitempty"`
	ExpectedForm    stdJSON.RawMessage `json:"expectedForm,omitempty"`
	WrongExpression stdJSON.RawMessage `json:"wrongExpression,omitempty"`
}
