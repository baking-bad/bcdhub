package transfer

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// ImplementationKey -
type ImplementationKey struct {
	Address    string
	Entrypoint string
	Name       string
}

type eventMap struct {
	m  map[ImplementationKey]contract_metadata.EventImplementation
	mx sync.RWMutex
}

func newEventMap() *eventMap {
	return &eventMap{
		m: make(map[ImplementationKey]contract_metadata.EventImplementation),
	}
}

// Load -
func (m *eventMap) Load(key ImplementationKey) (contract_metadata.EventImplementation, bool) {
	m.mx.RLock()
	data, ok := m.m[key]
	m.mx.RUnlock()
	return data, ok
}

// Store -
func (m *eventMap) Store(key ImplementationKey, value contract_metadata.EventImplementation) {
	m.mx.Lock()
	m.m[key] = value
	m.mx.Unlock()
}

// TokenEvents -
type TokenEvents struct {
	events    *eventMap
	updatedAt uint64
}

// EmptyTokenEvents -
func EmptyTokenEvents() *TokenEvents {
	return &TokenEvents{
		events: newEventMap(),
	}
}

// NewTokenEvents -
func NewTokenEvents(repo contract_metadata.Repository) (*TokenEvents, error) {
	tokenEvents := EmptyTokenEvents()
	return tokenEvents, tokenEvents.Update(repo)
}

// GetByOperation -
func (tokenEvents *TokenEvents) GetByOperation(operation operation.Operation) (contract_metadata.EventImplementation, string, bool) {
	if event, ok := tokenEvents.events.Load(ImplementationKey{
		Address:    operation.Destination.Address,
		Entrypoint: operation.Entrypoint.String(),
		Name:       tokenbalance.SingleAssetBalanceUpdates,
	}); ok {
		return event, tokenbalance.SingleAssetBalanceUpdates, ok
	}

	event, ok := tokenEvents.events.Load(ImplementationKey{
		Address:    operation.Destination.Address,
		Entrypoint: operation.Entrypoint.String(),
		Name:       tokenbalance.MultiAssetBalanceUpdates,
	})
	if ok {
		return event, tokenbalance.MultiAssetBalanceUpdates, ok
	}
	return event, "", ok
}

// Update -
func (tokenEvents *TokenEvents) Update(repo contract_metadata.Repository) error {
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
						tokenEvents.events.Store(ImplementationKey{
							Address:    token.Address,
							Entrypoint: entrypoint,
							Name:       events.NormalizeName(event.Name),
						}, implementation)

					}
				}

				if implementation.MichelsonExtendedStorageEvent != nil {
					for _, entrypoint := range implementation.MichelsonExtendedStorageEvent.Entrypoints {
						tokenEvents.events.Store(ImplementationKey{
							Address:    token.Address,
							Entrypoint: entrypoint,
							Name:       events.NormalizeName(event.Name),
						}, implementation)
					}
				}
			}
		}
	}
	return nil
}
