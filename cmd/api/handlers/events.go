package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
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
	if err := c.BindQuery(&pageReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	subscriptions, err := ctx.DB.ListSubscriptions(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	events, err := ctx.getEvents(subscriptions, pageReq.Size, pageReq.Offset)
	if ctx.handleError(c, err, 0) {
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
	if ctx.handleError(c, err, 0) {
		return
	}

	events, err := ctx.getMempoolEvents(subscriptions)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, events)
}

func (ctx *Context) getEvents(subscriptions []database.Subscription, size, offset int64) ([]models.Event, error) {
	subs := make([]models.SubscriptionRequest, len(subscriptions))
	for i := range subscriptions {
		subs[i] = models.SubscriptionRequest{
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

		if bcd.IsContract(subscriptions[i].Address) {
			contract := contract.NewEmptyContract(subscriptions[i].Network, subscriptions[i].Address)
			if err := ctx.Storage.GetByID(&contract); err != nil {
				return []models.Event{}, err
			}
			subs[i].Hash = contract.Hash
			subs[i].ProjectID = contract.ProjectID
		}
	}

	return ctx.Storage.GetEvents(subs, size, offset)
}

func (ctx *Context) getMempoolEvents(subscriptions []database.Subscription) ([]models.Event, error) {
	events := make([]models.Event, 0)

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

		aliases, err := ctx.TZIP.GetAliasesMap(sub.Network)
		if err != nil {
			if !ctx.Storage.IsRecordNotFound(err) {
				return nil, err
			}
			aliases = make(map[string]string)
		}

		for _, item := range res {
			status := item.Body.Status
			if status == consts.Applied {
				status = consts.Pending
			}

			op := core.EventOperation{
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

			op.SourceAlias = aliases[op.Source]
			op.DestinationAlias = aliases[op.Destination]
			op.Errors, err = tezerrors.ParseArray(item.Body.Errors)
			if err != nil {
				return nil, err
			}

			if bcd.IsContract(op.Destination) && item.Body.Protocol != "" {
				if len(item.Body.Parameters) > 0 {
					p := types.NewParameters(item.Body.Parameters)

					parameter, err := ctx.getParameterType(op.Destination, op.Network, item.Body.Protocol)
					if err != nil {
						return events, err
					}

					tree, err := parameter.FromParameters(p)
					if err != nil && op.Errors == nil {
						return events, err
					}

					op.Entrypoint = tree.Nodes[0].GetName()
				} else {
					op.Entrypoint = consts.DefaultEntrypoint
				}
			}

			event := models.Event{
				Type:    models.EventTypeMempool,
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
