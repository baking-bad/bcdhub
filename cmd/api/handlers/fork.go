package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/gin-gonic/gin"
)

// ForkContract -
func (ctx *Context) ForkContract(c *gin.Context) {
	var req forkRequest
	if err := c.BindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	response, err := ctx.buildStorageDataFromForkRequest(req)
	if err != nil {
		// TODO: validation
		ctx.handleError(c, err, 0)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) buildStorageDataFromForkRequest(req forkRequest) (gin.H, error) {
	var err error
	var scriptData []byte

	if req.Script != "" {
		scriptData = []byte(req.Script)
	} else {
		scriptData, err = ctx.getScriptBytes(req.Address, req.Network, "")
		if err != nil {
			return nil, err
		}
	}
	script, err := ast.NewScript(scriptData)
	if err != nil {
		return nil, err
	}

	storageType, err := script.StorageType()
	if err != nil {
		return nil, err
	}
	if err = storageType.FromJSONSchema(req.Storage); err != nil {
		return nil, err
	}

	storage, err := storageType.ToParameters("")
	if err != nil {
		return nil, err
	}

	return gin.H{
		"code":    scriptData,
		"storage": storage,
	}, nil
}
