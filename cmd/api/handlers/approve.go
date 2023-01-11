package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// ApproveSchema godoc
// @Summary Get schema for approvals
// @Description Get schema for approvals
// @Tags contract
// @ID get-contract-approve-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept json
// @Produce json
// @Success 200 {object} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/approve/schema/{tag} [get]
func ApproveSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getApproveSchemaRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		contract, err := ctx.Contracts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var has bool
		var schema []byte
		switch req.Tag {
		case 1:
			has = contract.Tags.Has(types.FA12Tag)
			schema = bcd.SchemaApproveFa1
		case 2:
			has = contract.Tags.Has(types.FA2Tag)
			schema = bcd.SchemaApproveFa2
		default:
			handleError(c, ctx.Storage, errors.New("invalid tag"), http.StatusBadRequest)
			return
		}

		if !has {
			handleError(c, ctx.Storage, ErrNotFAContract, http.StatusBadRequest)
			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, schema)
	}
}

// ApproveDataFa1 godoc
// @Summary Get data for approvals
// @Description Get data for approvals
// @Tags contract
// @ID get-contract-approve-data-1
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param body body approveDataFa1Request true "Request body"
// @Accept json
// @Produce json
// @Success 200 {object} ApproveResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/approve/data/1 [get]
func ApproveDataFa1() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var reqData approveDataFa1Request
		if err := c.BindJSON(&reqData); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		contract, err := ctx.Contracts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		if !contract.Tags.Has(types.FA12Tag) {
			handleError(c, ctx.Storage, ErrNotFAContract, 0)
			return
		}

		c.SecureJSON(http.StatusOK, approveDataFa1(ctx, reqData.Allowances))
	}
}

// ApproveDataFa2 godoc
// @Summary Get data for FA2 approvals
// @Description Get data for FA2 approvals
// @Tags contract
// @ID get-contract-approve-data-2
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param body body approveDataFa2Request true "Request body"
// @Accept json
// @Produce json
// @Success 200 {object} ApproveResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/approve/data/2 [get]
func ApproveDataFa2() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var reqData approveDataFa2Request
		if err := c.BindJSON(&reqData); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		contract, err := ctx.Contracts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		if !contract.Tags.Has(types.FA2Tag) {
			handleError(c, ctx.Storage, ErrNotFAContract, 0)
			return
		}

		c.SecureJSON(http.StatusOK, approveDataFa2(ctx, reqData.Allowances))
	}
}

const (
	allowFa1ValueTemplate  = `{"prim":"Pair","args":[{"string":"%s"},{"int":"%s"}]}`
	revokeFa1ValueTemplate = `{"prim":"Pair","args":[{"string":"%s"},{"int":"0"}]}`
	approveEntrypoint      = "approve"
)

func approveDataFa1(ctx *config.Context, allowances []allowanceFa1) *ApproveResponse {
	response := &ApproveResponse{
		Fa:      1,
		Allows:  make([]Parameters, 0),
		Revokes: make([]Parameters, 0),
	}

	if len(allowances) == 0 {
		return response
	}

	for i := range allowances {
		allowValue := fmt.Sprintf(allowFa1ValueTemplate, allowances[i].TokenContract, allowances[i].Allowance)
		revokeValue := fmt.Sprintf(revokeFa1ValueTemplate, allowances[i].TokenContract)

		response.Allows = append(response.Allows, Parameters{
			Entrypoint: approveEntrypoint,
			Value:      []byte(allowValue),
		})
		response.Revokes = append(response.Revokes, Parameters{
			Entrypoint: approveEntrypoint,
			Value:      []byte(revokeValue),
		})
	}

	return response
}

const (
	updateOperatorsEntrypoint = "update_operators"
	allowFa2ValueTemplate     = `{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"%s"},{"prim":"Pair","args":[{"string":"%s"},{"int":"%s"}]}]}]}`
	revokeFa2ValueTemplate    = `{"prim":"Right","args":[{"prim":"Pair","args":[{"string":"%s"},{"prim":"Pair","args":[{"string":"%s"},{"int":"%s"}]}]}]}`
)

func approveDataFa2(ctx *config.Context, allowances []allowanceFa2) *ApproveResponse {
	response := &ApproveResponse{
		Fa:      1,
		Allows:  make([]Parameters, 0),
		Revokes: make([]Parameters, 0),
	}

	if len(allowances) == 0 {
		return response
	}

	var (
		allowValue  = new(bytes.Buffer)
		revokeValue = new(bytes.Buffer)
	)
	allowValue.WriteByte('[')
	revokeValue.WriteByte('[')

	for i := range allowances {
		if i > 0 {
			allowValue.WriteByte(',')
			revokeValue.WriteByte(',')
		}

		allowValueStr := fmt.Sprintf(allowFa2ValueTemplate, allowances[i].Owner, allowances[i].TokenContract, allowances[i].TokenID)
		revokeValueStr := fmt.Sprintf(revokeFa2ValueTemplate, allowances[i].Owner, allowances[i].TokenContract, allowances[i].TokenID)

		allowValue.WriteString(allowValueStr)
		revokeValue.WriteString(revokeValueStr)

	}
	allowValue.WriteByte(']')
	revokeValue.WriteByte(']')

	response.Allows = append(response.Allows, Parameters{
		Entrypoint: updateOperatorsEntrypoint,
		Value:      allowValue.Bytes(),
	})
	response.Revokes = append(response.Revokes, Parameters{
		Entrypoint: updateOperatorsEntrypoint,
		Value:      revokeValue.Bytes(),
	})

	return response
}
