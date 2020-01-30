package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type offsetRequest struct {
	Offset int64 `form:"offset"`
	Limit  int64 `form:"limit"`
}

// GetContractOperations -
func (ctx *Context) GetContractOperations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var offsetReq offsetRequest
	if err := c.BindQuery(&offsetReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ops, err := ctx.ES.GetContractOperations(req.Address, offsetReq.Offset, 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := prepareOperations(ctx.ES, ops, req.Address)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func prepareOperations(es *elastic.Elastic, ops []models.Operation, address string) ([]Operation, error) {
	resp := make([]Operation, 0)
	for i := range ops {
		op := Operation{
			ID:       ops[i].ID,
			Protocol: ops[i].Protocol,
			Hash:     ops[i].Hash,
			Internal: ops[i].Internal,

			Level:         ops[i].Level,
			Kind:          ops[i].Kind,
			Source:        ops[i].Source,
			Fee:           ops[i].Fee,
			Counter:       ops[i].Counter,
			GasLimit:      ops[i].GasLimit,
			StorageLimit:  ops[i].StorageLimit,
			Amount:        ops[i].Amount,
			Destination:   ops[i].Destination,
			PublicKey:     ops[i].PublicKey,
			ManagerPubKey: ops[i].ManagerPubKey,
			Balance:       ops[i].Balance,
			Delegate:      ops[i].Delegate,

			Result: ops[i].Result,
		}

		if op.Kind != "transaction" {
			resp = append(resp, op)
			continue
		}
		if ops[i].Parameters != "" {
			metadata, err := getMetadata(es, address, "parameter", op.Level)
			if err != nil {
				panic(err)
			}

			params := gjson.Parse(ops[i].Parameters)

			op.Parameters, err = miguel.MichelineToMiguel(params, metadata)
			if err != nil {
				return nil, err
			}
		}

		if ops[i].DeffatedStorage != "" {
			metadata, err := getMetadata(es, address, "storage", op.Level)
			if err != nil {
				panic(err)
			}

			if err := insertBigMapDiffs(es, ops[i].DeffatedStorage, metadata, &op); err != nil {
				return nil, err
			}
		}

		resp = append(resp, op)
	}
	return resp, nil
}

func insertBigMapDiffs(es *elastic.Elastic, storage string, metadata meta.Metadata, op *Operation) error {
	bmd, err := es.GetBigMapDiffsByOperationID(op.ID)
	if err != nil {
		return err
	}

	data := bmd.Get("hits.hits.#._source")

	var richStorage gjson.Result
	if op.Level < consts.LevelBabylon {
		richStorage, err = getRichStorageAlpha(storage, data)
		if err != nil {
			return err
		}
	} else {
		richStorage, err = getRichStorageBabylon(storage, data)
		if err != nil {
			return err
		}
	}

	op.Storage, err = miguel.MichelineToMiguel(richStorage, metadata)
	if err != nil {
		return err
	}

	return nil
}
func getRichStorageAlpha(storage string, bmd gjson.Result) (gjson.Result, error) {
	p := miguel.GetGJSONPath("0")

	res := make([]interface{}, 0)
	for _, b := range bmd.Array() {
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 2)
		val := gjson.Parse(b.Get("value").String())
		args[0] = b.Get("key").Value()

		if b.Get("value").String() != "" {
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
func getRichStorageBabylon(storage string, bmd gjson.Result) (gjson.Result, error) {
	data := gjson.Parse(storage)

	for _, b := range bmd.Array() {
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 1)
		val := gjson.Parse(b.Get("value").String())
		args[0] = b.Get("key").Value()

		if b.Get("value").String() != "" {
			args = append(args, val.Value())
		}

		elt["args"] = args

		p := miguel.GetGJSONPath(b.Get("bin_path").String()[2:])
		value, err := sjson.Set(storage, p, []interface{}{elt})
		if err != nil {
			return gjson.Result{}, err
		}
		data = gjson.Parse(value)
	}

	return data, nil
}

func getMetadata(es *elastic.Elastic, address, tag string, level int64) (meta.Metadata, error) {
	if address == "" {
		return nil, fmt.Errorf("[getMetadata] Empty address")
	}

	data, err := es.GetByID(elastic.DocMetadata, address)
	if err != nil {
		return nil, err
	}

	network := meta.GetMetadataNetwork(level)
	path := fmt.Sprintf("_source.%s.%s", tag, network)
	metadata := data.Get(path).String()

	var res meta.Metadata
	if err := json.Unmarshal([]byte(metadata), &res); err != nil {
		return nil, err
	}

	return res, nil
}
