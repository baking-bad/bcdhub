package operations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
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
func (p Origination) Parse(data gjson.Result) ([]models.Model, error) {
	origination := operation.Operation{
		ID:            helpers.GenerateID(),
		Network:       p.network,
		Hash:          p.hash,
		Protocol:      p.head.Protocol,
		Level:         p.head.Level,
		Timestamp:     p.head.Timestamp,
		Kind:          data.Get("kind").String(),
		Initiator:     data.Get("source").String(),
		Source:        data.Get("source").String(),
		Fee:           data.Get("fee").Int(),
		Counter:       data.Get("counter").Int(),
		GasLimit:      data.Get("gas_limit").Int(),
		StorageLimit:  data.Get("storage_limit").Int(),
		Amount:        data.Get("balance").Int(),
		PublicKey:     data.Get("public_key").String(),
		ManagerPubKey: data.Get("manager_pubkey").String(),
		Delegate:      data.Get("delegate").String(),
		Parameters:    data.Get("parameters").String(),
		Script:        data.Get("script"),
		IndexedTime:   time.Now().UnixNano() / 1000,
		ContentIndex:  p.contentIdx,
	}

	if data.Get("nonce").Exists() {
		nonce := data.Get("nonce").Int()
		origination.Nonce = &nonce
	}

	p.fillInternal(&origination)

	operationMetadata := parseMetadata(data)
	origination.Result = &operationMetadata.Result
	origination.Status = origination.Result.Status
	origination.Errors = origination.Result.Errors
	origination.Destination = operationMetadata.Result.Originated

	origination.SetBurned(p.constants)

	originationModels := []models.Model{&origination}

	if origination.IsApplied() {
		appliedModels, err := p.appliedHandler(data, &origination)
		if err != nil {
			return nil, err
		}
		originationModels = append(originationModels, appliedModels...)
	}

	p.stackTrace.Add(origination)
	return originationModels, nil
}

func (p Origination) appliedHandler(item gjson.Result, origination *operation.Operation) ([]models.Model, error) {
	if !bcd.IsContract(origination.Destination) || !origination.IsApplied() {
		return nil, nil
	}

	models := make([]models.Model, 0)

	contractModels, err := p.contractParser.Parse(origination)
	if err != nil {
		return nil, err
	}
	models = append(models, contractModels...)

	rs, err := p.storageParser.Parse(item, origination)
	if err != nil {
		return nil, err
	}
	if !rs.Empty {
		origination.DeffatedStorage = rs.DeffatedStorage
		models = append(models, rs.Models...)
	}

	return models, nil
}

func (p Origination) fillInternal(tx *operation.Operation) {
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
