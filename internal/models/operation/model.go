package operation

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Operation -
type Operation struct {
	ID int64 `json:"-"`

	ContentIndex int64 `json:"content_index,omitempty" gorm:"index:new_opg_idx,default:0"`
	Level        int64 `json:"level" gorm:"index:idx_operations_level_network"`
	Counter      int64 `json:"counter,omitempty" gorm:"index:new_opg_idx"`
	Fee          int64 `json:"fee,omitempty"`
	GasLimit     int64 `json:"gas_limit,omitempty"`
	StorageLimit int64 `json:"storage_limit,omitempty"`
	Amount       int64 `json:"amount,omitempty"`

	ConsumedGas                        int64 `json:"consumed_gas,omitempty"`
	StorageSize                        int64 `json:"storage_size,omitempty"`
	PaidStorageSizeDiff                int64 `json:"paid_storage_size_diff,omitempty"`
	Burned                             int64 `json:"burned,omitempty"`
	AllocatedDestinationContractBurned int64 `json:"allocated_destination_contract_burned,omitempty"`

	Nonce      *int64        `json:"nonce,omitempty"`
	Network    types.Network `json:"network" gorm:"type:SMALLINT;index:idx_operations_level_network; index:operations_network_idx"`
	ProtocolID int64         `json:"protocol" gorm:"type:SMALLINT"`
	Hash       string        `json:"hash" gorm:"index:new_opg_idx;index:operations_hash_idx"`

	Timestamp       time.Time             `json:"timestamp"`
	Status          types.OperationStatus `json:"status" gorm:"type:SMALLINT"`
	Kind            types.OperationKind   `json:"kind" gorm:"type:SMALLINT"`
	Initiator       string                `json:"initiator"`
	Source          string                `json:"source" gorm:"index:source_idx"`
	Destination     string                `json:"destination,omitempty" gorm:"index:destination_idx"`
	Delegate        string                `json:"delegate,omitempty"`
	Entrypoint      types.NullString      `json:"entrypoint,omitempty" gorm:"index:operations_entrypoint_idx"`
	Parameters      []byte                `json:"parameters,omitempty"`
	DeffatedStorage []byte                `json:"deffated_storage"`

	Tags types.Tags `json:"tags,omitempty" gorm:"default:0"`

	Script []byte `json:"-"  gorm:"-"`

	Errors tezerrors.Errors `json:"errors,omitempty" gorm:"type:bytes"`

	AllocatedDestinationContract bool `json:"allocated_destination_contract,omitempty"`
	Internal                     bool `json:"internal" gorm:",default:false"`

	AST *ast.Script `json:"-" gorm:"-"`

	Transfers   []*transfer.Transfer     `json:"-"`
	BigMapDiffs []*bigmapdiff.BigMapDiff `json:"-"`
}

// GetID -
func (o *Operation) GetID() int64 {
	return o.ID
}

// GetIndex -
func (o *Operation) GetIndex() string {
	return "operations"
}

// Save -
func (o *Operation) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(o).Error
}

// LogFields -
func (o *Operation) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": o.Network.String(),
		"hash":    o.Hash,
		"block":   o.Level,
	}
}

// SetAllocationBurn -
func (o *Operation) SetAllocationBurn(constants protocol.Constants) {
	o.AllocatedDestinationContractBurned = 257 * constants.CostPerByte
}

// SetBurned -
func (o *Operation) SetBurned(constants protocol.Constants) {
	if o.Status != types.OperationStatusApplied {
		return
	}
	var burned int64

	if o.PaidStorageSizeDiff != 0 {
		burned += o.PaidStorageSizeDiff * constants.CostPerByte
	}

	if o.AllocatedDestinationContract {
		o.SetAllocationBurn(constants)
		burned += o.AllocatedDestinationContractBurned
	}

	o.Burned = burned
}

// IsEntrypoint -
func (o *Operation) IsEntrypoint(entrypoint string) bool {
	return o.Entrypoint.EqualString(entrypoint)
}

// IsOrigination -
func (o *Operation) IsOrigination() bool {
	return o.Kind == types.OperationKindOrigination || o.Kind == types.OperationKindOriginationNew
}

// IsTransaction -
func (o *Operation) IsTransaction() bool {
	return o.Kind == types.OperationKindTransaction
}

// IsApplied -
func (o *Operation) IsApplied() bool {
	return o.Status == types.OperationStatusApplied
}

// IsCall -
func (o *Operation) IsCall() bool {
	return bcd.IsContract(o.Destination) && len(o.Parameters) > 0
}

// EmptyTransfer -
func (o Operation) EmptyTransfer() *transfer.Transfer {
	return &transfer.Transfer{
		Network:     o.Network,
		Contract:    o.Destination,
		Status:      o.Status,
		Timestamp:   o.Timestamp,
		Level:       o.Level,
		Initiator:   o.Source,
		Entrypoint:  o.Entrypoint.String(),
		Amount:      decimal.Zero,
		OperationID: o.ID,
	}
}

// Result -
type Result struct {
	Status                       string             `json:"-"`
	ConsumedGas                  int64              `json:"consumed_gas,omitempty"`
	StorageSize                  int64              `json:"storage_size,omitempty"`
	PaidStorageSizeDiff          int64              `json:"paid_storage_size_diff,omitempty"`
	AllocatedDestinationContract bool               `json:"allocated_destination_contract,omitempty"`
	Originated                   string             `json:"-"`
	Errors                       []*tezerrors.Error `json:"-"`
}

// Stats -
type Stats struct {
	Count      int64
	LastAction time.Time
}

// Pageable -
type Pageable struct {
	Operations []Operation `json:"operations"`
	LastID     string      `json:"last_id"`
}

// TokenMethodUsageStats -
type TokenMethodUsageStats struct {
	Count       int64
	ConsumedGas int64
}

// TokenUsageStats -
type TokenUsageStats map[string]TokenMethodUsageStats
