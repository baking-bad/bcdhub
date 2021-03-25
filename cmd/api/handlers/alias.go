package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBySlug godoc
// @Summary Get contract by slug
// @Description Get contract by slug
// @Tags contract
// @ID get-contract-by-slug
// @Param slug path string true "Slug"
// @Accept  json
// @Produce  json
// @Success 200 {object} Alias
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/slug/{slug} [get]
func (ctx *Context) GetBySlug(c *gin.Context) {
	var req getBySlugRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	a, err := ctx.TZIP.GetBySlug(req.Slug)
	if ctx.handleError(c, err, 0) {
		return
	}
	if a == nil {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}
	var alias Alias
	alias.FromModel(a)
	c.JSON(http.StatusOK, alias)
}
