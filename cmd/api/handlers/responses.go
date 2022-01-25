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
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
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
	AllocatedDestinationContract       bool               `json:"allocated_destination_contract,omitempty" extensions:"x-nullable" example:"true"`
	Internal                           bool               `json:"internal"`
	Mempool                            bool               `json:"mempool"`
}

// FromModel -
func (o *Operation) FromModel(operation operation.Operation) {
	o.ID = operation.ID
	o.Hash = operation.Hash
	o.Network = operation.Network.String()
	o.Internal = operation.Internal
	o.Timestamp = operation.Timestamp.UTC()

	o.Level = operation.Level
	o.Kind = operation.Kind.String()
	o.Source = operation.Source.Address
	o.Fee = operation.Fee
	o.Counter = operation.Counter
	o.GasLimit = operation.GasLimit
	o.StorageLimit = operation.StorageLimit
	o.Amount = operation.Amount
	o.Destination = operation.Destination.Address
	o.Delegate = operation.Delegate.Address
	o.Status = operation.Status.String()
	o.Burned = operation.Burned
	o.Entrypoint = operation.Entrypoint.String()
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
		Network:   types.NewNetwork(o.Network),
		Internal:  o.Internal,
		Timestamp: o.Timestamp,
		Level:     o.Level,
		Kind:      types.NewOperationKind(o.Kind),
		Source: account.Account{
			Network: types.NewNetwork(o.Network),
			Address: o.Source,
			Type:    types.NewAccountType(o.Source),
		},
		Fee:          o.Fee,
		Counter:      o.Counter,
		GasLimit:     o.GasLimit,
		StorageLimit: o.StorageLimit,
		Amount:       o.Amount,
		Destination: account.Account{
			Network: types.NewNetwork(o.Network),
			Address: o.Destination,
			Type:    types.NewAccountType(o.Destination),
		},
		Delegate: account.Account{
			Network: types.NewNetwork(o.Network),
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
	Slug            string    `json:"slug,omitempty" extensions:"x-nullable"`

	SameCount    int64 `json:"same_count"`
	SimilarCount int64 `json:"similar_count"`
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
	c.Network = contract.Network.String()
	c.Tags = contract.Tags.ToArray()
	c.Timestamp = contract.Timestamp

	script := contract.Alpha
	if contract.BabylonID > 0 {
		script = contract.Babylon
	}

	c.Hash = script.Hash
	c.FailStrings = script.FailStrings
	c.Annotations = script.Annotations
	c.Entrypoints = script.Entrypoints
	c.ProjectID = script.ProjectID.String()
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

// FromModel -
func (a *Alias) FromModel(alias *contract_metadata.ContractMetadata) {
	a.Alias = alias.Name
	a.Address = alias.Address
	a.Network = alias.Network.String()
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
	p.Network = protocol.Network.String()
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
	b.Network = block.Network.String()
	b.Hash = block.Hash
	b.Level = block.Level
	b.Protocol = block.Protocol.Hash
	b.Predecessor = block.Predecessor
	b.ChainID = block.ChainID
	b.Timestamp = block.Timestamp
}

// SimilarContractsResponse -
type SimilarContractsResponse struct {
	Count     int               `json:"count"`
	Contracts []SimilarContract `json:"contracts"`
}

// SimilarContract -
type SimilarContract struct {
	*Contract
	Added   int64 `json:"added,omitempty" extensions:"x-nullable"`
	Removed int64 `json:"removed,omitempty" extensions:"x-nullable"`
}

// FromModel -
func (c *SimilarContract) FromModel(similar contract.Similar, diff CodeDiffResponse) {
	var contract Contract
	contract.FromModel(*similar.Contract)
	c.Contract = &contract

	c.Added = diff.Diff.Added
	c.Removed = diff.Diff.Removed
}

// SameContractsResponse -
type SameContractsResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}

// FromModel -
func (c *SameContractsResponse) FromModel(same contract.SameResponse, ctx *Context) {
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
	TokenID        uint64         `json:"token_id"`
	Amount         string         `json:"amount"`
	Counter        int64          `json:"counter"`
	Nonce          *int64         `json:"nonce,omitempty" extensions:"x-nullable"`
	Parent         string         `json:"parent,omitempty" extensions:"x-nullable"`
	Token          *TokenMetadata `json:"token,omitempty" extensions:"x-nullable"`
	Alias          string         `json:"alias,omitempty" extensions:"x-nullable"`
	InitiatorAlias string         `json:"initiator_alias,omitempty" extensions:"x-nullable"`
	FromAlias      string         `json:"from_alias,omitempty" extensions:"x-nullable"`
	ToAlias        string         `json:"to_alias,omitempty" extensions:"x-nullable"`
	Entrypoint     string         `json:"entrypoint,omitempty" extensions:"x-nullable"`
}

// TransferFromModel -
func TransferFromModel(model domains.Transfer) (t Transfer) {
	t.IndexedTime = model.ID
	t.Network = model.Network.String()
	t.Contract = model.Contract
	t.Initiator = model.Initiator.Address
	t.Status = model.Status.String()
	t.Timestamp = model.Timestamp.UTC()
	t.Level = model.Level
	t.From = model.From.Address
	t.To = model.To.Address
	t.TokenID = model.TokenID
	t.Amount = model.Amount.String()
	t.Parent = model.Parent.String()
	t.Entrypoint = model.Entrypoint
	t.Hash = model.Hash
	t.Counter = model.Counter
	t.Nonce = model.Nonce
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
	Name             string   `json:"name"`
	ShortDescription string   `json:"short_description"`
	FullDescription  string   `json:"full_description"`
	WebSite          string   `json:"website"`
	Slug             string   `json:"slug,omitempty" extensions:"x-nullable"`
	Authors          []string `json:"authors"`
	SocialLinks      []string `json:"social_links"`
	Interfaces       []string `json:"interfaces"`
	Categories       []string `json:"categories"`
	Soon             bool     `json:"soon"`
	Logo             string   `json:"logo"`
	Cover            string   `json:"cover,omitempty" extensions:"x-nullable"`
	Volume24Hours    float64  `json:"volume_24_hours,omitempty" extensions:"x-nullable"`

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
	Supply     string  `json:"supply,omitempty"`
	Transfered float64 `json:"transfered,omitempty"`
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

// TokenBalance -
type TokenBalance struct {
	TokenMetadata
	Balance string `json:"balance"`
}

// TokenBalances -
type TokenBalances struct {
	Balances []TokenBalance `json:"balances"`
	Total    int64          `json:"total"`
}

// TokenMetadata -
type TokenMetadata struct {
	Contract           string                 `json:"contract"`
	Network            string                 `json:"network"`
	Level              int64                  `json:"level,omitempty" extensions:"x-nullable"`
	Timestamp          *time.Time             `json:"timestamp,omitempty" extensions:"x-nullable"`
	TokenID            uint64                 `json:"token_id"`
	Symbol             string                 `json:"symbol,omitempty" extensions:"x-nullable"`
	Name               string                 `json:"name,omitempty" extensions:"x-nullable"`
	Decimals           *int64                 `json:"decimals,omitempty" extensions:"x-nullable"`
	Description        string                 `json:"description,omitempty" extensions:"x-nullable"`
	ArtifactURI        string                 `json:"artifact_uri,omitempty" extensions:"x-nullable"`
	DisplayURI         string                 `json:"display_uri,omitempty" extensions:"x-nullable"`
	ThumbnailURI       string                 `json:"thumbnail_uri,omitempty" extensions:"x-nullable"`
	ExternalURI        string                 `json:"external_uri,omitempty" extensions:"x-nullable"`
	Minter             string                 `json:"minter,omitempty" extensions:"x-nullable"`
	IsTransferable     bool                   `json:"is_transferable,omitempty" extensions:"x-nullable"`
	IsBooleanAmount    bool                   `json:"is_boolean_amount,omitempty" extensions:"x-nullable"`
	ShouldPreferSymbol bool                   `json:"should_prefer_symbol,omitempty" extensions:"x-nullable"`
	Creators           []string               `json:"creators,omitempty" extensions:"x-nullable"`
	Tags               []string               `json:"tags,omitempty" extensions:"x-nullable"`
	Formats            stdJSON.RawMessage     `json:"formats,omitempty" extensions:"x-nullable"`
	TokenInfo          map[string]interface{} `json:"token_info,omitempty" extensions:"x-nullable"`
	Volume24Hours      *float64               `json:"volume_24_hours,omitempty" extensions:"x-nullable"`
}

// TokenMetadataFromElasticModel -
func TokenMetadataFromElasticModel(model tokenmetadata.TokenMetadata, withTokenInfo bool) (tm TokenMetadata) {
	tm.TokenID = model.TokenID
	tm.Symbol = model.Symbol
	tm.Name = model.Name
	tm.Decimals = model.Decimals
	tm.Contract = model.Contract
	tm.Level = model.Level
	tm.Network = model.Network.String()
	tm.Description = model.Description
	tm.ArtifactURI = model.ArtifactURI
	tm.DisplayURI = model.DisplayURI
	tm.ThumbnailURI = model.ThumbnailURI
	tm.ExternalURI = model.ExternalURI
	tm.Minter = model.Minter
	tm.IsTransferable = model.IsTransferable
	tm.IsBooleanAmount = model.IsBooleanAmount
	tm.ShouldPreferSymbol = model.ShouldPreferSymbol
	tm.Creators = model.Creators
	tm.Tags = model.Tags
	tm.Formats = stdJSON.RawMessage(model.Formats)

	if !model.Timestamp.IsZero() {
		tm.Timestamp = &model.Timestamp
	}

	if withTokenInfo {
		tm.TokenInfo = model.Extras
	}
	return
}

// Empty -
func (tm TokenMetadata) Empty() bool {
	return tm.Symbol == "" && tm.Name == "" && tm.Decimals == nil && tm.TokenID == 0 &&
		tm.Description == "" && tm.ArtifactURI == "" && tm.DisplayURI == "" && tm.ThumbnailURI == "" &&
		tm.ExternalURI == "" && len(tm.Creators) == 0 && len(tm.Tags) == 0 && len(tm.Formats) == 0 && tm.Minter == ""
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

// TZIPResponse -
type TZIPResponse struct {
	Address     string                     `json:"address,omitempty"`
	Network     string                     `json:"network,omitempty"`
	DomainName  string                     `json:"domain,omitempty"`
	Extras      map[string]interface{}     `json:"extras,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	Version     string                     `json:"version,omitempty"`
	License     *contract_metadata.License `json:"license,omitempty"`
	Homepage    string                     `json:"homepage,omitempty"`
	Authors     []string                   `json:"authors,omitempty"`
	Interfaces  []string                   `json:"interfaces,omitempty"`
	Views       contract_metadata.Views    `json:"views,omitempty"`
	contract_metadata.TZIP20
}

// FromModel -
func (t *TZIPResponse) FromModel(model *contract_metadata.ContractMetadata, withViewsAndEvents bool) {
	t.DomainName = model.DomainName
	t.Extras = model.Extras
	t.Address = model.Address
	t.Network = model.Network.String()
	t.Name = model.Name
	t.Description = model.Description
	t.Version = model.Version
	t.Homepage = model.Homepage
	t.Authors = model.Authors
	t.Interfaces = model.Interfaces

	if !model.License.IsEmpty() {
		t.License = &model.License
	}

	if withViewsAndEvents {
		t.Views = model.Views
		t.TZIP20 = model.TZIP20
	}
}

// HeadResponse -
type HeadResponse struct {
	Network         string    `json:"network"`
	Level           int64     `json:"level"`
	Timestamp       time.Time `json:"time"`
	Protocol        string    `json:"protocol"`
	Total           int64     `json:"total"`
	ContractCalls   int64     `json:"contract_calls"`
	UniqueContracts int64     `json:"unique_contracts"`
	FACount         int64     `json:"fa_count"`
	Synced          bool      `json:"synced"`
}

// TokensCountWithMetadata -
type TokensCountWithMetadata struct {
	TZIPResponse
	Count int64    `json:"count"`
	Tags  []string `json:"contract_tags"`
}

// GlobalConstant -
type GlobalConstant struct {
	Network   types.Network      `json:"network"`
	Timestamp time.Time          `json:"timestamp"`
	Level     int64              `json:"level"`
	Address   string             `json:"address"`
	Value     stdJSON.RawMessage `json:"value,omitempty"`
}

// NewGlobalConstantFromModel -
func NewGlobalConstantFromModel(gc global_constant.GlobalConstant) GlobalConstant {
	return GlobalConstant{
		Network:   gc.Network,
		Timestamp: gc.Timestamp.UTC(),
		Level:     gc.Level,
		Address:   gc.Address,
		Value:     stdJSON.RawMessage(gc.Value),
	}
}
