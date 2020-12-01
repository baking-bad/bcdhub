package storage

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage/hash"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Alpha -
type Alpha struct{}

// NewAlpha -
func NewAlpha() *Alpha {
	return &Alpha{}
}

// ParseTransaction -
func (a *Alpha) ParseTransaction(content gjson.Result, _ meta.Metadata, operation models.Operation) (RichStorage, error) {
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
func (a *Alpha) ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	address := result.Get("originated_contracts.0").String()
	storage := content.Get("script.storage")

	var bmd []elastic.Model
	if bmMeta, ok := metadata["0/0"]; ok && bmMeta.Type == consts.BIGMAP {
		bigMapData := storage.Get("args.0")

		bmd = make([]elastic.Model, 0)
		for _, item := range bigMapData.Array() {
			keyHash, err := hash.Key(item.Get("args.0"))
			if err != nil {
				return RichStorage{Empty: true}, err
			}
			bmd = append(bmd, &models.BigMapDiff{
				ID:          helpers.GenerateID(),
				BinPath:     "0/0",
				Key:         item.Get("args.0").Value(),
				KeyHash:     keyHash,
				Value:       item.Get("args.1").String(),
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
	}

	res := storage.String()
	if len(bmd) > 0 {
		res, err = sjson.Set(res, "args.0", []interface{}{})
		if err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          bmd,
		DeffatedStorage: res,
	}, nil
}

// Enrich -
func (a *Alpha) Enrich(storage, sPrevStorage string, bmd []models.BigMapDiff, skipEmpty, unpack bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(storage), nil
	}

	p := newmiguel.GetGJSONPath("0")

	res := make([]interface{}, 0)
	for _, b := range bmd {
		if skipEmpty && b.Value == "" {
			continue
		}
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 2)
		args[0] = b.Key

		if b.Value != "" {
			var val gjson.Result
			var err error
			if unpack {
				val, err = stringer.Micheline(gjson.Parse(b.Value))
				if err != nil {
					return val, err
				}
			} else {
				val = gjson.Parse(b.Value)
			}
			args = append(args, val.Value())
		} else {
			args[1] = nil
		}

		elt["args"] = args
		res = append(res, elt)
	}
	value, err := sjson.Set(storage, p, res)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(value), nil
}

func (a *Alpha) getBigMapDiff(result gjson.Result, address string, operation models.Operation) []elastic.Model {
	bmd := make([]elastic.Model, 0)
	for _, item := range result.Get("big_map_diff").Array() {
		bmd = append(bmd, &models.BigMapDiff{
			ID:          helpers.GenerateID(),
			BinPath:     "0/0",
			Key:         item.Get("key").Value(),
			KeyHash:     item.Get("key_hash").String(),
			Value:       item.Get("value").String(),
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
