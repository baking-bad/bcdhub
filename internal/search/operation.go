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

	DelegateAlias string `json:"delegate_alias,omitempty"`

	ParameterStrings []string `json:"parameter_strings,omitempty"`
	StorageStrings   []string `json:"storage_strings,omitempty"`
}

// GetID -
func (o *Operation) GetID() string {
	return o.ID
}

// GetIndex -
func (o *Operation) GetIndex() string {
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

// Prepare -
func (o *Operation) Prepare(model models.Model) {
	op, ok := model.(*operation.Operation)
	if !ok {
		return
	}

	o.ID = helpers.GenerateID()
	o.Destination = op.Destination.Address
	o.DestinationAlias = op.Destination.Alias
	if op.Entrypoint.Valid {
		o.Entrypoint = op.Entrypoint.String()
	}
	o.Hash = op.Hash
	o.Initiator = op.Initiator.Address
	o.Internal = op.Internal
	o.Kind = op.Kind.String()
	o.Level = op.Level
	o.Network = op.Network.String()
	o.Source = op.Source.Address
	o.SourceAlias = op.Source.Alias
	o.Status = op.Status.String()
	o.Timestamp = op.Timestamp.UTC()
	o.Delegate = op.Delegate.Address
	o.DelegateAlias = op.Delegate.Alias

	if len(op.DeffatedStorage) > 0 {
		var tree ast.UntypedAST
		if err := json.Unmarshal(op.DeffatedStorage, &tree); err == nil {
			o.StorageStrings, err = tree.GetStrings(true)
			if err != nil {
				logger.Error().Err(err).Msg("GetStrings for storage")
			}
		} else {
			logger.Error().Err(err).Msg("GetStrings for storage")
		}
	}

	if op.Kind == modelTypes.OperationKindTransaction && len(op.Parameters) > 0 {
		params := types.NewParameters(op.Parameters)

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
}
