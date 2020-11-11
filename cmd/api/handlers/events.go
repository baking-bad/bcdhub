package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetEvents -
func (ctx *Context) GetEvents(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if handleError(c, err, 0) {
		return
	}

	events, err := ctx.getEvents(subscriptions, pageReq.Size, pageReq.Offset)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetMempoolEvents -
func (ctx *Context) GetMempoolEvents(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if handleError(c, err, 0) {
		return
	}

	events, err := ctx.getMempoolEvents(subscriptions)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, events)
}

func (ctx *Context) getEvents(subscriptions []database.Subscription, size, offset int64) ([]elastic.Event, error) {
	subs := make([]elastic.SubscriptionRequest, len(subscriptions))
	for i := range subscriptions {
		subs[i] = elastic.SubscriptionRequest{
			Address: subscriptions[i].Address,
			Network: subscriptions[i].Network,
			Alias:   subscriptions[i].Alias,

			WithSame:        subscriptions[i].WatchMask&WatchSame != 0,
			WithSimilar:     subscriptions[i].WatchMask&WatchSimilar != 0,
			WithMempool:     subscriptions[i].WatchMask&WatchMempool != 0,
			WithMigrations:  subscriptions[i].WatchMask&WatchMigrations != 0,
			WithDeployments: subscriptions[i].WatchMask&WatchDeployments != 0,
			WithCalls:       subscriptions[i].WatchMask&WatchCalls != 0,
			WithErrors:      subscriptions[i].WatchMask&WatchErrors != 0,
		}

		if helpers.IsContract(subscriptions[i].Address) {
			contract := models.NewEmptyContract(subscriptions[i].Network, subscriptions[i].Address)
			if err := ctx.ES.GetByID(&contract); err != nil {
				return []elastic.Event{}, err
			}
			subs[i].Hash = contract.Hash
			subs[i].ProjectID = contract.ProjectID
		}
	}

	return ctx.ES.GetEvents(subs, size, offset)
}

func (ctx *Context) getMempoolEvents(subscriptions []database.Subscription) ([]elastic.Event, error) {
	events := make([]elastic.Event, 0)

	for _, sub := range subscriptions {
		if sub.WatchMask&WatchMempool == 0 {
			continue
		}

		api, err := ctx.GetTzKTService(sub.Network)
		if err != nil {
			return events, err
		}

		res, err := api.GetMempool(sub.Address)
		if err != nil {
			return events, err
		}
		if len(res) == 0 {
			continue
		}

		for _, item := range res {
			status := item.Body.Status
			if status == consts.Applied {
				status = "pending" //nolint
			}

			op := elastic.EventOperation{
				Network:     sub.Network,
				Hash:        item.Body.Hash,
				Status:      status,
				Timestamp:   time.Unix(item.Body.Timestamp, 0).UTC(),
				Kind:        item.Body.Kind,
				Fee:         item.Body.Fee,
				Amount:      item.Body.Amount,
				Source:      item.Body.Source,
				Destination: item.Body.Destination,
			}

			op.SourceAlias = ctx.Aliases[op.Source]
			op.DestinationAlias = ctx.Aliases[op.Destination]
			op.Errors, err = cerrors.ParseArray(item.Body.Errors)
			if err != nil {
				return nil, err
			}

			if helpers.IsContract(op.Destination) && item.Body.Protocol != "" {
				if params := gjson.ParseBytes(item.Body.Parameters); params.Exists() {
					metadata, err := meta.GetMetadata(ctx.ES, op.Destination, consts.PARAMETER, item.Body.Protocol)
					if err != nil {
						return events, err
					}

					op.Entrypoint, err = metadata.GetByPath(params)
					if err != nil && op.Errors == nil {
						return events, err
					}
				} else {
					op.Entrypoint = consts.DefaultEntrypoint
				}
			}

			event := elastic.Event{
				Type:    elastic.EventTypeMempool,
				Address: sub.Address,
				Network: sub.Network,
				Alias:   sub.Alias,
				Body:    &op,
			}
			events = append(events, event)
		}
	}
	return events, nil
}
