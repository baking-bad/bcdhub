package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/aopoltorzhicky/bcdhub/internal/db/project"
	"net/http"
)

type getProjectRequest struct {
	ID int64 `uri:"id"`
}

// GetProject -
func (ctx *Context) GetProject(c *gin.Context) {
	var req getProjectRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	proj, err := project.Get(ctx.DB, req.ID)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, proj)
}
