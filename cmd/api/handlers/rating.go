package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContractRating godoc
// @Summary Get contract rating
// @Description Get contract rating
// @Tags contract
// @ID get-contract-rating
// @Param network path string true "Network"
// @Param address path string true "KT address"
// @Accept  json
// @Produce  json
// @Success 200 {array} SubRating
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/rating [get]
func (ctx *Context) GetContractRating(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntrID, err := ctx.ES.GetContractID(by)
	if handleError(c, err, 0) {
		return
	}

	rating, err := ctx.DB.GetSubscriptionRating(cntrID)
	if handleError(c, err, 0) {
		return
	}
	var subRating SubRating
	subRating.FromModel(rating)
	c.JSON(http.StatusOK, subRating)
}
