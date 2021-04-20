package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/gin-gonic/gin"
)

// GetContractTransfers godoc
// @Summary Show contract`s tokens transfers
// @Description Show contract`s tokens transfers.
// @Tags contract
// @ID get-contract-transfers
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param size query integer false "Transfers count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Param token_id query integer false "Token ID" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {object} TransferResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/{network}/{address}/transfers [get]
func (ctx *Context) GetContractTransfers(c *gin.Context) {
	var contractRequest getContractRequest
	if err := c.BindUri(&contractRequest); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var req getContractTransfers
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	tokenID := int64(-1)
	if req.TokenID != nil {
		tokenID = int64(*req.TokenID)
	}

	transfers, err := ctx.Transfers.Get(transfer.GetContext{
		Network:   contractRequest.Network,
		Contracts: []string{contractRequest.Address},
		Size:      req.Size,
		Offset:    req.Offset,
		TokenID:   tokenID,
	})
	if ctx.handleError(c, err, 0) {
		return
	}
	response, err := ctx.transfersPostprocessing(transfers, false)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) getTokenMetadata(network, contract string, tokenID int64) (tokenmetadata.TokenMetadata, error) {
	key := fmt.Sprintf("%s:%s:%d", network, contract, tokenID)
	item, err := ctx.Cache.Fetch(key, time.Minute*5, func() (interface{}, error) {
		return ctx.TokenMetadata.GetOne(network, contract, tokenID)
	})
	if err != nil {
		return tokenmetadata.TokenMetadata{}, err
	}
	return item.Value().(tokenmetadata.TokenMetadata), nil
}

func (ctx *Context) transfersPostprocessing(transfers transfer.Pageable, withLastID bool) (response TransferResponse, err error) {
	response.Total = transfers.Total
	response.Transfers = make([]Transfer, len(transfers.Transfers))
	if withLastID {
		response.LastID = transfers.LastID
	}

	for i := range transfers.Transfers {
		response.Transfers[i] = TransferFromElasticModel(transfers.Transfers[i])

		token, err := ctx.getTokenMetadata(response.Transfers[i].Network, response.Transfers[i].Contract, response.Transfers[i].TokenID)
		if err == nil {
			tm := TokenMetadataFromElasticModel(token, false)
			response.Transfers[i].Token = &TokenMetadata{
				Contract: tm.Contract,
				TokenID:  tm.TokenID,
				Symbol:   tm.Symbol,
				Name:     tm.Name,
				Decimals: tm.Decimals,
				Network:  tm.Network,
			}
		}

		response.Transfers[i].Alias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].Contract)
		response.Transfers[i].InitiatorAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].Initiator)
		response.Transfers[i].FromAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].From)
		response.Transfers[i].ToAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].To)
	}
	return
}
