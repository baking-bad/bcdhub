package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /slug/{slug} [get]
func (ctx *Context) GetBySlug(c *gin.Context) {
	var req getBySlugRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	a, err := ctx.ES.GetBySlug(req.Slug)
	if gorm.IsRecordNotFoundError(err) {
		handleError(c, err, http.StatusBadRequest)
		return
	}
	if handleError(c, err, 0) {
		return
	}
	var alias Alias
	alias.FromModel(a)
	c.JSON(http.StatusOK, alias)
}
