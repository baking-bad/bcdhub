package handlers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xeipuuv/gojsonschema"
)

const metadataBytesLimit = 65536

// UploadMetadata -
func (ctx *Context) UploadMetadata(c *gin.Context) {
	if c.Request.Body == nil {
		c.SecureJSON(http.StatusBadRequest, nil)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if len(body) > metadataBytesLimit {
		c.SecureJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("exceeded max upload limit of %d bytes", metadataBytesLimit)})
		return
	}

	schemaLoader := gojsonschema.NewStringLoader(ctx.TzipSchema)
	documentLoader := gojsonschema.NewStringLoader(string(body))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if !result.Valid() {
		c.SecureJSON(http.StatusBadRequest, result.Errors())
		return
	}

	response, err := ctx.Pinata.PinJSONToIPFS(bytes.NewBuffer(body))
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.SecureJSON(http.StatusOK, MetadataResponse{Hash: response.IpfsHash})
}

// ListMetadata -
func (ctx *Context) ListMetadata(c *gin.Context) {
	list, err := ctx.Pinata.PinList()
	if ctx.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.SecureJSON(http.StatusOK, list)
}

// DeleteMetadata -
func (ctx *Context) DeleteMetadata(c *gin.Context) {
	var req metadataRequest
	if err := c.BindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := ctx.Pinata.UnPin(req.Hash); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.SecureJSON(http.StatusOK, gin.H{"message": "metadata was successfully deleted"})
}
