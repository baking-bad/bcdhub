package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// Operation -
type Operation struct {
	ID string `json:"id"`

	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`

	Timestamp     time.Time `json:"timestamp"`
	Level         int64     `json:"level"`
	Kind          string    `json:"kind"`
	Source        string    `json:"source"`
	Fee           int64     `json:"fee,omitempty"`
	Counter       int64     `json:"counter,omitempty"`
	GasLimit      int64     `json:"gas_limit,omitempty"`
	StorageLimit  int64     `json:"storage_limit,omitempty"`
	Amount        int64     `json:"amount,omitempty"`
	Destination   string    `json:"destination,omitempty"`
	PublicKey     string    `json:"public_key,omitempty"`
	ManagerPubKey string    `json:"manager_pubkey,omitempty"`
	Balance       int64     `json:"balance,omitempty"`
	Delegate      string    `json:"delegate,omitempty"`
	Parameters    string    `json:"parameters,omitempty"`

	BalanceUpdates []BalanceUpdate  `json:"balance_updates,omitempty"`
	Result         *OperationResult `json:"result,omitempty"`

	DeffatedStorage string `json:"deffated_storage,omitempty"`
	ScrollID        string `json:"scroll_id"`
}

// ParseElasticJSON -
func (o *Operation) ParseElasticJSON(resp gjson.Result) {
	o.ID = resp.Get("_id").String()
	o.ScrollID = resp.Get("_scroll_id").String()

	o.Protocol = resp.Get("_source.protocol").String()
	o.Hash = resp.Get("_source.hash").String()
	o.Internal = resp.Get("_source.internal").Bool()
	o.Network = resp.Get("_source.network").String()
	o.Timestamp = resp.Get("_source.timestamp").Time().UTC()

	o.Level = resp.Get("_source.level").Int()
	o.Kind = resp.Get("_source.kind").String()
	o.Source = resp.Get("_source.source").String()
	o.Fee = resp.Get("_source.fee").Int()
	o.Counter = resp.Get("_source.counter").Int()
	o.GasLimit = resp.Get("_source.gas_limit").Int()
	o.StorageLimit = resp.Get("_source.storage_limit").Int()
	o.Amount = resp.Get("_source.amount").Int()
	o.Destination = resp.Get("_source.destination").String()
	o.PublicKey = resp.Get("_source.public_key").String()
	o.ManagerPubKey = resp.Get("_source.manager_pubkey").String()
	o.Balance = resp.Get("_source.balance").Int()
	o.Delegate = resp.Get("_source.delegate").String()
	o.Parameters = resp.Get("_source.parameters").String()

	var opResult OperationResult
	opResult.ParseElasticJSON(resp.Get("_source.result"))
	o.Result = &opResult

	o.DeffatedStorage = resp.Get("_source.deffated_storage").String()

	count := resp.Get("_source.balance_updates.#").Int()
	bu := make([]BalanceUpdate, count)
	for i, hit := range resp.Get("_source.balance_updates").Array() {
		var b BalanceUpdate
		b.ParseElasticJSON(hit)
		bu[i] = b
	}
	o.BalanceUpdates = bu
}

// BalanceUpdate -
type BalanceUpdate struct {
	Kind     string `json:"kind"`
	Contract string `json:"contract,omitempty"`
	Change   int64  `json:"change"`
	Category string `json:"category,omitempty"`
	Delegate string `json:"delegate,omitempty"`
	Cycle    int    `json:"cycle,omitempty"`
}

// ParseElasticJSON -
func (b *BalanceUpdate) ParseElasticJSON(data gjson.Result) {
	b.Kind = data.Get("kind").String()
	b.Contract = data.Get("contract").String()
	b.Change = data.Get("change").Int()
	b.Category = data.Get("category").String()
	b.Delegate = data.Get("delegate").String()
	b.Cycle = int(data.Get("cycle").Int())
}

// OperationResult -
type OperationResult struct {
	Status                       string `json:"status"`
	ConsumedGas                  int64  `json:"consumed_gas,omitempty"`
	StorageSize                  int64  `json:"storage_size,omitempty"`
	PaidStorageSizeDiff          int64  `json:"paid_storage_size_diff,omitempty"`
	AllocatedDestinationContract bool   `json:"allocated_destination_contract,omitempty"`
	Originated                   string `json:"-"`
	Errors                       string `json:"errors,omitempty"`

	BalanceUpdates []BalanceUpdate `json:"balance_updates,omitempty"`
}

// ParseElasticJSON -
func (o *OperationResult) ParseElasticJSON(data gjson.Result) {
	count := data.Get("balance_updates.#").Int()
	bu := make([]BalanceUpdate, count)
	for i, hit := range data.Get("balance_updates").Array() {
		var b BalanceUpdate
		b.ParseElasticJSON(hit)
		bu[i] = b
	}
	o.Status = data.Get("status").String()
	o.ConsumedGas = data.Get("consumed_gas").Int()
	o.StorageSize = data.Get("storage_size").Int()
	o.PaidStorageSizeDiff = data.Get("paid_storage_size_diff").Int()
	o.AllocatedDestinationContract = data.Get("allocated_destination_contract").Bool()
	o.Errors = data.Get("errors").String()
	o.BalanceUpdates = bu
}
