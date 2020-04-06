package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetFA12 -
func (ctx *Context) GetFA12(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	if pageReq.Size == 0 {
		pageReq.Size = 20
	}

	contracts, err := ctx.ES.GetFA12(req.Network, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, contractToFA12(contracts))
}

// GetFA12OperationsForAddress -
func (ctx *Context) GetFA12OperationsForAddress(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq cursorRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	operations, err := ctx.ES.GetFA12TransferOperations(req.Network, req.Address, cursorReq.LastID)
	if handleError(c, err, 0) {
		return
	}

	ops, err := operationToTransfer(ctx.ES, operations)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, ops)
}

func contractToFA12(contracts []models.Contract) []ContractFA12 {
	fa12 := make([]ContractFA12, len(contracts))
	for i := range contracts {
		fa12[i] = ContractFA12{
			Network:       contracts[i].Network,
			Level:         contracts[i].Level,
			Timestamp:     contracts[i].Timestamp,
			Address:       contracts[i].Address,
			Manager:       contracts[i].Manager,
			Delegate:      contracts[i].Delegate,
			Alias:         contracts[i].Alias,
			DelegateAlias: contracts[i].DelegateAlias,
		}
	}
	return fa12
}

func operationToTransfer(es *elastic.Elastic, po elastic.PageableOperations) (PageableTransfersFA12, error) {
	transfers := make([]TransferFA12, 0)
	contracts := map[string]bool{}
	metadatas := map[string]meta.Metadata{}

	for _, op := range po.Operations {
		key := op.Network + op.Destination
		isFA12, ok := contracts[key]
		if !ok {
			val, err := es.IsFA12Contract(op.Network, op.Destination)
			if err != nil {
				return PageableTransfersFA12{}, err
			}
			contracts[key] = val
			isFA12 = val
		}
		if !isFA12 {
			continue
		}
		if cerrors.HasParametersError(op.Errors) || cerrors.HasGasExhaustedError(op.Errors) {
			continue
		}

		transfer := TransferFA12{
			Network:   op.Network,
			Protocol:  op.Protocol,
			Hash:      op.Hash,
			Counter:   op.Counter,
			Status:    op.Status,
			Timestamp: op.Timestamp,
			Level:     op.Level,
		}

		metadata, ok := metadatas[key]
		if !ok {
			val, err := meta.GetMetadata(es, op.Destination, consts.PARAMETER, op.Protocol)
			if err != nil {
				return PageableTransfersFA12{}, fmt.Errorf("[operationToTransfer] Unknown metadata: %s", op.Destination)
			}
			metadatas[key] = val
			metadata = val
		}

		params := gjson.Parse(op.Parameters)
		parameters, err := newmiguel.ParameterToMiguel(params, metadata)
		if err != nil {
			return PageableTransfersFA12{}, err
		}
		if len(parameters.Children) != 3 {
			continue
		}
		transfer.From = parameters.Children[0].Value.(string)
		transfer.To = parameters.Children[1].Value.(string)
		amount, err := strconv.ParseInt(parameters.Children[2].Value.(string), 10, 64)
		if err != nil {
			return PageableTransfersFA12{}, err
		}
		transfer.Amount = amount

		transfers = append(transfers, transfer)
	}
	pt := PageableTransfersFA12{
		LastID:    po.LastID,
		Transfers: transfers,
	}
	return pt, nil
}
