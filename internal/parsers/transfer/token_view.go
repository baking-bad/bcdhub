package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
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
type TokenEvents map[ImplementationKey]tzip.EventImplementation

var tokens []tzip.TZIP

// NewTokenEvents -
func NewTokenEvents(repo tzip.Repository, storage models.GeneralRepository) (TokenEvents, error) {
	tokenEvents := make(TokenEvents)

	lastID, err := repo.GetLastIDWithEvents()
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 || tokens[0].ID != lastID {
		tokens, err = repo.GetWithEvents()
		if err != nil {
			if storage.IsRecordNotFound(err) {
				return tokenEvents, nil
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
				if implementation.MichelsonParameterEvent != nil {
					for _, entrypoint := range implementation.MichelsonParameterEvent.Entrypoints {
						tokenEvents[ImplementationKey{
							Address:    token.Address,
							Network:    token.Network,
							Entrypoint: entrypoint,
							Name:       events.NormalizeName(event.Name),
						}] = implementation

					}
				}

				if implementation.MichelsonExtendedStorageEvent != nil {
					for _, entrypoint := range implementation.MichelsonExtendedStorageEvent.Entrypoints {
						tokenEvents[ImplementationKey{
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

	return tokenEvents, nil
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
