package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
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

			res := gjson.ParseBytes(data)
			metadata, err := meta.ParseMetadata(res.Get("#(prim==\"storage\").args"))
			if err != nil {
				continue
			}

			schema, err := jsonschema.Create("0", metadata)
			if err != nil {
				continue
			}

			typedef, err := docstring.GetStorage(metadata)
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
