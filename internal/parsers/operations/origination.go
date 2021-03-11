package operations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
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
func (p Origination) Parse(data noderpc.Operation) ([]models.Model, error) {
	origination := operation.Operation{
		ID:            helpers.GenerateID(),
		Network:       p.network,
		Hash:          p.hash,
		Protocol:      p.head.Protocol,
		Level:         p.head.Level,
		Timestamp:     p.head.Timestamp,
		Kind:          data.Kind,
		Initiator:     data.Source,
		Source:        data.Source,
		Fee:           data.Fee,
		Counter:       data.Counter,
		GasLimit:      data.GasLimit,
		StorageLimit:  data.StorageLimit,
		Amount:        *data.Balance,
		PublicKey:     data.PublicKey,
		ManagerPubKey: data.ManagerPubKey,
		Delegate:      data.Delegate,
		Parameters:    string(data.Parameters),
		Nonce:         data.Nonce,
		IndexedTime:   time.Now().UnixNano() / 1000,
		ContentIndex:  p.contentIdx,
		Script:        data.Script,
	}

	p.fillInternal(&origination)

	result := parseOperationResult(&data)
	origination.Result = result
	origination.Status = origination.Result.Status
	origination.Errors = origination.Result.Errors
	origination.Destination = result.Originated

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

func (p Origination) appliedHandler(item noderpc.Operation, origination *operation.Operation) ([]models.Model, error) {
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
