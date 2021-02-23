package storage

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Alpha -
type Alpha struct{}

// NewAlpha -
func NewAlpha() *Alpha {
	return &Alpha{}
}

// ParseTransaction -
func (a *Alpha) ParseTransaction(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	address := content.Get("destination").String()

	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	return RichStorage{
		Models:          a.getBigMapDiff(result, address, operation),
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

// ParseOrigination -
func (a *Alpha) ParseOrigination(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	pair, ok := storage.Nodes[0].(*ast.Pair)
	if !ok {
		return RichStorage{Empty: true}, nil
	}
	bigMap, ok := pair.Args[0].(*ast.BigMap)
	if !ok {
		return RichStorage{Empty: true}, nil
	}

	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	address := result.Get("originated_contracts.0").String()

	var data ast.UntypedAST
	if err := json.UnmarshalFromString(content.Get("script.storage").String(), &data); err != nil {
		return RichStorage{Empty: true}, err
	}
	if err := storage.Settle(data); err != nil {
		return RichStorage{Empty: true}, err
	}

	bmd := make([]models.Model, 0)
	if err := bigMap.Data.Range(func(key, value ast.Comparable) (bool, error) {
		k := key.(ast.Node)
		keyHash, err := ast.BigMapKeyHashFromNode(k)
		if err != nil {
			return false, err
		}
		keyBytes, err := k.ToParameters()
		if err != nil {
			return false, err
		}

		var valBytes []byte
		if value != nil {
			v := value.(ast.Node)
			valBytes, err = v.ToParameters()
			if err != nil {
				return false, err
			}
		}
		bmd = append(bmd, &bigmapdiff.BigMapDiff{
			ID:          helpers.GenerateID(),
			Key:         keyBytes,
			KeyHash:     keyHash,
			Value:       valBytes,
			OperationID: operation.ID,
			Level:       operation.Level,
			Address:     address,
			IndexedTime: time.Now().UnixNano() / 1000,
			Network:     operation.Network,
			Timestamp:   operation.Timestamp,
			Protocol:    operation.Protocol,
			Ptr:         -1,
		})
		return false, nil
	}); err != nil {
		return RichStorage{Empty: true}, err
	}

	if len(bmd) > 0 {
		bigMap.Data = ast.NewOrderedMap()
	}

	b, err := storage.ToParameters(ast.DocsFull)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	return RichStorage{
		Models:          bmd,
		DeffatedStorage: string(b),
	}, nil
}

func (a *Alpha) getBigMapDiff(result gjson.Result, address string, operation operation.Operation) []models.Model {
	bmd := make([]models.Model, 0)
	for _, item := range result.Get("big_map_diff").Array() {
		bmd = append(bmd, &bigmapdiff.BigMapDiff{
			ID:          helpers.GenerateID(),
			Key:         []byte(item.Get("key").String()),
			KeyHash:     item.Get("key_hash").String(),
			Value:       []byte(item.Get("value").String()),
			OperationID: operation.ID,
			Level:       operation.Level,
			Address:     address,
			IndexedTime: time.Now().UnixNano() / 1000,
			Network:     operation.Network,
			Timestamp:   operation.Timestamp,
			Protocol:    operation.Protocol,
			Ptr:         -1,
		})
	}
	return bmd
}
