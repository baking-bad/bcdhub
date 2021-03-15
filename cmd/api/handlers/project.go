package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/gin-gonic/gin"
)

// GetSameContracts godoc
// @Summary Get same contracts
// @Description Get same contracts
// @Tags contract
// @ID get-contract-same
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param manager query string false "Manager"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Accept json
// @Produce json
// @Success 200 {object} SameContractsResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/same [get]
func (ctx *Context) GetSameContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var query sameContractRequest
	if err := c.BindQuery(&query); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	contract := contract.NewEmptyContract(req.Network, req.Address)
	err := ctx.Storage.GetByID(&contract)
	if ctx.handleError(c, err, 0) {
		return
	}

	sameContracts, err := ctx.Contracts.GetSameContracts(contract, query.Manager, query.Size, query.Offset)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	var response SameContractsResponse
	response.FromModel(sameContracts)

	c.JSON(http.StatusOK, response)
}

// GetSimilarContracts godoc
// @Summary Get similar contracts
// @Description Get similar contracts
// @Tags contract
// @ID get-contract-similar
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1) maximum(10)
// @Accept  json
// @Produce  json
// @Success 200 {object} SimilarContractsResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/similar [get]
func (ctx *Context) GetSimilarContracts(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	contract := contract.NewEmptyContract(req.Network, req.Address)
	err := ctx.Storage.GetByID(&contract)
	if ctx.handleError(c, err, 0) {
		return
	}

	similar, total, err := ctx.Contracts.GetSimilarContracts(contract, pageReq.Size, pageReq.Offset)
	if ctx.handleError(c, err, 0) {
		return
	}

	response := SimilarContractsResponse{
		Count:     total,
		Contracts: make([]SimilarContract, len(similar)),
	}
	for i := range similar {
		diff, err := ctx.getContractCodeDiff(
			CodeDiffLeg{Address: contract.Address, Network: contract.Network},
			CodeDiffLeg{Address: similar[i].Address, Network: similar[i].Network},
		)
		if ctx.handleError(c, err, 0) {
			return
		}
		response.Contracts[i].FromModel(similar[i], diff)
	}

	c.JSON(http.StatusOK, response)
}
