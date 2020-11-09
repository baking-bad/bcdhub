package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
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

	contract := models.NewEmptyContract(req.Network, req.Address)
	if err := ctx.ES.GetByID(&contract); handleError(c, err, 0) {
		return
	}
	res, err := ctx.contractPostprocessing(contract, c)
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
	var contract models.Contract

	for !helpers.StringInArray(contract.Network, ctx.Config.API.Networks) {
		cntr, err := ctx.ES.GetContractRandom()
		if handleError(c, err, 0) {
			return
		}
		contract = cntr
	}

	res, err := ctx.contractPostprocessing(contract, c)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
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

	if alias, err := ctx.ES.GetAlias(contract.Network, contract.Address); err == nil {
		res.Slug = alias.Slug
	} else if !elastic.IsRecordNotFound(err) {
		return res, err
	}

	tokenBalances, err := ctx.getAccountBalances(contract.Network, contract.Address)
	if err != nil {
		return res, err
	}
	res.Tokens = tokenBalances

	return res, nil
}
