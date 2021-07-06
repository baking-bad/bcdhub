package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// View -
type View struct {
	*TimeBased
	db   *gorm.DB
	name string
}

// NewViews -
func NewView(db *gorm.DB, name string, period time.Duration) *View {
	v := &View{
		name: name,
		db:   db,
	}
	v.TimeBased = NewTimeBased(v.refresh, period)
	return v
}

func (v *View) refresh() error {
	return v.db.Transaction(func(tx *gorm.DB) error {
		sql := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s;", v.name)
		return tx.Exec(sql).Error
	})
}
