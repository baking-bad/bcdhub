package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetUserProfile -
func (ctx *Context) GetUserProfile(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	count, err := ctx.DB.GetUserCompletedAssesments(user.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	compilationTasks, err := ctx.DB.CountCompilationTasks(user.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	verifications, err := ctx.DB.CountVerifications(user.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	deployments, err := ctx.DB.CountDeployments(user.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	profile := userProfile{
		Login:            user.Login,
		AvatarURL:        user.AvatarURL,
		MarkReadAt:       user.MarkReadAt,
		RegisteredAt:     user.CreatedAt,
		MarkedContracts:  count,
		CompilationTasks: compilationTasks,
		Verifications:    verifications,
		Deployments:      deployments,

		Subscriptions: PrepareSubscriptions(subscriptions),
	}

	c.JSON(http.StatusOK, profile)
}

// UserMarkAllRead -
func (ctx *Context) UserMarkAllRead(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req markReadRequest
	if err := c.ShouldBindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.Timestamp > time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp can't be in the future"})
		return
	}

	err := ctx.DB.UpdateUserMarkReadAt(userID, req.Timestamp)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
