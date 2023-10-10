package handlers

import (
	"encoding/hex"
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Error -
type Error struct {
	Message string `example:"text" json:"message"`
}

// Operation -
type Operation struct {
	ID                                 int64              `extensions:"x-nullable"     json:"id,omitempty"`
	Level                              int64              `extensions:"x-nullable"     json:"level,omitempty"`
	Fee                                int64              `extensions:"x-nullable"     json:"fee,omitempty"`
	Counter                            int64              `extensions:"x-nullable"     json:"counter,omitempty"`
	GasLimit                           int64              `extensions:"x-nullable"     json:"gas_limit,omitempty"`
	StorageLimit                       int64              `extensions:"x-nullable"     json:"storage_limit,omitempty"`
	Amount                             int64              `extensions:"x-nullable"     json:"amount,omitempty"`
	Balance                            int64              `extensions:"x-nullable"     json:"balance,omitempty"`
	Burned                             int64              `extensions:"x-nullable"     json:"burned,omitempty"`
	AllocatedDestinationContractBurned int64              `extensions:"x-nullable"     json:"allocated_destination_contract_burned,omitempty"`
	IndexedTime                        int64              `json:"-"`
	ContentIndex                       int64              `json:"content_index"`
	ConsumedGas                        int64              `example:"100"               extensions:"x-nullable"                                json:"consumed_gas,omitempty"`
	StorageSize                        int64              `example:"200"               extensions:"x-nullable"                                json:"storage_size,omitempty"`
	PaidStorageSizeDiff                int64              `example:"300"               extensions:"x-nullable"                                json:"paid_storage_size_diff,omitempty"`
	TicketUpdatesCount                 int                `json:"ticket_updates_count"`
	BigMapDiffsCount                   int                `json:"big_map_diffs_count"`
	Errors                             []*tezerrors.Error `extensions:"x-nullable"     json:"errors,omitempty"`
	Parameters                         interface{}        `extensions:"x-nullable"     json:"parameters,omitempty"`
	StorageDiff                        *ast.MiguelNode    `extensions:"x-nullable"     json:"storage_diff,omitempty"`
	Payload                            []*ast.MiguelNode  `extensions:"x-nullable"     json:"payload,omitempty"`
	RawMempool                         interface{}        `extensions:"x-nullable"     json:"rawMempool,omitempty"`
	Timestamp                          time.Time          `json:"timestamp"`
	Protocol                           string             `json:"protocol"`
	Hash                               string             `extensions:"x-nullable"     json:"hash,omitempty"`
	Network                            string             `json:"network"`
	Kind                               string             `json:"kind"`
	Source                             string             `extensions:"x-nullable"     json:"source,omitempty"`
	Destination                        string             `extensions:"x-nullable"     json:"destination,omitempty"`
	PublicKey                          string             `extensions:"x-nullable"     json:"public_key,omitempty"`
	ManagerPubKey                      string             `extensions:"x-nullable"     json:"manager_pubkey,omitempty"`
	Delegate                           string             `extensions:"x-nullable"     json:"delegate,omitempty"`
	Status                             string             `json:"status"`
	Entrypoint                         string             `extensions:"x-nullable"     json:"entrypoint,omitempty"`
	Tag                                string             `extensions:"x-nullable"     json:"tag,omitempty"`
	AllocatedDestinationContract       bool               `example:"true"              extensions:"x-nullable"                                json:"allocated_destination_contract,omitempty"`
	Internal                           bool               `json:"internal"`
	Mempool                            bool               `json:"mempool"`
	Storage                            stdJSON.RawMessage `json:"-"`
}

// FromModel -
func (o *Operation) FromModel(operation operation.Operation) {
	o.ID = operation.ID
	if len(operation.Hash) > 0 {
		o.Hash = encoding.MustEncodeOperationHash(operation.Hash)
	}
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
	o.TicketUpdatesCount = operation.TicketUpdatesCount
	o.BigMapDiffsCount = operation.BigMapDiffsCount
}

// ToModel -
func (o *Operation) ToModel() operation.Operation {
	var hash []byte
	if o.Hash != "" {
		hash = encoding.MustDecodeBase58(o.Hash)
	}
	return operation.Operation{
		ID:        o.ID,
		Hash:      hash,
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
	Tags        []string `extensions:"x-nullable" json:"tags,omitempty"`
	Hardcoded   []string `extensions:"x-nullable" json:"hardcoded,omitempty"`
	FailStrings []string `extensions:"x-nullable" json:"fail_strings,omitempty"`
	Annotations []string `extensions:"x-nullable" json:"annotations,omitempty"`
	Entrypoints []string `extensions:"x-nullable" json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `extensions:"x-nullable" json:"manager,omitempty"`
	Delegate string `extensions:"x-nullable" json:"delegate,omitempty"`

	FoundBy         string    `extensions:"x-nullable" json:"found_by,omitempty"`
	LastAction      time.Time `extensions:"x-nullable" json:"last_action,omitempty"`
	TxCount         int64     `extensions:"x-nullable" json:"tx_count,omitempty"`
	MigrationsCount int64     `extensions:"x-nullable" json:"migrations_count,omitempty"`
	Slug            string    `extensions:"x-nullable" json:"slug,omitempty"`
}

// FromModel -
func (c *Contract) FromModel(contract contract.Contract) {
	c.Address = contract.Account.Address
	c.Delegate = contract.Delegate.Address
	c.Level = contract.Level
	c.Manager = contract.Manager.Address
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

	SameCount int64 `json:"same_count"`
}

// RecentlyCalledContract -
type RecentlyCalledContract struct {
	ID              int64     `json:"id"`
	Address         string    `json:"address"`
	LastAction      time.Time `json:"last_action"`
	OperationsCount int64     `json:"operations_count"`
}

// FromModel -
func (c *RecentlyCalledContract) FromModel(account account.Account) {
	c.Address = account.Address
	c.ID = account.ID
	c.LastAction = account.LastAction
	c.OperationsCount = account.OperationsCount
}

// OperationResponse -
type OperationResponse struct {
	Operations []Operation `json:"operations"`
	LastID     string      `example:"1588640276994159" extensions:"x-nullable" json:"last_id,omitempty"`
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
	Address    string        `json:"address"`
	Network    string        `json:"network"`
	Ptr        int64         `json:"ptr"`
	ActiveKeys uint          `json:"active_keys"`
	TotalKeys  uint          `json:"total_keys"`
	Typedef    []ast.Typedef `extensions:"x-nullable" json:"typedef,omitempty"`
}

// Migration -
type Migration struct {
	Level        int64     `json:"level"`
	Timestamp    time.Time `json:"timestamp"`
	Hash         string    `extensions:"x-nullable" json:"hash,omitempty"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol"`
	Kind         string    `json:"kind"`
}

// BigMapDiffItem -
type BigMapDiffItem struct {
	Value     interface{} `json:"value"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}

// BigMapDiffByKeyResponse -
type BigMapDiffByKeyResponse struct {
	Key     interface{}      `extensions:"x-nullable" json:"key,omitempty"`
	KeyHash string           `json:"key_hash"`
	Values  []BigMapDiffItem `extensions:"x-nullable" json:"values,omitempty"`
	Total   int64            `json:"total"`
}

// BigMapKeyStateResponse -
type BigMapKeyStateResponse struct {
	Key             interface{} `extensions:"x-nullable"  json:"key,omitempty"`
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
	ContractsCount  int64      `example:"10"     json:"contracts_count"`
	OperationsCount int64      `example:"100"    json:"operations_count"`
	ContractCalls   uint64     `example:"100"    json:"contract_calls"`
	UniqueContracts uint64     `example:"100"    json:"unique_contracts"`
	FACount         uint64     `example:"100"    json:"fa_count"`
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
	DefaultModel ast.JSONModel   `extensions:"x-nullable" json:"default_model,omitempty"`
}

// GetErrorLocationResponse -
type GetErrorLocationResponse struct {
	Text        string `json:"text"`
	FailedRow   int    `json:"failed_row"`
	FirstRow    int    `json:"first_row"`
	StartColumn int    `json:"start_col"`
	EndColumn   int    `json:"end_col"`
}

// Protocol -
type Protocol struct {
	Hash       string `example:"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb" json:"hash"`
	StartLevel int64  `example:"851969"                                              json:"start_level"`
	EndLevel   int64  `example:"0"                                                   extensions:"x-nullable" json:"end_level,omitempty"`
	Alias      string `example:"Carthage"                                            json:"alias"`
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
	Network                      string    `example:"mainnet"                                             json:"network"`
	Hash                         string    `example:"BLyAEwaXShJuZasvUezHUfLqzZ48V8XrPvXF2wRaH15tmzEpsHT" json:"hash"`
	Level                        int64     `example:"100"                                                 json:"level"`
	Predecessor                  string    `example:"BMWVEwEYw9m5iaHzqxDfkPzZTV4rhkSouRh3DkVMVGkxZ3EVaNs" json:"predecessor"`
	ChainID                      string    `example:"NetXdQprcVkpaWU"                                     json:"chain_id"`
	Timestamp                    time.Time `example:"2018-06-30T18:05:27Z"                                json:"timestamp"`
	Protocol                     string    `example:"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY" json:"protocol"`
	CostPerByte                  int64     `example:"250"                                                 json:"cost_per_byte"`
	HardGasLimitPerOperation     int64     `example:"1040000"                                             json:"hard_gas_limit_per_operation"`
	HardStorageLimitPerOperation int64     `example:"60000"                                               json:"hard_storage_limit_per_operation"`
	TimeBetweenBlocks            int64     `example:"30"                                                  json:"time_between_blocks"`

	Stats *Stats `json:"stats,omitempty"`
}

// FromModel -
func (b *Block) FromModel(block block.Block) {
	b.Hash = block.Hash
	b.Level = block.Level
	b.Protocol = block.Protocol.Hash
	b.ChainID = block.Protocol.ChainID
	b.Timestamp = block.Timestamp
	if block.Protocol.Constants != nil {
		b.CostPerByte = block.Protocol.CostPerByte
		b.HardGasLimitPerOperation = block.Protocol.HardGasLimitPerOperation
		b.HardStorageLimitPerOperation = block.Protocol.HardStorageLimitPerOperation
		b.TimeBetweenBlocks = block.Protocol.TimeBetweenBlocks
	}
}

type Stats struct {
	ContractsCount              int `json:"contracts_count"`
	SmartRollupsCount           int `json:"smart_rollups_count"`
	GlobalConstantsCount        int `json:"global_constants_count"`
	OperationsCount             int `json:"operations_count"`
	EventsCount                 int `json:"events_count"`
	TransactionsCount           int `json:"tx_count"`
	OriginationsCount           int `json:"originations_count"`
	SrOriginationsCount         int `json:"sr_originations_count"`
	SrExecutesCount             int `json:"sr_executes_count"`
	RegisterGlobalConstantCount int `json:"register_global_constants_count"`
	TransferTicketsCount        int `json:"transfer_tickets_count"`
}

func NewStats(s stats.Stats) *Stats {
	return &Stats{
		ContractsCount:              s.ContractsCount,
		SmartRollupsCount:           s.SmartRollupsCount,
		GlobalConstantsCount:        s.GlobalConstantsCount,
		OperationsCount:             s.OperationsCount,
		EventsCount:                 s.EventsCount,
		TransactionsCount:           s.TransactionsCount,
		OriginationsCount:           s.OriginationsCount,
		SrOriginationsCount:         s.SrOriginationsCount,
		SrExecutesCount:             s.SrExecutesCount,
		RegisterGlobalConstantCount: s.RegisterGlobalConstantCount,
		TransferTicketsCount:        s.TransferTicketsCount,
	}
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
	Ptr     int64               `json:"ptr"`
	Items   []BigMapHistoryItem `extensions:"x-nullable" json:"items,omitempty"`
}

// BigMapHistoryItem -
type BigMapHistoryItem struct {
	Action         string    `json:"action"`
	SourcePtr      *int64    `extensions:"x-nullable" json:"source_ptr,omitempty"`
	DestinationPtr *int64    `extensions:"x-nullable" json:"destination_ptr,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	Level          int64     `json:"level"`
}

// ConfigResponse -
type ConfigResponse struct {
	Networks       []string          `json:"networks"`
	RPCEndpoints   map[string]string `json:"rpc_endpoints"`
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
	Address            string    `json:"address"`
	Balance            int64     `json:"balance"`
	OperationsCount    int64     `json:"operations_count"`
	MigrationsCount    int64     `json:"migrations_count"`
	EventsCount        int64     `json:"events_count"`
	TicketUpdatesCount int64     `json:"ticket_updates_count"`
	LastAction         time.Time `json:"last_action"`
	AccountType        string    `json:"account_type"`
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
	DefaultModel   interface{}     `extensions:"x-nullable"      json:"default_model,omitempty"`
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

// Head -
type Head struct {
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"time"`
	Protocol  string    `json:"protocol"`
	Synced    bool      `json:"synced"`

	network types.Network `json:"-"`
}

// Heads -
type HeadsByNetwork []Head

// Len -
func (a HeadsByNetwork) Len() int { return len(a) }

// Swap -
func (a HeadsByNetwork) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less -
func (a HeadsByNetwork) Less(i, j int) bool { return a[i].network < a[j].network }

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
	Type         []ast.Typedef   `extensions:"x-nullable" json:"type,omitempty"`
	Schema       *ast.JSONSchema `json:"schema"`
	DefaultModel ast.JSONModel   `extensions:"x-nullable" json:"default_model,omitempty"`
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
	var hash string
	if len(opg.Hash) > 0 {
		hash = encoding.MustEncodeOperationHash(opg.Hash)
	}
	return OPGResponse{
		LastID:       opg.LastID,
		ContentIndex: opg.ContentIndex,
		Counter:      opg.Counter,
		Level:        opg.Level,
		TotalCost:    opg.TotalCost,
		Flow:         opg.Flow,
		Internals:    opg.Internals,
		Hash:         hash,
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
	Payload   []*ast.MiguelNode `extensions:"x-nullable" json:"payload,omitempty"`
}

// NewEvent -
func NewEvent(o operation.Operation) (*Event, error) {
	if o.Kind != types.OperationKindEvent {
		return nil, nil
	}

	var hash string
	if len(o.Hash) > 0 {
		hash = encoding.MustEncodeOperationHash(o.Hash)
	}

	e := &Event{
		Hash:      hash,
		Status:    o.Status.String(),
		Timestamp: o.Timestamp,
		Level:     o.Level,
		Tag:       o.Tag.String(),
	}

	eventType, err := ast.NewTypedAstFromBytes(o.PayloadType)
	if err != nil {
		return nil, err
	}
	if err := eventType.SettleFromBytes(o.Payload); err != nil {
		return nil, err
	}
	eventMiguel, err := eventType.ToMiguel()
	if err != nil {
		return nil, err
	}
	e.Payload = eventMiguel
	return e, nil
}

// TicketUpdate -
type TicketUpdate struct {
	ID            int64           `json:"id"`
	Level         int64           `json:"level"`
	Timestamp     time.Time       `json:"timestamp"`
	Ticketer      string          `json:"ticketer"`
	Address       string          `json:"address"`
	Amount        string          `json:"amount"`
	OperationHash string          `json:"operation_hash"`
	ContentType   []ast.Typedef   `json:"content_type"`
	Content       *ast.MiguelNode `json:"content,omitempty"`
}

// NewTicketUpdateFromModel -
func NewTicketUpdateFromModel(update ticket.TicketUpdate) TicketUpdate {
	return TicketUpdate{
		ID:        update.ID,
		Timestamp: update.Timestamp.UTC(),
		Level:     update.Level,
		Ticketer:  update.Ticket.Ticketer.Address,
		Address:   update.Account.Address,
		Amount:    update.Amount.String(),
	}
}

// SmartRollup -
type SmartRollup struct {
	ID                    int64         `json:"id"`
	Level                 int64         `json:"level"`
	Timestamp             time.Time     `json:"timestamp"`
	Size                  uint64        `json:"size"`
	Address               string        `json:"address"`
	GenesisCommitmentHash string        `json:"genesis_commitment_hash"`
	PvmKind               string        `json:"pvm_kind"`
	Kernel                string        `json:"kernel"`
	Type                  []ast.Typedef `json:"type"`
}

// NewSmartRollup -
func NewSmartRollup(rollup smartrollup.SmartRollup) SmartRollup {
	kernel := hex.EncodeToString(rollup.Kernel)
	return SmartRollup{
		ID:                    rollup.ID,
		Level:                 rollup.Level,
		Timestamp:             rollup.Timestamp,
		Size:                  rollup.Size,
		Address:               rollup.Address.Address,
		GenesisCommitmentHash: rollup.GenesisCommitmentHash,
		PvmKind:               rollup.PvmKind,
		Kernel:                kernel,
	}
}

type TicketBalance struct {
	Ticketer    string          `json:"ticketer"`
	Amount      string          `json:"amount"`
	ContentType []ast.Typedef   `json:"content_type"`
	Content     *ast.MiguelNode `json:"content,omitempty"`
	TicketId    int64           `json:"ticket_id"`
}

func NewTicketBalance(balance ticket.Balance) TicketBalance {
	return TicketBalance{
		Ticketer: balance.Ticket.Ticketer.Address,
		Amount:   balance.Amount.String(),
		TicketId: balance.TicketId,
	}
}

type GlobalConstantItem struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      int64     `json:"level"`
	Address    string    `json:"address"`
	LinksCount uint64    `json:"links_count"`
}

func NewGlobalConstantItem(item contract.ListGlobalConstantItem) GlobalConstantItem {
	return GlobalConstantItem{
		Timestamp:  item.Timestamp,
		Level:      item.Level,
		Address:    item.Address,
		LinksCount: item.LinksCount,
	}
}
