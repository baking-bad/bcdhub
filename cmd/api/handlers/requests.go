package handlers

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

type getAccountRequest struct {
	Address string `uri:"address" binding:"required,address"`
	Network string `uri:"network" binding:"required,network"`
}

// NetworkID -
func (req getAccountRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type getContractRequest struct {
	Address string `uri:"address" binding:"required,contract"`
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

type withStatsRequest struct {
	Stats *bool `form:"stats,omitempty" binding:"omitempty"`
}

// HasStats -
func (req withStatsRequest) HasStats() bool {
	return req.Stats == nil || *req.Stats
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

type opgForAddressRequest struct {
	LastID int64  `form:"last_id" binding:"omitempty"`
	Size   uint64 `form:"size" binding:"min=0"`
}

type pageableRequest struct {
	Offset int64 `form:"offset" binding:"min=0"`
	Size   int64 `form:"size" binding:"min=0,bcd_max_size=10"`
}

// OPGRequest -
type OPGRequest struct {
	getByNetwork
	Hash string `uri:"hash" binding:"required,opg" example:"ooy4c6G2BZzybYEY3vRQ7WXGL63tFmamTeGTHdjUxhd6ckbSNnb"`
}

// OperationGroupContentRequest -
type OperationGroupContentRequest struct {
	OPGRequest

	Counter int64 `uri:"counter" binding:"required" example:"123456"`
}

// ImplicitOperationRequest -
type ImplicitOperationRequest struct {
	getByNetwork

	Counter int64 `uri:"counter" binding:"required" example:"123456"`
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

type bigMapSearchRequest struct {
	pageableRequest
	MaxLevel *int64 `form:"max_level,omitempty" binding:"omitempty,gt_int64_ptr=MinLevel"`
	MinLevel *int64 `form:"min_level,omitempty" binding:"omitempty"`
}

type opgRequest struct {
	WithMempool     bool `form:"with_mempool" binding:"omitempty"`
	WithStorageDiff bool `form:"with_storage_diff" binding:"omitempty"`
}

type getEntrypointDataRequest struct {
	Name   string                 `json:"name" binding:"required"`
	Data   map[string]interface{} `json:"data" binding:"required"`
	Format string                 `json:"format"`
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
	FillType       string  `form:"fill_type,omitempty" binding:"omitempty"`
	EntrypointName string  `form:"entrypoint" binding:"required"`
	Hash           string  `form:"hash,omitempty" binding:"omitempty"`
	Counter        *uint64 `form:"counter,omitempty" binding:"omitempty"`
}

type forkRequest struct {
	Address string `json:"address,omitempty" binding:"omitempty,address,required_with=Network"`
	Network string `json:"network,omitempty" binding:"omitempty,network,required_with=Address"`
	Script  string `json:"script,omitempty" binding:"omitempty"`

	Storage map[string]interface{} `json:"storage" binding:"required"`
}

// NetworkID -
func (req forkRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type storageRequest struct {
	Level int `form:"level" binding:"omitempty,gte=1"`
}

// GetTokenStatsRequest -
type GetTokenStatsRequest struct {
	Period    string `form:"period" binding:"oneof=all year month week day hour" example:"year"`
	Contracts string `form:"contracts"`
}

// Addresses -
func (req GetTokenStatsRequest) Addresses() []string {
	if req.Contracts == "" {
		return nil
	}
	return strings.Split(req.Contracts, ",")
}

type executeViewRequest struct {
	Data           map[string]interface{}       `json:"data" binding:"required"`
	Name           string                       `json:"name" binding:"required_if=Kind on-chain"`
	Implementation *int                         `json:"implementation" binding:"required_if=Kind on-chain"`
	Kind           ViewSchemaKind               `json:"kind" binding:"required"`
	Amount         int64                        `json:"amount,omitempty"`
	GasLimit       int64                        `json:"gas_limit,omitempty"`
	Source         string                       `json:"source,omitempty" binding:"omitempty,address"`
	Sender         string                       `json:"sender,omitempty" binding:"omitempty,address"`
	View           *contract.ViewImplementation `json:"view,omitempty" binding:"required_if=Kind off-chain"`
}

type getGlobalConstantRequest struct {
	Address string `uri:"address" binding:"required,global_constant"`
}

type globalConstantsListRequest struct {
	pageableRequest

	OrderBy string `form:"order_by" binding:"omitempty,oneof=level timestamp links_count address"`
	Sort    string `form:"sort" binding:"omitempty,oneof=asc desc"`
}

type globalConstantsContractsRequest struct {
	getGlobalConstantRequest
	pageableRequest
}

type getViewsArgs struct {
	Kind ViewSchemaKind `form:"kind" binding:"omitempty,oneof=off-chain on-chain"`
}

type findContract struct {
	Tags string `form:"tags" binding:"omitempty"`
}

type smartRollupListRequest struct {
	pageableRequest

	Sort string `form:"sort" binding:"omitempty,oneof=asc desc"`
}

type getSmartRollupRequest struct {
	Address string `uri:"address" binding:"required,smart_rollup"`
}
