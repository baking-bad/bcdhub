package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// TezosDomainsList godoc
// @Summary Show all tezos domains for network
// @Description Show all tezos domains for network
// @Tags domains
// @ID list-domains
// @Param network path string true "Network"
// @Param size query integer false "Transfers count" mininum(1)
// @Param offset query integer false "Offset" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {array} models.TezosDomain
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /domains/{network} [get]
func (ctx *Context) TezosDomainsList(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var args pageableRequest
	if err := c.BindQuery(&args); handleError(c, err, http.StatusBadRequest) {
		return
	}

	domains, err := ctx.ES.ListDomains(req.Network, args.Size, args.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, domains)
}

// ResolveDomain godoc
// @Summary Resolve domain
// @Description Resolve domain by address and vice versa
// @Tags domains
// @ID resolve-domain
// @Param network path string true "Network"
// @Param name query string false "Domain name"
// @Param address query string false "Address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} models.TezosDomain
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /domains/{network}/resolve [get]
func (ctx *Context) ResolveDomain(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var args resolveDomainRequest
	if err := c.BindQuery(&args); handleError(c, err, http.StatusBadRequest) {
		return
	}

	switch {
	case args.Name != "":
		td := models.TezosDomain{
			Network: req.Network,
			Name:    args.Name,
		}
		if err := ctx.ES.GetByID(&td); handleError(c, err, 0) {
			return
		}
		if td.Address == "" {
			handleError(c, errors.Errorf("Unknown domain name"), http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, td)
	case args.Address != "":
		td, err := ctx.ES.ResolveDomainByAddress(req.Network, args.Address)
		if handleError(c, err, 0) {
			return
		}
		c.JSON(http.StatusOK, td)
	default:
		handleError(c, errors.Errorf("Invalid resolve request: %##v", args), http.StatusBadRequest)
	}
}
