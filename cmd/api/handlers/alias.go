package handlers

import (
	"errors"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

// SetAlias -
func (ctx *Context) SetAlias(c *gin.Context) {
	var req aliasRequest
	if err := c.BindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.Address == "" || req.Alias == "" || req.Network == "" {
		handleError(c, errors.New("Inavlid request data"), http.StatusBadRequest)
		return
	}

	contract, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	})
	if handleError(c, err, 0) {
		return
	}

	project, err := ctx.ES.GetProject(contract.ProjectID)
	if handleError(c, err, 0) {
		return
	}

	project.Alias = req.Alias

	if _, err := ctx.ES.UpdateDoc(elastic.DocProjects, project.ID, project); handleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
