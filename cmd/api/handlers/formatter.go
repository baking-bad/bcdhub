package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	minLineSize = 20
	maxLineSize = 200
)

// GetFormatter -
func (ctx *Context) GetFormatter(c *gin.Context) {
	var req FormatterRequest
	if err := c.ShouldBindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.LineSize < minLineSize || req.LineSize > maxLineSize {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("linesize should be between [%v...%v]", minLineSize, maxLineSize)})
		return
	}

	if !gjson.Valid(req.Code) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json in code section"})
		return
	}

	parsedJSON := gjson.Parse(req.Code)
	result, err := formatter.MichelineToMichelson(parsedJSON, req.Inline, req.LineSize)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, result)
}
