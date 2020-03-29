package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContractMigrations -
func (ctx *Context) GetContractMigrations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	migrations, err := ctx.ES.GetMigrations(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, migrations)
}
