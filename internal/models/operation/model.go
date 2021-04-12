package operation

import (
	"errors"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Operation -
type Operation struct {
	ID int64 `json:"-"`

	ContentIndex int64 `json:"content_index,omitempty" gorm:",default:0"`
	Level        int64 `json:"level"`
	Counter      int64 `json:"counter,omitempty" gorm:"index:opg_idx"`
	Fee          int64 `json:"fee,omitempty"`
	GasLimit     int64 `json:"gas_limit,omitempty"`
	StorageLimit int64 `json:"storage_limit,omitempty"`
	Amount       int64 `json:"amount,omitempty"`

	ConsumedGas                        int64 `json:"consumed_gas,omitempty"`
	StorageSize                        int64 `json:"storage_size,omitempty"`
	PaidStorageSizeDiff                int64 `json:"paid_storage_size_diff,omitempty"`
	Burned                             int64 `json:"burned,omitempty"`
	AllocatedDestinationContractBurned int64 `json:"allocated_destination_contract_burned,omitempty"`

	Nonce    *int64 `json:"nonce,omitempty" gorm:"index:opg_idx"`
	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash" gorm:"index:opg_idx"`

	Timestamp        time.Time `json:"timestamp"`
	Status           string    `json:"status"`
	Kind             string    `json:"kind"`
	Initiator        string    `json:"initiator"`
	Source           string    `json:"source" gorm:"index:source_idx"`
	Destination      string    `json:"destination,omitempty" gorm:"index:destination_idx"`
	Delegate         string    `json:"delegate,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	DelegateAlias    string    `json:"delegate_alias,omitempty"`
	Parameters       []byte    `json:"parameters,omitempty"`
	DeffatedStorage  []byte    `json:"deffated_storage"`

	Tags pq.StringArray `json:"tags,omitempty" gorm:"type:text[]"`

	Script []byte `json:"-"  gorm:"-"`

	Errors           tezerrors.Errors `json:"errors,omitempty" gorm:"type:bytes"`
	ParameterStrings pq.StringArray   `json:"parameter_strings,omitempty" gorm:"type:text[]"`
	StorageStrings   pq.StringArray   `json:"storage_strings,omitempty" gorm:"type:text[]"`

	AllocatedDestinationContract bool `json:"allocated_destination_contract,omitempty"`
	Internal                     bool `json:"internal" gorm:",default:false"`

	AST *ast.Script `json:"-" gorm:"-"`
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

// GetQueues -
func (o *Operation) GetQueues() []string {
	return []string{"operations"}
}

// MarshalToQueue -
func (o *Operation) MarshalToQueue() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", o.ID)), nil
}

// LogFields -
func (o *Operation) LogFields() logrus.Fields {
	return logrus.Fields{
		"network": o.Network,
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
	if o.Status != consts.Applied {
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
	return o.Entrypoint == entrypoint
}

// IsOrigination -
func (o *Operation) IsOrigination() bool {
	return o.Kind == consts.Origination || o.Kind == consts.OriginationNew
}

// IsTransaction -
func (o *Operation) IsTransaction() bool {
	return o.Kind == consts.Transaction
}

// IsApplied -
func (o *Operation) IsApplied() bool {
	return o.Status == consts.Applied
}

// IsCall -
func (o *Operation) IsCall() bool {
	return bcd.IsContract(o.Destination) && len(o.Parameters) > 0
}

// HasTag -
func (o *Operation) HasTag(tag string) bool {
	for i := range o.Tags {
		if o.Tags[i] == tag {
			return true
		}
	}
	return false
}

// InitScript -
func (o *Operation) InitScript() (err error) {
	if o.Script == nil {
		return errors.New("Uninitialized script")
	}
	o.AST, err = ast.NewScriptWithoutCode(o.Script)
	return err
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
