package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/pack"
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

	contracts, err := ctx.ES.GetTokens(req.Network, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, contractToTokens(contracts))
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

	packedAddress, err := pack.Address(req.Address)
	if handleError(c, err, 0) {
		return
	}
	operations, err := ctx.ES.GetTokenTransferOperations(req.Network, req.Address, packedAddress, cursorReq.LastID, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	ops, err := operationToTransfer(ctx.ES, operations)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, ops)
}

func contractToTokens(contracts []models.Contract) []TokenContract {
	tokens := make([]TokenContract, len(contracts))
	for i := range contracts {
		tokens[i] = TokenContract{
			Network:       contracts[i].Network,
			Level:         contracts[i].Level,
			Timestamp:     contracts[i].Timestamp,
			Address:       contracts[i].Address,
			Manager:       contracts[i].Manager,
			Delegate:      contracts[i].Delegate,
			Alias:         contracts[i].Alias,
			DelegateAlias: contracts[i].DelegateAlias,
		}
		for _, tag := range contracts[i].Tags {
			if tag == consts.FA12Tag {
				tokens[i].Type = consts.FA12Tag
				break
			} else if tag == consts.FA1Tag {
				tokens[i].Type = consts.FA1Tag
			}
		}
	}
	return tokens
}

func operationToTransfer(es *elastic.Elastic, po elastic.PageableOperations) (PageableTokenTransfers, error) {
	transfers := make([]TokenTransfer, 0)
	contracts := map[string]bool{}
	metadatas := map[string]meta.Metadata{}

	for _, op := range po.Operations {
		key := op.Network + op.Destination
		isFA12, ok := contracts[key]
		if !ok {
			val, err := es.IsFAContract(op.Network, op.Destination)
			if err != nil {
				return PageableTokenTransfers{}, err
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

		transfer := TokenTransfer{
			Network:   op.Network,
			Contract:  op.Destination,
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
				return PageableTokenTransfers{}, fmt.Errorf("[operationToTransfer] Unknown metadata: %s", op.Destination)
			}
			metadatas[key] = val
			metadata = val
		}

		params := gjson.Parse(op.Parameters)
		parameters, err := newmiguel.ParameterToMiguel(params, metadata)
		if err != nil {
			return PageableTokenTransfers{}, err
		}

		if op.Entrypoint == "transfer" && len(parameters.Children) == 3 {
			transfer.From = parameters.Children[0].Value.(string)
			transfer.To = parameters.Children[1].Value.(string)
			amount, err := strconv.ParseInt(parameters.Children[2].Value.(string), 10, 64)
			if err != nil {
				return PageableTokenTransfers{}, err
			}
			transfer.Amount = amount
		} else if op.Entrypoint == "mint" && len(parameters.Children) == 2 {
			transfer.To = parameters.Children[0].Value.(string)
			amount, err := strconv.ParseInt(parameters.Children[1].Value.(string), 10, 64)
			if err != nil {
				return PageableTokenTransfers{}, err
			}
			transfer.Amount = amount
		} else {
			continue
		}

		transfers = append(transfers, transfer)
	}
	pt := PageableTokenTransfers{
		LastID:    po.LastID,
		Transfers: transfers,
	}
	return pt, nil
}
