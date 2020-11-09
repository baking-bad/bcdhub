package models

import (
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// BalanceUpdate -
type BalanceUpdate struct {
	ID            string `json:"-"`
	Change        int64  `json:"change"`
	Network       string `json:"network"`
	Contract      string `json:"contract"`
	OperationHash string `json:"hash"`
	ContentIndex  int64  `json:"content_index"`
	Nonce         *int64 `json:"nonce,omitempty"`
	Level         int64  `json:"level"`
}

// GetID -
func (b *BalanceUpdate) GetID() string {
	return b.ID
}

// GetIndex -
func (b *BalanceUpdate) GetIndex() string {
	return "balance_update"
}

// GetQueues -
func (b *BalanceUpdate) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BalanceUpdate) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (b *BalanceUpdate) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  b.Network,
		"contract": b.Contract,
		"change":   b.Change,
	}
}

// GetScores -
func (b *BalanceUpdate) GetScores(search string) []string {
	return []string{}
}

// FoundByName -
func (b *BalanceUpdate) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := b.GetScores("")
	return utils.GetFoundBy(keys, categories)
}
