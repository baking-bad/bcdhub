package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type getContractRequest struct {
	Address string `uri:"address"`
	Network string `uri:"network"`
}

// GetContract -
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntr, err := ctx.ES.GetContract(by)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res, err := ctx.setProfileInfo(cntr)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetRandomContract -
func (ctx *Context) GetRandomContract(c *gin.Context) {
	cntr, err := ctx.ES.GetRandomContract()
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, cntr)
}

func (ctx *Context) setProfileInfo(contract models.Contract) (Contract, error) {
	res := Contract{
		Contract: &contract,
	}
	if ctx.OAUTH.UserID == 0 {
		return res, nil
	}

	profile := ProfileInfo{}
	_, err := ctx.DB.GetSubscription(res.ID, "contract")
	if err == nil {
		profile.Subscribed = true
	} else {
		if !gorm.IsRecordNotFoundError(err) {
			return res, err
		}
	}
	res.Profile = &profile
	return res, nil
}
