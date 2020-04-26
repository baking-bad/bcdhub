package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetMempool -
func (ctx *Context) GetMempool(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	api, ok := ctx.TzKTSvcs[req.Network]
	if !ok {
		c.AbortWithError(500, fmt.Errorf("TzKT services does not support %s", req.Network))
		return
	}
	res, err := api.GetMempool(req.Address)
	if handleError(c, err, 0) {
		return
	}

	ret, err := ctx.prepareMempoolOperations(res, req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, ret)
}

func (ctx *Context) prepareMempoolOperations(res gjson.Result, address, network string) ([]Operation, error) {
	ret := make([]Operation, 0)
	if res.Get("#").Int() == 0 {
		return ret, nil
	}
	for _, item := range res.Array() {
		status := item.Get("status").String()
		if status == "applied" {
			status = "pending"
		}

		op := Operation{
			Protocol:  item.Get("protocol").String(),
			Hash:      item.Get("hash").String(),
			Network:   network,
			Timesatmp: time.Unix(item.Get("timestamp").Int(), 0).UTC(),

			Kind:         item.Get("kind").String(),
			Source:       item.Get("source").String(),
			Fee:          item.Get("fee").Int(),
			Counter:      item.Get("counter").Int(),
			GasLimit:     item.Get("gas_limit").Int(),
			StorageLimit: item.Get("storage_limit").Int(),
			Amount:       item.Get("amount").Int(),
			Destination:  item.Get("destination").String(),
			Mempool:      true,
			Status:       status,
		}

		op.Errors = cerrors.ParseArray(item.Get("errors"))

		if op.Kind != consts.Transaction {
			ret = append(ret, op)
			continue
		}
		params := item.Get("parameters")
		if strings.HasPrefix(op.Destination, "KT") && op.Protocol != "" {
			if params.Exists() {
				metadata, err := meta.GetMetadata(ctx.ES, address, consts.PARAMETER, op.Protocol)
				if err != nil {
					return nil, err
				}

				op.Entrypoint, err = metadata.GetByPath(params)
				if err != nil && op.Errors == nil {
					return nil, err
				}

				op.Parameters, err = newmiguel.ParameterToMiguel(params, metadata)
				if err != nil {
					if !cerrors.HasParametersError(op.Errors) {
						return nil, err
					}
				}
			} else {
				op.Entrypoint = "default"
			}
		}

		ret = append(ret, op)
	}

	// reverse array
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret, nil
}
