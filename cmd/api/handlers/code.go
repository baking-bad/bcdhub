package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
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

	symLink := meta.GetProtoSymLinkByLevel(req.Level, req.Network)
	code, err := ctx.getContractCode(req.Network, req.Address, symLink)
	if handleError(c, err, 0) {
		return
	}

	versions, err := ctx.getContractVersions(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, CodeResponse{
		Code:           code,
		CurrentVersion: symLink,
		Versions:       versions,
	})
}

func (ctx *Context) getContractVersions(network, address string) ([]CodeVersion, error) {
	protocols, err := ctx.ES.GetContractVersions(network, address)
	if err != nil {
		return nil, err
	}
	if len(protocols) == 0 {
		return []CodeVersion{
			CodeVersion{
				Name:  consts.MetadataBabylon,
				Level: 0,
			},
		}, nil
	}

	versions := make([]CodeVersion, len(protocols))
	for i := len(protocols) - 1; i > -1; i-- {
		if protocols[i] == consts.Vesting {
			versions[i] = CodeVersion{
				Name:  protocols[i],
				Level: 1,
			}
		} else {
			protoSymLink, err := meta.GetProtoSymLink(protocols[i])
			if err != nil {
				return nil, err
			}
			level := meta.GetLevelByProtoSymLink(protoSymLink, network)
			versions[i] = CodeVersion{
				Name:  protoSymLink,
				Level: level,
			}
			if i == 0 {
				versions = append([]CodeVersion{
					CodeVersion{
						Name:  consts.MetadataAlpha,
						Level: consts.LevelBabylon - 1,
					}}, versions...)
			}
		}
	}
	return versions, nil
}

// GetDiff -
func (ctx *Context) GetDiff(c *gin.Context) {
	var req getDiffRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	d, err := ctx.getDiff(req.SourceAddress, req.SourceNetwork, req.DestinationAddress, req.DestinationNetwork, consts.MetadataBabylon, consts.MetadataBabylon)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, d)
}

func (ctx *Context) getContractCode(network, address, protoSymLink string) (string, error) {
	contract, err := ctx.getContractCodeJSON(network, address, protoSymLink)
	if err != nil {
		return "", err
	}

	code := contract.Get("code")
	return formatter.MichelineToMichelson(code, false, formatter.DefLineSize)
}

func (ctx *Context) getContractCodeJSON(network, address, protoSymLink string) (res gjson.Result, err error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return res, fmt.Errorf("Unknown network %s", network)
	}
	contract, err := contractparser.GetContract(rpc, address, network, protoSymLink, ctx.Dir)
	if err != nil {
		return
	}
	if !contract.IsArray() && !contract.IsObject() {
		return res, fmt.Errorf("Unknown contract: %s", address)
	}

	// return macros.FindMacros(contractJSON)
	return contract, nil
}

func (ctx *Context) getDiff(srcAddress, srcNetwork, destAddress, destNetwork string, protoSymLinkSrc, protoSymLinkDest string) (res formatter.DiffResult, err error) {
	srcCode, err := ctx.getContractCodeJSON(srcNetwork, srcAddress, protoSymLinkSrc)
	if err != nil {
		return
	}
	destCode, err := ctx.getContractCodeJSON(destNetwork, destAddress, protoSymLinkDest)
	if err != nil {
		return
	}
	a := srcCode.Get("code")
	b := destCode.Get("code")
	res, err = formatter.Diff(a, b)
	if err != nil {
		return
	}
	res.NameLeft = fmt.Sprintf("%s [%s]", srcAddress, srcNetwork)
	res.NameRight = fmt.Sprintf("%s [%s]", destAddress, destNetwork)
	return
}
