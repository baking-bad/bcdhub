package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetBigMap -
func (ctx *Context) GetBigMap(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.ES.GetBigMap(req.Address, req.Ptr)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMap(bm, req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBigMapByKeyHash -
func (ctx *Context) GetBigMapByKeyHash(c *gin.Context) {
	var req getBigMapByKeyHashRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.ES.GetBigMapDiffByPtrAndKeyHash(req.Address, req.Ptr, req.KeyHash)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapItem(bm, req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctx *Context) prepareBigMap(data []elastic.BigMapDiff, network, address string) ([]BigMapResponseItem, error) {
	babyMeta, err := meta.GetMetadata(ctx.ES, address, consts.STORAGE, consts.HashBabylon)
	if err != nil {
		return nil, err
	}

	res := make([]BigMapResponseItem, len(data))
	for i := range data {
		metadata := babyMeta
		if network == consts.Mainnet && data[i].Level < consts.LevelBabylon {
			alphaMeta, err := meta.GetMetadata(ctx.ES, address, consts.STORAGE, consts.Hash1)
			if err != nil {
				return nil, err
			}
			metadata = alphaMeta
		}

		var value interface{}
		if data[i].Value != "" {
			val := gjson.Parse(data[i].Value)
			valueData, err := newmiguel.BigMapToMiguel(val, data[i].BinPath+"/v", metadata)
			if err != nil {
				return nil, err
			}
			value = valueData
		}
		var key interface{}
		if data[i].Key != "" {
			val := gjson.Parse(data[i].Key)
			keyData, err := newmiguel.BigMapToMiguel(val, data[i].BinPath+"/k", metadata)
			if err != nil {
				return nil, err
			}
			key = keyData
		}

		res[i] = BigMapResponseItem{
			Item: BigMapItem{
				Key:     key,
				KeyHash: data[i].KeyHash,
				Level:   data[i].Level,
				Value:   value,
			},
			Count: data[i].Count,
		}
	}
	return res, nil
}

func (ctx *Context) prepareBigMapItem(data []models.BigMapDiff, network, address string) (res []BigMapItem, err error) {
	alphaMeta, err := meta.GetMetadata(ctx.ES, address, consts.STORAGE, consts.Hash1)
	if err != nil {
		return
	}

	babyMeta, err := meta.GetMetadata(ctx.ES, address, consts.STORAGE, consts.HashBabylon)
	if err != nil {
		return
	}

	res = make([]BigMapItem, len(data))
	for i := range data {
		var value interface{}
		if data[i].Value != "" {
			val := gjson.Parse(data[i].Value)
			metadata := babyMeta
			if network == consts.Mainnet && data[i].Level < consts.LevelBabylon {
				metadata = alphaMeta
			}
			value, err = newmiguel.BigMapToMiguel(val, data[i].BinPath+"/v", metadata)
			if err != nil {
				return
			}
		}

		res[i] = BigMapItem{
			Key:     data[i].Key,
			KeyHash: data[i].KeyHash,
			Level:   data[i].Level,
			Value:   value,
		}

	}
	return
}
