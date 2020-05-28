package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// GetEvents -
func (e *Elastic) GetEvents(subscriptions []SubscriptionRequest, size, offset int64) ([]Event, error) {
	if len(subscriptions) == 0 {
		return []Event{}, nil
	}

	if size == 0 || size > 50 {
		size = defaultSize
	}

	shouldItems := make([]qItem, 0)
	indicesMap := make(map[string]struct{})

	for i := range subscriptions {
		contract, err := e.GetContract(map[string]interface{}{
			"address": subscriptions[i].Address,
			"network": subscriptions[i].Network,
		})
		eventContracts, err := e.getRelatedContracts(subscriptions[i], contract)
		if err != nil {
			return nil, err
		}
		items := getEventsQuery(subscriptions[i], eventContracts, indicesMap)
		if err != nil {
			return nil, err
		}
		shouldItems = append(shouldItems, items...)
	}

	indices := make([]string, 0)
	for ind := range indicesMap {
		indices = append(indices, ind)
	}
	if len(indices) == 0 {
		return []Event{}, nil
	}

	return e.getEvents(shouldItems, indices, size, offset)
}

func (e *Elastic) getEvents(shouldItems []qItem, indices []string, size, offset int64) ([]Event, error) {
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
		event, err := parseEvent(hit)
		if err != nil {
			return nil, err
		}
		events[i] = event
	}

	return events, nil
}

func parseEvent(hit gjson.Result) (Event, error) {
	index := hit.Get("_index").String()
	switch index {
	case DocOperations:
		var event EventOperation
		event.ParseElasticJSON(hit)
		return Event{
			Index: index,
			Body:  event,
		}, nil
	case DocMigrations:
		var event EventMigration
		event.ParseElasticJSON(hit)
		return Event{
			Index: index,
			Body:  event,
		}, nil
	default:
		return Event{}, fmt.Errorf("[parseEvent] Invalid reponse type: %s", index)
	}
}

func getEventsQuery(subscription SubscriptionRequest, contracts []EventContract, indices map[string]struct{}) []qItem {
	shouldItems := make([]qItem, 0)

	if item := getEventsWatchMigrations(subscription, contracts); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocMigrations] = struct{}{}
	}
	if item := getEventsWatchDeployments(subscription, contracts); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}
	if item := getEventsWatchCalls(subscription, contracts); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}
	if item := getEventsWatchErrors(subscription, contracts); item != nil {
		shouldItems = append(shouldItems, item)
		indices[DocOperations] = struct{}{}
	}

	return shouldItems
}

func getEventsWatchMigrations(subscription SubscriptionRequest, contracts []EventContract) qItem {
	if !subscription.WithMigrations {
		return nil
	}

	items := make([]qItem, len(contracts))
	for i := range contracts {
		items[i] = boolQ(
			filter(
				term("network.keyword", contracts[i].Network),
				term("address.keyword", contracts[i].Address),
			),
		)
	}

	return boolQ(
		should(items...),
		minimumShouldMatch(1),
	)
}

func getEventsWatchDeployments(subscription SubscriptionRequest, contracts []EventContract) qItem {
	if !subscription.WithDeployments {
		return nil
	}

	items := make([]qItem, len(contracts))
	for i := range contracts {
		items[i] = boolQ(
			filter(
				term("kind.keyword", "origination"),
				term("network.keyword", contracts[i].Network),
				term("source.keyword", contracts[i].Address),
			),
		)
	}

	return boolQ(
		should(items...),
		minimumShouldMatch(1),
	)
}

func getEventsWatchCalls(subscription SubscriptionRequest, contracts []EventContract) qItem {
	if !subscription.WithCalls {
		return nil
	}

	items := make([]qItem, len(contracts))
	for i := range contracts {
		items[i] = boolQ(
			filter(
				term("kind.keyword", "transaction"),
				term("network.keyword", contracts[i].Network),
				term("destination.keyword", contracts[i].Address),
			),
		)
	}

	return boolQ(
		should(items...),
		minimumShouldMatch(1),
	)
}

func getEventsWatchErrors(subscription SubscriptionRequest, contracts []EventContract) qItem {
	if !subscription.WithErrors {
		return nil
	}

	items := make([]qItem, len(contracts))
	for i := range contracts {
		items[i] = boolQ(
			filter(
				term("network.keyword", contracts[i].Network),
				boolQ(
					should(
						term("source.keyword", contracts[i].Address),
						term("destination.keyword", contracts[i].Address),
					),
					minimumShouldMatch(1),
				),
			),
			notMust(
				term("status.keyword", "applied"),
			),
		)
	}

	return boolQ(
		should(items...),
		minimumShouldMatch(1),
	)
}

func (e *Elastic) getRelatedContracts(subscription SubscriptionRequest, contract models.Contract) ([]EventContract, error) {
	query, err := getRelatedContractsQuery(subscription, contract)
	if err != nil {
		return nil, err
	}

	var contracts []models.Contract
	if err := e.GetAllByQuery(query, &contracts); err != nil {
		return nil, err
	}

	result := make([]EventContract, len(contracts))
	for i := range contracts {
		result[i] = EventContract{
			SubscriptionID: subscription.ID,
			Address:        contracts[i].Address,
			Network:        contracts[i].Network,
		}
	}

	return result, nil
}

func getRelatedContractsQuery(subscription SubscriptionRequest, contract models.Contract) (base, error) {
	shouldItems := []qItem{
		boolQ(
			filter(
				term("network.keyword", contract.Network),
				term("address.keyword", contract.Address),
			),
		),
	}

	// Filter subscription mask contracts
	if len(subscription.Address) < 2 {
		return nil, fmt.Errorf("Invalid subscription address: %s %s", subscription.Network, subscription.Address)
	}

	switch subscription.Address[:2] {
	case "KT":
		if item := getSubscriptionWithSame(subscription, contract); item != nil {
			shouldItems = append(shouldItems, item)
		}
		if item := getSubscriptionWithSimilar(subscription, contract); item != nil {
			shouldItems = append(shouldItems, item)
		}
	default:
		if item := getSubscriptionWithDeployed(subscription, subscription.Address); item != nil {
			shouldItems = append(shouldItems, item)
		}
	}

	return newQuery().Query(
		boolQ(
			should(
				shouldItems...,
			),
			minimumShouldMatch(1),
		),
	).All(), nil
}

func getSubscriptionWithSame(subscription SubscriptionRequest, contract models.Contract) qItem {
	if !subscription.WithSame {
		return nil
	}

	return boolQ(
		filter(term("hash.keyword", contract.Hash)),
	)
}

func getSubscriptionWithSimilar(subscription SubscriptionRequest, contract models.Contract) qItem {
	if !subscription.WithSimilar {
		return nil
	}
	return boolQ(
		filter(term("project_id.keyword", contract.ProjectID)),
		notMust(term("hash.keyword", contract.Hash)),
	)
}

func getSubscriptionWithDeployed(subscription SubscriptionRequest, address string) qItem {
	if !subscription.WithDeployed {
		return nil
	}
	return boolQ(
		filter(
			term("manager.keyword", address),
			term("network.keyword", subscription.Network),
		),
	)
}
