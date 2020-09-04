package handlers

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/compiler/filesgenerator"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/gin-gonic/gin"
)

// DeployContract -
func (ctx *Context) DeployContract(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	form, err := c.MultipartForm()
	if handleError(c, err, 0) {
		logger.Error(err)
		return
	}

	task := database.CompilationTask{
		UserID: user.ID,
		Kind:   compilation.KindDeployment,
		Status: compilation.StatusPending,
	}

	err = ctx.DB.CreateCompilationTask(&task)
	if handleError(c, err, 0) {
		return
	}

	err = ctx.runDeployment(task.ID, form)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": compilation.StatusPending})
}

func (ctx *Context) runDeployment(taskID uint, form *multipart.Form) error {
	dir := filepath.Join(ctx.SharePath, "/compilations")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if ctx.handleCompilationError(taskID, err) {
			return err
		}
	}

	tempDir, err := ioutil.TempDir(dir, "deployment")
	if ctx.handleCompilationError(taskID, err) {
		return err
	}

	files, err := filesgenerator.FromUpload(form, tempDir)
	if ctx.handleCompilationError(taskID, err) {
		return err
	}

	data := compilation.Task{
		ID:    taskID,
		Kind:  compilation.KindDeployment,
		Files: files,
		Dir:   tempDir,
	}

	err = ctx.MQPublisher.Send(mq.ChannelNew, data, data)
	if ctx.handleCompilationError(taskID, err) {
		return err
	}

	return nil
}

// FinalizeDeploy -
func (ctx *Context) FinalizeDeploy(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req deploymentRequest
	if err := c.ShouldBindJSON(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	t, err := ctx.DB.GetCompilationTask(req.TaskID)
	if handleError(c, err, 0) {
		return
	}

	var results []database.CompilationTaskResult
	for _, r := range t.Results {
		if r.Status != compilation.StatusSuccess || r.ID == req.ResultID {
			results = append(results, r)
		}
	}

	task := database.CompilationTask{
		UserID:  user.ID,
		Address: req.Address,
		Network: req.Network,
		Kind:    compilation.KindVerification,
		Status:  compilation.StatusSuccess,
		Results: results,
	}

	err = ctx.DB.CreateCompilationTask(&task)
	if handleError(c, err, 0) {
		return
	}

	contract, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	})
	if handleError(c, err, 0) {
		return
	}

	err = ctx.MQPublisher.Send(mq.ChannelNew, &contract, contract.GetID())
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": compilation.StatusSuccess})
}
