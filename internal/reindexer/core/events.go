package core

import (
	"strings"

	constants "github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/restream/reindexer"
)

// GetEvents -
func (r *Reindexer) GetEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	if len(subscriptions) == 0 {
		return []models.Event{}, nil
	}

	if size == 0 {
		size = DefaultSize
	}

	events := make([]models.Event, 0)
	contractEvents, err := r.getContractEvents(subscriptions, size, offset)
	if err != nil {
		return nil, err
	}
	events = append(events, contractEvents...)

	operationEvents, err := r.getOperationEvents(subscriptions, size, offset)
	if err != nil {
		return nil, err
	}
	events = append(events, operationEvents...)

	migrationEvents, err := r.getMigrationEvents(subscriptions, size, offset)
	if err != nil {
		return nil, err
	}
	events = append(events, migrationEvents...)
	return events, nil
}

func (r *Reindexer) getContractEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	tx, err := r.BeginTx(models.DocContracts)
	if err != nil {
		return nil, err
	}

	events := make([]models.Event, 0)
	for _, subscription := range subscriptions {
		if !subscription.WithSame && !subscription.WithSimilar {
			continue
		}

		query := tx.Query()
		getSubscriptionWithSame(subscription, query)
		getSubscriptionWithSimilar(subscription, query)
		it := query.Limit(int(size)).Offset(int(offset)).Exec()
		defer it.Close()

		if it.Error() != nil {
			if err := tx.Rollback(); err != nil {
				return nil, err
			}
			return nil, it.Error()
		}

		for it.Next() {
			var event EventContract
			it.NextObj(&event)
			res := models.Event{
				Body:    event,
				Network: subscription.Network,
				Address: subscription.Address,
				Alias:   subscription.Alias,
			}

			if event.Hash == subscription.Hash {
				res.Type = models.EventTypeSame
			} else {
				res.Type = models.EventTypeSimilar
			}
			events = append(events, res)
		}
	}

	return events, tx.Commit()
}

func (r *Reindexer) getOperationEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	tx, err := r.BeginTx(models.DocOperations)
	if err != nil {
		return nil, err
	}

	events := make([]models.Event, 0)
	for _, subscription := range subscriptions {
		if !subscription.WithCalls && !subscription.WithErrors && !subscription.WithDeployments {
			continue
		}
		query := tx.Query()
		getEventsWatchCalls(subscription, query)
		getEventsWatchErrors(subscription, query)
		getEventsWatchDeployments(subscription, query)
		it := query.Limit(int(size)).Offset(int(offset)).Exec()
		defer it.Close()

		if it.Error() != nil {
			if err := tx.Rollback(); err != nil {
				return nil, err
			}
			return nil, it.Error()
		}

		for it.Next() {
			var event EventOperation
			it.NextObj(&event)
			res := models.Event{
				Body:    event,
				Network: subscription.Network,
				Address: subscription.Address,
			}

			switch {
			case event.Status != constants.Applied:
				res.Type = models.EventTypeError
			case event.Source == subscription.Address && event.Kind == constants.Origination:
				res.Type = models.EventTypeDeploy
			case event.Source == subscription.Address && event.Kind == constants.Transaction:
				res.Type = models.EventTypeCall
			case event.Destination == subscription.Address && event.Kind == constants.Transaction:
				res.Type = models.EventTypeInvoke
			}
			events = append(events, res)
		}
	}
	return events, tx.Commit()
}

func (r *Reindexer) getMigrationEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	tx, err := r.BeginTx(models.DocMigrations)
	if err != nil {
		return nil, err
	}
	events := make([]models.Event, 0)
	for _, subscription := range subscriptions {
		if !subscription.WithMigrations {
			continue
		}
		it := tx.Query().
			Match("kind", constants.MigrationBootstrap, constants.MigrationLambda, constants.MigrationUpdate).
			Match("network", subscription.Network).
			Match("address", subscription.Address).
			Limit(int(size)).Offset(int(offset)).Exec()
		defer it.Close()

		if it.Error() != nil {
			if err := tx.Rollback(); err != nil {
				return nil, err
			}
			return nil, it.Error()
		}

		for it.Next() {
			var event EventOperation
			it.NextObj(&event)
			events = append(events, models.Event{
				Body:    event,
				Network: subscription.Network,
				Address: subscription.Address,
				Type:    models.EventTypeMigration,
				Alias:   subscription.Alias,
			})
		}
	}
	return events, tx.Commit()
}

func getEventsWatchDeployments(subscription models.SubscriptionRequest, query *reindexer.Query) {
	if !subscription.WithDeployments {
		return
	}

	query.
		Match("kind", constants.Origination).
		Match("network", subscription.Network).
		Match("source", subscription.Address)
}

func getEventsWatchCalls(subscription models.SubscriptionRequest, query *reindexer.Query) {
	if !subscription.WithCalls {
		return
	}

	addressKeyword := "destination"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source"
	}
	query.
		Match("kind", constants.Transaction).
		Match("network", subscription.Network).
		Match(addressKeyword, subscription.Address).
		Match("status", constants.Applied)
}

func getEventsWatchErrors(subscription models.SubscriptionRequest, query *reindexer.Query) {
	if !subscription.WithErrors {
		return
	}

	addressKeyword := "destination"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source"
	}

	query.
		Match("network", subscription.Network).
		Match(addressKeyword, subscription.Address).
		Match("status", constants.Applied)
}

func getSubscriptionWithSame(subscription models.SubscriptionRequest, query *reindexer.Query) {
	if !subscription.WithSame {
		return
	}

	query.
		Match("hash", subscription.Hash).
		Not().
		Match("address", subscription.Address)
}

func getSubscriptionWithSimilar(subscription models.SubscriptionRequest, query *reindexer.Query) {
	if !subscription.WithSimilar {
		return
	}

	query.
		Match("project_id", subscription.ProjectID).
		Not().
		Match("hash", subscription.Hash).
		Not().
		Match("address", subscription.Address)
}
