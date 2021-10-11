package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGlobalConstant godoc
// @Summary Get global constant
// @Description Get global constant
// @Tags contract
// @ID get-global-constant
// @Param network path string true "network"
// @Param address path string true "expr address of constant" minlength(54) maxlength(54)
// @Accept json
// @Produce json
// @Success 200 {array} GlobalConstant
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/global_constants/{network}/{address} [get]
func (ctx *Context) GetGlobalConstant(c *gin.Context) {
	var req getGlobalConstantRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	constant, err := ctx.GlobalConstants.Get(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, NewGlobalConstantFromModel(constant))
}
