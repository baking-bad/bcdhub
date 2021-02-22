package handlers

import "strings"

type getContractRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
}

type getContractCodeRequest struct {
	getContractRequest
	Protocol string `form:"protocol,omitempty"`
	Level    int64  `form:"level,omitempty"`
}

type networkQueryRequest struct {
	Network string `form:"network,omitempty" binding:"omitempty,network"`
}

// CodeDiffLeg -
type CodeDiffLeg struct {
	Address  string `json:"address" binding:"required,address"`
	Network  string `json:"network" binding:"required,network"`
	Protocol string `json:"protocol,omitempty"`
	Level    int64  `json:"level,omitempty"`
}

// CodeDiffRequest -
type CodeDiffRequest struct {
	Left  CodeDiffLeg `json:"left" binding:"required"`
	Right CodeDiffLeg `json:"right" binding:"required"`
}

type getBigMapRequest struct {
	Network string `uri:"network" binding:"required,network"`
	Ptr     int64  `uri:"ptr" binding:"min=0"`
}

type getBigMapByKeyHashRequest struct {
	Network string `uri:"network" binding:"required,network"`
	Ptr     int64  `uri:"ptr" binding:"min=0"`
	KeyHash string `uri:"key_hash" binding:"required"`
}

// OauthRequest -
type OauthRequest struct {
	State string `form:"state"`
	Code  string `form:"code"`
}

// OauthParams -
type OauthParams struct {
	Provider string `uri:"provider"`
}

type operationsRequest struct {
	LastID          string `form:"last_id" binding:"omitempty,numeric"`
	From            uint   `form:"from" binding:"omitempty"`
	To              uint   `form:"to" binding:"omitempty,gtfield=From"`
	Size            uint64 `form:"size" binding:"min=0"`
	Status          string `form:"status" binding:"omitempty,status"`
	Entrypoints     string `form:"entrypoints" binding:"omitempty,excludesall=\"'"`
	WithStorageDiff bool   `form:"with_storage_diff"`
}

type pageableRequest struct {
	Offset int64 `form:"offset" binding:"min=0"`
	Size   int64 `form:"size" binding:"min=0,max=10000"`
}

type cursorRequest struct {
	LastID string `form:"last_id" binding:"omitempty,numeric"`
	Size   int64  `form:"size" binding:"min=0,max=10000"`
}

type searchRequest struct {
	Text      string `form:"q" binding:"required,search"`
	Fields    string `form:"f,omitempty"`
	Networks  string `form:"n,omitempty"`
	Offset    uint   `form:"o,omitempty"`
	DateFrom  uint   `form:"s,omitempty"`
	DateTo    uint   `form:"e,omitempty"`
	Grouping  uint   `form:"g,omitempty"`
	Indices   string `form:"i,omitempty"`
	Languages string `form:"l,omitempty"`
}

// Subscription flags
const (
	WatchSame uint = 1 << iota
	WatchSimilar
	WatchMempool
	WatchMigrations
	WatchDeployments
	WatchCalls
	WatchErrors
	SentryEnabled
)

type subRequest struct {
	getContractRequest
	Alias            string `json:"alias"`
	WatchSame        bool   `json:"watch_same"`
	WatchSimilar     bool   `json:"watch_similar"`
	WatchMempool     bool   `json:"watch_mempool"`
	WatchMigrations  bool   `json:"watch_migrations"`
	WatchDeployments bool   `json:"watch_deployments"`
	WatchCalls       bool   `json:"watch_calls"`
	WatchErrors      bool   `json:"watch_errors"`
	SentryEnabled    bool   `json:"sentry_enabled"`
	SentryDSN        string `json:"sentry_dsn,omitempty"`
}

func newSubscriptionWithMask(mask uint) Subscription {
	return Subscription{
		WatchSame:        mask&WatchSame != 0,
		WatchSimilar:     mask&WatchSimilar != 0,
		WatchMempool:     mask&WatchMempool != 0,
		WatchMigrations:  mask&WatchMigrations != 0,
		WatchDeployments: mask&WatchDeployments != 0,
		WatchCalls:       mask&WatchCalls != 0,
		WatchErrors:      mask&WatchErrors != 0,
		SentryEnabled:    mask&SentryEnabled != 0,
	}
}

func (s subRequest) getMask() uint {
	var b uint

	if s.WatchSame {
		b |= WatchSame
	}

	if s.WatchSimilar {
		b |= WatchSimilar
	}

	if s.WatchMempool {
		b |= WatchMempool
	}

	if s.WatchMigrations {
		b |= WatchMigrations
	}

	if s.WatchDeployments {
		b |= WatchDeployments
	}

	if s.WatchCalls {
		b |= WatchCalls
	}

	if s.WatchErrors {
		b |= WatchErrors
	}

	if s.SentryEnabled {
		b |= SentryEnabled
	}

	return b
}

type sameContractRequest struct {
	pageableRequest
	Manager string `form:"manager,omitempty"`
}

type voteRequest struct {
	SourceAddress      string `json:"src" binding:"required,address"`
	SourceNetwork      string `json:"src_network" binding:"required,network"`
	DestinationAddress string `json:"dest" binding:"required,address"`
	DestinationNetwork string `json:"dest_network" binding:"required,network"`
	Vote               uint   `json:"vote" binding:"oneof=1 2"`
}

// OPGRequest -
type OPGRequest struct {
	Hash string `uri:"hash" binding:"required,opg" example:"ooy4c6G2BZzybYEY3vRQ7WXGL63tFmamTeGTHdjUxhd6ckbSNnb"`
}

// FormatterRequest -
type FormatterRequest struct {
	Inline   bool   `form:"inline"`
	LineSize int    `form:"lineSize"`
	Code     string `form:"code"`
}

type getByNetwork struct {
	Network string `uri:"network" binding:"required,network" example:"mainnet"`
}

type getTokensByVersion struct {
	Network string `uri:"network" binding:"required,network" example:"mainnet"`
	Version string `uri:"faversion" binding:"required,faversion" example:"fa2"`
}

type bigMapSearchRequest struct {
	pageableRequest
	Search   string `form:"q"`
	MaxLevel *int64 `form:"max_level,omitempty" binding:"omitempty,gt_int64_ptr=MinLevel"`
	MinLevel *int64 `form:"min_level,omitempty" binding:"omitempty"`
}

type opgRequest struct {
	WithMempool bool `form:"with_mempool"`
}

type getEntrypointDataRequest struct {
	Name   string                 `json:"name" binding:"required"`
	Data   map[string]interface{} `json:"data" binding:"required"`
	Format string                 `json:"format"`
}

type getSeriesRequest struct {
	Name    string `form:"name" binding:"oneof=contract operation paid_storage_size_diff consumed_gas volume users token_volume" example:"contract"`
	Period  string `form:"period" binding:"oneof=year month week day" example:"year"`
	Address string `form:"address,omitempty" binding:"omitempty"`
}

type getBySlugRequest struct {
	Slug string `uri:"slug"  binding:"required"`
}

type getOperationByIDRequest struct {
	ID string `uri:"id" binding:"required"`
}

type runOperationRequest struct {
	Data   map[string]interface{} `json:"data" binding:"required"`
	Name   string                 `json:"name" binding:"required"`
	Amount int64                  `json:"amount,omitempty"`
	Source string                 `json:"source,omitempty" binding:"omitempty,address"`
}

type runCodeRequest struct {
	Data     map[string]interface{} `json:"data" binding:"required"`
	Name     string                 `json:"name" binding:"required"`
	Amount   int64                  `json:"amount,omitempty"`
	GasLimit int64                  `json:"gas_limit,omitempty"`
	Source   string                 `json:"source,omitempty" binding:"omitempty,address"`
	Sender   string                 `json:"sender,omitempty" binding:"omitempty,address"`
}

type markReadRequest struct {
	Timestamp int64 `json:"timestamp"`
}

type storageSchemaRequest struct {
	FillType string `form:"fill_type,omitempty" binding:"omitempty,fill_type"`
}

type entrypointSchemaRequest struct {
	FillType       string `form:"fill_type,omitempty" binding:"omitempty"`
	EntrypointName string `form:"entrypoint" binding:"required"`
}

type forkRequest struct {
	Address string `json:"address" binding:"required_with=Network,omitempty,address"`
	Network string `json:"network" binding:"required_with=Address,omitempty,network"`
	Script  string `json:"script" binding:"omitempty"`

	Storage map[string]interface{} `json:"storage" binding:"required"`
}

type storageRequest struct {
	Level int `form:"level" binding:"omitempty,gte=1"`
}

// GetTokenStatsRequest -
type GetTokenStatsRequest struct {
	Period    string `form:"period" binding:"oneof=all year month week day" example:"year"`
	Contracts string `form:"contracts"`
}

// Addresses -
func (req GetTokenStatsRequest) Addresses() []string {
	if req.Contracts == "" {
		return nil
	}
	return strings.Split(req.Contracts, ",")
}

type getTokenSeriesRequest struct {
	Contract string `form:"contract" binding:"required,address"`
	Period   string `form:"period" binding:"oneof=year month week day" example:"year"`
	TokenID  uint   `form:"token_id"`
	Slug     string `form:"slug" binding:"required"`
}

type verificationRequest struct {
	getContractRequest
	Account string `json:"account"`
	Repo    string `json:"repo"`
	Ref     string `json:"ref"`
}

type deploymentRequest struct {
	OperationHash string `json:"operation_hash" binding:"required"`
	TaskID        uint   `json:"task_id" binding:"required"`
	ResultID      uint   `json:"result_id"`
}

type compilationRequest struct {
	Limit  uint `form:"limit" binding:"omitempty,min=0"`
	Offset uint `form:"offset" binding:"omitempty,min=0"`
}

type compilationTasksRequest struct {
	compilationRequest
	Kind string `form:"kind" binding:"omitempty,compilation_kind"`
}

type publicReposRequest struct {
	Login string `form:"login" binding:"required"`
}

type publicRefsRequest struct {
	Owner string `form:"owner" binding:"required"`
	Repo  string `form:"repo" binding:"required"`
}

type getDappRequest struct {
	Slug string `uri:"slug" binding:"required"`
}

type getContractTransfers struct {
	pageableRequest
	TokenID *uint `form:"token_id"  binding:"omitempty,min=0"`
}

type getTransfersRequest struct {
	cursorRequest
	Start     uint   `form:"start"  binding:"omitempty,min=1"`
	End       uint   `form:"end"  binding:"omitempty,min=1,gtfield=Start"`
	Contracts string `form:"contracts"  binding:"omitempty"`
	Sort      string `form:"sort" binding:"omitempty,oneof=asc desc"`
	TokenID   *int64 `form:"token_id" binding:"omitempty,min=0"`
}

type getTokenHolders struct {
	TokenID *int64 `form:"token_id" binding:"min=0"`
}

type resolveDomainRequest struct {
	Name    string `form:"name" binding:"omitempty"`
	Address string `form:"address" binding:"omitempty"`
}

type metadataRequest struct {
	Hash string `json:"hash" binding:"required"`
}

type executeViewRequest struct {
	Data           map[string]interface{} `json:"data" binding:"required"`
	Name           string                 `json:"name" binding:"required"`
	Implementation *int                   `json:"implementation" binding:"required"`
	Amount         int64                  `json:"amount,omitempty"`
	GasLimit       int64                  `json:"gas_limit,omitempty"`
	Source         string                 `json:"source,omitempty" binding:"omitempty,address"`
	Sender         string                 `json:"sender,omitempty" binding:"omitempty,address"`
}
