package handlers

import (
	"net/http"

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
			Level:        data[i].Level,
			Timestamp:    data[i].Timestamp,
			Hash:         data[i].Hash,
			Protocol:     data[i].Protocol,
			PrevProtocol: data[i].Protocol,
			Kind:         data[i].Kind,
		}
	}
	return result
}
