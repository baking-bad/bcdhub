package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
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

// NewTokenEvents -
func NewTokenEvents(repo tzip.Repository, storage models.GeneralRepository) (TokenEvents, error) {
	views := make(TokenEvents)
	tokens, err := repo.GetWithEvents()
	if err != nil {
		if storage.IsRecordNotFound(err) {
			return views, nil
		}
		return nil, err
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

// NewInitialStorageEvents -
func NewInitialStorageEvents(repo tzip.Repository, storage models.GeneralRepository) (TokenEvents, error) {
	views := make(TokenEvents)
	tokens, err := repo.GetWithEvents()
	if err != nil {
		if storage.IsRecordNotFound(err) {
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
				if !implementation.MichelsonInitialStorageEvent.Empty() {
					views[ImplementationKey{
						Address: token.Address,
						Network: token.Network,
						Name:    events.NormalizeName(view.Name),
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
		Name:       events.SingleAssetBalanceUpdates,
	}]; ok {
		return event, events.SingleAssetBalanceUpdates, ok
	}

	event, ok := tokenEvents[ImplementationKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
		Name:       events.MultiAssetBalanceUpdates,
	}]
	if ok {
		return event, events.MultiAssetBalanceUpdates, ok
	}
	return event, "", ok
}
