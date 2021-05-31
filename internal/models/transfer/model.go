package transfer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Transfer -
type Transfer struct {
	ID           int64           `json:"-"`
	Network      types.Network   `json:"network" gorm:"index:transfers_network_idx;index:transfers_token_idx"`
	Contract     string          `json:"contract" gorm:"index:transfers_token_idx"`
	Initiator    string          `json:"initiator"`
	Hash         string          `json:"hash"`
	Status       string          `json:"status"`
	Timestamp    time.Time       `json:"timestamp" gorm:"index:transfers_timestamp_idx"`
	Level        int64           `json:"level" gorm:"index:transfers_network_idx"`
	From         string          `json:"from" gorm:"index:transfers_from_idx"`
	To           string          `json:"to" gorm:"index:transfers_to_idx"`
	TokenID      uint64          `json:"token_id" gorm:"type:numeric(50,0);index:transfers_token_idx"`
	Amount       decimal.Decimal `json:"amount" gorm:"type:numeric(100,0)"`
	AmountString string          `json:"amount_string"`
	Counter      int64           `json:"counter"`
	Nonce        *int64          `json:"nonce,omitempty"`
	Parent       string          `json:"parent,omitempty"`
	Entrypoint   string          `json:"entrypoint,omitempty"`
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
		"network":  t.Network.String(),
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
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
	t.AmountString = t.Amount.String()
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
		Amount:     decimal.Zero,
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
		Balance:  decimal.Zero,
	}
	switch {
	case from && rollback:
		tb.Address = t.From
		tb.Balance = t.Amount
	case !from && rollback:
		tb.Address = t.To
		tb.Balance = t.Amount.Neg()
	case from && !rollback:
		tb.Address = t.From
		tb.Balance = t.Amount.Neg()
	case !from && !rollback:
		tb.Address = t.To
		tb.Balance = t.Amount
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
	Balance decimal.Decimal
	Address string
	TokenID uint64
}
