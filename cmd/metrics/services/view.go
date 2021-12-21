package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
)

// View -
type View struct {
	*TimeBased
	db   pg.DBI
	name string
}

// NewViews -
func NewView(db pg.DBI, name string, period time.Duration) *View {
	v := &View{
		name: name,
		db:   db,
	}
	v.TimeBased = NewTimeBased(v.refresh, period)
	return v
}

func (v *View) refresh(ctx context.Context) error {
	return v.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		sql := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s;", v.name)
		_, err := tx.Exec(sql)
		return err
	})
}
