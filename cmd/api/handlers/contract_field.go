package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type getContractFieldRequest struct {
	ID        int64  `uri:"id"`
	FieldName string `uri:"field"`
}

type getContractByNetworkAndAddressFieldRequest struct {
	getContractByNetworkAndAddressRequest
	FieldName string `uri:"field"`
}

var columnsContract = []string{
	"id",
	"level",
	"timestemp",
	"network",
	"balance",
	"manager",
	"delegate",
	"address",
	"kind",
	"script",
	"project_id",
	"language",
}

// GetContractField -
func (ctx *Context) GetContractField(c *gin.Context) {
	var req getContractFieldRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	field, err := getTableColumn(ctx.DB, "contracts", req.FieldName, fmt.Sprintf("id = %d", req.ID), columnsContract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.FieldName == "script" {
		var script map[string]interface{}
		if err := json.Unmarshal(field.([]byte), &script); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, script)
		return
	}

	c.JSON(http.StatusOK, field)
}

// GetContractByNetworkAndAddressField -
func (ctx *Context) GetContractByNetworkAndAddressField(c *gin.Context) {
	var req getContractByNetworkAndAddressFieldRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	field, err := getTableColumn(ctx.DB, "contracts", req.FieldName, fmt.Sprintf("network = '%s' AND address = '%s'", req.Network, req.Address), columnsContract)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.FieldName == "script" {
		var script map[string]interface{}
		if err := json.Unmarshal(field.([]byte), &script); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, script)
		return
	}

	c.JSON(http.StatusOK, field)
}
