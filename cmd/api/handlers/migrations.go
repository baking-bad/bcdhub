package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

// GetContractMigrations -
func (ctx *Context) GetContractMigrations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	migrations, err := ctx.ES.GetMigrations(req.Network, req.Address)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, prepareMigrations(migrations))
}

func prepareMigrations(data []models.Migration) []Migration {
	result := make([]Migration, len(data))
	for i := range data {
		result[i] = Migration{
			Level:     data[i].Level,
			Timestamp: data[i].Timestamp,
			Hash:      data[i].Hash,
			Protocol:  data[i].Protocol,
			Vesting:   data[i].Vesting,
		}
	}
	return result
}

// GetMigration -
func (ctx *Context) GetMigration(c *gin.Context) {
	var req getMigrationRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	symLink, err := meta.GetProtoSymLink(req.Protocol)
	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	prevSymLink := getPreviousSymLink(symLink)
	if prevSymLink == "" {
		c.JSON(http.StatusOK, nil)
		return
	}

	codeDiff, err := ctx.getDiff(req.Address, req.Network, req.Address, req.Network, prevSymLink, symLink)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, codeDiff)
}

func getPreviousSymLink(currentSymLink string) string {
	switch currentSymLink {
	case consts.MetadataBabylon:
		return consts.MetadataAlpha
	default:
		return ""
	}
}
