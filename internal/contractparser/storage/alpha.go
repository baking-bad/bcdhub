package storage

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/storage/hash"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Alpha -
type Alpha struct {
	es *elastic.Elastic
}

// NewAlpha -
func NewAlpha(es *elastic.Elastic) Alpha {
	return Alpha{
		es: es,
	}
}

// ParseTransaction -
func (a Alpha) ParseTransaction(content gjson.Result, protocol string, level int64, operationID string) (RichStorage, error) {
	address := content.Get("destination").String()

	m, err := meta.GetMetadata(a.es, address, consts.Mainnet, "storage", protocol)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	bm, err := a.getBigMapDiff(result, operationID, address, level, m)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

// ParseOrigination -
func (a Alpha) ParseOrigination(content gjson.Result, protocol string, level int64, operationID string) (RichStorage, error) {
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	address := result.Get("originated_contracts.0").String()
	storage := content.Get("script.storage")

	m, err := meta.GetMetadata(a.es, address, consts.MetadataAlpha, "storage", protocol)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	var bmd []models.BigMapDiff
	if bmMeta, ok := m["0/0"]; ok && bmMeta.Type == consts.BIGMAP {
		bigMapData := storage.Get("args.0")

		bmd = make([]models.BigMapDiff, 0)
		for _, item := range bigMapData.Array() {
			keyHash, err := hash.Key(item.Get("args.0"))
			if err != nil {
				return RichStorage{Empty: true}, err
			}
			bmd = append(bmd, models.BigMapDiff{
				BinPath:     "0/0",
				Key:         item.Get("args.0").Value(),
				KeyHash:     keyHash,
				Value:       item.Get("args.1").String(),
				OperationID: operationID,
				Level:       level,
				Address:     address,
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
		BigMapDiffs:     bmd,
		DeffatedStorage: res,
	}, nil
}

// Enrich -
func (a Alpha) Enrich(storage string, bmd gjson.Result, skipEmpty bool) (gjson.Result, error) {
	if bmd.IsArray() && len(bmd.Array()) == 0 {
		return gjson.Parse(storage), nil
	}

	p := miguel.GetGJSONPath("0")

	res := make([]interface{}, 0)
	for _, b := range bmd.Array() {
		if skipEmpty && b.Get("value").String() == "" {
			continue
		}
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 1)
		args[0] = b.Get("key").Value()

		sVal := b.Get("value").String()
		if sVal != "" {
			val := gjson.Parse(sVal)
			args = append(args, val.Value())
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

func (a Alpha) getBigMapDiff(result gjson.Result, operationID, address string, level int64, m meta.Metadata) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)
	for _, item := range result.Get("big_map_diff").Array() {
		bmd = append(bmd, models.BigMapDiff{
			BinPath:     "0/0",
			Key:         item.Get("key").Value(),
			KeyHash:     item.Get("key_hash").String(),
			Value:       item.Get("value").String(),
			OperationID: operationID,
			Level:       level,
			Address:     address,
		})
	}
	return bmd, nil
}
