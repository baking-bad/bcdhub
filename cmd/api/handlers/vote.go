package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/classification/metrics"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var model = []metrics.Metric{
	metrics.NewManager(),
	metrics.NewArray("Tags"),
	metrics.NewArray("FailStrings"),
	metrics.NewArray("Annotations"),
	metrics.NewBool("Language"),
	metrics.NewArray("Entrypoints"),
	metrics.NewFingerprintLength("parameter"),
	metrics.NewFingerprintLength("storage"),
	metrics.NewFingerprintLength("code"),
	metrics.NewFingerprint("parameter"),
	metrics.NewFingerprint("storage"),
	metrics.NewFingerprint("code"),
}

// Vote -
func (ctx *Context) Vote(c *gin.Context) {
	var req voteRequest
	if err := c.BindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	a, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.SourceAddress,
		"network": req.SourceNetwork,
	})
	if handleError(c, err, 0) {
		return
	}

	b, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.DestinationAddress,
		"network": req.DestinationNetwork,
	})
	if handleError(c, err, 0) {
		return
	}

	userID := CurrentUserID(c)
	if err := ctx.DB.CreateOrUpdateAssessment(a.Address, a.Network, b.Address, b.Network, userID, req.Vote); handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, "")
}

// GetNextDiffTask -
func (ctx *Context) GetNextDiffTask(c *gin.Context) {
	userID := CurrentUserID(c)
	a, err := ctx.DB.GetNextAssessmentWithValue(userID, 10)
	if gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusOK, nil)
		return
	}
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, a)
}
