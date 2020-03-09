package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/cerrors"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/tzkt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetMempool -
func (ctx *Context) GetMempool(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	api := tzkt.NewServicesTzKT(tzkt.TzKTServices, req.Network, time.Second*time.Duration(10))
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
			continue
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

			Result: &models.OperationResult{
				Status: status,
			},
		}

		op.Errors = cerrors.ParseArray(item.Get("errors"))

		if op.Kind != consts.Transaction {
			ret = append(ret, op)
			continue
		}
		params := item.Get("parameters").String()
		if params != "" && strings.HasPrefix(op.Destination, "KT") && op.Protocol != "" {
			metadata, err := meta.GetMetadata(ctx.ES, op.Destination, op.Network, "parameter", op.Protocol)
			if err != nil {
				return nil, err
			}

			paramsJSON := gjson.Parse(params)

			op.Entrypoint, err = metadata.GetByPath(paramsJSON)
			if err != nil && op.Errors == nil {
				return nil, err
			}

			op.Parameters, err = miguel.MichelineToMiguel(paramsJSON, metadata)
			if err != nil {
				if !cerrors.HasParametersError(op.Errors) {
					return nil, err
				}
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
