package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// ImplementationKey -
type ImplementationKey struct {
	Address    string
	Network    types.Network
	Entrypoint string
	Name       string
}

// TokenEvents -
type TokenEvents struct {
	events    map[ImplementationKey]tzip.EventImplementation
	updatedAt uint64
}

// EmptyTokenEvents -
func EmptyTokenEvents() *TokenEvents {
	return &TokenEvents{
		events: make(map[ImplementationKey]tzip.EventImplementation),
	}
}

// NewTokenEvents -
func NewTokenEvents(repo tzip.Repository) (*TokenEvents, error) {
	tokenEvents := EmptyTokenEvents()
	return tokenEvents, tokenEvents.Update(repo)
}

// GetByOperation -
func (tokenEvents *TokenEvents) GetByOperation(operation operation.Operation) (tzip.EventImplementation, string, bool) {
	if event, ok := tokenEvents.events[ImplementationKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
		Name:       tokenbalance.SingleAssetBalanceUpdates,
	}]; ok {
		return event, tokenbalance.SingleAssetBalanceUpdates, ok
	}

	event, ok := tokenEvents.events[ImplementationKey{
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

// Update -
func (tokenEvents *TokenEvents) Update(repo tzip.Repository) error {
	tokens, err := repo.GetWithEvents(tokenEvents.updatedAt)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		tokenEvents.updatedAt = token.UpdatedAt
		if len(token.Events) == 0 {
			continue
		}

		for _, event := range token.Events {
			for _, implementation := range event.Implementations {
				if implementation.MichelsonParameterEvent != nil {
					for _, entrypoint := range implementation.MichelsonParameterEvent.Entrypoints {
						tokenEvents.events[ImplementationKey{
							Address:    token.Address,
							Network:    token.Network,
							Entrypoint: entrypoint,
							Name:       events.NormalizeName(event.Name),
						}] = implementation

					}
				}

				if implementation.MichelsonExtendedStorageEvent != nil {
					for _, entrypoint := range implementation.MichelsonExtendedStorageEvent.Entrypoints {
						tokenEvents.events[ImplementationKey{
							Address:    token.Address,
							Network:    token.Network,
							Entrypoint: entrypoint,
							Name:       events.NormalizeName(event.Name),
						}] = implementation
					}
				}
			}
		}
	}
	return nil
}
