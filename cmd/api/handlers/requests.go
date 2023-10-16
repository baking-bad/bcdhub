package handlers

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

type getAccountRequest struct {
	Address string `binding:"required,address" uri:"address"`
	Network string `binding:"required,network" uri:"network"`
}

// NetworkID -
func (req getAccountRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type getContractRequest struct {
	Address string `binding:"required,contract" uri:"address"`
	Network string `binding:"required,network"  uri:"network"`
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
	Stats *bool `binding:"omitempty" form:"stats,omitempty"`
}

// HasStats -
func (req withStatsRequest) HasStats() bool {
	return req.Stats == nil || *req.Stats
}

// CodeDiffLeg -
type CodeDiffLeg struct {
	Address  string        `binding:"required,address" json:"address"`
	Network  types.Network `binding:"required,network" json:"network"`
	Protocol string        `json:"protocol,omitempty"`
	Level    int64         `json:"level,omitempty"`
}

// CodeDiffRequest -
type CodeDiffRequest struct {
	Left  CodeDiffLeg `binding:"required" json:"left"`
	Right CodeDiffLeg `binding:"required" json:"right"`
}

type getBigMapRequest struct {
	Network string `binding:"required,network" uri:"network"`
	Ptr     int64  `binding:"min=0"            uri:"ptr"`
}

// NetworkID -
func (req getBigMapRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type getBigMapByKeyHashRequest struct {
	Network string `binding:"required,network" uri:"network"`
	Ptr     int64  `binding:"min=0"            uri:"ptr"`
	KeyHash string `binding:"required"         uri:"key_hash"`
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
	LastID int64  `binding:"omitempty" form:"last_id"`
	Size   uint64 `binding:"min=0"     form:"size"`
}

type pageableRequest struct {
	Offset int64 `binding:"min=0"                 form:"offset"`
	Size   int64 `binding:"min=0,bcd_max_size=10" form:"size"`
}

// OPGRequest -
type OPGRequest struct {
	getByNetwork
	Hash string `binding:"required,opg" example:"ooy4c6G2BZzybYEY3vRQ7WXGL63tFmamTeGTHdjUxhd6ckbSNnb" uri:"hash"`
}

// OperationGroupContentRequest -
type OperationGroupContentRequest struct {
	OPGRequest

	Counter int64 `binding:"required" example:"123456" uri:"counter"`
}

// ImplicitOperationRequest -
type ImplicitOperationRequest struct {
	getByNetwork

	Counter int64 `binding:"required" example:"123456" uri:"counter"`
}

// FormatterRequest -
type FormatterRequest struct {
	Inline   bool   `form:"inline"`
	LineSize int    `form:"lineSize"`
	Code     string `form:"code"`
}

type getByNetwork struct {
	Network string `binding:"required,network" example:"mainnet" uri:"network"`
}

// NetworkID -
func (req getByNetwork) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type bigMapSearchRequest struct {
	pageableRequest
	MaxLevel *int64 `binding:"omitempty,gt_int64_ptr=MinLevel" form:"max_level,omitempty"`
	MinLevel *int64 `binding:"omitempty"                       form:"min_level,omitempty"`
}

type opgRequest struct {
	WithMempool     bool `binding:"omitempty" form:"with_mempool"`
	WithStorageDiff bool `binding:"omitempty" form:"with_storage_diff"`
}

type getEntrypointDataRequest struct {
	Name   string                 `binding:"required" json:"name"`
	Data   map[string]interface{} `binding:"required" json:"data"`
	Format string                 `json:"format"`
}

type getOperationByIDRequest struct {
	ID int64 `binding:"required" uri:"id"`
}

type runOperationRequest struct {
	Data   map[string]interface{} `binding:"required"          json:"data"`
	Name   string                 `binding:"required"          json:"name"`
	Amount int64                  `json:"amount,omitempty"`
	Source string                 `binding:"omitempty,address" json:"source,omitempty"`
}

type runCodeRequest struct {
	Data     map[string]interface{} `binding:"required"          json:"data"`
	Name     string                 `binding:"required"          json:"name"`
	Amount   int64                  `json:"amount,omitempty"`
	GasLimit int64                  `json:"gas_limit,omitempty"`
	Source   string                 `binding:"omitempty,address" json:"source,omitempty"`
	Sender   string                 `binding:"omitempty,address" json:"sender,omitempty"`
}

type storageSchemaRequest struct {
	FillType string `binding:"omitempty,fill_type" form:"fill_type,omitempty"`
}

type entrypointSchemaRequest struct {
	FillType       string  `binding:"omitempty" form:"fill_type,omitempty"`
	EntrypointName string  `binding:"required"  form:"entrypoint"`
	Hash           string  `binding:"omitempty" form:"hash,omitempty"`
	Counter        *uint64 `binding:"omitempty" form:"counter,omitempty"`
}

type forkRequest struct {
	Address string `binding:"omitempty,address,required_with=Network" json:"address,omitempty"`
	Network string `binding:"omitempty,network,required_with=Address" json:"network,omitempty"`
	Script  string `binding:"omitempty"                               json:"script,omitempty"`

	Storage map[string]interface{} `binding:"required" json:"storage"`
}

// NetworkID -
func (req forkRequest) NetworkID() types.Network {
	return types.NewNetwork(req.Network)
}

type storageRequest struct {
	Level int `binding:"omitempty,gte=1" form:"level"`
}

// GetTokenStatsRequest -
type GetTokenStatsRequest struct {
	Period    string `binding:"oneof=all year month week day hour" example:"year" form:"period"`
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
	Data           map[string]interface{}       `binding:"required"                   json:"data"`
	Name           string                       `binding:"required_if=Kind on-chain"  json:"name"`
	Implementation *int                         `binding:"required_if=Kind on-chain"  json:"implementation"`
	Kind           ViewSchemaKind               `binding:"required"                   json:"kind"`
	Amount         int64                        `json:"amount,omitempty"`
	GasLimit       int64                        `json:"gas_limit,omitempty"`
	Source         string                       `binding:"omitempty,address"          json:"source,omitempty"`
	Sender         string                       `binding:"omitempty,address"          json:"sender,omitempty"`
	View           *contract.ViewImplementation `binding:"required_if=Kind off-chain" json:"view,omitempty"`
}

type getGlobalConstantRequest struct {
	Address string `binding:"required,global_constant" uri:"address"`
}

type globalConstantsListRequest struct {
	pageableRequest

	OrderBy string `binding:"omitempty,oneof=level timestamp links_count address" form:"order_by"`
	Sort    string `binding:"omitempty,oneof=asc desc"                            form:"sort"`
}

type globalConstantsContractsRequest struct {
	getGlobalConstantRequest
	pageableRequest
}

type getViewsArgs struct {
	Kind ViewSchemaKind `binding:"omitempty,oneof=off-chain on-chain" form:"kind"`
}

type findContract struct {
	Tags string `binding:"omitempty" form:"tags"`
}

type smartRollupListRequest struct {
	pageableRequest

	Sort string `binding:"omitempty,oneof=asc desc" form:"sort"`
}

type getSmartRollupRequest struct {
	Address string `binding:"required,smart_rollup" uri:"address"`
}

type ticketBalancesRequest struct {
	pageableRequest
	WithoutZeroBalances bool `binding:"omitempty" form:"skip_empty"`
}

type ticketUpdatesRequest struct {
	pageableRequest

	Account  string  `binding:"omitempty,address" form:"account"`
	TicketId *uint64 `binding:"omitempty"         form:"ticket_id"`
}
