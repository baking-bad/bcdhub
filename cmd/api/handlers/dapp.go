package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetDAppList -
func GetDAppList() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		dapps, err := ctx.DApps.All()
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusOK, []interface{}{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		results := make([]DApp, len(dapps))
		for i := range dapps {
			result, err := appendDAppInfo(ctx, dapps[i], false)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			results[i] = result
		}

		c.SecureJSON(http.StatusOK, results)
	}
}

// GetDApp -
func GetDApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getDappRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		dapp, err := ctx.DApps.Get(req.Slug)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		response, err := appendDAppInfo(ctx, dapp, true)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

// GetDexTokens -
func GetDexTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getDappRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		dapp, err := ctx.DApps.Get(req.Slug)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}
		if !helpers.StringInArray("DEX", dapp.Categories) {
			handleError(c, ctx.Storage, errors.New("dapp is not DEX"), http.StatusBadRequest)
			return
		}

		if len(dapp.DexTokens) == 0 {
			c.SecureJSON(http.StatusOK, dapp.DexTokens)
			return
		}

		dexTokens := make([]TokenMetadata, 0)

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
				handleError(c, ctx.Storage, err, 0)
				return
			}

			initiators := make(map[string]struct{})
			entrypoints := make(map[string]struct{})
			for _, c := range dapp.Contracts {
				initiators[c.Address] = struct{}{}
				for i := range c.Entrypoint {
					entrypoints[c.Entrypoint[i]] = struct{}{}
				}
			}

			initiatorsArr := make([]string, 0)
			for address := range initiators {
				initiatorsArr = append(initiatorsArr, address)
			}

			entrypointsArr := make([]string, 0)
			for entrypoint := range entrypoints {
				entrypointsArr = append(entrypointsArr, entrypoint)
			}

			vol, err := ctx.Transfers.GetToken24HoursVolume(types.Mainnet, token.Contract, initiatorsArr, entrypointsArr, token.TokenID)
			if err != nil {
				if ctx.Storage.IsRecordNotFound(err) {
					continue
				}
				handleError(c, ctx.Storage, err, 0)
				return
			}

			for i := range tokenMetadata {
				tm := TokenMetadataFromElasticModel(tokenMetadata[i], true)
				tm.Volume24Hours = &vol
				dexTokens = append(dexTokens, tm)
			}
		}
		c.SecureJSON(http.StatusOK, dexTokens)
	}
}

// GetDexTezosVolume -
func GetDexTezosVolume() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getDappRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		dapp, err := ctx.DApps.Get(req.Slug)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusNoContent, gin.H{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		if !helpers.StringInArray("DEX", dapp.Categories) {
			handleError(c, ctx.Storage, errors.New("dapp is not DEX"), http.StatusBadRequest)
			return
		}

		if len(dapp.Contracts) == 0 {
			c.SecureJSON(http.StatusOK, 0)
		}

		var volume float64
		for _, address := range dapp.Contracts {
			vol, err := ctx.Operations.GetContract24HoursVolume(types.Mainnet, address.Address, address.Entrypoint)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			volume += vol
		}
		c.SecureJSON(http.StatusOK, volume)
	}
}

func appendDAppInfo(ctx *config.Context, dapp dapp.DApp, withDetails bool) (DApp, error) {
	result := DApp{
		Name:             dapp.Name,
		ShortDescription: dapp.ShortDescription,
		FullDescription:  dapp.FullDescription,
		WebSite:          dapp.WebSite,
		Slug:             dapp.Slug,
		Authors:          dapp.Authors,
		SocialLinks:      dapp.SocialLinks,
		Interfaces:       dapp.Interfaces,
		Categories:       dapp.Categories,
		Soon:             dapp.Soon,
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
		if len(dapp.Contracts) > 0 {
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
					Address:     contract.Account.Address,
					Alias:       contract.Account.Alias,
					ReleaseDate: contract.Timestamp.UTC(),
				})

				if address.WithTokens {
					metadata, err := ctx.TokenMetadata.GetAll(tokenmetadata.GetContext{
						Contract: address.Address,
						Network:  types.Mainnet,
						TokenID:  nil,
					})
					if err != nil {
						return result, err
					}
					tokens, err := addSupply(ctx, metadata)
					if err != nil {
						return result, err
					}
					result.Tokens = append(result.Tokens, tokens...)
				}
			}
		}
	}

	return result, nil
}
