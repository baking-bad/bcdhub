package models

//go:generate mockgen -source=$GOFILE -destination=mock/model.go -package=mock -typed
type Model interface {
	GetID() int64
	TableName() string
}
