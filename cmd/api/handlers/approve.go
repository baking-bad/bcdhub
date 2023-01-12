package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// ApproveSchema -
func ApproveSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, gin.MIMEJSON, bcd.SchemaApprove)
	}
}

// ApproveData -
func ApproveData() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)

		var reqData approveDataRequest
		if err := c.BindJSON(&reqData); handleError(c, ctxs.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		response, err := approveData(reqData.Allowances)
		if handleError(c, ctxs.Any().Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

const (
	allowFa1ValueTemplate  = `{"prim":"Pair","args":[{"string":"%s"},{"int":"%d"}]}`
	revokeFa1ValueTemplate = `{"prim":"Pair","args":[{"string":"%s"},{"int":"0"}]}`
	approveEntrypoint      = "approve"
)

func approveData(allowances []tokenType) (ApproveResponse, error) {
	response := ApproveResponse{
		Fa12: make([]Approves, 0),
		Fa2:  make([]Approves, 0),
	}

	for i := range allowances {
		switch allowances[i].Type.TokenType {
		case 1:
			approve := Approves{
				Allows:  make([]Parameters, 0),
				Revokes: make([]Parameters, 0),
			}
			allow, revoke, err := approveDataFa1(allowances[i].Type)
			if err != nil {
				return response, err
			}
			approve.Allows = append(approve.Allows, allow)
			approve.Revokes = append(approve.Revokes, revoke)
			response.Fa12 = append(response.Fa12, approve)
		case 2:
			approve := Approves{
				Allows:  make([]Parameters, 0),
				Revokes: make([]Parameters, 0),
			}
			allow, revoke, err := approveDataFa2(allowances[i].Type)
			if err != nil {
				return response, err
			}
			approve.Allows = append(approve.Allows, allow)
			approve.Revokes = append(approve.Revokes, revoke)
			response.Fa2 = append(response.Fa2, approve)
		default:
			return response, errors.New("invalid token type")
		}
	}

	return response, nil
}

func approveDataFa1(allowance allowance) (Parameters, Parameters, error) {
	if allowance.Allowance == nil {
		return Parameters{}, Parameters{}, errors.New("empty allowance value for FA1.2 token")
	}
	allowValue := fmt.Sprintf(allowFa1ValueTemplate, allowance.TokenContract, *allowance.Allowance)
	revokeValue := fmt.Sprintf(revokeFa1ValueTemplate, allowance.TokenContract)

	return Parameters{
			Entrypoint:  approveEntrypoint,
			Value:       []byte(allowValue),
			Destination: allowance.TokenContract,
		}, Parameters{
			Entrypoint:  approveEntrypoint,
			Value:       []byte(revokeValue),
			Destination: allowance.TokenContract,
		}, nil
}

const (
	updateOperatorsEntrypoint = "update_operators"
	allowFa2ValueTemplate     = `{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"%s"},{"prim":"Pair","args":[{"string":"%s"},{"int":"%d"}]}]}]}`
	revokeFa2ValueTemplate    = `{"prim":"Right","args":[{"prim":"Pair","args":[{"string":"%s"},{"prim":"Pair","args":[{"string":"%s"},{"int":"%d"}]}]}]}`
)

func approveDataFa2(allowance allowance) (Parameters, Parameters, error) {
	if allowance.Owner == nil || allowance.TokenID == nil {
		return Parameters{}, Parameters{}, errors.New("empty owner or token id field for FA2 token")
	}
	allowValueStr := fmt.Sprintf(allowFa2ValueTemplate, *allowance.Owner, allowance.TokenContract, *allowance.TokenID)
	revokeValueStr := fmt.Sprintf(revokeFa2ValueTemplate, *allowance.Owner, allowance.TokenContract, *allowance.TokenID)

	return Parameters{
			Entrypoint:  updateOperatorsEntrypoint,
			Value:       []byte("[" + allowValueStr + "]"),
			Destination: allowance.TokenContract,
		}, Parameters{
			Entrypoint:  updateOperatorsEntrypoint,
			Value:       []byte("[" + revokeValueStr + "]"),
			Destination: allowance.TokenContract,
		}, nil
}
