package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetBigMap godoc
// @Summary Get big map info by pointer
// @Description Get big map info by pointer
// @Tags bigmap
// @ID get-bigmap
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {array} GetBigMapResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /bigmap/{network}/{ptr} [get]
func (ctx *Context) GetBigMap(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.ES.GetBigMapKeys(req.Ptr, req.Network, "", 10000, 0)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMap(bm)
	if handleError(c, err, 0) {
		return
	}

	if response.Address == "" {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBigMapKeys godoc
// @Summary Get big map keys by pointer
// @Description Get big map keys by pointer
// @Tags bigmap
// @ID get-bigmap-keys
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param search query string false "Search string"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {array} BigMapResponseItem
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /bigmap/{network}/{ptr}/keys [get]
func (ctx *Context) GetBigMapKeys(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq bigMapSearchRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.ES.GetBigMapKeys(req.Ptr, req.Network, pageReq.Search, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapKeys(bm)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBigMapByKeyHash godoc
// @Summary Get big map diffs by pointer and key hash
// @Description Get big map diffs by pointer and key hash
// @Tags bigmap
// @ID get-bigmap-keyhash
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param key_hash path string true "Key hash in big map" minlength(54) maxlength(54)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} BigMapDiffByKeyResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /bigmap/{network}/{ptr}/keys/{key_hash} [get]
func (ctx *Context) GetBigMapByKeyHash(c *gin.Context) {
	var req getBigMapByKeyHashRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, total, err := ctx.ES.GetBigMapDiffByPtrAndKeyHash(req.Ptr, req.Network, req.KeyHash, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapItem(bm, req.KeyHash)
	if handleError(c, err, 0) {
		return
	}

	response.Total = total
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) prepareBigMap(data []elastic.BigMapDiff) (res GetBigMapResponse, err error) {
	if len(data) == 0 {
		return
	}
	res.Address = data[0].Address
	res.Network = data[0].Network
	res.Ptr = data[0].Ptr
	for i := range data {
		if data[i].Value != "" {
			res.ActiveKeys++
		}
	}
	return
}

func (ctx *Context) prepareBigMapKeys(data []elastic.BigMapDiff) ([]BigMapResponseItem, error) {
	if len(data) == 0 {
		return []BigMapResponseItem{}, nil
	}

	contractMetadata, err := meta.GetContractMetadata(ctx.ES, data[0].Address)
	if err != nil {
		return nil, err
	}

	res := make([]BigMapResponseItem, len(data))
	for i := range data {
		key, value, keyString, err := prepareItem(data[i], contractMetadata)
		if err != nil {
			return nil, err
		}

		res[i] = BigMapResponseItem{
			Item: BigMapItem{
				Key:       key,
				KeyHash:   data[i].KeyHash,
				KeyString: keyString,
				Level:     data[i].Level,
				Value:     value,
				Timestamp: data[i].Timestamp,
			},
			Count: data[i].Count,
		}
	}
	return res, nil
}

func (ctx *Context) prepareBigMapItem(data []elastic.BigMapDiff, keyHash string) (res BigMapDiffByKeyResponse, err error) {
	if len(data) == 0 {
		return
	}

	contractMetadata, err := meta.GetContractMetadata(ctx.ES, data[0].Address)
	if err != nil {
		return
	}

	var key, value interface{}
	values := make([]BigMapDiffItem, len(data))
	for i := range data {
		key, value, _, err = prepareItem(data[i], contractMetadata)
		if err != nil {
			return
		}

		values[i] = BigMapDiffItem{
			Level:     data[i].Level,
			Value:     value,
			Timestamp: data[i].Timestamp,
		}

	}
	res.Values = values
	res.KeyHash = keyHash
	res.Key = key
	return
}

func prepareItem(item elastic.BigMapDiff, contractMetadata *meta.ContractMetadata) (interface{}, interface{}, string, error) {
	var protoSymLink string
	protoSymLink, err := meta.GetProtoSymLink(item.Protocol)
	if err != nil {
		return nil, nil, "", err
	}

	metadata, ok := contractMetadata.Storage[protoSymLink]
	if !ok {
		err = fmt.Errorf("Unknown metadata: %s", protoSymLink)
		return nil, nil, "", err
	}

	binPath := item.BinPath
	if protoSymLink == "alpha" {
		binPath = "0/0"
	}

	var value interface{}
	if item.Value != "" {
		val := gjson.Parse(item.Value)
		value, err = newmiguel.BigMapToMiguel(val, binPath+"/v", metadata)
		if err != nil {
			return nil, nil, "", err
		}
	}
	var key interface{}
	var keyString string
	if item.Key != "" {
		val := gjson.Parse(item.Key)
		key, err = newmiguel.BigMapToMiguel(val, binPath+"/k", metadata)
		if err != nil {
			return nil, nil, "", err
		}
		keyString = stringer.Stringify(val)
	}
	return key, value, keyString, err
}
