package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
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

		if strings.HasPrefix(subscriptions[i].Address, "KT") {
			contract, err := ctx.ES.GetContract(map[string]interface{}{
				"address": subscriptions[i].Address,
				"network": subscriptions[i].Network,
			})
			if err != nil {
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
		if res.Get("#").Int() == 0 {
			continue
		}

		for _, item := range res.Array() {
			status := item.Get("status").String()
			if status == consts.Applied {
				status = "pending" //nolint
			}

			op := elastic.EventOperation{
				Network:     sub.Network,
				Hash:        item.Get("hash").String(),
				Status:      status,
				Timestamp:   time.Unix(item.Get("timestamp").Int(), 0).UTC(),
				Kind:        item.Get("kind").String(),
				Fee:         item.Get("fee").Int(),
				Amount:      item.Get("amount").Int(),
				Source:      item.Get("source").String(),
				Destination: item.Get("destination").String(),
			}

			op.SourceAlias = ctx.Aliases[op.Source]
			op.DestinationAlias = ctx.Aliases[op.Destination]
			op.Errors, err = cerrors.ParseArray([]byte(item.Get("errors").Raw))
			if err != nil {
				return nil, err
			}

			protocol := item.Get("protocol").String()

			if strings.HasPrefix(op.Destination, "KT") && protocol != "" {
				if params := item.Get("parameters"); params.Exists() {
					metadata, err := meta.GetMetadata(ctx.ES, op.Destination, consts.PARAMETER, protocol)
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
