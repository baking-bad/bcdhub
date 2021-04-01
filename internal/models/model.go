package models

import (
	"github.com/baking-bad/bcdhub/internal/mq"
	"gorm.io/gorm"
)

// Model -
type Model interface {
	mq.IMessage

	GetID() int64
	GetIndex() string
	Save(tx *gorm.DB) error
}
