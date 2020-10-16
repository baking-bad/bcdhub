package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetFA godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard
// @Description Get all contracts that implement FA1/FA1.2 standard
// @Tags tokens
// @ID get-fa-all
// @Param network path string true "Network"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(100)
// @Accept json
// @Produce json
// @Success 200 {object} PageableTokenContracts
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network} [get]
func (ctx *Context) GetFA(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq pageableRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	if cursorReq.Size == 0 {
		cursorReq.Size = 20
	}
	contracts, total, err := ctx.ES.GetTokens(req.Network, "", cursorReq.Offset, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	tokens, err := ctx.contractToTokens(contracts, req.Network, "")
	if handleError(c, err, 0) {
		return
	}
	tokens.Total = total

	c.JSON(http.StatusOK, tokens)
}

// GetFAByVersion godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard by version
// @Description Get all contracts that implement FA1/FA1.2 standard by version
// @Tags tokens
// @ID get-fa-version
// @Param network path string true "Network"
// @Param faversion path string true "FA token version" Enums(fa1, fa12, fa2)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(100)
// @Accept json
// @Produce json
// @Success 200 {object} PageableTokenContracts
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network}/version/{faversion} [get]
func (ctx *Context) GetFAByVersion(c *gin.Context) {
	var req getTokensByVersion
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq pageableRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	if cursorReq.Size == 0 {
		cursorReq.Size = 20
	}
	contracts, total, err := ctx.ES.GetTokens(req.Network, req.Version, cursorReq.Offset, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	tokens, err := ctx.contractToTokens(contracts, req.Network, req.Version)
	if handleError(c, err, 0) {
		return
	}
	tokens.Total = total
	c.JSON(http.StatusOK, tokens)
}

// GetFA12OperationsForAddress godoc
// @Summary Get all token transfers (FA1/FA1.2) where given address is src/dst
// @Description Get all token transfers (FA1/FA1.2) where given address is src/dst
// @Tags tokens
// @ID get-token-transfers
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param last_id query string false "Last transfer ID"
// @Param size query integer false "Requested count" mininum(1) maximum(100)
// @Param sort query string false "Sort: one of `asc` and `desc`"
// @Param start query integer false "Timestamp in seconds" mininum(1)
// @Param end query integer false "Timestamp in seconds" mininum(1)
// @Param contracts query string false "Comma-separated list of contracts which tokens will be requested"
// @Param token_id query integer false "Token ID" mininum(0)
// @Accept json
// @Produce json
// @Success 200 {object} elastic.TransfersResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network}/transfers/{address} [get]
func (ctx *Context) GetFA12OperationsForAddress(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var ctxReq getTransfersRequest
	if err := c.BindQuery(&ctxReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var contracts []string
	if ctxReq.Contracts != "" {
		contracts = strings.Split(ctxReq.Contracts, ",")
	}

	tokenID := int64(-1)
	if ctxReq.TokenID != nil {
		tokenID = *ctxReq.TokenID
	}

	transfers, err := ctx.ES.GetTransfers(elastic.GetTransfersContext{
		Network:   req.Network,
		Address:   req.Address,
		Contracts: contracts,
		Start:     ctxReq.Start * 1000,
		End:       ctxReq.End * 1000,
		LastID:    ctxReq.LastID,
		SortOrder: ctxReq.Sort,
		Size:      ctxReq.Size,
		TokenID:   tokenID,
	})
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, transfers)
}

// GetTokenVolumeSeries godoc
// @Summary Get volume series for token
// @Description Get volume series for token
// @Tags tokens
// @ID get-token-series
// @Param network path string true "Network"
// @Param period query string true "One of periods"  Enums(year, month, week, day)
// @Param addresses path string true "Comma-separated contract addresses"
// @Param contract path string true "KT address" minlength(36) maxlength(36)
// @Param token_id query int true "Token ID" minimum(0)
// @Accept json
// @Produce  json
// @Success 200 {object} Series
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network}/series [get]
func (ctx *Context) GetTokenVolumeSeries(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var reqArgs getTokenSeriesRequest
	if err := c.BindQuery(&reqArgs); handleError(c, err, http.StatusBadRequest) {
		return
	}

	series, err := ctx.ES.GetTokenVolumeSeries(req.Network, reqArgs.Period, []string{reqArgs.Contract}, reqArgs.GetAddresses(), reqArgs.TokenID)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, series)
}

func (ctx *Context) contractToTokens(contracts []models.Contract, network, version string) (PageableTokenContracts, error) {
	tokens := make([]TokenContract, len(contracts))
	addresses := make([]string, len(contracts))
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
			Balance:       contracts[i].Balance,
			TxCount:       contracts[i].TxCount,
			LastAction:    contracts[i].LastAction.Time,
		}
		for _, tag := range contracts[i].Tags {
			if tag == consts.FA2Tag {
				tokens[i].Type = consts.FA2Tag
				break
			}

			if tag == consts.FA12Tag {
				tokens[i].Type = consts.FA12Tag
				break
			} else if tag == consts.FA1Tag {
				tokens[i].Type = consts.FA1Tag
			}
		}
		addresses[i] = tokens[i].Address
	}

	if version != "" {
		interfaceVersion, ok := ctx.Interfaces[version]
		if !ok {
			return PageableTokenContracts{}, errors.Errorf("Unknown interface version: %s", version)
		}
		methods := make([]string, len(interfaceVersion.Entrypoints))
		for i := range interfaceVersion.Entrypoints {
			methods[i] = interfaceVersion.Entrypoints[i].Name
		}

		stats, err := ctx.ES.GetTokensStats(network, addresses, methods)
		if err != nil {
			return PageableTokenContracts{}, err
		}

		for i := range tokens {
			tokens[i].Methods = make(map[string]TokenMethodStats)
			stat, ok := stats[tokens[i].Address]
			if !ok {
				for _, method := range methods {
					tokens[i].Methods[method] = TokenMethodStats{}
				}
				continue
			}

			for _, method := range methods {
				s, ok := stat[method]
				if !ok {
					tokens[i].Methods[method] = TokenMethodStats{}
					continue
				}
				tokens[i].Methods[method] = TokenMethodStats{
					CallCount:          s.Count,
					AverageConsumedGas: s.ConsumedGas,
				}
			}
		}
	}

	return PageableTokenContracts{
		Tokens: tokens,
	}, nil
}

// GetContractTokens godoc
// @Summary List contract tokens
// @Description List contract tokens
// @Tags contract
// @ID get-contract-token
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {array} Token
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/tokens [get]
func (ctx *Context) GetContractTokens(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	tokens, err := ctx.getTokens(req.Network, req.Address)
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			handleError(c, err, 0)
		} else {
			c.JSON(http.StatusOK, []interface{}{})
		}
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (ctx *Context) getTokens(network, address string) ([]Token, error) {
	metadata, err := ctx.ES.GetTokenMetadata(elastic.GetTokenMetadataContext{
		Contract: address,
		Network:  network,
		TokenID:  -1,
	})
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	tokens := make([]Token, 0)
	for _, token := range metadata {
		supply, err := ctx.ES.GetTokenSupply(network, address, token.TokenID)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, Token{
			token, supply,
		})
	}
	return tokens, nil
}
