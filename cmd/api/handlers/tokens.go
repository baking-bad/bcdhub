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

// GetFA godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard
// @Description Get all contracts that implement FA1/FA1.2 standard
// @Tags tokens
// @ID get-tokens
// @Param network path string true "Network"
// @Param offset query integer false "Offset (deprecated)"
// @Param last_id query string false "Last ID"
// @Param size query integer false "Requested count"
// @Accept json
// @Produce json
// @Success 200 {array} TokenContract
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network} [get]
func (ctx *Context) GetFA(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq cursorRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	if cursorReq.Size == 0 {
		cursorReq.Size = 20
	}
	var lastID int64
	if cursorReq.LastID != "" {
		var err error
		lastID, err = strconv.ParseInt(cursorReq.LastID, 10, 64)
		if handleError(c, err, http.StatusBadRequest) {
			return
		}
	}
	contracts, err := ctx.ES.GetTokens(req.Network, "", lastID, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	tokens, err := ctx.contractToTokens(contracts, req.Network, "")
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// GetFAByVersion godoc
// @Summary Get all contracts that implement FA1/FA1.2 standard by version
// @Description Get all contracts that implement FA1/FA1.2 standard by version
// @Tags tokens
// @ID get-tokens
// @Param network path string true "Network"
// @Param faversion path string true "FA token version" Enums(fa1, fa12, fa2)
// @Param offset query integer false "Offset (deprecated)"
// @Param last_id query string false "Last ID"
// @Param size query integer false "Requested count"
// @Accept json
// @Produce json
// @Success 200 {array} TokenContract
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network}/version/{faversion} [get]
func (ctx *Context) GetFAByVersion(c *gin.Context) {
	var req getTokensByVersion
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq cursorRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}
	if cursorReq.Size == 0 {
		cursorReq.Size = 20
	}
	var lastID int64
	if cursorReq.LastID != "" {
		var err error
		lastID, err = strconv.ParseInt(cursorReq.LastID, 10, 64)
		if handleError(c, err, http.StatusBadRequest) {
			return
		}
	}
	contracts, err := ctx.ES.GetTokens(req.Network, req.Version, lastID, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	tokens, err := ctx.contractToTokens(contracts, req.Network, req.Version)
	if handleError(c, err, 0) {
		return
	}
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
// @Param size query integer false "Requested count" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {object} PageableTokenTransfers
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /tokens/{network}/transfers/{address} [get]
func (ctx *Context) GetFA12OperationsForAddress(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var cursorReq cursorRequest
	if err := c.BindQuery(&cursorReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	operations, err := ctx.ES.GetTokenTransferOperations(req.Network, req.Address, cursorReq.LastID, cursorReq.Size)
	if handleError(c, err, 0) {
		return
	}

	ops, err := operationToTransfer(ctx.ES, operations)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, ops)
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
			return PageableTokenContracts{}, fmt.Errorf("Unknown interface version: %s", version)
		}
		methods := make([]string, len(interfaceVersion))
		for i := range interfaceVersion {
			methods[i] = interfaceVersion[i].Name
		}

		stats, err := ctx.ES.GetTokensStats(network, addresses, methods)
		if err != nil {
			return PageableTokenContracts{}, err
		}

		for i := range tokens {
			stat, ok := stats[tokens[i].Address]
			if !ok {
				continue
			}
			tokens[i].Methods = make(map[string]TokenMethodStats)
			for method, value := range stat {
				tokens[i].Methods[method] = TokenMethodStats{
					CallCount:          value.Count,
					AverageConsumedGas: value.ConsumedGas,
				}
			}
		}
	}

	var lastID int64
	if len(contracts) > 0 {
		lastID = contracts[len(contracts)-1].LastAction.UTC().Unix()
	}
	return PageableTokenContracts{
		Tokens: tokens,
		LastID: lastID,
	}, nil
}

func operationToTransfer(es elastic.IElastic, po elastic.PageableOperations) (PageableTokenTransfers, error) {
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
			Source:    op.Source,
			Nonce:     op.Nonce,
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
