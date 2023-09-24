package models

// Model -
type Model interface {
	GetID() int64
	GetIndex() string
	PartitionBy() string
}
