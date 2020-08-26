package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Transfer -
type Transfer struct {
	ID          string    `json:"-"`
	IndexedTime int64     `json:"indexed_time"`
	Network     string    `json:"network"`
	Contract    string    `json:"contract"`
	Alias       string    `json:"alias,omitempty"`
	Hash        string    `json:"hash"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Level       int64     `json:"level"`
	From        string    `json:"from"`
	FromAlias   string    `json:"from_alias,omitempty"`
	To          string    `json:"to"`
	ToAlias     string    `json:"to_alias,omitempty"`
	TokenID     int64     `json:"token_id"`
	Amount      int64     `json:"amount"`
}

// ParseElasticJSON -
func (t *Transfer) ParseElasticJSON(resp gjson.Result) {
	t.ID = resp.Get("_id").String()
	t.IndexedTime = resp.Get("_source.indexed_time").Int()
	t.Network = resp.Get("_source.network").String()
	t.Contract = resp.Get("_source.contract").String()
	t.Alias = resp.Get("_source.alias").String()
	t.Hash = resp.Get("_source.hash").String()
	t.Status = resp.Get("_source.status").String()
	t.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	t.Level = resp.Get("_source.level").Int()
	t.From = resp.Get("_source.from").String()
	t.FromAlias = resp.Get("_source.from_alias").String()
	t.To = resp.Get("_source.to").String()
	t.ToAlias = resp.Get("_source.to_alias").String()
	t.Amount = resp.Get("_source.amount").Int()
	t.TokenID = resp.Get("_source.token_id").Int()
}

// GetID -
func (t *Transfer) GetID() string {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfer"
}

// GetQueue -
func (t *Transfer) GetQueue() string {
	return "transfers"
}

// GetScores -
func (t *Transfer) GetScores(search string) []string {
	return []string{
		"contract^8",
		"hash^7",
		"from^7",
		"to^6",
	}
}

// FoundByName -
func (t *Transfer) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := t.GetScores("")
	return getFoundBy(keys, categories)
}

func newTransfer(o *Operation) *Transfer {
	return &Transfer{
		ID:          helpers.GenerateID(),
		IndexedTime: o.IndexedTime,
		Network:     o.Network,
		Contract:    o.Destination,
		Hash:        o.Hash,
		Status:      o.Status,
		Timestamp:   o.Timestamp,
		Level:       o.Level,
	}
}

func getAddress(data gjson.Result) (string, error) {
	if data.Get("string").Exists() {
		return data.Get("string").String(), nil
	}

	if data.Get("bytes").Exists() {
		return unpack.Address(data.Get("bytes").String())
	}
	return "", errors.Errorf("Unknown address data: %s", data.Raw)
}

// CreateTransfers -
func CreateTransfers(o *Operation) ([]*Transfer, error) {
	if o.Entrypoint != "transfer" {
		return nil, nil
	}

	parameters := getParameters(o.Parameters)

	transfers := make([]*Transfer, 0)
	for i := range o.Tags {
		switch o.Tags[i] {
		case consts.FA12Tag:
			transfer := newTransfer(o)
			fromAddr, err := getAddress(parameters.Get("args.0"))
			if err != nil {
				return nil, err
			}
			toAddr, err := getAddress(parameters.Get("args.1.args.0"))
			if err != nil {
				return nil, err
			}
			transfer.From = fromAddr
			transfer.To = toAddr
			transfer.Amount = parameters.Get("args.1.args.1.int").Int()
			transfers = append(transfers, transfer)
			return transfers, nil
		case consts.FA2Tag:
			for _, from := range parameters.Array() {
				fromAddr, err := getAddress(from.Get("args.0"))
				if err != nil {
					return nil, err
				}
				for _, to := range from.Get("args.1").Array() {
					toAddr, err := getAddress(from.Get("args.0"))
					if err != nil {
						return nil, err
					}
					transfer := newTransfer(o)
					transfer.From = fromAddr
					transfer.To = toAddr
					transfer.Amount = to.Get("args.1.args.1").Int()
					transfer.TokenID = to.Get("args.1.args.0").Int()
					transfers = append(transfers, transfer)
				}
			}
			return transfers, nil
		default:
		}
	}
	return nil, nil
}

func getParameters(str string) gjson.Result {
	parameters := gjson.Parse(str)
	if !parameters.Get("value").Exists() {
		return parameters
	}
	parameters = parameters.Get("value")
	for end := false; !end; {
		prim := parameters.Get("prim|@lower").String()
		end = prim != consts.LEFT && prim != consts.RIGHT
		if !end {
			parameters = parameters.Get("args.0")
		}
	}
	return parameters
}
