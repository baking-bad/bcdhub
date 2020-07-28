package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ForkContract -
func (ctx *Context) ForkContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var forkReq forkRequest
	if err := c.BindJSON(&forkReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if handleError(c, err, 0) {
		return
	}

	script, err := rpc.GetScriptJSON(req.Address, 0)
	if handleError(c, err, 0) {
		return
	}

	storage, err := ctx.buildStorageMicheline(req.Network, req.Address, "0", forkReq.Storage, false)
	if handleError(c, err, 0) {
		return
	}
	response := gin.H{
		"code":    script.Get("code").Value(),
		"storage": storage.Get("value").Value(),
	}
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) buildStorageMicheline(network, address, binPath string, data map[string]interface{}, needValidate bool) (gjson.Result, error) {
	metadata, err := getStorageMetadata(ctx.ES, address, network)
	if err != nil {
		return gjson.Result{}, err
	}

	return metadata.BuildEntrypointMicheline(binPath, data, needValidate)
}
