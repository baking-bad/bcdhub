package operation

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Operation -
type Operation struct {
	ID int64 `json:"-"`

	IndexedTime  int64 `json:"indexed_time"`
	ContentIndex int64 `json:"content_index,omitempty" gorm:",default:0"`

	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal" gorm:",default:false"`
	Nonce    *int64 `json:"nonce,omitempty"`

	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Level            int64     `json:"level"`
	Kind             string    `json:"kind"`
	Initiator        string    `json:"initiator"`
	Source           string    `json:"source"`
	Fee              int64     `json:"fee,omitempty"`
	Counter          int64     `json:"counter,omitempty"`
	GasLimit         int64     `json:"gas_limit,omitempty"`
	StorageLimit     int64     `json:"storage_limit,omitempty"`
	Amount           int64     `json:"amount,omitempty"`
	Destination      string    `json:"destination,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	Parameters       []byte    `json:"parameters,omitempty"`
	DeffatedStorage  []byte    `json:"deffated_storage"`
	DelegateAlias    string    `json:"delegate_alias,omitempty"`

	ConsumedGas                        int64            `json:"consumed_gas,omitempty"`
	StorageSize                        int64            `json:"storage_size,omitempty"`
	PaidStorageSizeDiff                int64            `json:"paid_storage_size_diff,omitempty"`
	AllocatedDestinationContract       bool             `json:"allocated_destination_contract,omitempty"`
	Errors                             tezerrors.Errors `json:"errors,omitempty" gorm:"type:bytes"`
	Burned                             int64            `json:"burned,omitempty"`
	AllocatedDestinationContractBurned int64            `json:"allocated_destination_contract_burned,omitempty"`

	Tags pq.StringArray `json:"tags,omitempty" gorm:"type:text[]"`

	Script []byte `json:"-"  gorm:"-"`

	ParameterStrings pq.StringArray `json:"parameter_strings,omitempty" gorm:"type:text[]"`
	StorageStrings   pq.StringArray `json:"storage_strings,omitempty" gorm:"type:text[]"`
}

// GetID -
func (o *Operation) GetID() int64 {
	return o.ID
}

// GetIndex -
func (o *Operation) GetIndex() string {
	return "operations"
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
