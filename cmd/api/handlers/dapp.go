package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// GetDAppList -
func (ctx *Context) GetDAppList(c *gin.Context) {
	dapps, err := ctx.DApps.All()
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.SecureJSON(http.StatusOK, []interface{}{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	results := make([]DApp, len(dapps))
	for i := range dapps {
		result, err := ctx.appendDAppInfo(dapps[i], false, false)
		if ctx.handleError(c, err, 0) {
			return
		}
		results[i] = result
	}

	c.SecureJSON(http.StatusOK, results)
}

// GetDApp -
func (ctx *Context) GetDApp(c *gin.Context) {
	var req getDappRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	dapp, err := ctx.DApps.Get(req.Slug)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	response, err := ctx.appendDAppInfo(dapp, req.WithDetails, true)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, response)
}

func (ctx *Context) appendDAppInfo(dapp dapp.DApp, withDetails bool, withContracts bool) (DApp, error) {
	result := DApp{
		Name:              dapp.Name,
		ShortDescription:  dapp.ShortDescription,
		FullDescription:   dapp.FullDescription,
		WebSite:           dapp.WebSite,
		Slug:              dapp.Slug,
		AgoraReviewPostID: dapp.AgoraReviewPostID,
		AgoraQAPostID:     dapp.AgoraQAPostID,
		Authors:           dapp.Authors,
		SocialLinks:       dapp.SocialLinks,
		Interfaces:        dapp.Interfaces,
		Categories:        dapp.Categories,
		Soon:              dapp.Soon,
	}

	if len(dapp.Pictures) > 0 {
		screenshots := make([]Screenshot, 0)
		for _, pic := range dapp.Pictures {
			switch pic.Type {
			case "logo":
				result.Logo = pic.Link
			case "cover":
				result.Cover = pic.Link
			default:
				screenshots = append(screenshots, Screenshot{
					Type: pic.Type,
					Link: pic.Link,
				})
			}
		}

		result.Screenshots = screenshots
	}

	if withContracts && len(dapp.Contracts) > 0 {
		result.Contracts = make([]DAppContract, 0)

		for _, address := range dapp.Contracts {
			contract, err := ctx.Contracts.Get(types.Mainnet, address.Address)
			if err != nil {
				if ctx.Storage.IsRecordNotFound(err) {
					continue
				}
				return result, err
			}
			result.Contracts = append(result.Contracts, DAppContract{
				Network:     contract.Network.String(),
				Address:     contract.Address,
				Alias:       ctx.CachedAlias(contract.Network, contract.Address),
				ReleaseDate: contract.Timestamp.UTC(),
			})
		}
	}

	if withDetails {
		if len(dapp.DexTokens) > 0 {
			result.DexTokens = make([]TokenMetadata, 0)

			for _, token := range dapp.DexTokens {
				tokenMetadata, err := ctx.TokenMetadata.GetAll(tokenmetadata.GetContext{
					Contract: token.Contract,
					Network:  types.Mainnet,
					TokenID:  &token.TokenID,
				})
				if err != nil {
					if ctx.Storage.IsRecordNotFound(err) {
						continue
					}
					return result, err
				}

				var initiators, entrypoints []string
				for _, c := range dapp.Contracts {
					initiators = append(initiators, c.Address)
					entrypoints = append(entrypoints, c.Entrypoint...)
				}

				vol, err := ctx.Transfers.GetToken24HoursVolume(types.Mainnet, token.Contract, initiators, entrypoints, token.TokenID)
				if err != nil {
					if ctx.Storage.IsRecordNotFound(err) {
						continue
					}
					return result, err
				}

				for i := range tokenMetadata {
					tm := TokenMetadataFromElasticModel(tokenMetadata[i], true)
					tm.Volume24Hours = &vol
					result.DexTokens = append(result.DexTokens, tm)
				}
			}
		}

		for _, address := range dapp.Contracts {
			if address.WithTokens {
				metadata, err := ctx.TokenMetadata.GetAll(tokenmetadata.GetContext{
					Contract: address.Address,
					Network:  types.Mainnet,
					TokenID:  nil,
				})
				if err != nil {
					return result, err
				}
				tokens, err := ctx.addSupply(metadata)
				if err != nil {
					return result, err
				}
				result.Tokens = append(result.Tokens, tokens...)
			}

			if helpers.StringInArray("DEX", dapp.Categories) {
				vol, err := ctx.Operations.GetContract24HoursVolume(types.Mainnet, address.Address, address.Entrypoint)
				if err != nil {
					return result, err
				}

				result.Volume24Hours += vol
			}
		}
	}

	return result, nil
}
