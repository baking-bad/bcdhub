package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GetHead godoc
// @Summary Show indexer head
// @Description Get indexer head for each network
// @Tags head
// @ID get-indexer-head
// @Accept json
// @Produce json
// @Success 200 {array} Head
// @Failure 500 {object} Error
// @Router /v1/head [get]
func GetHead() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)

		body := make([]Head, 0)
		for network, ctx := range ctxs {
			block, err := ctx.Blocks.Last(c.Request.Context())
			if err != nil {
				if ctx.Storage.IsRecordNotFound(err) {
					continue
				}
				handleError(c, ctx.Storage, err, 0)
				return
			}

			head, err := getHead(ctx, network, block)
			if err != nil {
				log.Warn().Str("network", network.String()).Err(err).Msg("head api")
				continue
			}

			body = append(body, head)
		}

		sort.Sort(HeadsByNetwork(body))

		c.SecureJSON(http.StatusOK, body)
	}
}

// GetHeadByNetwork godoc
// @Summary Show indexer head for the network
// @Description Get indexer head for the network
// @Tags head
// @ID get-indexer-head-by-network
// @Param network path string true "Network"
// @Accept json
// @Produce json
// @Success 200 {object} Head
// @Failure 500 {object} Error
// @Router /v1/head/{network} [get]
func GetHeadByNetwork() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		block, err := ctx.Blocks.Last(c.Request.Context())
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		head, err := getHead(ctx, ctx.Network, block)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, head)
	}
}

func getHead(ctx *config.Context, network types.Network, block block.Block) (Head, error) {
	var found bool
	for j := range ctx.Config.API.Networks {
		if network.String() == ctx.Config.API.Networks[j] {
			found = true
			break
		}
	}

	if !found {
		return Head{}, errors.New("unknown network")
	}

	return Head{
		network:   network,
		Network:   network.String(),
		Level:     block.Level,
		Timestamp: block.Timestamp.UTC(),
		Protocol:  block.Protocol.Hash,
		Synced:    !block.Timestamp.UTC().Add(2 * time.Minute).Before(time.Now().UTC()),
	}, nil
}
