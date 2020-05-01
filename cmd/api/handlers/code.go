package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/macros"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractCodeRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.Protocol == "" {
		state, err := ctx.ES.CurrentState(req.Network)
		if handleError(c, err, 0) {
			return
		}
		req.Protocol = state.Protocol
	}

	code, err := ctx.getContractCodeJSON(req.Network, req.Address, req.Protocol, req.Level)
	if handleError(c, err, 0) {
		return
	}

	collapsed, err := macros.Collapse(code)
	if err != nil {
		logger.Error(err)
		collapsed = code
	}

	resp, err := formatter.MichelineToMichelson(collapsed, false, formatter.DefLineSize)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDiff -
func (ctx *Context) GetDiff(c *gin.Context) {
	var req CodeDiffRequest
	if err := c.BindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	resp, err := ctx.getContractCodeDiff(req.Left, req.Right)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ctx *Context) getContractCodeJSON(network, address, protocol string, fallbackLevel int64) (res gjson.Result, err error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return res, fmt.Errorf("Unknown network %s", network)
	}

	contract, err := contractparser.GetContract(rpc, address, network, protocol, ctx.Dir, fallbackLevel)
	if err != nil {
		return
	}
	if !contract.IsArray() && !contract.IsObject() {
		return res, fmt.Errorf("Unknown contract: %s", address)
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
				state, err := ctx.ES.CurrentState(leg.Network)
				if err != nil {
					return res, err
				}
				leg.Protocol = state.Protocol
				currentProtocols[leg.Network] = state.Protocol
			} else {
				leg.Protocol = protocol
			}
		}
		code, err := ctx.getContractCodeJSON(leg.Network, leg.Address, leg.Protocol, leg.Level)
		if err != nil {
			return res, err
		}
		sides[i] = code
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
