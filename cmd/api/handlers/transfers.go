package handlers

import (
	"net/http"

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

	tokenID := new(uint64)
	if req.TokenID != nil {
		tokenID = req.TokenID
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

type tokenKey struct {
	Network  string
	Contract string
	TokenID  uint64
}

func (ctx *Context) transfersPostprocessing(transfers transfer.Pageable, withLastID bool) (response TransferResponse, err error) {
	response.Total = transfers.Total
	response.Transfers = make([]Transfer, len(transfers.Transfers))
	if withLastID {
		response.LastID = transfers.LastID
	}

	mapTokens := make(map[tokenKey]*TokenMetadata)
	tokens, err := ctx.TokenMetadata.GetAll()
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			return
		}
	} else {
		for i := range tokens {
			mapTokens[tokenKey{
				Network:  tokens[i].Network,
				Contract: tokens[i].Contract,
				TokenID:  tokens[i].TokenID,
			}] = &TokenMetadata{
				Contract: tokens[i].Contract,
				TokenID:  tokens[i].TokenID,
				Symbol:   tokens[i].Symbol,
				Name:     tokens[i].Name,
				Decimals: tokens[i].Decimals,
				Network:  tokens[i].Network,
			}
		}
	}

	for i := range transfers.Transfers {
		token := mapTokens[tokenKey{
			Network:  transfers.Transfers[i].Network,
			Contract: transfers.Transfers[i].Contract,
			TokenID:  transfers.Transfers[i].TokenID,
		}]

		response.Transfers[i] = TransferFromElasticModel(transfers.Transfers[i])
		response.Transfers[i].Token = token
		response.Transfers[i].Alias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].Contract)
		response.Transfers[i].InitiatorAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].Initiator)
		response.Transfers[i].FromAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].From)
		response.Transfers[i].ToAlias = ctx.getAlias(transfers.Transfers[i].Network, transfers.Transfers[i].To)
	}
	return
}
