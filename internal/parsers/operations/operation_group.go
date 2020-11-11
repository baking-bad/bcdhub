package operations

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
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
func (opg Group) Parse(data gjson.Result) ([]elastic.Model, error) {
	parsedModels := make([]elastic.Model, 0)

	opg.hash = data.Get("hash").String()
	helpers.SetTagSentry("hash", opg.hash)

	for idx, item := range data.Get("contents").Array() {
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
func (content Content) Parse(data gjson.Result) ([]elastic.Model, error) {
	if !content.needParse(data) {
		return nil, nil
	}

	models := make([]elastic.Model, 0)

	kind := data.Get("kind").String()
	switch kind {
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
		return nil, errors.Errorf("Invalid operation kind: %s", kind)
	}

	internalModels, err := content.parseInternal(data)
	if err != nil {
		return nil, err
	}
	models = append(models, internalModels...)

	return models, nil
}

func (content Content) needParse(item gjson.Result) bool {
	kind := item.Get("kind").String()
	source := item.Get("source").String()
	destination := item.Get("destination").String()
	prefixCondition := helpers.IsContract(source) || helpers.IsContract(destination)
	transactionCondition := kind == consts.Transaction && prefixCondition
	originationCondition := (kind == consts.Origination || kind == consts.OriginationNew) && item.Get("script").Exists()
	return originationCondition || transactionCondition
}

func (content Content) parseInternal(data gjson.Result) ([]elastic.Model, error) {
	path := "metadata.internal_operation_results"
	if !data.Get(path).Exists() {
		path = "metadata.internal_operations"
		if !data.Get(path).Exists() {
			return nil, nil
		}
	}

	internalModels := make([]elastic.Model, 0)
	for _, internal := range data.Get(path).Array() {
		parsedModels, err := content.Parse(internal)
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
