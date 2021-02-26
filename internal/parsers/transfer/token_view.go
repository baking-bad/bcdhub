package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// ImplementationKey -
type ImplementationKey struct {
	Address    string
	Network    string
	Entrypoint string
	Name       string
}

// TokenEvents -
type TokenEvents map[ImplementationKey]tzip.EventImplementation

var tokens []tzip.TZIP

// NewTokenEvents -
func NewTokenEvents(repo tzip.Repository, storage models.GeneralRepository) (TokenEvents, error) {
	views := make(TokenEvents)

	count, err := repo.GetWithEventsCounts()
	if err != nil {
		return nil, err
	}
	if int(count) != len(tokens) {
		tokens, err = repo.GetWithEvents()
		if err != nil {
			if storage.IsRecordNotFound(err) {
				return views, nil
			}
			return nil, err
		}
	}

	for _, token := range tokens {
		if len(token.Events) == 0 {
			continue
		}

		for _, event := range token.Events {
			for _, implementation := range event.Implementations {
				for _, entrypoint := range implementation.MichelsonParameterEvent.Entrypoints {
					views[ImplementationKey{
						Address:    token.Address,
						Network:    token.Network,
						Entrypoint: entrypoint,
						Name:       events.NormalizeName(event.Name),
					}] = implementation

				}

				for _, entrypoint := range implementation.MichelsonExtendedStorageEvent.Entrypoints {
					views[ImplementationKey{
						Address:    token.Address,
						Network:    token.Network,
						Entrypoint: entrypoint,
						Name:       events.NormalizeName(event.Name),
					}] = implementation
				}
			}
		}
	}

	return views, nil
}

// GetByOperation -
func (tokenEvents TokenEvents) GetByOperation(operation operation.Operation) (tzip.EventImplementation, string, bool) {
	if event, ok := tokenEvents[ImplementationKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
		Name:       tokenbalance.SingleAssetBalanceUpdates,
	}]; ok {
		return event, tokenbalance.SingleAssetBalanceUpdates, ok
	}

	event, ok := tokenEvents[ImplementationKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
		Name:       tokenbalance.MultiAssetBalanceUpdates,
	}]
	if ok {
		return event, tokenbalance.MultiAssetBalanceUpdates, ok
	}
	return event, "", ok
}
