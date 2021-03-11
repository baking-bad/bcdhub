package core

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	constants "github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// GetEvents -
func (e *Elastic) GetEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	if len(subscriptions) == 0 {
		return []models.Event{}, nil
	}

	if size == 0 || size > 50 { // TODO: ???
		size = consts.DefaultSize
	}

	shouldItems := make([]Item, 0)
	indicesMap := make(map[string]struct{})

	for i := range subscriptions {
		items := getEventsQuery(subscriptions[i], indicesMap)
		shouldItems = append(shouldItems, items...)
	}

	indices := make([]string, 0)
	for ind := range indicesMap {
		indices = append(indices, ind)
	}
	if len(indices) == 0 {
		return []models.Event{}, nil
	}

	return e.getEvents(subscriptions, shouldItems, indices, size, offset)
}

func (e *Elastic) getEvents(subscriptions []models.SubscriptionRequest, shouldItems []Item, indices []string, size, offset int64) ([]models.Event, error) {
	query := NewQuery()
	if len(shouldItems) != 0 {
		query.Query(
			Bool(
				Should(shouldItems...),
				MinimumShouldMatch(1),
			),
		)
	}
	query.Sort("timestamp", "desc").Size(size).From(offset)

	var response SearchResponse
	if err := e.Query(indices, query, &response); err != nil {
		return nil, err
	}

	hits := response.Hits.Hits
	events := make([]models.Event, len(hits))
	for i := range hits {
		event, err := parseEvent(subscriptions, hits[i])
		if err != nil {
			return nil, err
		}
		events[i] = event
	}

	return events, nil
}

func (m *EventMigration) makeEvent(subscriptions []models.SubscriptionRequest) (models.Event, error) {
	res := models.Event{
		Type:    models.EventTypeMigration,
		Address: m.Address,
		Network: m.Network,
		Body:    m,
	}
	for i := range subscriptions {
		if m.Network == subscriptions[i].Network && m.Address == subscriptions[i].Address {
			res.Alias = subscriptions[i].Alias
			return res, nil
		}
	}
	return models.Event{}, errors.Errorf("Couldn't find a matching subscription for %v", m)
}

func (o *EventOperation) makeEvent(subscriptions []models.SubscriptionRequest) (models.Event, error) {
	res := models.Event{
		Network: o.Network,
		Body:    o,
	}
	for i := range subscriptions {
		if o.Network != subscriptions[i].Network {
			continue
		}
		if o.Source != subscriptions[i].Address && o.Destination != subscriptions[i].Address {
			continue
		}

		res.Address = subscriptions[i].Address
		res.Alias = subscriptions[i].Alias

		switch {
		case o.Status != "applied":
			res.Type = models.EventTypeError
		case o.Source == subscriptions[i].Address && o.Kind == "origination":
			res.Type = models.EventTypeDeploy
		case o.Source == subscriptions[i].Address && o.Kind == "transaction":
			res.Type = models.EventTypeCall
		case o.Destination == subscriptions[i].Address && o.Kind == "transaction":
			res.Type = models.EventTypeInvoke
		}

		return res, nil
	}
	return models.Event{}, errors.Errorf("Couldn't find a matching subscription for %v", o)
}

func (c *EventContract) makeEvent(subscriptions []models.SubscriptionRequest) (models.Event, error) {
	res := models.Event{
		Body: c,
	}
	for i := range subscriptions {
		if c.Hash == subscriptions[i].Hash || c.ProjectID == subscriptions[i].ProjectID {
			res.Network = subscriptions[i].Network
			res.Address = subscriptions[i].Address
			res.Alias = subscriptions[i].Alias

			if c.Hash == subscriptions[i].Hash {
				res.Type = models.EventTypeSame
			} else {
				res.Type = models.EventTypeSimilar
			}
			return res, nil
		}
	}
	return models.Event{}, errors.Errorf("Couldn't find a matching subscription for %v", c)
}

func parseEvent(subscriptions []models.SubscriptionRequest, hit Hit) (models.Event, error) {
	switch hit.Index {
	case models.DocOperations:
		var event EventOperation
		if err := json.Unmarshal(hit.Source, &event); err != nil {
			return models.Event{}, err
		}
		return event.makeEvent(subscriptions)
	case models.DocMigrations:
		var event EventMigration
		if err := json.Unmarshal(hit.Source, &event); err != nil {
			return models.Event{}, err
		}
		return event.makeEvent(subscriptions)
	case models.DocContracts:
		var event EventContract
		if err := json.Unmarshal(hit.Source, &event); err != nil {
			return models.Event{}, err
		}
		return event.makeEvent(subscriptions)
	default:
		return models.Event{}, errors.Errorf("[parseEvent] Invalid reponse type: %s", hit.Index)
	}
}

func getEventsQuery(subscription models.SubscriptionRequest, indices map[string]struct{}) []Item {
	shouldItems := make([]Item, 0)

	if item := getEventsWatchCalls(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[models.DocOperations] = struct{}{}
	}
	if item := getEventsWatchErrors(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[models.DocOperations] = struct{}{}
	}
	if item := getEventsWatchDeployments(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[models.DocOperations] = struct{}{}
	}

	if bcd.IsContract(subscription.Address) {
		if item := getEventsWatchMigrations(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[models.DocMigrations] = struct{}{}
		}
		if item := getSubscriptionWithSame(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[models.DocContracts] = struct{}{}
		}
		if item := getSubscriptionWithSimilar(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[models.DocContracts] = struct{}{}
		}
	}

	return shouldItems
}

func getEventsWatchMigrations(subscription models.SubscriptionRequest) Item {
	if !subscription.WithMigrations {
		return nil
	}

	return Bool(
		Filter(
			In("kind.keyword", []string{constants.MigrationBootstrap, constants.MigrationLambda, constants.MigrationUpdate}),
			Term("network.keyword", subscription.Network),
			Term("address.keyword", subscription.Address),
		),
	)
}

func getEventsWatchDeployments(subscription models.SubscriptionRequest) Item {
	if !subscription.WithDeployments {
		return nil
	}

	return Bool(
		Filter(
			Term("kind.keyword", "origination"),
			Term("network.keyword", subscription.Network),
			Term("source.keyword", subscription.Address),
		),
	)
}

func getEventsWatchCalls(subscription models.SubscriptionRequest) Item {
	if !subscription.WithCalls {
		return nil
	}

	addressKeyword := "destination.keyword"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source.keyword"
	}

	return Bool(
		Filter(
			Term("kind.keyword", "transaction"),
			Term("status.keyword", "applied"),
			Term("network.keyword", subscription.Network),
			Term(addressKeyword, subscription.Address),
		),
	)
}

func getEventsWatchErrors(subscription models.SubscriptionRequest) Item {
	if !subscription.WithErrors {
		return nil
	}

	addressKeyword := "destination.keyword"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source.keyword"
	}

	return Bool(
		Filter(
			Term("network.keyword", subscription.Network),
			Term(addressKeyword, subscription.Address),
		),
		MustNot(
			Term("status.keyword", "applied"),
		),
	)
}

func getSubscriptionWithSame(subscription models.SubscriptionRequest) Item {
	if !subscription.WithSame {
		return nil
	}

	return Bool(
		Filter(Term("hash.keyword", subscription.Hash)),
		MustNot(Term("address.keyword", subscription.Address)),
	)
}

func getSubscriptionWithSimilar(subscription models.SubscriptionRequest) Item {
	if !subscription.WithSimilar {
		return nil
	}
	return Bool(
		Filter(Term("project_id.keyword", subscription.ProjectID)),
		MustNot(Term("hash.keyword", subscription.Hash)),
		MustNot(Term("address.keyword", subscription.Address)),
	)
}
