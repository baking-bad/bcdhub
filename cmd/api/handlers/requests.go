package handlers

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

type getContractRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
}

// NetworkID -
func (req getContractRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type getContractCodeRequest struct {
	getContractRequest
	Protocol string `form:"protocol,omitempty"`
	Level    int64  `form:"level,omitempty"`
}

type networkQueryRequest struct {
	Network types.Network `form:"network,omitempty" binding:"omitempty,network"`
}

// CodeDiffLeg -
type CodeDiffLeg struct {
	Address  string        `json:"address" binding:"required,address"`
	Network  types.Network `json:"network" binding:"required,network"`
	Protocol string        `json:"protocol,omitempty"`
	Level    int64         `json:"level,omitempty"`
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

// NetworkID -
func (req getBigMapRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
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
	Size   int64 `form:"size" binding:"min=0,bcd_max_size=10"`
}

type cursorRequest struct {
	LastID string `form:"last_id" binding:"omitempty,numeric"`
	Size   int64  `form:"size" binding:"min=0,bcd_max_size=10"`
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

type sameContractRequest struct {
	pageableRequest
	Manager string `form:"manager,omitempty"`
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

// NetworkID -
func (req getByNetwork) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type getTokensByVersion struct {
	Network string `uri:"network" binding:"required,network" example:"mainnet"`
	Version string `uri:"faversion" binding:"required,faversion" example:"fa2"`
}

// NetworkID -
func (req getTokensByVersion) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
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
	Period  string `form:"period" binding:"oneof=year month week day hour" example:"year"`
	Address string `form:"address,omitempty" binding:"omitempty"`
}

func (req getSeriesRequest) isCached() bool {
	return req.Period == "month" && (req.Name == "contract" || req.Name == "operation" || req.Name == "paid_storage_size_diff" || req.Name == "consumed_gas")
}

type getBySlugRequest struct {
	Slug string `uri:"slug"  binding:"required"`
}

type getOperationByIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
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

type storageSchemaRequest struct {
	FillType string `form:"fill_type,omitempty" binding:"omitempty,fill_type"`
}

type entrypointSchemaRequest struct {
	FillType       string `form:"fill_type,omitempty" binding:"omitempty"`
	EntrypointName string `form:"entrypoint" binding:"required"`
}

type forkRequest struct {
	Address string        `json:"address" binding:"required_with=Network,omitempty,address"`
	Network types.Network `json:"network" binding:"required_with=Address,omitempty,network"`
	Script  string        `json:"script" binding:"omitempty"`

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
	TokenID  uint64 `form:"token_id"`
	Slug     string `form:"slug" binding:"required"`
}

type getDappRequest struct {
	Slug string `uri:"slug" binding:"required"`
}

type getContractTransfers struct {
	pageableRequest
	TokenID *uint64 `form:"token_id"  binding:"omitempty,min=0"`
}

type getTransfersRequest struct {
	cursorRequest
	Start     uint    `form:"start"  binding:"omitempty,min=1"`
	End       uint    `form:"end"  binding:"omitempty,min=1,gtfield=Start"`
	Contracts string  `form:"contracts"  binding:"omitempty"`
	Sort      string  `form:"sort" binding:"omitempty,oneof=asc desc"`
	TokenID   *uint64 `form:"token_id" binding:"omitempty,min=0"`
}

type byTokenIDRequest struct {
	TokenID *uint64 `form:"token_id" binding:"min=0"`
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

type minMaxLevel struct {
	MaxLevel int64 `form:"max_level,omitempty" binding:"omitempty,gt_int64_ptr=MinLevel"`
	MinLevel int64 `form:"min_level,omitempty" binding:"omitempty"`
}

type tokenRequest struct {
	pageableRequest
	minMaxLevel
	TokenID *uint64 `form:"token_id" binding:"omitempty"`
}

type tokenBalanceRequest struct {
	pageableRequest
	Contract string `form:"contract" binding:"omitempty,address"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=token_id balance"`
}

type batchAddressRequest struct {
	Address string `form:"address" binding:"required"`
}

type tokenMetadataRequest struct {
	tokenRequest
	Creator  string `form:"creator" binding:"omitempty"`
	Contract string `form:"contract" binding:"omitempty,address"`
}
