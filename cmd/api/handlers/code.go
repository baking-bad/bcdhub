package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/macros"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/logger"
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
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractCodeRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.Protocol == "" {
		state, err := ctx.Blocks.Last(req.Network)
		if ctx.handleError(c, err, 0) {
			return
		}
		req.Protocol = state.Protocol
	}

	code, err := ctx.getContractCodeJSON(req.Network, req.Address, req.Protocol)
	if ctx.handleError(c, err, 0) {
		return
	}

	collapsed, err := macros.Collapse(code, macros.GetAllFamilies())
	if err != nil {
		logger.Error(err)
		collapsed = code
	}

	resp, err := formatter.MichelineToMichelson(collapsed, false, formatter.DefLineSize)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDiff godoc
// @Summary Get diff between two contracts
// @Description Get diff between two contracts
// @Tags contract
// @ID get-diff
// @Param body body CodeDiffRequest true "Request body"
// @Accept  json
// @Produce  json
// @Success 200 {object} CodeDiffResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/diff [post]
func (ctx *Context) GetDiff(c *gin.Context) {
	var req CodeDiffRequest
	if err := c.BindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	resp, err := ctx.getContractCodeDiff(req.Left, req.Right)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ctx *Context) getContractCodeJSON(network, address, protocol string) (res gjson.Result, err error) {
	data, err := fetch.Contract(address, network, protocol, ctx.SharePath)
	if err != nil {
		return
	}
	contract := gjson.ParseBytes(data)
	if !contract.IsArray() && !contract.IsObject() {
		return res, errors.Errorf("Unknown contract: %s", address)
	}

	// return macros.FindMacros(contractJSON)
	return contract.Get("code"), nil
}

func (ctx *Context) getContractCodeDiff(left, right CodeDiffLeg) (res CodeDiffResponse, err error) {
	currentProtocols := make(map[string]string, 2)
	sides := make([]gjson.Result, 2)

	for i, leg := range []*CodeDiffLeg{&left, &right} {
		if leg.Protocol == "" {
			protocol, ok := currentProtocols[leg.Network]
			if !ok {
				state, err := ctx.Blocks.Last(leg.Network)
				if err != nil {
					return res, err
				}
				leg.Protocol = state.Protocol
				currentProtocols[leg.Network] = state.Protocol
			} else {
				leg.Protocol = protocol
			}
		}
		code, err := ctx.getContractCodeJSON(leg.Network, leg.Address, leg.Protocol)
		if err != nil {
			return res, err
		}
		collapsed, err := macros.Collapse(code, macros.GetAllFamilies())
		if err != nil {
			return res, err
		}
		sides[i] = collapsed
	}

	diff, err := formatter.Diff(sides[0], sides[1])
	if err != nil {
		return res, err
	}

	res.Left = left
	res.Right = right
	res.Diff = diff
	return res, nil
}
