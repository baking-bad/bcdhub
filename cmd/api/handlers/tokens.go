package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast/interfaces"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// GetFA godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard
// @Description Get all contracts that implement FA1/FA1.2 standard
// @Tags tokens
// @ID get-fa-all
// @Param network path string true "Network"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(10)
// @Accept json
// @Produce json
// @Success 200 {object} PageableTokenContracts
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/tokens/{network} [get]
func GetFA() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getByNetwork
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var cursorReq pageableRequest
		if err := c.BindQuery(&cursorReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		contracts, total, err := ctx.Contracts.GetTokens("", cursorReq.Offset, cursorReq.Size)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		tokens, err := contractToTokens(ctx, contracts, "")
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		tokens.Total = total

		c.SecureJSON(http.StatusOK, tokens)
	}
}

// GetFAByVersion godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard by version
// @Description Get all contracts that implement FA1/FA1.2 standard by version
// @Tags tokens
// @ID get-fa-version
// @Param network path string true "Network"
// @Param faversion path string true "FA token version" Enums(fa1, fa12, fa2)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" minimum(0) maximum(10)
// @Accept json
// @Produce json
// @Success 200 {object} PageableTokenContracts
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/tokens/{network}/version/{faversion} [get]
func GetFAByVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getTokensByVersion
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var cursorReq pageableRequest
		if err := c.BindQuery(&cursorReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		if req.Version == "fa12" {
			req.Version = consts.FA12Tag
		}
		contracts, total, err := ctx.Contracts.GetTokens(req.Version, cursorReq.Offset, cursorReq.Size)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		tokens, err := contractToTokens(ctx, contracts, req.Version)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		tokens.Total = total
		c.SecureJSON(http.StatusOK, tokens)
	}
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
// @Success 200 {object} TransferResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/tokens/{network}/transfers/{address} [get]
func GetFA12OperationsForAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var ctxReq getTransfersRequest
		if err := c.BindQuery(&ctxReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var contracts []string
		if ctxReq.Contracts != "" {
			contracts = strings.Split(ctxReq.Contracts, ",")
		}

		acc, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		transfers, err := ctx.Domains.Transfers(transfer.GetContext{
			AccountID: acc.ID,
			Contracts: contracts,
			Start:     ctxReq.Start,
			End:       ctxReq.End,
			LastID:    ctxReq.LastID,
			SortOrder: ctxReq.Sort,
			Size:      ctxReq.Size,
			TokenID:   ctxReq.TokenID,
		})
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, transfersPostprocessing(ctx, transfers, true))
	}
}

// GetTokenVolumeSeries godoc
// @Summary Get volume series for token
// @Description Get volume series for token
// @Tags tokens
// @ID get-token-series
// @Param network path string true "Network"
// @Param period query string true "One of periods"  Enums(year, month, week, day)
// @Param contract path string true "KT address" minlength(36) maxlength(36)
// @Param token_id query int true "Token ID" minimum(0)
// @Param slug query string true "DApp slug"
// @Accept json
// @Produce  json
// @Success 200 {object} SeriesFloat
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/tokens/{network}/series [get]
func GetTokenVolumeSeries() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getByNetwork
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var args getTokenSeriesRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		dapp, err := ctx.DApps.Get(args.Slug)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		series, err := ctx.Transfers.GetTokenVolumeSeries(args.Period, []string{args.Contract}, dapp.Contracts, args.TokenID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, series)
	}
}

func contractToTokens(ctx *config.Context, contracts []contract.Contract, version string) (PageableTokenContracts, error) {
	tokens := make([]TokenContract, len(contracts))
	addresses := make([]string, len(contracts))
	for i := range contracts {
		tokens[i] = TokenContract{
			Level:         contracts[i].Level,
			Timestamp:     contracts[i].Timestamp,
			Address:       contracts[i].Account.Address,
			Manager:       contracts[i].Manager.Address,
			Delegate:      contracts[i].Delegate.Address,
			Alias:         contracts[i].Account.Alias,
			DelegateAlias: contracts[i].Delegate.Alias,
			LastAction:    contracts[i].LastAction,
			TxCount:       contracts[i].TxCount,
		}

		switch {
		case contracts[i].Tags.Has(types.FA2Tag):
			tokens[i].Type = consts.FA2Tag
		case contracts[i].Tags.Has(types.FA12Tag):
			tokens[i].Type = consts.FA12Tag
		case contracts[i].Tags.Has(types.FA1Tag):
			tokens[i].Type = consts.FA1Tag
		}

		addresses[i] = tokens[i].Address
	}

	if version != "" {
		methods, err := interfaces.GetMethods(version)
		if err != nil {
			return PageableTokenContracts{}, err
		}

		stats, err := ctx.Operations.GetTokensStats(addresses, methods)
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
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1) maximum(10)
// @Param max_level query integer false "Maximum token`s creation level (less than or equal)" mininum(1)
// @Param min_level query integer false "Minimum token`s creation level (greater than)" mininum(1)
// @Param token_id query integer false "Token ID" mininum(0)
// @Accept  json
// @Produce  json
// @Success 200 {array} TokenMetadata
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/tokens [get]
func GetContractTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var pageReq tokenRequest
		if err := c.BindQuery(&pageReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		metadata, err := ctx.TokenMetadata.Get([]tokenmetadata.GetContext{{
			Contract: req.Address,
			TokenID:  pageReq.TokenID,
			MinLevel: pageReq.MinLevel,
			MaxLevel: pageReq.MaxLevel,
		}}, pageReq.Size, pageReq.Offset)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusOK, []TokenMetadata{})
			} else {
				handleError(c, ctx.Storage, err, 0)
			}
			return
		}

		result := make([]TokenMetadata, len(metadata))
		for i := range metadata {
			result[i] = TokenMetadataFromModel(metadata[i], false)
			result[i].Network = req.Network
		}
		c.SecureJSON(http.StatusOK, result)
	}
}

// GetContractTokensCount godoc
// @Summary Get contract`s tokens count
// @Description Get contract`s tokens count
// @Tags contract
// @ID get-contract-token-count
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} CountResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/tokens/count [get]
func GetContractTokensCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		count, err := ctx.TokenMetadata.Count([]tokenmetadata.GetContext{{
			Contract: req.Address,
		}})
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, CountResponse{
			count,
		})
	}
}

func getTokensWithSupply(ctx *config.Context, getCtx tokenmetadata.GetContext, size, offset int64) ([]Token, error) {
	metadata, err := ctx.TokenMetadata.Get([]tokenmetadata.GetContext{getCtx}, size, offset)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			return []Token{}, nil
		}
		return nil, err
	}

	return addSupply(ctx, metadata)
}

func addSupply(ctx *config.Context, metadata []tokenmetadata.TokenMetadata) ([]Token, error) {
	tokens := make([]Token, 0)
	for _, token := range metadata {
		tokenMetadata := TokenMetadataFromModel(token, true)
		tokenMetadata.Network = ctx.Network.String()
		t := Token{
			TokenMetadata: tokenMetadata,
		}

		supply, err := ctx.TokenBalances.TokenSupply(token.Contract, token.TokenID)
		if err != nil {
			return nil, err
		}
		t.Supply = supply

		transferred, err := ctx.Transfers.GetTransfered(token.Contract, token.TokenID)
		if err != nil {
			return nil, err
		}
		t.Transferred = transferred

		tokens = append(tokens, t)
	}

	return tokens, nil
}

// GetTokenHolders godoc
// @Summary List token holders
// @Description List token holders
// @Tags contract
// @ID get-token-holders
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param token_id query int true "Token ID" minimum(0)
// @Accept  json
// @Produce  json
// @Success 200 {array} gin.H
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/tokens/holders [get]
func GetTokenHolders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var reqArgs byTokenIDRequest
		if err := c.BindQuery(&reqArgs); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		balances, err := ctx.TokenBalances.GetHolders(req.Address, *reqArgs.TokenID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		result := make(map[string]string)
		for i := range balances {
			result[balances[i].Account.Address] = balances[i].Balance.String()
		}

		c.SecureJSON(http.StatusOK, result)
	}
}

// GetTokenMetadata godoc
// @Summary List token metadata
// @Description List token metadata
// @Tags tokens
// @ID list-token-metadata
// @Param network path string true "Network"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1) maximum(10)
// @Param max_level query integer false "Maximum token`s creation level (less than or equal)" mininum(1)
// @Param min_level query integer false "Minimum token`s creation level (greater than)" mininum(1)
// @Param creator query string false "Creator name" maxlength(25)
// @Param contract query string false "KT address" minlength(36) maxlength(36)
// @Param token_id query integer false "Token ID" mininum(0)
// @Accept  json
// @Produce  json
// @Success 200 {array} Token
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/tokens/{network}/metadata [get]
func GetTokenMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getByNetwork
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		var queryParams tokenMetadataRequest
		if err := c.BindQuery(&queryParams); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		tokens, err := getTokensWithSupply(ctx, tokenmetadata.GetContext{
			Contract: queryParams.Contract,
			MinLevel: queryParams.MinLevel,
			MaxLevel: queryParams.MaxLevel,
			Creator:  queryParams.Creator,
			TokenID:  queryParams.TokenID,
		}, queryParams.Size, queryParams.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, tokens)
	}
}
