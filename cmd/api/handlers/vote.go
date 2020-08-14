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

// GetTasks -
func (ctx *Context) GetTasks(c *gin.Context) {
	var req pageableRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	userID := CurrentUserID(c)

	assesments, err := ctx.DB.GetAssessmentsWithValue(userID, database.AssessmentUndefined, uint(req.Size))
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) && handleError(c, err, 0) {
			return
		}
		assesments = make([]database.Assessments, 0)
	}

	c.JSON(http.StatusOK, assesments)
}

// GenerateTasks -
func (ctx *Context) GenerateTasks(c *gin.Context) {
	userID := CurrentUserID(c)
	tasks, err := ctx.ES.GetDiffTasks()
	if handleError(c, err, 0) {
		return
	}
	assesments := make([]database.Assessments, 0)
	for i := 0; i < len(tasks) && len(assesments) < 10; i++ {
		a := database.Assessments{
			Address1:   tasks[i].Address1,
			Network1:   tasks[i].Network1,
			Address2:   tasks[i].Address2,
			Network2:   tasks[i].Network2,
			UserID:     userID,
			Assessment: database.AssessmentUndefined,
		}
		if err := ctx.DB.CreateAssessment(&a); handleError(c, err, 0) {
			return
		}
		if a.Assessment == database.AssessmentUndefined {
			assesments = append(assesments, a)
		}
	}
	c.JSON(http.StatusOK, assesments)
}
