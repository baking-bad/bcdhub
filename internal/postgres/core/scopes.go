package core

import (
	"time"

	"github.com/go-pg/pg/v10/orm"
)

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

// OrderByLevelDesc -
func OrderByLevelDesc(db *orm.Query) *orm.Query {
	return db.Order("level desc")
}

// IsApplied -
func IsApplied(db *orm.Query) *orm.Query {
	return db.Where("status = 1")
}

// Token -
func Token(contract string, tokenID uint64) func(db *orm.Query) *orm.Query {
	return func(db *orm.Query) *orm.Query {
		return db.Where("contract = ?", contract).
			Where("token_id = ?", tokenID)
	}
}

// EmptyRelation -
var EmptyRelation = func(q *orm.Query) (*orm.Query, error) {
	return q, nil
}

// TimestampFilter -
type TimestampFilter struct {
	Gt  time.Time
	Gte time.Time
	Lt  time.Time
	Lte time.Time
}

// Apply -
func (tf TimestampFilter) Apply(q *orm.Query) *orm.Query {
	if q == nil {
		return q
	}

	if !tf.Gt.IsZero() {
		q = q.Where("timestamp > ?", tf.Gt)
	}
	if !tf.Gte.IsZero() {
		q = q.Where("timestamp >- ?", tf.Gte)
	}
	if !tf.Lt.IsZero() {
		q = q.Where("timestamp < ?", tf.Lt)
	}
	if !tf.Lte.IsZero() {
		q = q.Where("timestamp <= ?", tf.Lte)
	}

	return q
}
