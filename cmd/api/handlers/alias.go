package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetAlias -
func (ctx *Context) SetAlias(c *gin.Context) {
	var request []aliasRequest
	if err := c.BindJSON(&request); handleError(c, err, http.StatusBadRequest) {
		return
	}

	for _, req := range request {
		if req.Address == "" || req.Alias == "" || req.Network == "" {
			handleError(c, errors.New("Inavlid request data"), http.StatusBadRequest)
			return
		}

		if err := ctx.DB.CreateAlias(req.Alias, req.Address, req.Network); handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetBySlug -
func (ctx *Context) GetBySlug(c *gin.Context) {
	var req getBySlugRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	a, err := ctx.DB.GetBySlug(req.Slug)
	if gorm.IsRecordNotFoundError(err) {
		handleError(c, err, http.StatusBadRequest)
		return
	}
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, a)
}
