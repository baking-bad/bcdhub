package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetContract -
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	by := map[string]interface{}{
		"address": req.Address,
		"network": req.Network,
	}
	cntr, err := ctx.ES.GetContract(by)
	if handleError(c, err, 0) {
		return
	}
	res, err := ctx.setProfileInfo(cntr)
	if handleError(c, err, 0) {
		return
	}
	if err := ctx.setContractSlug(&res); handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetRandomContract -
func (ctx *Context) GetRandomContract(c *gin.Context) {
	cntr, err := ctx.ES.GetRandomContract()
	if err != nil {
		if strings.Contains(err.Error(), "Unknown contract") {
			c.AbortWithStatus(204)
		} else {
			handleError(c, err, 0)
		}
	} else {
		c.JSON(http.StatusOK, cntr)
	}
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

func (ctx *Context) setContractSlug(contract *Contract) error {
	a, err := ctx.DB.GetAlias(contract.Address, contract.Network)
	if err != nil {
		return err
	}
	contract.Slug = a.Slug
	return nil
}
