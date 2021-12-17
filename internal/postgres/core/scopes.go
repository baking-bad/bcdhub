package core

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10/orm"
)

// Network -
func Network(network types.Network) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("network = ?", network)
	}
}

// Address -
func Address(address string) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("address = ?", address)
	}
}

// Contract -
func Contract(address string) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("contract = ?", address)
	}
}

// NetworkAndAddress -
func NetworkAndAddress(network types.Network, address string) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("network = ?", network).Where("address = ?", address)
	}
}

// NetworkAndContract -
func NetworkAndContract(network types.Network, address string) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("network = ?", network).Where("contract = ?", address)
	}
}

// OrderByLevelDesc -
func OrderByLevelDesc(db *orm.Query) *orm.Query {
	return db.Order("level desc")
}

// IsApplied -
func IsApplied(db *orm.Query) *orm.Query {
	return db.Where("status = 1")
}

// Token -
func Token(network types.Network, contract string, tokenID uint64) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("network = ?", network).
			Where("contract = ?", contract).
			Where("token_id = ?", tokenID)
	}
}

// EmptyRelation -
var EmptyRelation = func(q *orm.Query) (*orm.Query, error) {
	return q, nil
}
