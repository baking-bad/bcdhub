package operations

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Group -
type Group struct {
	*ParseParams
}

// NewGroup -
func NewGroup(params *ParseParams) Group {
	return Group{params}
}

// Parse -
func (opg Group) Parse(data noderpc.OperationGroup) ([]models.Model, error) {
	parsedModels := make([]models.Model, 0)

	opg.hash = data.Hash
	helpers.SetTagSentry("hash", opg.hash)

	for idx, item := range data.Contents {
		opg.contentIdx = int64(idx)

		contentParser := NewContent(opg.ParseParams)
		models, err := contentParser.Parse(item)
		if err != nil {
			return nil, err
		}
		parsedModels = append(parsedModels, models...)
		contentParser.clear()
	}

	return parsedModels, nil
}

// Content -
type Content struct {
	*ParseParams
}

// NewContent -
func NewContent(params *ParseParams) Content {
	return Content{params}
}

// Parse -
func (content Content) Parse(data noderpc.Operation) ([]models.Model, error) {
	if !content.needParse(data) {
		return nil, nil
	}

	models := make([]models.Model, 0)

	switch data.Kind {
	case consts.Origination, consts.OriginationNew:
		originationModels, err := NewOrigination(content.ParseParams).Parse(data)
		if err != nil {
			return nil, err
		}
		models = append(models, originationModels...)
	case consts.Transaction:
		txModels, err := NewTransaction(content.ParseParams).Parse(data)
		if err != nil {
			return nil, err
		}
		models = append(models, txModels...)
	default:
		return nil, errors.Errorf("Invalid operation kind: %s", data.Kind)
	}

	internalModels, err := content.parseInternal(data)
	if err != nil {
		return nil, err
	}
	models = append(models, internalModels...)

	return models, nil
}

func (content Content) needParse(item noderpc.Operation) bool {
	var destination string
	if item.Destination != nil {
		destination = *item.Destination
	}
	prefixCondition := bcd.IsContract(item.Source) || bcd.IsContract(destination)
	transactionCondition := item.Kind == consts.Transaction && prefixCondition
	originationCondition := (item.Kind == consts.Origination || item.Kind == consts.OriginationNew) && item.Script != nil
	return originationCondition || transactionCondition
}

func (content Content) parseInternal(data noderpc.Operation) ([]models.Model, error) {
	if data.Metadata == nil {
		return nil, nil
	}
	internals := data.Metadata.Internal
	if internals == nil {
		internals = data.Metadata.InternalOperations
		if internals == nil {
			return nil, nil
		}
	}

	internalModels := make([]models.Model, 0)
	for i := range internals {
		parsedModels, err := content.Parse(internals[i])
		if err != nil {
			return nil, err
		}
		internalModels = append(internalModels, parsedModels...)
	}
	return internalModels, nil
}

func (content *Content) clear() {
	content.main = nil
}
