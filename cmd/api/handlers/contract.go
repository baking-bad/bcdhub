package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetContract godoc
// @Summary Get contract info
// @Description Get full contract info
// @Tags contract
// @ID get-contract
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address} [get]
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
	res, err := ctx.contractPostprocessing(cntr, CurrentUserID(c))
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetRandomContract godoc
// @Summary Show random contract
// @Description Get random contract with 2 or more operations
// @Tags contract
// @ID get-random-contract
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 500 {object} Error
// @Router /pick_random [get]
func (ctx *Context) GetRandomContract(c *gin.Context) {
	cntr, err := ctx.ES.GetRandomContract()
	if handleError(c, err, 0) {
		return
	}
	res, err := ctx.contractPostprocessing(cntr)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctx *Context) contractPostprocessing(cntr models.Contract, userID uint) (Contract, error) {
	res, err := ctx.setProfileInfo(cntr, userID)
	if err != nil {
		return res, err
	}
	err = ctx.setContractSlug(&res)
	return res, err
}

func (ctx *Context) setProfileInfo(contract models.Contract, userID uint) (Contract, error) {
	var res Contract
	res.FromModel(contract)

	if userID == 0 {
		return res, nil
	}

	profile := ProfileInfo{}
	if _, err := ctx.DB.GetSubscription(res.Address, res.Network); err == nil {
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
