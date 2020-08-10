package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetContract godoc
// @Summary Get contract info
// @Description Get full contract info
// @Tags contract
// @ID get-contract
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address} [get]
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntr, err := ctx.ES.GetContract(by)
	if handleError(c, err, 0) {
		return
	}
	res, err := ctx.contractPostprocessing(cntr, c)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetRandomContract godoc
// @Summary Show random contract
// @Description Get random contract with 2 or more operations
// @Tags contract
// @ID get-random-contract
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 500 {object} Error
// @Router /pick_random [get]
func (ctx *Context) GetRandomContract(c *gin.Context) {
	cntr, err := ctx.ES.GetContractRandom()
	if handleError(c, err, 0) {
		return
	}

	res, err := ctx.contractPostprocessing(cntr, c)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetContractTransfers godoc
// @Summary Show contract`s tokens transfers
// @Description Show contract`s tokens transfers
// @Tags contract
// @ID get-contract-transfers
// @Param size query integer false "Transfers count" mininum(1)
// @Param offset query integer false "Offset" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {object} TransferResponse
// @Failure 500 {object} Error
// @Router /{network}/{address}/transfers [get]
func (ctx *Context) GetContractTransfers(c *gin.Context) {
	var req pageableRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var contractRequest getContractRequest
	if err := c.BindUri(&contractRequest); handleError(c, err, http.StatusBadRequest) {
		return
	}

	transfers, err := ctx.ES.GetContractTransfers(contractRequest.Network, contractRequest.Address, req.Size, req.Offset)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, transfers)
}

func (ctx *Context) contractPostprocessing(contract models.Contract, c *gin.Context) (Contract, error) {
	var res Contract
	res.FromModel(contract)

	if userID, err := ctx.getUserFromToken(c); err == nil && userID != 0 {
		if sub, err := ctx.DB.GetSubscription(userID, res.Address, res.Network); err == nil {
			subscription := PrepareSubscription(sub)
			res.Subscription = &subscription
		} else if !gorm.IsRecordNotFoundError(err) {
			return res, err
		}
	}

	if totalSubscribed, err := ctx.DB.GetSubscriptionsCount(res.Address, res.Network); err == nil {
		res.TotalSubscribed = totalSubscribed
	} else {
		return res, err
	}

	if alias, err := ctx.DB.GetAlias(contract.Address, contract.Network); err == nil {
		res.Slug = alias.Slug
	} else {
		return res, err
	}

	return res, nil
}
