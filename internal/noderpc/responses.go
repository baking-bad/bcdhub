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
	MinimalBlockDelay            *int64           `json:"minimal_block_delay,omitempty,string"`
}

// BlockDelay -
func (c Constants) BlockDelay() int64 {
	switch {
	case c.MinimalBlockDelay != nil:
		return *c.MinimalBlockDelay
	case len(c.TimeBetweenBlocks) > 0:
		return c.TimeBetweenBlocks[0]
	default:
		return 30
	}
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
	Kind         string             `json:"kind"`
	Source       string             `json:"source"`
	Destination  *string            `json:"destination,omitempty"`
	Delegate     string             `json:"delegate,omitempty"`
	Fee          int64              `json:"fee,string"`
	Counter      int64              `json:"counter,string"`
	Balance      *int64             `json:"balance,omitempty,string"`
	GasLimit     int64              `json:"gas_limit,string"`
	StorageLimit int64              `json:"storage_limit,string"`
	Amount       *int64             `json:"amount,omitempty,string"`
	Nonce        *int64             `json:"nonce,omitempty"`
	Parameters   stdJSON.RawMessage `json:"parameters,omitempty"`
	Metadata     *OperationMetadata `json:"metadata,omitempty"`
	Result       *OperationResult   `json:"result,omitempty"`
	Script       stdJSON.RawMessage `json:"script,omitempty"`
	Value        stdJSON.RawMessage `json:"value,omitempty"`
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

// LightOperationGroup -
type LightOperationGroup struct {
	Protocol  string           `json:"protocol"`
	ChainID   string           `json:"chain_id"`
	Hash      string           `json:"hash"`
	Branch    string           `json:"branch"`
	Signature string           `json:"signature"`
	Contents  []LightOperation `json:"contents"`
}

// LightOperation -
type LightOperation struct {
	Raw         stdJSON.RawMessage `json:"-"`
	Kind        string             `json:"kind"`
	Source      string             `json:"source"`
	Destination *string            `json:"destination,omitempty"`
}

// UnmarshalJSON -
func (op *LightOperation) UnmarshalJSON(data []byte) error {
	op.Raw = data
	type buf LightOperation
	return json.Unmarshal(data, (*buf)(op))
}

// Script -
type Script struct {
	Code    *ast.Script        `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// GetSettledStorage -
func (s *Script) GetSettledStorage() (*ast.TypedAst, error) {
	typ, err := s.Code.StorageType()
	if err != nil {
		return nil, err
	}
	var data ast.UntypedAST
	if err := json.Unmarshal(s.Storage, &data); err != nil {
		return nil, err
	}
	err = typ.Settle(data)
	return typ, err
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
	GlobalAddress                string             `json:"global_address,omitempty"`
	LazyStorageDiff              []LazyStorageDiff  `json:"lazy_storage_diff,omitempty"`
	OriginatedRollup             string             `json:"originated_rollup,omitempty"`
}

// LazyStorageDiff -
type LazyStorageDiff struct {
	LazyStorageDiffKind
	Diff *Diff `json:"omitempty"`
}

// LazyStorageDiffKind -
type LazyStorageDiffKind struct {
	Kind string             `json:"kind"`
	ID   int64              `json:"id,string"`
	Raw  stdJSON.RawMessage `json:"diff,omitempty"`
}

// UnmarshalJSON -
func (lsd *LazyStorageDiff) UnmarshalJSON(data []byte) error {
	lsd.LazyStorageDiffKind = LazyStorageDiffKind{}
	if err := json.Unmarshal(data, &lsd.LazyStorageDiffKind); err != nil {
		return err
	}

	switch lsd.Kind {
	case "big_map":
		lsd.Diff = &Diff{
			BigMap: new(LazyBigMapDiff),
		}
		return json.Unmarshal(lsd.Raw, lsd.Diff.BigMap)
	case "sapling_state":
		lsd.Diff = &Diff{
			SaplingState: new(LazySaplingStateDiff),
		}
		return json.Unmarshal(lsd.Raw, lsd.Diff.SaplingState)
	}
	return nil
}

// Diff -
type Diff struct {
	BigMap       *LazyBigMapDiff
	SaplingState *LazySaplingStateDiff
}

// LazySaplingStateDiff -
type LazySaplingStateDiff struct {
	Action   string                 `json:"action"`
	Updates  LazySaplingStateUpdate `json:"updates"`
	Source   *int64                 `json:"source,omitempty,string"`
	MemoSize *int64                 `json:"memo_size,omitempty"`
}

// LazyBigMapUpdate -
type LazySaplingStateUpdate struct {
	CommitmentsAndCiphertexts []CommitmentsAndCiphertexts `json:"commitments_and_ciphertexts"`
	Nullifiers                []string                    `json:"nullifiers"`
}

// CommitmentsAndCiphertexts -
type CommitmentsAndCiphertexts struct {
	Commitment string
	CipherText CipherText
}

// UnmarshalJSON -
func (c *CommitmentsAndCiphertexts) UnmarshalJSON(data []byte) error {
	buf := []any{&c.Commitment, &c.CipherText}
	return json.Unmarshal(data, &buf)
}

// CipherText -
type CipherText struct {
	CV         string `json:"cv"`
	EPK        string `json:"epk"`
	PayloadEnc string `json:"payload_enc"`
	NonceEnc   string `json:"nonce_enc"`
	PayloadOut string `json:"payload_out"`
	NonceOut   string `json:"nonce_out"`
}

// LazyBigMapDiff -
type LazyBigMapDiff struct {
	Action    string             `json:"action"`
	Updates   []LazyBigMapUpdate `json:"updates"`
	Source    *int64             `json:"source,omitempty,string"`
	KeyType   stdJSON.RawMessage `json:"key_type,omitempty"`
	ValueType stdJSON.RawMessage `json:"value_type,omitempty"`
}

// LazyBigMapUpdate -
type LazyBigMapUpdate struct {
	KeyHash string             `json:"key_hash"`
	Key     stdJSON.RawMessage `json:"key"`
	Value   stdJSON.RawMessage `json:"value,omitempty"`
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
	Operations       []Operation        `json:"operations"`
	Storage          stdJSON.RawMessage `json:"storage"`
	LazyStorageDiffs []LazyStorageDiff  `json:"lazy_storage_diff,omitempty"`
}

// RunCodeError -
type RunCodeError struct {
	ID   string `json:"id"`
	Kind string `json:"kind,omitempty"`
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

// BigMap -
type BigMap struct {
	KeyType    *ast.TypedAst `json:"key_type"`
	ValueType  *ast.TypedAst `json:"value_type"`
	TotalBytes uint64        `json:"total_bytes,string"`
}

// Metadata -
type Metadata struct {
	Protocol        string `json:"protocol"`
	NextProtocol    string `json:"next_protocol"`
	TestChainStatus struct {
		Status string `json:"status"`
	} `json:"test_chain_status"`
	MaxOperationsTTL       int `json:"max_operations_ttl"`
	MaxOperationDataLength int `json:"max_operation_data_length"`
	MaxBlockHeaderLength   int `json:"max_block_header_length"`
	MaxOperationListLength []struct {
		MaxSize int `json:"max_size"`
		MaxOp   int `json:"max_op,omitempty"`
	} `json:"max_operation_list_length"`
	Baker            string    `json:"baker"`
	LevelInfo        LevelInfo `json:"level_info"`
	VotingPeriodInfo struct {
		VotingPeriod struct {
			Index         int    `json:"index"`
			Kind          string `json:"kind"`
			StartPosition int    `json:"start_position"`
		} `json:"voting_period"`
		Position  int `json:"position"`
		Remaining int `json:"remaining"`
	} `json:"voting_period_info"`
	NonceHash                 string                     `json:"nonce_hash"`
	ConsumedGas               string                     `json:"consumed_gas"`
	Deactivated               []interface{}              `json:"deactivated"`
	BalanceUpdates            []BalanceUpdate            `json:"balance_updates"`
	LiquidityBakingEscapeEma  int                        `json:"liquidity_baking_escape_ema"`
	ImplicitOperationsResults []ImplicitOperationsResult `json:"implicit_operations_results"`
}

// BalanceUpdate -
type BalanceUpdate struct {
	Kind     string `json:"kind"`
	Contract string `json:"contract,omitempty"`
	Change   string `json:"change"`
	Origin   string `json:"origin"`
	Category string `json:"category,omitempty"`
	Delegate string `json:"delegate,omitempty"`
	Cycle    int64  `json:"cycle,omitempty"`
}

// ImplicitOperationsResult -
type ImplicitOperationsResult struct {
	Kind                string             `json:"kind"`
	BalanceUpdates      []BalanceUpdate    `json:"balance_updates"`
	OriginatedContracts []string           `json:"originated_contracts,omitempty"`
	StorageSize         int64              `json:"storage_size,string"`
	PaidStorageSizeDiff int64              `json:"paid_storage_size_diff,string"`
	Storage             stdJSON.RawMessage `json:"storage,omitempty"`
	ConsumedGas         int64              `json:"consumed_gas,string,omitempty"`
	ConsumedMilligas    int64              `json:"consumed_milligas,string,omitempty"`
}

// LevelInfo -
type LevelInfo struct {
	Level              int64 `json:"level"`
	LevelPosition      int64 `json:"level_position"`
	Cycle              int64 `json:"cycle"`
	CyclePosition      int64 `json:"cycle_position"`
	ExpectedCommitment bool  `json:"expected_commitment"`
}
