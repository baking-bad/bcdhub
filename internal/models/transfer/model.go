package transfer

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Transfer -
type Transfer struct {
	ID           int64     `json:"-"`
	IndexedTime  int64     `json:"indexed_time"`
	Network      string    `json:"network"`
	Contract     string    `json:"contract"`
	Initiator    string    `json:"initiator"`
	Hash         string    `json:"hash"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	Level        int64     `json:"level"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	TokenID      uint64    `json:"token_id" gorm:"type:numeric(50,0)"`
	Amount       float64   `json:"amount" gorm:"type:numeric(100,0)"`
	AmountStr    string    `json:"amount_str"`
	AmountBigInt *big.Int  `json:"-" gorm:"-"`
	Counter      int64     `json:"counter"`
	Nonce        *int64    `json:"nonce,omitempty"`
	Parent       string    `json:"parent,omitempty"`
}

// GetID -
func (t *Transfer) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfers"
}

// GetQueues -
func (t *Transfer) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (t *Transfer) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (t *Transfer) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  t.Network,
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
}

// EmptyTransfer -
func EmptyTransfer(o operation.Operation) *Transfer {
	return &Transfer{
		IndexedTime:  o.IndexedTime,
		Network:      o.Network,
		Contract:     o.Destination,
		Hash:         o.Hash,
		Status:       o.Status,
		Timestamp:    o.Timestamp,
		Level:        o.Level,
		Initiator:    o.Source,
		AmountBigInt: big.NewInt(0),
		Counter:      o.Counter,
		Nonce:        o.Nonce,
	}
}

// GetFromTokenBalanceID -
func (t *Transfer) GetFromTokenBalanceID() string {
	if t.From != "" {
		return fmt.Sprintf("%s_%s_%s_%d", t.Network, t.From, t.Contract, t.TokenID)
	}
	return ""
}

// GetToTokenBalanceID -
func (t *Transfer) GetToTokenBalanceID() string {
	if t.To != "" {
		return fmt.Sprintf("%s_%s_%s_%d", t.Network, t.To, t.Contract, t.TokenID)
	}
	return ""
}

// MakeTokenBalanceUpdate -
func (t *Transfer) MakeTokenBalanceUpdate(from, rollback bool) *tokenbalance.TokenBalance {
	tb := &tokenbalance.TokenBalance{
		Network:  t.Network,
		Contract: t.Contract,
		TokenID:  t.TokenID,
		Value:    big.NewInt(0),
	}
	switch {
	case from && rollback:
		tb.Address = t.From
		tb.Value.Set(t.AmountBigInt)
	case !from && rollback:
		tb.Address = t.To
		tb.Value.Neg(t.AmountBigInt)
	case from && !rollback:
		tb.Address = t.From
		tb.Value.Neg(t.AmountBigInt)
	case !from && !rollback:
		tb.Address = t.To
		tb.Value.Set(t.AmountBigInt)
	}
	return tb
}

// TokenSupply -
type TokenSupply struct {
	Supply     float64 `json:"supply"`
	Transfered float64 `json:"transfered"`
}

// Pageable -
type Pageable struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id"`
}

// BeforeScan -
func (t *Transfer) AfterFind(tx *gorm.DB) error {
	return t.unmarshal()
}

// BeforeInsert -
func (t *Transfer) BeforeCreate(tx *gorm.DB) error {
	return t.marshal()
}

// BeforeUpdate -
func (t *Transfer) BeforeUpdate(tx *gorm.DB) error {
	return t.marshal()
}

func (t *Transfer) unmarshal() error {
	t.AmountBigInt = big.NewInt(0)

	if _, ok := t.AmountBigInt.SetString(t.AmountStr, 10); !ok {
		return fmt.Errorf("Can't set balance value: %s", t.AmountStr)
	}
	return nil
}

func (t *Transfer) marshal() error {
	if t.AmountBigInt == nil {
		return fmt.Errorf("Nil balance value")
	}
	t.AmountStr = t.AmountBigInt.String()
	val, err := strconv.ParseFloat(t.AmountStr, 64)
	if err != nil {
		return err
	}
	t.Amount = val
	return nil
}

// UnmarshalJSON -
func (t *Transfer) UnmarshalJSON(data []byte) error {
	type buf Transfer
	if err := json.Unmarshal(data, (*buf)(t)); err != nil {
		return err
	}
	return t.unmarshal()
}

// MarshalJSON -
func (t *Transfer) MarshalJSON() ([]byte, error) {
	if err := t.marshal(); err != nil {
		return nil, err
	}
	type buf Transfer
	return json.Marshal((*buf)(t))
}

// SetAmountFromString -
func (t *Transfer) SetAmountFromString(val string) error {
	amount, ok := big.NewInt(0).SetString(val, 10)
	if !ok {
		return fmt.Errorf("cant create fa2 transfer amount for %s", val)
	}
	t.AmountBigInt = amount
	return nil
}
