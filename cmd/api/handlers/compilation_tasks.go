package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListCompilationTasks -
func (ctx *Context) ListCompilationTasks(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	_, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	var ctReq compilationTasksRequest
	if err := c.BindQuery(&ctReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	tasks, err := ctx.DB.ListCompilationTasks(userID, ctReq.Limit, ctReq.Offset, ctReq.Kind)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, tasks)
}
