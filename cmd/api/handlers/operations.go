package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/storage"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/r3labs/diff"
	"github.com/tidwall/gjson"
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

	size := int64(10)
	ops, err := ctx.ES.GetContractOperations(req.Network, req.Address, offsetReq.Offset, size)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := prepareOperations(ctx.ES, ops, req.Address, req.Network)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func prepareOperations(es *elastic.Elastic, ops []models.Operation, address, network string) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		op := Operation{
			ID:        ops[i].ID,
			Protocol:  ops[i].Protocol,
			Hash:      ops[i].Hash,
			Network:   ops[i].Network,
			Internal:  ops[i].Internal,
			Timesatmp: ops[i].Timestamp,

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
			if err := setStorageDiff(es, address, network, ops[i].DeffatedStorage, &op); err != nil {
				return nil, err
			}
		}

		if op.Kind != consts.Transaction {
			resp[i] = op
			continue
		}
		if ops[i].Parameters != "" {
			metadata, err := meta.GetMetadata(es, address, network, "parameter", op.Protocol)
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

func setStorageDiff(es *elastic.Elastic, address, network string, storage string, op *Operation) error {
	metadata, err := meta.GetMetadata(es, address, network, "storage", op.Protocol)
	if err != nil {
		return err
	}
	bmd, err := es.GetBigMapDiffsByOperationID(op.ID)
	if err != nil {
		return err
	}
	store, err := enrichStorage(storage, bmd, op.Protocol)
	if err != nil {
		return err
	}
	currentStorage, err := miguel.MichelineToMiguel(store, metadata)
	if err != nil {
		return err
	}

	var prevStorage interface{}
	prev, err := es.GetPreviousOperation(address, op.Network, op.Level)
	if err == nil {
		prevBmd := bmd
		if len(bmd.Array()) > 0 {
			prevBmd, err = getPrevBmd(es, bmd, op.Level)
			if err != nil {
				return err
			}
		}
		prevStore, err := enrichStorage(prev.DeffatedStorage, prevBmd, op.Protocol)
		if err != nil {
			return err
		}
		prevStorage, err = miguel.MichelineToMiguel(prevStore, metadata)
		if err != nil {
			return err
		}
	} else {
		if !strings.Contains(err.Error(), "Operation not found") {
			return err
		}

		if currentStorage == nil {
			return nil
		}
		switch reflect.ValueOf(currentStorage).Kind() {
		case reflect.Map:
			prevStorage = map[string]string{}
		case reflect.Slice:
			prevStorage = make([]interface{}, 0)
		default:
			return fmt.Errorf("Unsupported storage type: %T %v", currentStorage, currentStorage)
		}

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

func enrichStorage(s string, bmd gjson.Result, protocol string) (gjson.Result, error) {
	if len(bmd.Array()) == 0 {
		return gjson.Parse(s), nil
	}

	var parser storage.Parser
	if helpers.StringInArray(protocol, []string{
		consts.HashBabylon, consts.HashCarthage, consts.HashZeroBabylon,
	}) {
		parser = storage.NewBabylon(nil, nil)
	} else {
		parser = storage.NewAlpha(nil)
	}
	return parser.Enrich(s, bmd)
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
		idx, err := strconv.Atoi(path[0])
		if err == nil {
			if val.Kind() != reflect.Map {
				val = val.Index(idx).Elem()
			}
		}
		if !val.IsValid() {
			return nil
		}
		if val.Kind() != reflect.Map {
			return fmt.Errorf("Unsupported change type: %v %v", val, val.Kind())
		}
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
