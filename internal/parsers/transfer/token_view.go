package transfer

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// TokenKey -
type TokenKey struct {
	Address    string
	Network    string
	Entrypoint string
}

// TokenEvents -
type TokenEvents map[TokenKey]tzip.EventImplementation

// NewTokenViews -
func NewTokenViews(es elastic.IElastic) (TokenEvents, error) {
	views := make(TokenEvents)
	tokens, err := es.GetTZIPWithViews()
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			return views, nil
		}
		return nil, err
	}

	for _, token := range tokens {
		if len(token.Events) == 0 {
			continue
		}

		for _, view := range token.Events {
			for _, implementation := range view.Implementations {
				for _, entrypoint := range implementation.MichelsonParameterEvent.Entrypoints {
					views[TokenKey{
						Address:    token.Address,
						Network:    token.Network,
						Entrypoint: entrypoint,
					}] = implementation
				}
			}
		}
	}

	return views, nil
}

// Get -
func (events TokenEvents) Get(address, network, entrypoint string) (tzip.EventImplementation, bool) {
	view, ok := events[TokenKey{
		Address:    address,
		Network:    network,
		Entrypoint: entrypoint,
	}]
	return view, ok
}

// GetByOperation -
func (events TokenEvents) GetByOperation(operation models.Operation) (tzip.EventImplementation, bool) {
	event, ok := events[TokenKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
	}]
	return event, ok
}
