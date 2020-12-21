package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetMempool godoc
// @Summary Get contract mempool operations
// @Description Get contract mempool operations
// @Tags contract
// @ID get-contract-mempool
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/mempool [get]
func (ctx *Context) GetMempool(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	api, err := ctx.GetTzKTService(req.Network)
	if err != nil {
		c.JSON(http.StatusNoContent, []Operation{})
		return
	}

	res, err := api.GetMempool(req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	ret, err := ctx.prepareMempoolOperations(res, req.Address, req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, ret)
}

func (ctx *Context) prepareMempoolOperations(res []tzkt.MempoolOperation, address, network string) ([]Operation, error) {
	ret := make([]Operation, len(res))
	if len(res) == 0 {
		return ret, nil
	}

	aliases, err := ctx.ES.GetAliasesMap(network)
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return nil, err
		}
		aliases = make(map[string]string)
	}

	for i := len(res) - 1; i >= 0; i-- {
		status := res[i].Body.Status
		if status == consts.Applied {
			status = "pending"
		}

		op := Operation{
			Protocol:  res[i].Body.Protocol,
			Hash:      res[i].Body.Hash,
			Network:   network,
			Timestamp: time.Unix(res[i].Body.Timestamp, 0).UTC(),

			Kind:         res[i].Body.Kind,
			Source:       res[i].Body.Source,
			Fee:          res[i].Body.Fee,
			Counter:      res[i].Body.Counter,
			GasLimit:     res[i].Body.GasLimit,
			StorageLimit: res[i].Body.StorageLimit,
			Amount:       res[i].Body.Amount,
			Destination:  res[i].Body.Destination,
			Mempool:      true,
			Status:       status,
			RawMempool:   res[i].Body,
		}

		op.SourceAlias = aliases[op.Source]
		op.DestinationAlias = aliases[op.Destination]
		errs, err := cerrors.ParseArray(res[i].Body.Errors)
		if err != nil {
			return nil, err
		}
		op.Errors = errs

		if op.Kind != consts.Transaction {
			ret = append(ret, op)
			continue
		}

		if helpers.IsContract(op.Destination) && op.Protocol != "" {
			params := gjson.ParseBytes(res[i].Body.Parameters)
			if params.Exists() {
				metadata, err := meta.GetMetadata(ctx.Schema, address, consts.PARAMETER, op.Protocol)
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
				op.Entrypoint = consts.DefaultEntrypoint
			}
		}

		ret[i] = op
	}

	return ret, nil
}
