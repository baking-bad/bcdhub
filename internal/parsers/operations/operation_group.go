package operations

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
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
func (opg Group) Parse(data noderpc.OperationGroup) (*parsers.Result, error) {
	result := parsers.NewResult()

	opg.hash = data.Hash
	helpers.SetTagSentry("hash", opg.hash)

	for idx, item := range data.Contents {
		opg.contentIdx = int64(idx)

		contentParser := NewContent(opg.ParseParams)
		contentResult, err := contentParser.Parse(item)
		if err != nil {
			return nil, err
		}
		result.Merge(contentResult)
		contentParser.clear()
	}

	return result, nil
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
func (content Content) Parse(data noderpc.Operation) (*parsers.Result, error) {
	if !content.needParse(data) {
		return nil, nil
	}
	result := parsers.NewResult()

	switch data.Kind {
	case consts.Origination, consts.OriginationNew:
		originationResult, err := NewOrigination(content.ParseParams).Parse(data)
		if err != nil {
			return nil, err
		}
		result.Merge(originationResult)
	case consts.Transaction:
		txResult, err := NewTransaction(content.ParseParams).Parse(data)
		if err != nil {
			return nil, err
		}
		result.Merge(txResult)
	case consts.RegisterGlobalConstant:
		txResult, err := NewRegisterGlobalConstant(content.ParseParams).Parse(data)
		if err != nil {
			return nil, err
		}
		result.Merge(txResult)
	default:
		return nil, errors.Errorf("Invalid operation kind: %s", data.Kind)
	}

	internalResult, err := content.parseInternal(data)
	if err != nil {
		return nil, err
	}
	result.Merge(internalResult)

	return result, nil
}

func (content Content) needParse(item noderpc.Operation) bool {
	var destination string
	if item.Destination != nil {
		destination = *item.Destination
	}
	prefixCondition := bcd.IsContract(item.Source) || bcd.IsContract(destination)
	transactionCondition := item.Kind == consts.Transaction && prefixCondition
	originationCondition := (item.Kind == consts.Origination || item.Kind == consts.OriginationNew) && item.Script != nil
	registerGlobalConstantCondition := item.Kind == consts.RegisterGlobalConstant
	return originationCondition || transactionCondition || registerGlobalConstantCondition
}

func (content Content) parseInternal(data noderpc.Operation) (*parsers.Result, error) {
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

	result := parsers.NewResult()
	for i := range internals {
		parsedModels, err := content.Parse(internals[i])
		if err != nil {
			return nil, err
		}
		result.Merge(parsedModels)
	}
	return result, nil
}

func (content *Content) clear() {
	content.main = nil
}
