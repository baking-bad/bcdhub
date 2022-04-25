package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
)

// Operation -
type Operation struct {
	ID       string `json:"-"`
	Network  string `json:"network"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`

	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Level            int64     `json:"level"`
	Kind             string    `json:"kind"`
	Initiator        string    `json:"initiator"`
	Source           string    `json:"source"`
	Destination      string    `json:"destination,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	FoundBy          string    `json:"found_by,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	PaidStorageDiff  int64     `json:"paid_storage_diff,omitempty"`
	ConsumedGas      int64     `json:"consumed_gas,omitempty"`

	DelegateAlias string `json:"delegate_alias,omitempty"`

	ParameterStrings []string `json:"parameter_strings,omitempty"`
	StorageStrings   []string `json:"storage_strings,omitempty"`
}

// NewOperation -
func NewOperation(network modelTypes.Network, model *operation.Operation) Operation {
	var o Operation
	o.ID = helpers.GenerateID()
	o.Destination = model.Destination.Address
	o.DestinationAlias = model.Destination.Alias
	if model.Entrypoint.Valid {
		o.Entrypoint = model.Entrypoint.String()
	}
	o.Hash = model.Hash
	o.Initiator = model.Initiator.Address
	o.Internal = model.Internal
	o.Kind = model.Kind.String()
	o.Level = model.Level
	o.Network = network.String()
	o.Source = model.Source.Address
	o.SourceAlias = model.Source.Alias
	o.Status = model.Status.String()
	o.Timestamp = model.Timestamp.UTC()
	o.Delegate = model.Delegate.Address
	o.DelegateAlias = model.Delegate.Alias
	o.PaidStorageDiff = model.PaidStorageSizeDiff
	o.ConsumedGas = model.ConsumedGas

	if len(model.DeffatedStorage) > 0 {
		var tree ast.UntypedAST
		if err := json.Unmarshal(model.DeffatedStorage, &tree); err == nil {
			o.StorageStrings, err = tree.GetStrings(true)
			if err != nil {
				logger.Error().Err(err).Msg("GetStrings for storage")
			}
		} else {
			logger.Error().Err(err).Msg("GetStrings for storage")
		}
	}

	if model.Kind == modelTypes.OperationKindTransaction && len(model.Parameters) > 0 {
		params := types.NewParameters(model.Parameters)

		var tree ast.UntypedAST
		if err := json.Unmarshal(params.Value, &tree); err == nil {
			o.ParameterStrings, err = tree.GetStrings(true)
			if err != nil {
				logger.Error().Err(err).Msg("GetStrings for storage")
			}
		} else {
			logger.Error().Err(err).Msg("GetStrings for storage")
		}
	}
	return o
}

// GetID -
func (o Operation) GetID() string {
	return o.ID
}

// GetIndex -
func (o Operation) GetIndex() string {
	return models.DocOperations
}

// GetScores -
func (o Operation) GetScores(search string) []string {
	return []string{
		"hash^10",
		"entrypoint^8",
		"errors.with^5",
		"errors.id^5",
		"source_alias",
	}
}

// GetFields -
func (o Operation) GetFields() []string {
	return []string{
		"entrypoint",
		"errors.with",
		"errors.id",
		"source_alias",
		"hash",
	}
}

// Parse  -
func (o Operation) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return &Item{
		Type:       o.GetIndex(),
		Value:      o.Hash,
		Body:       &o,
		Highlights: highlight,
		Network:    o.Network,
	}, nil
}
