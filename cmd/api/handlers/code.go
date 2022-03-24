package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// GetContractCode godoc
// @Summary Get contract code
// @Description Get contract code
// @Tags contract
// @ID get-contract-code
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param protocol query string false "Protocol"
// @Param level query integer false "Level"
// @Accept  json
// @Produce  json
// @Success 200 {string} string
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/code [get]
func GetContractCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractCodeRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		if err := c.BindQuery(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		if req.Protocol == "" {
			state, err := ctx.Cache.CurrentBlock()
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			proto, err := ctx.Cache.ProtocolByID(state.ProtocolID)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			req.Protocol = proto.Hash
		}

		code, err := getContractCodeJSON(ctx, req.Address, req.Protocol)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		resp, err := formatter.MichelineToMichelson(code, false, formatter.DefLineSize)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

func getContractCodeJSON(ctx *config.Context, address string, protocol string) (res gjson.Result, err error) {
	symLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return res, err
	}
	script, err := ctx.Cache.ScriptBytes(address, symLink)
	if err != nil {
		return res, err
	}
	contract := gjson.ParseBytes(script)
	if !contract.IsArray() && !contract.IsObject() {
		return res, errors.Errorf("Unknown contract: %s", address)
	}

	return contract, nil
}
