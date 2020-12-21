package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/gin-gonic/gin"
)

// GetContractTransfers godoc
// @Summary Show contract`s tokens transfers
// @Description Show contract`s tokens transfers.
// @Tags contract
// @ID get-contract-transfers
// @Param size query integer false "Transfers count" mininum(1)
// @Param offset query integer false "Offset" mininum(1)
// @Param token_id query integer false "Token ID" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {object} TransferResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /{network}/{address}/transfers [get]
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
	response, err := ctx.transfersPostprocessing(transfers)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, response)
}

type tokenKey struct {
	Network  string
	Contract string
	TokenID  int64
}

func (ctx *Context) transfersPostprocessing(transfers transfer.Pageable) (response TransferResponse, err error) {
	response.Total = transfers.Total
	response.Transfers = make([]Transfer, len(transfers.Transfers))

	mapTokens := make(map[tokenKey]*TokenMetadata)
	tokens, err := ctx.TZIP.GetTokenMetadata(tzip.GetTokenMetadataContext{
		TokenID: -1,
	})
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			return
		}
	} else {
		for i := range tokens {
			mapTokens[tokenKey{
				Network:  tokens[i].Network,
				Contract: tokens[i].Address,
				TokenID:  tokens[i].TokenID,
			}] = &TokenMetadata{
				Contract: tokens[i].Address,
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
		response.Transfers[i] = Transfer{&transfers.Transfers[i], token}
	}
	return
}
