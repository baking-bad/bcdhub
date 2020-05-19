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
// @Summary Get big map by pointer
// @Description Get contract rating
// @Tags contract
// @ID get-contract-bigmap
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {array} BigMapResponseItem
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/bigmap/{ptr} [get]
func (ctx *Context) GetBigMap(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq bigMapSearchRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.ES.GetBigMap(req.Address, req.Ptr, pageReq.Search, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMap(bm, req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBigMapByKeyHash godoc
// @Summary Get big map diffs by pointer and key hash
// @Description Get big map diffs by pointer and key hash
// @Tags contract
// @ID get-contract-bigmap-keyhash
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param ptr path integer true "Big map pointer"
// @Param key_hash path string true "Key hash in big map" minlength(54) maxlength(54)
// @Accept json
// @Produce json
// @Success 200 {array} BigMapDiffByKeyResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/bigmap/{ptr}/{key_hash} [get]
func (ctx *Context) GetBigMapByKeyHash(c *gin.Context) {
	var req getBigMapByKeyHashRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, total, err := ctx.ES.GetBigMapDiffByPtrAndKeyHash(req.Address, req.Ptr, req.KeyHash, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapItem(bm, req.Network, req.Address, req.KeyHash)
	if handleError(c, err, 0) {
		return
	}

	response.Total = total
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) prepareBigMap(data []elastic.BigMapDiff, network, address string) (res []BigMapResponseItem, err error) {
	contractMetadata, err := meta.GetContractMetadata(ctx.ES, address)
	if err != nil {
		return
	}

	res = make([]BigMapResponseItem, len(data))
	for i := range data {
		var protoSymLink string
		protoSymLink, err = meta.GetProtoSymLink(data[i].Protocol)
		if err != nil {
			return
		}

		metadata, ok := contractMetadata.Storage[protoSymLink]
		if !ok {
			err = fmt.Errorf("Unknown metadata: %s", protoSymLink)
			return
		}

		var value interface{}
		if data[i].Value != "" {
			val := gjson.Parse(data[i].Value)
			value, err = newmiguel.BigMapToMiguel(val, data[i].BinPath+"/v", metadata)
			if err != nil {
				return
			}
		}
		var key interface{}
		var keyString string
		if data[i].Key != "" {
			val := gjson.Parse(data[i].Key)
			key, err = newmiguel.BigMapToMiguel(val, data[i].BinPath+"/k", metadata)
			if err != nil {
				return
			}
			keyString = stringer.Stringify(val)
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
	return
}

func (ctx *Context) prepareBigMapItem(data []elastic.BigMapDiff, network, address, keyHash string) (res BigMapDiffByKeyResponse, err error) {
	contractMetadata, err := meta.GetContractMetadata(ctx.ES, address)
	if err != nil {
		return
	}

	var key interface{}
	values := make([]BigMapDiffItem, len(data))
	for i := range data {
		var protoSymLink string
		protoSymLink, err = meta.GetProtoSymLink(data[i].Protocol)
		if err != nil {
			return
		}

		metadata, ok := contractMetadata.Storage[protoSymLink]
		if !ok {
			err = fmt.Errorf("Unknown metadata: %s", protoSymLink)
			return
		}

		var value interface{}
		if data[i].Value != "" {
			val := gjson.Parse(data[i].Value)
			value, err = newmiguel.BigMapToMiguel(val, data[i].BinPath+"/v", metadata)
			if err != nil {
				return
			}
		}

		if i == 0 {
			if data[i].Key != "" {
				val := gjson.Parse(data[i].Key)
				key, err = newmiguel.BigMapToMiguel(val, data[i].BinPath+"/k", metadata)
				if err != nil {
					return
				}
			}
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
