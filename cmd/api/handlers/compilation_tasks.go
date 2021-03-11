package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/database"
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
	if ctx.handleError(c, err, 0) {
		return
	}

	var ctReq compilationTasksRequest
	if err := c.BindQuery(&ctReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	tasks, err := ctx.DB.ListCompilationTasks(userID, ctReq.Limit, ctReq.Offset, ctReq.Kind)
	if ctx.handleError(c, err, 0) {
		return
	}

	addSchemaToResults(tasks)

	c.JSON(http.StatusOK, tasks)
}

func addSchemaToResults(tasks []database.CompilationTask) {
	for i, t := range tasks {
		if t.Kind != compilation.KindDeployment || t.Status != compilation.StatusSuccess {
			continue
		}

		for j, r := range t.Results {
			if len(r.Script.RawMessage) == 0 {
				continue
			}

			data, err := r.Script.MarshalJSON()
			if err != nil {
				continue
			}

			script, err := ast.NewScript(data)
			if err != nil {
				continue
			}

			storage, err := script.StorageType()
			if err != nil {
				continue
			}

			schema, err := storage.ToJSONSchema()
			if err != nil {
				continue
			}

			typedef, err := storage.Docs(ast.DocsFull)
			if err != nil {
				continue
			}

			if len(typedef) > 0 {
				tasks[i].Results[j].Typedef = typedef[0]
			}

			tasks[i].Results[j].Schema = schema
		}
	}
}
