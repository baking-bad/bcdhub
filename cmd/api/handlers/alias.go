package handlers

import (
	"errors"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

type aliasRequest struct {
	Address string `form:"address"`
	Network string `form:"network"`
	Alias   string `form:"alias"`
}

// SetAlias -
func (ctx *Context) SetAlias(c *gin.Context) {
	var req aliasRequest

	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Address == "" || req.Alias == "" || req.Network == "" {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("Inavlid request data"))
		return
	}

	contract, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	project, err := ctx.ES.GetProject(contract.ProjectID)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	project.Alias = req.Alias

	if _, err := ctx.ES.UpdateDoc(elastic.DocProjects, project.ID, project); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
