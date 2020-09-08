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
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/gin-gonic/gin"
)

// ListDeployments -
func (ctx *Context) ListDeployments(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	_, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	var ctReq compilationRequest
	if err := c.BindQuery(&ctReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	deployments, err := ctx.DB.ListDeployments(userID, ctReq.Limit, ctReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// CreateDeployment -
func (ctx *Context) CreateDeployment(c *gin.Context) {
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

	if err = ctx.DB.CreateCompilationTask(&task); handleError(c, err, 0) {
		return
	}

	if err = ctx.runDeployment(task.ID, form); handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": compilation.StatusPending})
}

func (ctx *Context) runDeployment(taskID uint, form *multipart.Form) error {
	dir := filepath.Join(ctx.SharePath, "/compilations")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); ctx.handleCompilationError(taskID, err) {
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

	if err = ctx.MQPublisher.Send(mq.ChannelNew, data, data); ctx.handleCompilationError(taskID, err) {
		return err
	}

	return nil
}

// FinalizeDeployment -
func (ctx *Context) FinalizeDeployment(c *gin.Context) {
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

	task, err := ctx.DB.GetCompilationTask(req.TaskID)
	if handleError(c, err, 0) {
		return
	}

	paths := make([]string, len(task.Results))
	for i := range task.Results {
		paths[i] = task.Results[i].AWSPath
	}

	d := database.Deployment{
		UserID:            user.ID,
		CompilationTaskID: req.TaskID,
		OperationHash:     req.OperationHash,
		Sources:           paths,
	}

	err = ctx.DB.CreateDeployment(&d)
	if handleError(c, err, 0) {
		return
	}

	op, err := ctx.ES.GetOperations(
		map[string]interface{}{
			"hash": req.OperationHash,
		},
		0,
		true,
	)

	if !elastic.IsRecordNotFound(err) && handleError(c, err, 0) {
		return
	}

	if len(op) != 0 {
		h := metrics.New(ctx.ES, ctx.DB)

		err := h.SetOperationDeployment(&op[0])
		if handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": compilation.StatusSuccess})
}
