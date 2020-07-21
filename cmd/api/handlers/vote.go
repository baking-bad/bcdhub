package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

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

	assessment := database.Assessments{
		Address1:   a.Address,
		Network1:   a.Network,
		Address2:   b.Address,
		Network2:   b.Network,
		UserID:     CurrentUserID(c),
		Assessment: req.Vote,
	}
	if err := ctx.DB.CreateOrUpdateAssessment(&assessment); handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, "")
}

// GetNextDiffTask -
func (ctx *Context) GetNextDiffTask(c *gin.Context) {
	userID := CurrentUserID(c)
	a, err := ctx.DB.GetNextAssessmentWithValue(userID, database.AssessmentUndefined)
	if gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusOK, nil)
		return
	}
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, a)
}
