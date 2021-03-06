package handlers

import (
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Error -
type Error struct {
	Message string `json:"message" example:"text"`
}

// Operation -
type Operation struct {
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
	Errors                             []*tezerrors.Error `json:"errors,omitempty" extensions:"x-nullable"`
	Result                             *OperationResult   `json:"result,omitempty" extensions:"x-nullable"`
	Parameters                         interface{}        `json:"parameters,omitempty" extensions:"x-nullable"`
	StorageDiff                        interface{}        `json:"storage_diff,omitempty" extensions:"x-nullable"`
	RawMempool                         interface{}        `json:"rawMempool,omitempty" extensions:"x-nullable"`
	Timestamp                          time.Time          `json:"timestamp"`
	ID                                 string             `json:"id,omitempty" extensions:"x-nullable"`
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
	Internal                           bool               `json:"internal"`
	Mempool                            bool               `json:"mempool"`
}

// FromModel -
func (o *Operation) FromModel(operation operation.Operation) {
	o.ID = operation.ID
	o.Protocol = operation.Protocol
	o.Hash = operation.Hash
	o.Network = operation.Network
	o.Internal = operation.Internal
	o.Timestamp = operation.Timestamp

	o.Level = operation.Level
	o.Kind = operation.Kind
	o.Source = operation.Source
	o.SourceAlias = operation.SourceAlias
	o.Fee = operation.Fee
	o.Counter = operation.Counter
	o.GasLimit = operation.GasLimit
	o.StorageLimit = operation.StorageLimit
	o.Amount = operation.Amount
	o.Destination = operation.Destination
	o.DestinationAlias = operation.DestinationAlias
	o.PublicKey = operation.PublicKey
	o.ManagerPubKey = operation.ManagerPubKey
	o.Delegate = operation.Delegate
	o.Status = operation.Status
	o.Burned = operation.Burned
	o.Entrypoint = operation.Entrypoint
	o.IndexedTime = operation.IndexedTime
	o.ContentIndex = operation.ContentIndex
	o.AllocatedDestinationContractBurned = operation.AllocatedDestinationContractBurned
}

// ToModel -
func (o *Operation) ToModel() operation.Operation {
	var result *operation.Result
	if o.Result != nil {
		result = o.Result.ToModel()
	}
	return operation.Operation{
		ID:        o.ID,
		Protocol:  o.Protocol,
		Hash:      o.Hash,
		Network:   o.Network,
		Internal:  o.Internal,
		Timestamp: o.Timestamp,

		Level:            o.Level,
		Kind:             o.Kind,
		Source:           o.Source,
		SourceAlias:      o.SourceAlias,
		Fee:              o.Fee,
		Counter:          o.Counter,
		GasLimit:         o.GasLimit,
		StorageLimit:     o.StorageLimit,
		Amount:           o.Amount,
		Destination:      o.Destination,
		DestinationAlias: o.DestinationAlias,
		PublicKey:        o.PublicKey,
		ManagerPubKey:    o.ManagerPubKey,
		Delegate:         o.Delegate,
		Status:           o.Status,
		Burned:           o.Burned,
		Entrypoint:       o.Entrypoint,
		IndexedTime:      o.IndexedTime,

		Result: result,
	}
}

// OperationResult -
type OperationResult struct {
	ConsumedGas                  int64 `json:"consumed_gas,omitempty" extensions:"x-nullable" example:"100"`
	StorageSize                  int64 `json:"storage_size,omitempty" extensions:"x-nullable" example:"200"`
	PaidStorageSizeDiff          int64 `json:"paid_storage_size_diff,omitempty" extensions:"x-nullable" example:"300"`
	AllocatedDestinationContract bool  `json:"allocated_destination_contract,omitempty" extensions:"x-nullable" example:"true"`
}

// FromModel -
func (r *OperationResult) FromModel(result *operation.Result) {
	if result == nil || r == nil {
		return
	}
	r.AllocatedDestinationContract = result.AllocatedDestinationContract
	r.ConsumedGas = result.ConsumedGas
	r.PaidStorageSizeDiff = result.PaidStorageSizeDiff
	r.StorageSize = result.StorageSize
}

// ToModel -
func (r *OperationResult) ToModel() *operation.Result {
	if r == nil {
		return nil
	}

	return &operation.Result{
		AllocatedDestinationContract: r.AllocatedDestinationContract,
		ConsumedGas:                  r.ConsumedGas,
		PaidStorageSizeDiff:          r.PaidStorageSizeDiff,
		StorageSize:                  r.StorageSize,
	}
}

// Contract -
type Contract struct {
	ID        string    `json:"id"`
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language,omitempty" extensions:"x-nullable"`

	Hash        string   `json:"hash"`
	Tags        []string `json:"tags,omitempty" extensions:"x-nullable"`
	Hardcoded   []string `json:"hardcoded,omitempty" extensions:"x-nullable"`
	FailStrings []string `json:"fail_strings,omitempty" extensions:"x-nullable"`
	Annotations []string `json:"annotations,omitempty" extensions:"x-nullable"`
	Entrypoints []string `json:"entrypoints,omitempty" extensions:"x-nullable"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty" extensions:"x-nullable"`
	Delegate string `json:"delegate,omitempty" extensions:"x-nullable"`

	ProjectID       string    `json:"project_id,omitempty" extensions:"x-nullable"`
	FoundBy         string    `json:"found_by,omitempty" extensions:"x-nullable"`
	LastAction      time.Time `json:"last_action,omitempty" extensions:"x-nullable"`
	TxCount         int64     `json:"tx_count,omitempty" extensions:"x-nullable"`
	MigrationsCount int64     `json:"migrations_count,omitempty" extensions:"x-nullable"`
	Alias           string    `json:"alias,omitempty" extensions:"x-nullable"`
	DelegateAlias   string    `json:"delegate_alias,omitempty" extensions:"x-nullable"`

	Subscription       *Subscription `json:"subscription,omitempty" extensions:"x-nullable"`
	TotalSubscribed    int           `json:"total_subscribed"`
	Slug               string        `json:"slug,omitempty" extensions:"x-nullable"`
	Verified           bool          `json:"verified,omitempty" extensions:"x-nullable"`
	VerificationSource string        `json:"verification_source,omitempty" extensions:"x-nullable"`

	Tokens []TokenBalance `json:"tokens"`
}

// FromModel -
func (c *Contract) FromModel(contract contract.Contract) {
	c.Address = contract.Address
	c.Alias = contract.Alias
	c.Annotations = contract.Annotations
	c.Delegate = contract.Delegate
	c.DelegateAlias = contract.DelegateAlias
	c.Entrypoints = contract.Entrypoints
	c.FailStrings = contract.FailStrings
	c.FoundBy = contract.FoundBy
	c.Hardcoded = contract.Hardcoded
	c.Hash = contract.Hash
	c.ID = contract.GetID()
	c.Language = contract.Language
	c.TxCount = contract.TxCount
	c.LastAction = contract.LastAction

	c.Level = contract.Level
	c.Manager = contract.Manager
	c.MigrationsCount = contract.MigrationsCount
	c.Network = contract.Network
	c.ProjectID = contract.ProjectID
	c.Tags = contract.Tags
	c.Timestamp = contract.Timestamp
	c.Verified = contract.Verified
	c.VerificationSource = contract.VerificationSource
}

// Subscription -
type Subscription struct {
	Address          string    `json:"address"`
	Network          string    `json:"network"`
	Alias            string    `json:"alias,omitempty" extensions:"x-nullable"`
	SubscribedAt     time.Time `json:"subscribed_at"`
	WatchSame        bool      `json:"watch_same"`
	WatchSimilar     bool      `json:"watch_similar"`
	WatchMempool     bool      `json:"watch_mempool"`
	WatchMigrations  bool      `json:"watch_migrations"`
	WatchDeployments bool      `json:"watch_deployments"`
	WatchCalls       bool      `json:"watch_calls"`
	WatchErrors      bool      `json:"watch_errors"`
	SentryEnabled    bool      `json:"sentry_enabled"`
	SentryDSN        string    `json:"sentry_dsn,omitempty" extensions:"x-nullable"`
}

// Event -
type Event struct {
	Event string    `json:"event"`
	Date  time.Time `json:"date"`
}

// OperationResponse -
type OperationResponse struct {
	Operations []Operation `json:"operations"`
	LastID     string      `json:"last_id,omitempty" extensions:"x-nullable" example:"1588640276994159"`
}

type userProfile struct {
	Login            string    `json:"login"`
	AvatarURL        string    `json:"avatar_url"`
	MarkReadAt       time.Time `json:"mark_read_at"`
	RegisteredAt     time.Time `json:"registered_at"`
	MarkedContracts  int       `json:"marked_contracts"`
	CompilationTasks int64     `json:"compilation_tasks"`
	Verifications    int64     `json:"verifications"`
	Deployments      int64     `json:"deployments"`

	Subscriptions []Subscription `json:"subscriptions"`
}

// BigMapItem -
type BigMapItem struct {
	Key       interface{} `json:"key"`
	Value     interface{} `json:"value"`
	KeyHash   string      `json:"key_hash"`
	KeyString string      `json:"key_string"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
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

// CodeDiffResponse -
type CodeDiffResponse struct {
	Left  CodeDiffLeg          `json:"left"`
	Right CodeDiffLeg          `json:"right"`
	Diff  formatter.DiffResult `json:"diff"`
}

// NetworkStats -
type NetworkStats struct {
	ContractsCount  int64            `json:"contracts_count" example:"10"`
	OperationsCount int64            `json:"operations_count" example:"100"`
	Protocols       []Protocol       `json:"protocols"`
	Languages       map[string]int64 `json:"languages"`
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

// FromModel -
func (a *Alias) FromModel(alias *tzip.TZIP) {
	a.Alias = alias.Name
	a.Address = alias.Address
	a.Network = alias.Network
	a.Slug = alias.Slug
}

// Protocol -
type Protocol struct {
	Hash       string `json:"hash" example:"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb"`
	Network    string `json:"network" example:"mainnet"`
	StartLevel int64  `json:"start_level" example:"851969"`
	EndLevel   int64  `json:"end_level" example:"0"`
	Alias      string `json:"alias" example:"Carthage"`
}

// FromModel -
func (p *Protocol) FromModel(protocol protocol.Protocol) {
	p.Hash = protocol.Hash
	p.Network = protocol.Network
	p.StartLevel = protocol.StartLevel
	p.EndLevel = protocol.EndLevel
	p.Alias = protocol.Alias
}

// Block -
type Block struct {
	Network     string    `json:"network" example:"mainnet"`
	Hash        string    `json:"hash" example:"BLyAEwaXShJuZasvUezHUfLqzZ48V8XrPvXF2wRaH15tmzEpsHT"`
	Level       int64     `json:"level" example:"24"`
	Predecessor string    `json:"predecessor" example:"BMWVEwEYw9m5iaHzqxDfkPzZTV4rhkSouRh3DkVMVGkxZ3EVaNs"`
	ChainID     string    `json:"chain_id" example:"NetXdQprcVkpaWU"`
	Protocol    string    `json:"protocol" example:"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY"`
	Timestamp   time.Time `json:"timestamp" example:"2018-06-30T18:05:27Z"`
}

// FromModel -
func (b *Block) FromModel(block block.Block) {
	b.Network = block.Network
	b.Hash = block.Hash
	b.Level = block.Level
	b.Predecessor = block.Predecessor
	b.ChainID = block.ChainID
	b.Protocol = block.Protocol
	b.Timestamp = block.Timestamp
}

// LightContract -
type LightContract struct {
	Address  string    `json:"address"`
	Network  string    `json:"network"`
	Deployed time.Time `json:"deploy_time"`
}

// FromModel -
func (c *LightContract) FromModel(light contract.Light) {
	c.Address = light.Address
	c.Network = light.Network
	c.Deployed = light.Deployed
}

// SimilarContractsResponse -
type SimilarContractsResponse struct {
	Count     int               `json:"count"`
	Contracts []SimilarContract `json:"contracts"`
}

// SimilarContract -
type SimilarContract struct {
	*Contract
	Count   int64 `json:"count"`
	Added   int64 `json:"added,omitempty" extensions:"x-nullable"`
	Removed int64 `json:"removed,omitempty" extensions:"x-nullable"`
}

// FromModel -
func (c *SimilarContract) FromModel(similar contract.Similar, diff CodeDiffResponse) {
	var contract Contract
	contract.FromModel(*similar.Contract)
	c.Contract = &contract

	c.Count = similar.Count
	c.Added = diff.Diff.Added
	c.Removed = diff.Diff.Removed
}

// SameContractsResponse -
type SameContractsResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}

// FromModel -
func (c *SameContractsResponse) FromModel(same contract.SameResponse) {
	c.Count = same.Count

	c.Contracts = make([]Contract, len(same.Contracts))
	for i := range same.Contracts {
		var contract Contract
		contract.FromModel(same.Contracts[i])
		c.Contracts[i] = contract
	}
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

// Transfer -
type Transfer struct {
	IndexedTime    int64          `json:"indexed_time"`
	Network        string         `json:"network"`
	Contract       string         `json:"contract"`
	Initiator      string         `json:"initiator"`
	Hash           string         `json:"hash"`
	Status         string         `json:"status"`
	Timestamp      time.Time      `json:"timestamp"`
	Level          int64          `json:"level"`
	From           string         `json:"from"`
	To             string         `json:"to"`
	TokenID        int64          `json:"token_id"`
	Amount         string         `json:"amount"`
	Counter        int64          `json:"counter"`
	Nonce          *int64         `json:"nonce,omitempty" extensions:"x-nullable"`
	Parent         string         `json:"parent,omitempty" extensions:"x-nullable"`
	Token          *TokenMetadata `json:"token,omitempty" extensions:"x-nullable"`
	Alias          string         `json:"alias,omitempty" extensions:"x-nullable"`
	InitiatorAlias string         `json:"initiator_alias,omitempty" extensions:"x-nullable"`
	FromAlias      string         `json:"from_alias,omitempty" extensions:"x-nullable"`
	ToAlias        string         `json:"to_alias,omitempty" extensions:"x-nullable"`
}

// TransferFromElasticModel -
func TransferFromElasticModel(model transfer.Transfer) (t Transfer) {
	t.IndexedTime = model.IndexedTime
	t.Network = model.Network
	t.Contract = model.Contract
	t.Initiator = model.Initiator
	t.Hash = model.Hash
	t.Status = model.Status
	t.Timestamp = model.Timestamp
	t.Level = model.Level
	t.From = model.From
	t.To = model.To
	t.TokenID = model.TokenID
	t.Amount = model.AmountStr
	t.Counter = model.Counter
	t.Nonce = model.Nonce
	t.Parent = model.Parent
	return
}

// TransferResponse -
type TransferResponse struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id,omitempty" extensions:"x-nullable"`
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

// DApp -
type DApp struct {
	Name              string   `json:"name"`
	ShortDescription  string   `json:"short_description"`
	FullDescription   string   `json:"full_description"`
	WebSite           string   `json:"website"`
	Slug              string   `json:"slug,omitempty" extensions:"x-nullable"`
	AgoraReviewPostID int64    `json:"agora_review_post_id,omitempty" extensions:"x-nullable"`
	AgoraQAPostID     int64    `json:"agora_qa_post_id,omitempty" extensions:"x-nullable"`
	Authors           []string `json:"authors"`
	SocialLinks       []string `json:"social_links"`
	Interfaces        []string `json:"interfaces"`
	Categories        []string `json:"categories"`
	Soon              bool     `json:"soon"`
	Logo              string   `json:"logo"`
	Cover             string   `json:"cover,omitempty" extensions:"x-nullable"`
	Volume24Hours     float64  `json:"volume_24_hours,omitempty" extensions:"x-nullable"`

	Screenshots []Screenshot    `json:"screenshots,omitempty" extensions:"x-nullable"`
	Contracts   []DAppContract  `json:"contracts,omitempty" extensions:"x-nullable"`
	DexTokens   []TokenMetadata `json:"dex_tokens,omitempty" extensions:"x-nullable"`
	Tokens      []Token         `json:"tokens,omitempty" extensions:"x-nullable"`
}

// DAppContract -
type DAppContract struct {
	Network     string    `json:"network"`
	Address     string    `json:"address"`
	Alias       string    `json:"alias,omitempty" extensions:"x-nullable"`
	ReleaseDate time.Time `json:"release_date"`
}

// Screenshot -
type Screenshot struct {
	Type string `json:"type"`
	Link string `json:"link"`
}

// Token -
type Token struct {
	TokenMetadata
	transfer.TokenSupply
}

// AccountInfo -
type AccountInfo struct {
	Address    string         `json:"address"`
	Network    string         `json:"network"`
	Alias      string         `json:"alias,omitempty" extensions:"x-nullable"`
	Balance    int64          `json:"balance"`
	TxCount    int64          `json:"tx_count"`
	LastAction time.Time      `json:"last_action"`
	Tokens     []TokenBalance `json:"tokens"`
}

// TokenBalance -
type TokenBalance struct {
	TokenMetadata
	Balance string `json:"balance"`
}

// TokenMetadata -
type TokenMetadata struct {
	Contract      string                 `json:"contract"`
	Network       string                 `json:"network"`
	Level         int64                  `json:"level,omitempty" extensions:"x-nullable"`
	TokenID       int64                  `json:"token_id"`
	Symbol        string                 `json:"symbol,omitempty" extensions:"x-nullable"`
	Name          string                 `json:"name,omitempty" extensions:"x-nullable"`
	Decimals      *int64                 `json:"decimals,omitempty" extensions:"x-nullable"`
	TokenInfo     map[string]interface{} `json:"token_info,omitempty" extensions:"x-nullable"`
	Volume24Hours *float64               `json:"volume_24_hours,omitempty" extensions:"x-nullable"`
}

// TokenMetadataFromElasticModel -
func TokenMetadataFromElasticModel(model tokenmetadata.TokenMetadata, withTokenInfo bool) (tm TokenMetadata) {
	tm.TokenID = model.TokenID
	tm.Symbol = model.Symbol
	tm.Name = model.Name
	tm.Decimals = model.Decimals
	tm.Contract = model.Contract
	tm.Level = model.Level
	tm.Network = model.Network
	if withTokenInfo {
		tm.TokenInfo = model.Extras
	}
	return
}

// DomainsResponse -
type DomainsResponse struct {
	Domains []tezosdomain.TezosDomain `json:"domains"`
	Total   int64                     `json:"total"`
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
	Description    string          `json:"description"`
	Schema         *ast.JSONSchema `json:"schema"`
	DefaultModel   interface{}     `json:"default_model,omitempty" extensions:"x-nullable"`
}

// ForkResponse -
type ForkResponse struct {
	Script  stdJSON.RawMessage `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}
