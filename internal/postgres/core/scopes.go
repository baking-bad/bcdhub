package core

import (
	"time"

	"github.com/uptrace/bun"
)

// Address -
func Address(query *bun.SelectQuery, address string) *bun.SelectQuery {
	return query.Where("address = ?", address)
}

// Contract -
func Contract(query *bun.SelectQuery, address string) *bun.SelectQuery {
	return query.Where("contract = ?", address)
}

// OrderByLevelDesc -
func OrderByLevelDesc(db *bun.SelectQuery) *bun.SelectQuery {
	return db.Order("level desc")
}

// IsApplied -
func IsApplied(db *bun.SelectQuery) *bun.SelectQuery {
	return db.Where("status = 1")
}

// Token -
func Token(query *bun.SelectQuery, contract string, tokenID uint64) *bun.SelectQuery {
	return query.Where("contract = ?", contract).Where("token_id = ?", tokenID)
}

// EmptyRelation -
var EmptyRelation = func(q *bun.Query) (*bun.Query, error) {
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
func (tf TimestampFilter) Apply(q *bun.SelectQuery) *bun.SelectQuery {
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
