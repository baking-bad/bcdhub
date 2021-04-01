package core

import (
	"sort"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetEvents -
func (p *Postgres) GetEvents(subscriptions []models.SubscriptionRequest, size, offset int64) ([]models.Event, error) {
	if len(subscriptions) == 0 {
		return []models.Event{}, nil
	}

	queries := make(map[string]*gorm.DB)

	for i := range subscriptions {
		subQueries := make(map[string]*gorm.DB)
		getEventsQuery(p, subscriptions[i], subQueries)

		for table, sq := range subQueries {
			makePartOfQuery(p, table, sq, queries)
		}
	}

	events := make([]models.Event, 0)
	limit := p.GetPageSize(size)
	for table, q := range queries {
		q.Limit(limit).Offset(int(offset)).Order("timestamp desc")

		switch table {
		case models.DocContracts:
			var contracts []EventContract
			if err := q.Find(&contracts).Error; err != nil {
				return nil, err
			}

			for i := range contracts {
				e, err := contracts[i].makeEvent(subscriptions)
				if err != nil {
					return nil, err
				}
				events = append(events, e)
			}
		case models.DocOperations:
			var ops []EventOperation
			if err := q.Find(&ops).Error; err != nil {
				return nil, err
			}

			for i := range ops {
				e, err := ops[i].makeEvent(subscriptions)
				if err != nil {
					return nil, err
				}
				events = append(events, e)
			}
		case models.DocMigrations:
			var migrations []EventMigration
			if err := q.Find(&migrations).Error; err != nil {
				return nil, err
			}

			for i := range migrations {
				e, err := migrations[i].makeEvent(subscriptions)
				if err != nil {
					return nil, err
				}
				events = append(events, e)
			}
		}
	}

	sort.Sort(models.ByTimestamp(events))

	return events, nil
}

func (m *EventMigration) makeEvent(subscriptions []models.SubscriptionRequest) (models.Event, error) {
	res := models.Event{
		Type:      models.EventTypeMigration,
		Address:   m.Address,
		Network:   m.Network,
		Timestamp: m.Timestamp,
		Body:      m,
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
		Network:   o.Network,
		Timestamp: o.Timestamp,
		Body:      o,
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
		Body:      c,
		Timestamp: c.Timestamp,
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

func getEventsQuery(p *Postgres, subscription models.SubscriptionRequest, queries map[string]*gorm.DB) {

	if item := getEventsWatchCalls(p, subscription); item != nil {
		makePartOfQuery(p, models.DocOperations, item, queries)
	}
	if item := getEventsWatchErrors(p, subscription); item != nil {
		makePartOfQuery(p, models.DocOperations, item, queries)
	}
	if item := getEventsWatchDeployments(p, subscription); item != nil {
		makePartOfQuery(p, models.DocOperations, item, queries)
	}

	if bcd.IsContract(subscription.Address) {
		if item := getEventsWatchMigrations(p, subscription); item != nil {
			makePartOfQuery(p, models.DocMigrations, item, queries)
		}
		if item := getSubscriptionWithSame(p, subscription); item != nil {
			makePartOfQuery(p, models.DocContracts, item, queries)
		}
		if item := getSubscriptionWithSimilar(p, subscription); item != nil {
			makePartOfQuery(p, models.DocContracts, item, queries)
		}
	}
}

func makePartOfQuery(p *Postgres, table string, item *gorm.DB, queries map[string]*gorm.DB) {
	if q, ok := queries[table]; ok {
		q.Or(item)
	} else {
		queries[table] = p.DB.Table(table).Where(item)
	}
}

func getEventsWatchMigrations(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithMigrations {
		return nil
	}

	return p.DB.Where("kind IN (?)", []string{consts.MigrationBootstrap, consts.MigrationLambda, consts.MigrationUpdate}).
		Where("network = ?", subscription.Network).
		Where("address = ?", subscription.Address)
}

func getEventsWatchDeployments(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithDeployments {
		return nil
	}

	return p.DB.Where("kind = ?", consts.Origination).
		Where("network = ?", subscription.Network).
		Where("source = ?", subscription.Address)
}

func getEventsWatchCalls(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithCalls {
		return nil
	}

	addressKeyword := "destination = ?"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source = ?"
	}

	return p.DB.Where("kind = ?", consts.Transaction).
		Where("status = ?", consts.Applied).
		Where("network = ?", subscription.Network).
		Where(addressKeyword, subscription.Address)
}

func getEventsWatchErrors(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithErrors {
		return nil
	}

	addressKeyword := "destination = ?"
	if strings.HasPrefix(subscription.Address, "tz") {
		addressKeyword = "source = ?"
	}

	return p.DB.Where("status != ?", consts.Applied).
		Where("network = ?", subscription.Network).
		Where(addressKeyword, subscription.Address)
}

func getSubscriptionWithSame(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithSame {
		return nil
	}

	return p.DB.Where("hash = ?", subscription.Hash).
		Where("address != ?", subscription.Address)
}

func getSubscriptionWithSimilar(p *Postgres, subscription models.SubscriptionRequest) *gorm.DB {
	if !subscription.WithSimilar {
		return nil
	}

	return p.DB.Where("project_id = ?", subscription.ProjectID).
		Where("hash != ?", subscription.Hash).
		Where("address != ?", subscription.Address)
}
