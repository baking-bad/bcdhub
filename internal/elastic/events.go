package elastic

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// GetEvents -
func (e *Elastic) GetEvents(subscriptions []SubscriptionRequest, size, offset int64) ([]Event, error) {
	if len(subscriptions) == 0 {
		return []Event{}, nil
	}

	if size == 0 || size > 50 { // TODO: ???
		size = defaultSize
	}

	shouldItems := make([]qItem, 0)
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
		return []Event{}, nil
	}

	return e.getEvents(subscriptions, shouldItems, indices, size, offset)
}

func (e *Elastic) getEvents(subscriptions []SubscriptionRequest, shouldItems []qItem, indices []string, size, offset int64) ([]Event, error) {
	query := newQuery()
	if len(shouldItems) != 0 {
		query.Query(
			boolQ(
				should(shouldItems...),
				minimumShouldMatch(1),
			),
		)
	}
	query.Sort("timestamp", "desc").Size(size).From(offset)

	response, err := e.query(indices, query)
	if err != nil {
		return nil, err
	}

	hits := response.Get("hits.hits").Array()
	events := make([]Event, len(hits))
	for i, hit := range hits {
		event, err := parseEvent(subscriptions, hit)
		if err != nil {
			return nil, err
		}
		events[i] = event
	}

	return events, nil
}

func (m *EventMigration) makeEvent(subscriptions []SubscriptionRequest) (Event, error) {
	res := Event{
		Type:    EventTypeMigration,
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
	return Event{}, errors.Errorf("Couldn't find a matching subscription for %v", m)
}

func (o *EventOperation) makeEvent(subscriptions []SubscriptionRequest) (Event, error) {
	res := Event{
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
			res.Type = EventTypeError
		case o.Source == subscriptions[i].Address && o.Kind == "origination":
			res.Type = EventTypeDeploy
		case o.Source == subscriptions[i].Address && o.Kind == "transaction":
			res.Type = EventTypeCall
		case o.Destination == subscriptions[i].Address && o.Kind == "transaction":
			res.Type = EventTypeInvoke
		}

		return res, nil
	}
	return Event{}, errors.Errorf("Couldn't find a matching subscription for %v", o)
}

func (c *EventContract) makeEvent(subscriptions []SubscriptionRequest) (Event, error) {
	res := Event{
		Body: c,
	}
	for i := range subscriptions {
		if c.Hash == subscriptions[i].Hash || c.ProjectID == subscriptions[i].ProjectID {
			res.Network = subscriptions[i].Network
			res.Address = subscriptions[i].Address
			res.Alias = subscriptions[i].Alias

			if c.Hash == subscriptions[i].Hash {
				res.Type = EventTypeSame
			} else {
				res.Type = EventTypeSimilar
			}
			return res, nil
		}
	}
	return Event{}, errors.Errorf("Couldn't find a matching subscription for %v", c)
}

func parseEvent(subscriptions []SubscriptionRequest, hit gjson.Result) (Event, error) {
	index := hit.Get("_index").String()
	switch index {
	case DocOperations:
		var event EventOperation
		event.ParseElasticJSON(hit)
		return event.makeEvent(subscriptions)
	case DocMigrations:
		var event EventMigration
		event.ParseElasticJSON(hit)
		return event.makeEvent(subscriptions)
	case DocContracts:
		var event EventContract
		event.ParseElasticJSON(hit)
		return event.makeEvent(subscriptions)
	default:
		return Event{}, errors.Errorf("[parseEvent] Invalid reponse type: %s", index)
	}
}

func getEventsQuery(subscription SubscriptionRequest, indices map[string]struct{}) []qItem {
	shouldItems := make([]qItem, 0)

	if item := getEventsWatchCalls(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}
	if item := getEventsWatchErrors(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}
	if item := getEventsWatchDeployments(subscription); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}

	if strings.HasPrefix(subscription.Address, "KT") {
		if item := getEventsWatchMigrations(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[DocMigrations] = struct{}{}
		}
		if item := getSubscriptionWithSame(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[DocContracts] = struct{}{}
		}
		if item := getSubscriptionWithSimilar(subscription); item != nil {
			shouldItems = append(shouldItems, item)
			indices[DocContracts] = struct{}{}
		}
	}

	return shouldItems
}

func getEventsWatchMigrations(subscription SubscriptionRequest) qItem {
	if !subscription.WithMigrations {
		return nil
	}

	return boolQ(
		filter(
			in("kind.keyword", []string{consts.MigrationBootstrap, consts.MigrationLambda, consts.MigrationUpdate}),
			term("network.keyword", subscription.Network),
			term("address.keyword", subscription.Address),
		),
	)
}

func getEventsWatchDeployments(subscription SubscriptionRequest) qItem {
	if !subscription.WithDeployments {
		return nil
	}

	return boolQ(
		filter(
			term("kind.keyword", "origination"),
			term("network.keyword", subscription.Network),
			term("source.keyword", subscription.Address),
		),
	)
}

func getEventsWatchCalls(subscription SubscriptionRequest) qItem {
	if !subscription.WithCalls {
		return nil
	}

	addressKeyword := "destination.keyword"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source.keyword"
	}

	return boolQ(
		filter(
			term("kind.keyword", "transaction"),
			term("status.keyword", "applied"),
			term("network.keyword", subscription.Network),
			term(addressKeyword, subscription.Address),
		),
	)
}

func getEventsWatchErrors(subscription SubscriptionRequest) qItem {
	if !subscription.WithErrors {
		return nil
	}

	addressKeyword := "destination.keyword"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source.keyword"
	}

	return boolQ(
		filter(
			term("network.keyword", subscription.Network),
			term(addressKeyword, subscription.Address),
		),
		notMust(
			term("status.keyword", "applied"),
		),
	)
}

func getSubscriptionWithSame(subscription SubscriptionRequest) qItem {
	if !subscription.WithSame {
		return nil
	}

	return boolQ(
		filter(term("hash.keyword", subscription.Hash)),
		notMust(term("address.keyword", subscription.Address)),
	)
}

func getSubscriptionWithSimilar(subscription SubscriptionRequest) qItem {
	if !subscription.WithSimilar {
		return nil
	}
	return boolQ(
		filter(term("project_id.keyword", subscription.ProjectID)),
		notMust(term("hash.keyword", subscription.Hash)),
		notMust(term("address.keyword", subscription.Address)),
	)
}
