package operation

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// Operation -
type Operation struct {
	ID string `json:"-"`

	IndexedTime  int64 `json:"indexed_time"`
	ContentIndex int64 `json:"content_index,omitempty"`

	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`
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
	PublicKey        string    `json:"public_key,omitempty"`
	ManagerPubKey    string    `json:"manager_pubkey,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	Parameters       string    `json:"parameters,omitempty"`
	FoundBy          string    `json:"found_by,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`

	Result                             *Result            `json:"result,omitempty"`
	Errors                             []*tezerrors.Error `json:"errors,omitempty"`
	Burned                             int64              `json:"burned,omitempty"`
	AllocatedDestinationContractBurned int64              `json:"allocated_destination_contract_burned,omitempty"`

	DeffatedStorage string       `json:"deffated_storage"`
	Script          gjson.Result `json:"-"`

	DelegateAlias string `json:"delegate_alias,omitempty"`

	ParameterStrings []string `json:"parameter_strings,omitempty"`
	StorageStrings   []string `json:"storage_strings,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// GetID -
func (o *Operation) GetID() string {
	return o.ID
}

// GetIndex -
func (o *Operation) GetIndex() string {
	return "operation"
}

// GetQueues -
func (o *Operation) GetQueues() []string {
	return []string{"operations"}
}

// MarshalToQueue -
func (o *Operation) MarshalToQueue() ([]byte, error) {
	return []byte(o.ID), nil
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

	if o.Result == nil {
		return
	}

	var burned int64

	if o.Result.PaidStorageSizeDiff != 0 {
		burned += o.Result.PaidStorageSizeDiff * constants.CostPerByte
	}

	if o.Result.AllocatedDestinationContract {
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
	return bcd.IsContract(o.Destination) && o.Parameters != ""
}

// GetScriptSection -
func (o *Operation) GetScriptSection(name string) gjson.Result {
	return o.Script.Get(fmt.Sprintf("code.#(prim==\"%s\").args.0", name))
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
