package transfer

import (
	"fmt"
	"math/big"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Transfer -
type Transfer struct {
	ID           int64     `json:"-"`
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
	AmountString string    `json:"amount_string"`
	Counter      int64     `json:"counter"`
	Nonce        *int64    `json:"nonce,omitempty"`
	Parent       string    `json:"parent,omitempty"`
	Entrypoint   string    `json:"entrypoint,omitempty"`

	Value *big.Int `json:"-" gorm:"-"`
}

// GetID -
func (t *Transfer) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfers"
}

// Save -
func (t *Transfer) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(t).Error
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

// AfterFind -
func (t *Transfer) AfterFind(tx *gorm.DB) (err error) {
	return t.unmarshal()
}

// BeforeCreate -
func (t *Transfer) BeforeCreate(tx *gorm.DB) (err error) {
	return t.marshal()
}

// BeforeUpdate -
func (t *Transfer) BeforeUpdate(tx *gorm.DB) (err error) {
	return t.marshal()
}

func (t *Transfer) marshal() error {
	if t.Value == nil {
		return errors.New("Nil amount in transfer")
	}
	t.AmountString = t.Value.String()
	t.Amount, _ = new(big.Float).SetInt(t.Value).Float64()
	return nil
}

func (t *Transfer) unmarshal() error {
	if t.Value == nil {
		t.Value = big.NewInt(0)
	}

	if _, ok := t.Value.SetString(t.AmountString, 10); !ok {
		return errors.Errorf("Invalid amount in transfer: %s", t.AmountString)
	}

	return nil
}

// EmptyTransfer -
func EmptyTransfer(o operation.Operation) *Transfer {
	return &Transfer{
		Network:    o.Network,
		Contract:   o.Destination,
		Hash:       o.Hash,
		Status:     o.Status,
		Timestamp:  o.Timestamp,
		Level:      o.Level,
		Initiator:  o.Source,
		Counter:    o.Counter,
		Nonce:      o.Nonce,
		Entrypoint: o.Entrypoint,
		Value:      new(big.Int),
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
		tb.Value.Set(t.Value)
	case !from && rollback:
		tb.Address = t.To
		tb.Value.Neg(t.Value)
	case from && !rollback:
		tb.Address = t.From
		tb.Value.Neg(t.Value)
	case !from && !rollback:
		tb.Address = t.To
		tb.Value.Set(t.Value)
	}
	return tb
}

// Pageable -
type Pageable struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id"`
}

// Balance -
type Balance struct {
	Balance *big.Int
	Address string
	TokenID uint64
}
