package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers/enrichment"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/r3labs/diff"
	"github.com/tidwall/gjson"
)

type offsetRequest struct {
	Offset int64 `form:"offset"`
	Limit  int64 `form:"limit"`
}

var enrichments = []enrichment.Enrichment{
	enrichment.Babylon{},
	enrichment.Alpha{},
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

	size := int64(10)
	ops, err := ctx.ES.GetContractOperations(req.Network, req.Address, offsetReq.Offset, size)
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
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		op := Operation{
			ID:       ops[i].ID,
			Protocol: ops[i].Protocol,
			Hash:     ops[i].Hash,
			Network:  ops[i].Network,
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

		if ops[i].DeffatedStorage != "" {
			if err := setStorageDiff(es, address, ops[i].DeffatedStorage, &op); err != nil {
				return nil, err
			}
		}

		if op.Kind != "transaction" {
			resp[i] = op
			continue
		}
		if ops[i].Parameters != "" {
			metadata, err := getMetadata(es, address, "parameter", op.Level)
			if err != nil {
				return nil, err
			}

			params := gjson.Parse(ops[i].Parameters)

			op.Parameters, err = miguel.MichelineToMiguel(params, metadata)
			if err != nil {
				return nil, err
			}
		}
		resp[i] = op
	}
	return resp, nil
}

func setStorageDiff(es *elastic.Elastic, address string, storage string, op *Operation) error {
	metadata, err := getMetadata(es, address, "storage", op.Level)
	if err != nil {
		return err
	}
	bmd, err := es.GetBigMapDiffsByOperationID(op.ID)
	if err != nil {
		return err
	}
	store, err := enrichStorage(storage, bmd, op.Level)
	if err != nil {
		return err
	}
	currentStorage, err := miguel.MichelineToMiguel(store, metadata)
	if err != nil {
		return err
	}

	prev, err := es.GetPreviousOperation(address, op.Network, op.Level)
	if err != nil {
		if !strings.Contains(err.Error(), "Operation not found") {
			return err
		}

		store := gjson.Parse(storage)
		op.StorageDiff, err = miguel.MichelineToMiguel(store, metadata)
		if err != nil {
			return err
		}
		return nil
	}

	prevBmd, err := getPrevBmd(es, bmd, op.Level)
	if err != nil {
		return err
	}
	prevStore, err := enrichStorage(prev.DeffatedStorage, prevBmd, op.Level)
	if err != nil {
		return err
	}
	prevStorage, err := miguel.MichelineToMiguel(prevStore, metadata)
	if err != nil {
		return err
	}

	changelog, err := diff.Diff(prevStorage, currentStorage)
	if err != nil {
		return err
	}
	if err := applyChanges(changelog, currentStorage); err != nil {
		return err
	}

	op.StorageDiff = currentStorage
	return nil
}

func enrichStorage(storage string, bmd gjson.Result, level int64) (gjson.Result, error) {
	for _, e := range enrichments {
		if e.Level() < level {
			return e.Do(storage, bmd)
		}
	}
	return gjson.Result{}, nil
}

func getPrevBmd(es *elastic.Elastic, bmd gjson.Result, level int64) (gjson.Result, error) {
	keys := make([]string, 0)
	for _, b := range bmd.Array() {
		keys = append(keys, b.Get("key_hash").String())
	}

	return es.GetBigMapDiffsByKeyHash(keys, level)
}

func applyChanges(changelog diff.Changelog, v interface{}) error {
	for _, c := range changelog {
		if err := applyChange(c.Path, c.From, c.Type, v); err != nil {
			return err
		}
	}
	return nil
}

func applyChange(path []string, from interface{}, typ string, v interface{}) error {
	val := reflect.ValueOf(v)
	if len(path) == 1 {
		val.SetMapIndex(reflect.ValueOf("from"), reflect.ValueOf(from))
		val.SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf(typ))
		return nil
	}
	var field reflect.Value
	if val.Kind() == reflect.Map {
		field = val.MapIndex(reflect.ValueOf(path[0]))
	} else if val.Kind() == reflect.Slice {
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			return err
		}
		field = val.Index(idx)
	}
	return applyChange(path[1:], from, typ, field.Interface())

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
