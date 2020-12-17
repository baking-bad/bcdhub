package models

import "github.com/baking-bad/bcdhub/internal/mq"

// Model -
type Model interface {
	mq.IMessage

	GetID() string
	GetIndex() string
}
