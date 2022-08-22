package handlers

import (
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Error -
type Error struct {
	Message string `json:"message" example:"text"`
}

// Operation -
type Operation struct {
	ID                                 int64              `json:"id,omitempty" extensions:"x-nullable"`
	Level                              int64              `json:"level,omitempty" extensions:"x-nullable"`
	Fee                                int64              `json:"fee,omitempty" extensions:"x-nullable"`
	Counter                            int64              `json:"counter,omitempty" extensions:"x-nullable"`
	GasLimit                           int64              `json:"gas_limit,omitempty" extensions:"x-nullable"`
	StorageLimit                       int64              `json:"storage_limit,omitempty" extensions:"x-nullable"`
	Amount                             int64              `json:"amount,omitempty" extensions:"x-nullable"`
	Balance                            int64              `json:"balance,omitempty" extensions:"x-nullable"`
	Burned                             int64              `json:"burned,omitempty" extensions:"x-nullable"`
	AllocatedDestinationContractBurned int64              `json:"allocated_destination_contract_burned,omitempty" extensions:"x-nullable"`
	IndexedTime                        int64              `json:"-"`
	ContentIndex                       int64              `json:"content_index"`
	ConsumedGas                        int64              `json:"consumed_gas,omitempty" extensions:"x-nullable" example:"100"`
	StorageSize                        int64              `json:"storage_size,omitempty" extensions:"x-nullable" example:"200"`
	PaidStorageSizeDiff                int64              `json:"paid_storage_size_diff,omitempty" extensions:"x-nullable" example:"300"`
	Errors                             []*tezerrors.Error `json:"errors,omitempty" extensions:"x-nullable"`
	Parameters                         interface{}        `json:"parameters,omitempty" extensions:"x-nullable"`
	StorageDiff                        *ast.MiguelNode    `json:"storage_diff,omitempty" extensions:"x-nullable"`
	Event                              []*ast.MiguelNode  `json:"event,omitempty" extensions:"x-nullable"`
	RawMempool                         interface{}        `json:"rawMempool,omitempty" extensions:"x-nullable"`
	Timestamp                          time.Time          `json:"timestamp"`
	Protocol                           string             `json:"protocol"`
	Hash                               string             `json:"hash,omitempty" extensions:"x-nullable"`
	Network                            string             `json:"network"`
	Kind                               string             `json:"kind"`
	Source                             string             `json:"source,omitempty" extensions:"x-nullable"`
	SourceAlias                        string             `json:"source_alias,omitempty" extensions:"x-nullable"`
	Destination                        string             `json:"destination,omitempty" extensions:"x-nullable"`
	DestinationAlias                   string             `json:"destination_alias,omitempty" extensions:"x-nullable"`
	PublicKey                          string             `json:"public_key,omitempty" extensions:"x-nullable"`
	ManagerPubKey                      string             `json:"manager_pubkey,omitempty" extensions:"x-nullable"`
	Delegate                           string             `json:"delegate,omitempty" extensions:"x-nullable"`
	Status                             string             `json:"status"`
	Entrypoint                         string             `json:"entrypoint,omitempty" extensions:"x-nullable"`
	Tag                                string             `json:"tag,omitempty" extensions:"x-nullable"`
	AllocatedDestinationContract       bool               `json:"allocated_destination_contract,omitempty" extensions:"x-nullable" example:"true"`
	Internal                           bool               `json:"internal"`
	Mempool                            bool               `json:"mempool"`
	Storage                            stdJSON.RawMessage `json:"-"`
}

// FromModel -
func (o *Operation) FromModel(operation operation.Operation) {
	o.ID = operation.ID
	o.Hash = operation.Hash
	o.Internal = operation.Internal
	o.Timestamp = operation.Timestamp.UTC()

	o.Level = operation.Level
	o.Kind = operation.Kind.String()
	o.Source = operation.Source.Address
	o.Fee = operation.Fee
	if o.Hash != "" {
		o.Counter = operation.Counter
	}
	o.GasLimit = operation.GasLimit
	o.StorageLimit = operation.StorageLimit
	o.Amount = operation.Amount
	o.Destination = operation.Destination.Address
	o.Delegate = operation.Delegate.Address
	o.Status = operation.Status.String()
	o.Burned = operation.Burned
	o.Entrypoint = operation.Entrypoint.String()
	o.Tag = operation.Tag.String()
	o.ContentIndex = operation.ContentIndex
	o.AllocatedDestinationContractBurned = operation.AllocatedDestinationContractBurned
	o.ConsumedGas = operation.ConsumedGas
	o.StorageSize = operation.StorageSize
	o.PaidStorageSizeDiff = operation.PaidStorageSizeDiff
	o.AllocatedDestinationContract = operation.AllocatedDestinationContract
}

// ToModel -
func (o *Operation) ToModel() operation.Operation {
	return operation.Operation{
		ID:        o.ID,
		Hash:      o.Hash,
		Internal:  o.Internal,
		Timestamp: o.Timestamp,
		Level:     o.Level,
		Kind:      types.NewOperationKind(o.Kind),
		Source: account.Account{
			Address: o.Source,
			Type:    types.NewAccountType(o.Source),
		},
		Fee:          o.Fee,
		Counter:      o.Counter,
		GasLimit:     o.GasLimit,
		StorageLimit: o.StorageLimit,
		Amount:       o.Amount,
		Destination: account.Account{
			Address: o.Destination,
			Type:    types.NewAccountType(o.Destination),
		},
		Delegate: account.Account{
			Address: o.Delegate,
			Type:    types.NewAccountType(o.Delegate),
		},
		Status: types.NewOperationStatus(o.Status),
		Burned: o.Burned,
		Entrypoint: types.NullString{
			Str:   o.Entrypoint,
			Valid: o.Entrypoint != "",
		},
		AllocatedDestinationContract: o.AllocatedDestinationContract,
		ConsumedGas:                  o.ConsumedGas,
		StorageSize:                  o.StorageSize,
		PaidStorageSizeDiff:          o.PaidStorageSizeDiff,
	}
}

// Contract -
type Contract struct {
	ID        int64     `json:"id"`
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`

	Hash        string   `json:"hash,omitempty"`
	Tags        []string `json:"tags,omitempty" extensions:"x-nullable"`
	Hardcoded   []string `json:"hardcoded,omitempty" extensions:"x-nullable"`
	FailStrings []string `json:"fail_strings,omitempty" extensions:"x-nullable"`
	Annotations []string `json:"annotations,omitempty" extensions:"x-nullable"`
	Entrypoints []string `json:"entrypoints,omitempty" extensions:"x-nullable"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty" extensions:"x-nullable"`
	Delegate string `json:"delegate,omitempty" extensions:"x-nullable"`

	FoundBy         string    `json:"found_by,omitempty" extensions:"x-nullable"`
	LastAction      time.Time `json:"last_action,omitempty" extensions:"x-nullable"`
	TxCount         int64     `json:"tx_count,omitempty" extensions:"x-nullable"`
	MigrationsCount int64     `json:"migrations_count,omitempty" extensions:"x-nullable"`
	Alias           string    `json:"alias,omitempty" extensions:"x-nullable"`
	DelegateAlias   string    `json:"delegate_alias,omitempty" extensions:"x-nullable"`
	Slug            string    `json:"slug,omitempty" extensions:"x-nullable"`
}

// FromModel -
func (c *Contract) FromModel(contract contract.Contract) {
	c.Address = contract.Account.Address
	c.Alias = contract.Account.Alias
	c.Delegate = contract.Delegate.Address
	c.DelegateAlias = contract.Delegate.Alias
	c.TxCount = contract.TxCount
	c.LastAction = contract.LastAction

	c.Level = contract.Level
	c.Manager = contract.Manager.Address
	c.MigrationsCount = contract.MigrationsCount
	c.Tags = contract.Tags.ToArray()
	c.Timestamp = contract.Timestamp

	script := contract.Alpha
	switch {
	case contract.BabylonID > 0:
		script = contract.Babylon
	case contract.JakartaID > 0:
		script = contract.Jakarta
	}

	c.Hash = script.Hash
	c.FailStrings = script.FailStrings
	c.Annotations = script.Annotations
	c.Entrypoints = script.Entrypoints
	c.ID = contract.ID
}

// ContractWithStats -
type ContractWithStats struct {
	Contract

	SameCount   int64 `json:"same_count"`
	EventsCount int   `json:"events_count"`
}

// RecentlyCalledContract -
type RecentlyCalledContract struct {
	ID         int64     `json:"id"`
	Address    string    `json:"address"`
	Alias      string    `json:"alias,omitempty" extensions:"x-nullable"`
	LastAction time.Time `json:"last_action,omitempty" extensions:"x-nullable"`
	TxCount    int64     `json:"tx_count,omitempty" extensions:"x-nullable"`
}

// FromModel -
func (c *RecentlyCalledContract) FromModel(contract contract.Contract) {
	c.Address = contract.Account.Address
	c.Alias = contract.Account.Alias
	c.TxCount = contract.TxCount
	c.LastAction = contract.LastAction
	c.ID = contract.ID
}

// OperationResponse -
type OperationResponse struct {
	Operations []Operation `json:"operations"`
	LastID     string      `json:"last_id,omitempty" extensions:"x-nullable" example:"1588640276994159"`
}

// BigMapItem -
type BigMapItem struct {
	Key       interface{} `json:"key"`
	KeyHash   string      `json:"key_hash"`
	KeyString string      `json:"key_string"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
	IsActive  bool        `json:"is_active"`
}

// BigMapResponseItem -
type BigMapResponseItem struct {
	Item  BigMapItem `json:"data"`
	Count int64      `json:"count"`
}

// GetBigMapResponse -
type GetBigMapResponse struct {
	Address       string        `json:"address"`
	Network       string        `json:"network"`
	Ptr           int64         `json:"ptr"`
	ActiveKeys    uint          `json:"active_keys"`
	TotalKeys     uint          `json:"total_keys"`
	ContractAlias string        `json:"contract_alias,omitempty" extensions:"x-nullable"`
	Typedef       []ast.Typedef `json:"typedef,omitempty" extensions:"x-nullable"`
}

// Migration -
type Migration struct {
	Level        int64     `json:"level"`
	Timestamp    time.Time `json:"timestamp"`
	Hash         string    `json:"hash,omitempty" extensions:"x-nullable"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol"`
	Kind         string    `json:"kind"`
}

// TokenContract -
type TokenContract struct {
	Network       string                      `json:"network"`
	Level         int64                       `json:"level"`
	Timestamp     time.Time                   `json:"timestamp"`
	LastAction    time.Time                   `json:"last_action"`
	Address       string                      `json:"address"`
	Manager       string                      `json:"manager,omitempty" extensions:"x-nullable"`
	Delegate      string                      `json:"delegate,omitempty" extensions:"x-nullable"`
	Alias         string                      `json:"alias,omitempty" extensions:"x-nullable"`
	DelegateAlias string                      `json:"delegate_alias,omitempty" extensions:"x-nullable"`
	Type          string                      `json:"type"`
	Balance       int64                       `json:"balance"`
	TxCount       int64                       `json:"tx_count"`
	Methods       map[string]TokenMethodStats `json:"methods,omitempty" extensions:"x-nullable"`
}

// TokenMethodStats -
type TokenMethodStats struct {
	CallCount          int64 `json:"call_count"`
	AverageConsumedGas int64 `json:"average_consumed_gas"`
}

// PageableTokenContracts -
type PageableTokenContracts struct {
	Total  int64           `json:"total"`
	Tokens []TokenContract `json:"tokens"`
}

// TokenTransfer -
type TokenTransfer struct {
	Contract  string    `json:"contract"`
	Network   string    `json:"network"`
	Protocol  string    `json:"protocol"`
	Hash      string    `json:"hash"`
	Counter   int64     `json:"counter,omitempty" extensions:"x-nullable"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	From      string    `json:"from,omitempty" extensions:"x-nullable"`
	To        string    `json:"to"`
	Amount    int64     `json:"amount"`
	Source    string    `json:"source"`
	Nonce     *int64    `json:"nonce,omitempty" extensions:"x-nullable"`
}

// PageableTokenTransfers -
type PageableTokenTransfers struct {
	Transfers []TokenTransfer `json:"transfers"`
	LastID    string          `json:"last_id,omitempty" extensions:"x-nullable"`
}

// BigMapDiffItem -
type BigMapDiffItem struct {
	Value     interface{} `json:"value"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}

// BigMapDiffByKeyResponse -
type BigMapDiffByKeyResponse struct {
	Key     interface{}      `json:"key,omitempty" extensions:"x-nullable"`
	KeyHash string           `json:"key_hash"`
	Values  []BigMapDiffItem `json:"values,omitempty" extensions:"x-nullable"`
	Total   int64            `json:"total"`
}

// BigMapKeyStateResponse -
type BigMapKeyStateResponse struct {
	Key             interface{} `json:"key,omitempty" extensions:"x-nullable"`
	KeyHash         string      `json:"key_hash"`
	KeyString       string      `json:"key_string"`
	Value           interface{} `json:"value"`
	LastUpdateLevel int64       `json:"last_update_level"`
	LastUpdateTime  time.Time   `json:"last_update_time"`
	Removed         bool        `json:"removed"`
	UpdatesCount    int64       `json:"updates_count"`
}

// CodeDiffResponse -
type CodeDiffResponse struct {
	Left  CodeDiffLeg          `json:"left"`
	Right CodeDiffLeg          `json:"right"`
	Diff  formatter.DiffResult `json:"diff"`
}

// NetworkStats -
type NetworkStats struct {
	ContractsCount  int64      `json:"contracts_count" example:"10"`
	OperationsCount int64      `json:"operations_count" example:"100"`
	ContractCalls   uint64     `json:"contract_calls" example:"100"`
	UniqueContracts uint64     `json:"unique_contracts" example:"100"`
	FACount         uint64     `json:"fa_count" example:"100"`
	Protocols       []Protocol `json:"protocols"`
}

// SearchBigMapDiff -
type SearchBigMapDiff struct {
	Ptr       int64              `json:"ptr"`
	Key       string             `json:"key"`
	KeyHash   string             `json:"key_hash"`
	Value     stdJSON.RawMessage `json:"value"`
	Level     int64              `json:"level"`
	Address   string             `json:"address"`
	Network   string             `json:"network"`
	Timestamp time.Time          `json:"timestamp"`
	FoundBy   string             `json:"found_by"`
}

// EntrypointSchema ;
type EntrypointSchema struct {
	ast.EntrypointType
	Schema       *ast.JSONSchema `json:"schema"`
	DefaultModel ast.JSONModel   `json:"default_model,omitempty" extensions:"x-nullable"`
}

// GetErrorLocationResponse -
type GetErrorLocationResponse struct {
	Text        string `json:"text"`
	FailedRow   int    `json:"failed_row"`
	FirstRow    int    `json:"first_row"`
	StartColumn int    `json:"start_col"`
	EndColumn   int    `json:"end_col"`
}

// Alias -
type Alias struct {
	Alias   string `json:"alias" example:"Contract alias"`
	Network string `json:"network" example:"babylonnet"`
	Address string `json:"address" example:"KT1CeekjGVRc5ASmgWDc658NBExetoKNuiqz"`
	Slug    string `json:"slug" example:"contract_slug"`
}

// Protocol -
type Protocol struct {
	Hash       string `json:"hash" example:"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb"`
	StartLevel int64  `json:"start_level" example:"851969"`
	EndLevel   int64  `json:"end_level,omitempty" example:"0" extensions:"x-nullable"`
	Alias      string `json:"alias" example:"Carthage"`
}

// FromModel -
func (p *Protocol) FromModel(protocol protocol.Protocol) {
	p.Hash = protocol.Hash
	p.StartLevel = protocol.StartLevel
	p.EndLevel = protocol.EndLevel
	p.Alias = protocol.Alias
}

// Block -
type Block struct {
	Network                      string    `json:"network" example:"mainnet"`
	Hash                         string    `json:"hash" example:"BLyAEwaXShJuZasvUezHUfLqzZ48V8XrPvXF2wRaH15tmzEpsHT"`
	Level                        int64     `json:"level" example:"100"`
	Predecessor                  string    `json:"predecessor" example:"BMWVEwEYw9m5iaHzqxDfkPzZTV4rhkSouRh3DkVMVGkxZ3EVaNs"`
	ChainID                      string    `json:"chain_id" example:"NetXdQprcVkpaWU"`
	Timestamp                    time.Time `json:"timestamp" example:"2018-06-30T18:05:27Z"`
	Protocol                     string    `json:"protocol" example:"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY"`
	CostPerByte                  int64     `json:"cost_per_byte" example:"250"`
	HardGasLimitPerOperation     int64     `json:"hard_gas_limit_per_operation" example:"1040000"`
	HardStorageLimitPerOperation int64     `json:"hard_storage_limit_per_operation" example:"60000"`
	TimeBetweenBlocks            int64     `json:"time_between_blocks" example:"30"`
}

// FromModel -
func (b *Block) FromModel(block block.Block) {
	b.Hash = block.Hash
	b.Level = block.Level
	b.Protocol = block.Protocol.Hash
	b.ChainID = block.Protocol.ChainID
	b.Timestamp = block.Timestamp
	b.CostPerByte = block.Protocol.CostPerByte
	b.HardGasLimitPerOperation = block.Protocol.HardGasLimitPerOperation
	b.HardStorageLimitPerOperation = block.Protocol.HardStorageLimitPerOperation
	b.TimeBetweenBlocks = block.Protocol.TimeBetweenBlocks
}

// SameContractsResponse -
type SameContractsResponse struct {
	Count     int64               `json:"count"`
	Contracts []ContractWithStats `json:"contracts"`
}

// Series -
type Series [][]int64

// SeriesFloat -
type SeriesFloat [][]float64

// BigMapHistoryResponse -
type BigMapHistoryResponse struct {
	Address string              `json:"address"`
	Network string              `json:"network"`
	Ptr     int64               `json:"ptr"`
	Items   []BigMapHistoryItem `json:"items,omitempty" extensions:"x-nullable"`
}

// BigMapHistoryItem -
type BigMapHistoryItem struct {
	Action         string    `json:"action"`
	SourcePtr      *int64    `json:"source_ptr,omitempty" extensions:"x-nullable"`
	DestinationPtr *int64    `json:"destination_ptr,omitempty" extensions:"x-nullable"`
	Timestamp      time.Time `json:"timestamp"`
}

// ConfigResponse -
type ConfigResponse struct {
	Networks       []string          `json:"networks"`
	RPCEndpoints   map[string]string `json:"rpc_endpoints"`
	TzKTEndpoints  map[string]string `json:"tzkt_endpoints"`
	SentryDSN      string            `json:"sentry_dsn"`
	OauthEnabled   bool              `json:"oauth_enabled"`
	GaEnabled      bool              `json:"ga_enabled"`
	MempoolEnabled bool              `json:"mempool_enabled"`
	SandboxMode    bool              `json:"sandbox_mode"`
}

// Screenshot -
type Screenshot struct {
	Type string `json:"type"`
	Link string `json:"link"`
}

// AccountInfo -
type AccountInfo struct {
	Address    string    `json:"address"`
	Network    string    `json:"network"`
	Alias      string    `json:"alias,omitempty" extensions:"x-nullable"`
	Balance    int64     `json:"balance"`
	TxCount    int64     `json:"tx_count"`
	LastAction time.Time `json:"last_action"`
}

// CountResponse -
type CountResponse struct {
	Count int64 `json:"count"`
}

// MetadataResponse -
type MetadataResponse struct {
	Hash string `json:"hash"`
}

// ViewSchema ;
type ViewSchema struct {
	Type           []ast.Typedef   `json:"typedef"`
	Name           string          `json:"name"`
	Implementation int             `json:"implementation"`
	Description    string          `json:"description,omitempty"`
	Schema         *ast.JSONSchema `json:"schema"`
	DefaultModel   interface{}     `json:"default_model,omitempty" extensions:"x-nullable"`
	Error          string          `json:"error,omitempty"`
	Kind           ViewSchemaKind  `json:"kind"`
}

// ViewSchemaKind -
type ViewSchemaKind string

// ViewSchemaKind
const (
	OffchainView ViewSchemaKind = "off-chain"
	OnchainView  ViewSchemaKind = "on-chain"
	EmptyView    ViewSchemaKind = ""
)

// ForkResponse -
type ForkResponse struct {
	Script  stdJSON.RawMessage `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// HeadResponse -
type HeadResponse struct {
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"time"`
	Protocol  string    `json:"protocol"`
	Synced    bool      `json:"synced"`
}

// GlobalConstant -
type GlobalConstant struct {
	Timestamp time.Time          `json:"timestamp"`
	Level     int64              `json:"level"`
	Address   string             `json:"address"`
	Value     stdJSON.RawMessage `json:"value,omitempty"`
	Code      string             `json:"code,omitempty"`
}

// NewGlobalConstantFromModel -
func NewGlobalConstantFromModel(gc contract.GlobalConstant) GlobalConstant {
	return GlobalConstant{
		Timestamp: gc.Timestamp.UTC(),
		Level:     gc.Level,
		Address:   gc.Address,
		Value:     stdJSON.RawMessage(gc.Value),
	}
}

// CodeFromMichelsonResponse -
type CodeFromMichelsonResponse struct {
	Script  stdJSON.RawMessage       `json:"script"`
	Storage CodeFromMichelsonStorage `json:"storage"`
}

// CodeFromMichelsonStorage -
type CodeFromMichelsonStorage struct {
	Type         []ast.Typedef   `json:"type,omitempty" extensions:"x-nullable"`
	Schema       *ast.JSONSchema `json:"schema"`
	DefaultModel ast.JSONModel   `json:"default_model,omitempty" extensions:"x-nullable"`
}

// OPGResponse -
type OPGResponse struct {
	LastID       int64     `json:"last_id"`
	ContentIndex int64     `json:"content_index"`
	Counter      int64     `json:"counter"`
	Level        int64     `json:"level"`
	TotalCost    int64     `json:"total_cost"`
	Flow         int64     `json:"flow"`
	Internals    int       `json:"internals"`
	Hash         string    `json:"hash"`
	Entrypoint   string    `json:"entrypoint"`
	Timestamp    time.Time `json:"timestamp"`
	Status       string    `json:"status"`
	Kind         string    `json:"kind"`
}

// NewOPGResponse -
func NewOPGResponse(opg operation.OPG) OPGResponse {
	return OPGResponse{
		LastID:       opg.LastID,
		ContentIndex: opg.ContentIndex,
		Counter:      opg.Counter,
		Level:        opg.Level,
		TotalCost:    opg.TotalCost,
		Flow:         opg.Flow,
		Internals:    opg.Internals,
		Hash:         opg.Hash,
		Entrypoint:   opg.Entrypoint,
		Timestamp:    opg.Timestamp,
		Status:       opg.Status.String(),
		Kind:         opg.Kind.String(),
	}
}

// Event -
type Event struct {
	Hash      string            `json:"hash"`
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Level     int64             `json:"level"`
	Tag       string            `json:"tag"`
	Payload   []*ast.MiguelNode `json:"payload,omitempty" extensions:"x-nullable"`
}

// NewEvent -
func NewEvent(o operation.Operation) (*Event, error) {
	if !o.IsEvent() {
		return nil, nil
	}

	e := &Event{
		Hash:      o.Hash,
		Status:    o.Status.String(),
		Timestamp: o.Timestamp,
		Level:     o.Level,
		Tag:       o.Tag.String(),
	}

	eventType, err := ast.NewTypedAstFromBytes(o.EventType)
	if err != nil {
		return nil, err
	}
	if err := eventType.SettleFromBytes(o.EventPayload); err != nil {
		return nil, err
	}
	eventMiguel, err := eventType.ToMiguel()
	if err != nil {
		return nil, err
	}
	e.Payload = eventMiguel
	return e, nil
}
