package models

import "github.com/go-pg/pg/v10"

// Model -
type Model interface {
	GetID() int64
	GetIndex() string
	Save(tx pg.DBI) error
}
