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
func (opg Group) Parse(data noderpc.LightOperationGroup) (*parsers.Result, error) {
	result := parsers.NewResult()

	opg.hash = data.Hash
	helpers.SetTagSentry("hash", opg.hash)

	for idx, item := range data.Contents {
		opg.contentIdx = int64(idx)

		if !opg.needParse(item) {
			continue
		}

		var operation noderpc.Operation
		if err := json.Unmarshal(item.Raw, &operation); err != nil {
			return nil, err
		}

		contentParser := NewContent(opg.ParseParams)
		if err := contentParser.Parse(operation, result); err != nil {
			return nil, err
		}
		contentParser.clear()
	}

	return result, nil
}

func (Group) needParse(item noderpc.LightOperation) bool {
	var destination string
	if item.Destination != nil {
		destination = *item.Destination
	}
	prefixCondition := bcd.IsContract(item.Source) || bcd.IsContract(destination)
	transactionCondition := item.Kind == consts.Transaction && prefixCondition
	originationCondition := (item.Kind == consts.Origination || item.Kind == consts.OriginationNew)
	registerGlobalConstantCondition := item.Kind == consts.RegisterGlobalConstant
	return originationCondition || transactionCondition || registerGlobalConstantCondition
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
func (content Content) Parse(data noderpc.Operation, result *parsers.Result) error {
	switch data.Kind {
	case consts.Origination, consts.OriginationNew:
		if err := NewOrigination(content.ParseParams).Parse(data, result); err != nil {
			return err
		}
	case consts.Transaction:
		if err := NewTransaction(content.ParseParams).Parse(data, result); err != nil {
			return err
		}
	case consts.RegisterGlobalConstant:
		if err := NewRegisterGlobalConstant(content.ParseParams).Parse(data, result); err != nil {
			return err
		}
	default:
		return errors.Errorf("Invalid operation kind: %s", data.Kind)
	}

	if err := content.parseInternal(data, result); err != nil {
		return err
	}

	return nil
}

func (content Content) parseInternal(data noderpc.Operation, result *parsers.Result) error {
	if data.Metadata == nil {
		return nil
	}
	internals := data.Metadata.Internal
	if internals == nil {
		internals = data.Metadata.InternalOperations
		if internals == nil {
			return nil
		}
	}

	for i := range internals {
		if err := content.Parse(internals[i], result); err != nil {
			return err
		}
	}
	return nil
}

func (content *Content) clear() {
	content.main = nil
}
