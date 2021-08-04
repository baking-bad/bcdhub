package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
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
		ctx.handleError(c, err, 0)
		return
	}
	c.SecureJSON(http.StatusOK, response)
}

func (ctx *Context) buildStorageDataFromForkRequest(req forkRequest) (*ForkResponse, error) {
	var err error
	var scriptData []byte

	if req.Script != "" {
		scriptData = []byte(req.Script)
	} else {
		scriptData, err = ctx.getScriptBytes(req.NetworkID(), req.Address, "")
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

	if storageType.Nodes[0].IsPrim(consts.PAIR) {
		req.Storage = map[string]interface{}{
			storageType.Nodes[0].GetName(): req.Storage,
		}
	}
	if err = storageType.FromJSONSchema(req.Storage); err != nil {
		return nil, err
	}

	storage, err := storageType.ToParameters("")
	if err != nil {
		return nil, err
	}

	return &ForkResponse{
		Script:  scriptData,
		Storage: storage,
	}, nil
}
