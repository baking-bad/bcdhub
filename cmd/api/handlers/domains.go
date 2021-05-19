package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// TezosDomainsList godoc
// @Summary Show all tezos domains for network
// @Description Show all tezos domains for network
// @Tags domains
// @ID list-domains
// @Param network path string true "Network"
// @Param size query integer false "Transfers count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {object} DomainsResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/domains/{network} [get]
func (ctx *Context) TezosDomainsList(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var args pageableRequest
	if err := c.BindQuery(&args); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	domains, err := ctx.TezosDomains.ListDomains(types.NewNetwork(req.Network), args.Size, args.Offset)
	if ctx.handleError(c, err, 0) {
		return
	}

	arr := make([]TezosDomain, 0, len(domains.Domains))
	for _, domain := range domains.Domains {
		var td TezosDomain
		td.FromModel(domain)
		arr = append(arr, td)
	}

	response := DomainsResponse{
		Domains: arr,
		Total:   domains.Total,
	}

	c.JSON(http.StatusOK, response)
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
// @Success 200 {object} TezosDomain
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/domains/{network}/resolve [get]
func (ctx *Context) ResolveDomain(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var args resolveDomainRequest
	if err := c.BindQuery(&args); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	switch {
	case args.Name != "":
		domain := tezosdomain.TezosDomain{
			Network: req.NetworkID(),
			Name:    args.Name,
		}
		if err := ctx.Storage.GetByID(&domain); err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.JSON(http.StatusNoContent, gin.H{})
				return
			}
			ctx.handleError(c, err, 0)
			return
		}
		if domain.Address == "" {
			ctx.handleError(c, errors.Errorf("Unknown domain name"), http.StatusBadRequest)
			return
		}
		var resp TezosDomain
		resp.FromModel(domain)
		c.JSON(http.StatusOK, resp)
	case args.Address != "":
		domain, err := ctx.TezosDomains.ResolveDomainByAddress(req.NetworkID(), args.Address)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.JSON(http.StatusNoContent, gin.H{})
				return
			}
			ctx.handleError(c, err, 0)
			return
		}
		var resp TezosDomain
		resp.FromModel(*domain)
		c.JSON(http.StatusOK, domain)
	default:
		ctx.handleError(c, errors.Errorf("Invalid resolve request: %##v", args), http.StatusBadRequest)
	}
}
