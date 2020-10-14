package operations

import (
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Origination -
type Origination struct {
	*ParseParams
}

// NewOrigination -
func NewOrigination(params *ParseParams) Origination {
	return Origination{params}
}

// Parse -
func (p Origination) Parse(data gjson.Result) ([]elastic.Model, error) {
	origination := models.Operation{
		ID:             helpers.GenerateID(),
		Network:        p.network,
		Hash:           p.hash,
		Protocol:       p.head.Protocol,
		Level:          p.head.Level,
		Timestamp:      p.head.Timestamp,
		Kind:           data.Get("kind").String(),
		Initiator:      data.Get("source").String(),
		Source:         data.Get("source").String(),
		Fee:            data.Get("fee").Int(),
		Counter:        data.Get("counter").Int(),
		GasLimit:       data.Get("gas_limit").Int(),
		StorageLimit:   data.Get("storage_limit").Int(),
		Amount:         data.Get("balance").Int(),
		PublicKey:      data.Get("public_key").String(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		Script:         data.Get("script"),
		BalanceUpdates: NewBalanceUpdate("metadata").Parse(data),
		IndexedTime:    time.Now().UnixNano() / 1000,
		ContentIndex:   p.contentIdx,
	}

	p.fillInternal(&origination)

	operationMetadata := parseMetadata(data)
	origination.Result = &operationMetadata.Result
	origination.BalanceUpdates = append(origination.BalanceUpdates, operationMetadata.BalanceUpdates...)
	origination.Status = origination.Result.Status
	origination.Errors = origination.Result.Errors
	origination.Destination = operationMetadata.Result.Originated

	origination.SetBurned(p.constants)

	originationModels := []elastic.Model{&origination}

	switch origination.Status {
	case consts.Applied:
		appliedModels, err := p.appliedHandler(data, &origination)
		if err != nil {
			return nil, err
		}
		originationModels = append(originationModels, appliedModels...)
	}

	return originationModels, nil
}

func (p Origination) appliedHandler(item gjson.Result, origination *models.Operation) ([]elastic.Model, error) {
	if !strings.HasPrefix(origination.Destination, "KT") || origination.Status == consts.Applied {
		return nil, nil
	}

	models := make([]elastic.Model, 0)

	contractModels, err := p.contractParser.Parse(*origination)
	if err != nil {
		return nil, err
	}
	models = append(models, contractModels...)

	metadata, err := p.contractParser.GetContractMetadata(origination.Destination)
	if err != nil {
		metadata, err = meta.GetContractMetadata(p.es, origination.Destination)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				return nil, nil
			}
			return nil, err
		}
	}

	rs, err := NewRichStorage(p.es, p.rpc, origination, metadata).Parse(item)
	if err != nil {
		return nil, err
	}
	if !rs.Empty {
		origination.DeffatedStorage = rs.DeffatedStorage
		models = append(models, rs.Models...)
	}
	return models, nil
}

func (p Origination) fillInternal(tx *models.Operation) {
	if p.main == nil {
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}
