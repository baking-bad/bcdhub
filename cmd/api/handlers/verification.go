package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/compiler/filesgenerator"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/providers"
	"github.com/gin-gonic/gin"
)

// ListVerifications -
func (ctx *Context) ListVerifications(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	_, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	var ctReq compilationRequest
	if err := c.BindQuery(&ctReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	verifications, err := ctx.DB.ListVerifications(userID, ctReq.Limit, ctReq.Offset)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, verifications)
}

// CreateVerification -
func (ctx *Context) CreateVerification(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req verificationRequest
	if err := c.ShouldBindJSON(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	provider, err := providers.NewPublic(user.Provider)
	if ctx.handleError(c, err, 0) {
		return
	}

	task := database.CompilationTask{
		UserID:  user.ID,
		Address: req.Address,
		Network: req.NetworkID(),
		Account: req.Account,
		Repo:    req.Repo,
		Ref:     req.Ref,
		Kind:    compilation.KindVerification,
		Status:  compilation.StatusPending,
	}

	err = ctx.DB.CreateCompilationTask(&task)
	if ctx.handleError(c, err, 0) {
		return
	}

	go ctx.runVerification(task.ID, provider.ArchivePath(req.Account, req.Repo, req.Ref))

	c.JSON(http.StatusOK, gin.H{"status": compilation.StatusPending})
}

func (ctx *Context) runVerification(taskID uint, sourceURL string) {
	dir := filepath.Join(ctx.SharePath, "/compilations")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if ctx.handleCompilationError(taskID, err) {
			return
		}
	}

	tempDir, err := ioutil.TempDir(dir, "verification")
	if ctx.handleCompilationError(taskID, err) {
		return
	}

	files, err := filesgenerator.FromRepo(sourceURL, tempDir)
	if ctx.handleCompilationError(taskID, err) {
		return
	}

	data := compilation.Task{
		ID:    taskID,
		Kind:  compilation.KindVerification,
		Files: files,
		Dir:   tempDir,
	}

	err = ctx.MQ.Send(data)
	if ctx.handleCompilationError(taskID, err) {
		return
	}
}

func (ctx *Context) handleCompilationError(taskID uint, err error) bool {
	if err == nil {
		return false
	}

	logger.Error(err)

	if err := ctx.DB.UpdateTaskStatus(taskID, compilation.StatusError); err != nil {
		logger.Error(err)
	}

	return true
}
