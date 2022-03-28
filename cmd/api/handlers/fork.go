package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// ForkContract -
func ForkContract(ctxs config.Contexts) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req forkRequest
		if err := c.BindJSON(&req); handleError(c, ctxs.Any().Storage, err, http.StatusBadRequest) {
			return
		}
		ctx, err := ctxs.Get(req.NetworkID())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		response, err := buildStorageDataFromForkRequest(ctx, req)
		if err != nil {
			handleError(c, ctx.Storage, err, 0)
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

func buildStorageDataFromForkRequest(ctx *config.Context, req forkRequest) (*ForkResponse, error) {
	var err error
	var scriptData []byte

	if req.Script != "" {
		scriptData = []byte(req.Script)
	} else {
		scriptData, err = getScriptBytes(ctx, req.Address, "")
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
