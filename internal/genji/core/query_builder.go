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

// Lt -
type Lt struct {
	cond strings.Builder
}

// NewLt -
func NewLt(field string, value interface{}) *Lt {
	lt := &Lt{}
	lt.cond.WriteString(fmt.Sprintf("%s <= %v", field, value))
	return lt
}

// String -
func (lt *Lt) String() string {
	return lt.cond.String()
}

// Lte -
type Lte struct {
	cond strings.Builder
}

// NewLte -
func NewLte(field string, value interface{}) *Lte {
	lte := &Lte{}
	lte.cond.WriteString(fmt.Sprintf("%s <= %v", field, value))
	return lte
}

// String -
func (lte *Lte) String() string {
	return lte.cond.String()
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

// Gte -
type Gte struct {
	cond strings.Builder
}

// NewGte -
func NewGte(field string, value interface{}) *Gte {
	gte := &Gte{}
	gte.cond.WriteString(fmt.Sprintf("%s > %v", field, value))
	return gte
}

// String -
func (gte *Gte) String() string {
	return gte.cond.String()
}

// IsNotNull -
type IsNotNull struct {
	cond strings.Builder
}

// NewIsNotNull -
func NewIsNotNull(field string) *IsNotNull {
	inn := &IsNotNull{}
	inn.cond.WriteString(fmt.Sprintf("%s IS NOT NULL", field))
	return inn
}

// String -
func (inn *IsNotNull) String() string {
	return inn.cond.String()
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
		if and.cond.Len() > 1 {
			and.cond.WriteString(" AND ")
		}
		and.cond.WriteString(conditions[i].String())
	}
	and.cond.WriteByte(')')
	return and
}

// NewAndFromMap -
func NewAndFromMap(conditions map[string]interface{}) *And {
	and := &And{}
	and.cond.WriteByte('(')
	for field, value := range conditions {
		if and.cond.Len() > 1 {
			and.cond.WriteString(" AND ")
		}
		and.cond.WriteString(NewEq(field, value).String())
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
		if or.cond.Len() > 1 {
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

// In -
type In struct {
	cond strings.Builder
}

// NewIn -
func NewIn(field string, values ...string) *In {
	in := &In{}
	in.cond.WriteString(fmt.Sprintf("%s IN (", field))
	in.cond.WriteString(strings.Join(values, ", "))
	in.cond.WriteByte(')')
	return in
}

// String -
func (in *In) String() string {
	return in.cond.String()
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
func (b *Builder) Select(table string, fields ...string) *Builder {
	b.command.WriteString(fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), table))
	return b
}

// SelectAll -
func (b *Builder) SelectAll(table string) *Builder {
	b.Select(table, "*")
	return b
}

// Insert -
func (b *Builder) Insert(table string) *Builder {
	b.command.WriteString(fmt.Sprintf("INSERT INTO %s", table))
	return b
}

// Count -
func (b *Builder) Count(table string) *Builder {
	b.command.WriteString(fmt.Sprintf("SELECT COUNT(id) FROM %s", table))
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
		if b.where.Len() > 1 {
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
		if b.where.Len() > 1 {
			b.where.WriteString(" OR ")
		}
		b.where.WriteString(conditions[i].String())
	}
	b.where.WriteByte(')')
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

// Limit -
func (b *Builder) Limit(val int64) *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString(fmt.Sprintf("LIMIT %d", val))
	return b
}

// Offset -
func (b *Builder) Offset(val int64) *Builder {
	if b.additional.Len() > 0 {
		b.additional.WriteByte(' ')
	}
	b.additional.WriteString(fmt.Sprintf("OFFSET %d", val))
	return b
}

// End -
func (b *Builder) End() *Builder {
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

// String -
func (b *Builder) String() string {
	return b.query.String()
}
