package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/gin-gonic/gin"
)

// GetDAppList -
func (ctx *Context) GetDAppList(c *gin.Context) {
	dapps, err := ctx.TZIP.GetDApps()
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	results := make([]DApp, len(dapps))
	for i := range dapps {
		result, err := ctx.appendDAppInfo(&dapps[i], false)
		if ctx.handleError(c, err, 0) {
			return
		}
		results[i] = result
	}

	c.JSON(http.StatusOK, results)
}

// GetDApp -
func (ctx *Context) GetDApp(c *gin.Context) {
	var req getDappRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	dapp, err := ctx.TZIP.GetDAppBySlug(req.Slug)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	response, err := ctx.appendDAppInfo(dapp, true)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctx *Context) appendDAppInfo(dapp *tzip.DApp, withDetails bool) (DApp, error) {
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

	if withDetails {
		if len(dapp.DexTokens) > 0 {
			result.DexTokens = make([]TokenMetadata, 0)

			for _, token := range dapp.DexTokens {
				tokenMetadata, err := ctx.TokenMetadata.Get(tokenmetadata.GetContext{
					Contract: token.Contract,
					Network:  consts.Mainnet,
					TokenID:  token.TokenID,
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
					entrypoints = append(entrypoints, c.DexVolumeEntrypoints...)
				}

				vol, err := ctx.Transfers.GetToken24HoursVolume(consts.Mainnet, token.Contract, initiators, entrypoints, token.TokenID)
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

		if len(dapp.Contracts) > 0 {
			result.Contracts = make([]DAppContract, 0)

			for _, address := range dapp.Contracts {
				contract := contract.NewEmptyContract(consts.Mainnet, address.Address)
				if err := ctx.Storage.GetByID(&contract); err != nil {
					return result, err
				}
				result.Contracts = append(result.Contracts, DAppContract{
					Network:     contract.Network,
					Address:     contract.Address,
					Alias:       contract.Alias,
					ReleaseDate: contract.Timestamp.UTC(),
				})

				tokens, err := ctx.getTokens(consts.Mainnet, address.Address)
				if err != nil {
					return result, err
				}
				result.Tokens = append(result.Tokens, tokens...)

				if helpers.StringInArray("DEX", dapp.Categories) {
					vol, err := ctx.Operations.GetContract24HoursVolume(consts.Mainnet, address.Address, address.DexVolumeEntrypoints)
					if err != nil {
						return result, err
					}

					result.Volume24Hours += vol
				}
			}
		}
	}

	return result, nil
}
