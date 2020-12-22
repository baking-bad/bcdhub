package core

import (
	"fmt"
	"strings"
)

// sizes
const (
	MaxQuerySize = 10000
	MinQuerySize = 0
)

// Eq -
type Eq struct {
	cond strings.Builder
}

// NewEq -
func NewEq(field string, value interface{}) *Eq {
	eq := &Eq{}
	eq.cond.WriteString(fmt.Sprintf("%s = %v", field, value))
	return eq
}

// String -
func (eq *Eq) String() string {
	return eq.cond.String()
}

// Gt -
type Gt struct {
	cond strings.Builder
}

// NewGt -
func NewGt(field string, value interface{}) *Gt {
	gt := &Gt{}
	gt.cond.WriteString(fmt.Sprintf("%s > %v", field, value))
	return gt
}

// String -
func (gt *Gt) String() string {
	return gt.cond.String()
}

// And -
type And struct {
	cond strings.Builder
}

// NewAnd -
func NewAnd(conditions ...fmt.Stringer) *And {
	and := &And{}
	and.cond.WriteByte('(')
	for i := range conditions {
		if and.cond.Len() > 0 {
			and.cond.WriteString(" AND ")
		}
		and.cond.WriteString(conditions[i].String())
	}
	and.cond.WriteByte(')')
	return and
}

// String -
func (and *And) String() string {
	return and.cond.String()
}

// Or -
type Or struct {
	cond strings.Builder
}

// NewOr -
func NewOr(conditions ...fmt.Stringer) *Or {
	or := &Or{}
	or.cond.WriteByte('(')
	for i := range conditions {
		if or.cond.Len() > 0 {
			or.cond.WriteString(" OR ")
		}
		or.cond.WriteString(conditions[i].String())
	}
	or.cond.WriteByte(')')
	return or
}

// String -
func (or *Or) String() string {
	return or.cond.String()
}

// Builder -
type Builder struct {
	query      strings.Builder
	command    strings.Builder
	additional strings.Builder
	where      strings.Builder
}

// NewBuilder -
func NewBuilder() *Builder {
	return &Builder{}
}

// Select -
func (b *Builder) Select(table, fields string) *Builder {
	b.command.WriteString(fmt.Sprintf("SELECT %s FROM %s", fields, table))
	return b
}

// SelectAll -
func (b *Builder) SelectAll(table string) *Builder {
	b.Select(table, "*")
	return b
}

// Delete -
func (b *Builder) Delete(table string) *Builder {
	b.command.WriteString(fmt.Sprintf("DELETE FROM %s", table))
	return b
}

// Drop -
func (b *Builder) Drop(table string) *Builder {
	b.command.WriteString(fmt.Sprintf("DROP TABLE %s", table))
	return b
}

// And -
func (b *Builder) And(conditions ...fmt.Stringer) *Builder {
	b.where.WriteByte('(')
	for i := range conditions {
		if b.where.Len() > 0 {
			b.where.WriteString(" AND ")
		}
		b.where.WriteString(conditions[i].String())
	}
	b.where.WriteByte(')')
	return b
}

// Or -
func (b *Builder) Or(conditions ...fmt.Stringer) *Builder {
	b.where.WriteByte('(')
	for i := range conditions {
		if b.where.Len() > 0 {
			b.where.WriteString(" OR ")
		}
		b.where.WriteString(conditions[i].String())
	}
	b.where.WriteByte(')')
	return b
}

// Next -
func (b *Builder) Next() *Builder {
	b.query.WriteString(b.command.String())
	if b.where.Len() > 0 {
		b.query.WriteString(" WHERE ")
		b.query.WriteString(b.where.String())
	}
	if b.additional.Len() > 0 {
		b.query.WriteByte(' ')
		b.query.WriteString(b.additional.String())
	}
	b.query.WriteByte(';')

	b.where.Reset()
	b.command.Reset()
	b.additional.Reset()

	return b
}

// GroupBy -
func (b *Builder) GroupBy(field string) *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString(fmt.Sprintf("GROUP BY %s", field))
	return b
}

// SortAsc -
func (b *Builder) SortAsc(field string) *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString(fmt.Sprintf("ORDER BY %s ASC", field))
	return b
}

// SortDesc -
func (b *Builder) SortDesc(field string) *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString(fmt.Sprintf("ORDER BY %s DESC", field))
	return b
}

// One -
func (b *Builder) One() *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString("TOP 1")
	return b
}

// String -
func (b *Builder) String() string {
	return b.query.String()
}
