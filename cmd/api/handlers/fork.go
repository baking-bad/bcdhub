package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ForkContract -
func (ctx *Context) ForkContract(c *gin.Context) {
	var req forkRequest
	if err := c.BindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	response, err := ctx.buildStorageDataFromForkRequest(req)
	if err != nil {
		var code int
		if errors.As(err, &meta.ValidationError{}) || errors.As(err, &meta.RequiredError{}) {
			code = http.StatusBadRequest
		}
		ctx.handleError(c, err, code)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) buildStorageDataFromForkRequest(req forkRequest) (gin.H, error) {
	var err error
	var script gjson.Result
	var metadata meta.Metadata

	if req.Script != "" {
		script = gjson.Parse(req.Script)
		metadata, err = meta.ParseMetadata(script.Get("#(prim==\"storage\").args"))
		if err != nil {
			return nil, err
		}

	} else {
		rpc, err := ctx.GetRPC(req.Network)
		if err != nil {
			return nil, err
		}
		script, err = rpc.GetScriptJSON(req.Address, 0)
		if err != nil {
			return nil, err
		}
		metadata, err = ctx.getStorageMetadata(req.Address, req.Network)
		if err != nil {
			return nil, err
		}
		script = script.Get("code")
	}

	storage, err := metadata.BuildEntrypointMicheline("0", req.Storage, false)
	if err != nil {
		return nil, err
	}

	return gin.H{
		"code":    script.Value(),
		"storage": storage.Get("value").Value(),
	}, nil
}
