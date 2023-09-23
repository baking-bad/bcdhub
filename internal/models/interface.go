package models

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=mock/general.go -package=mock -typed
type GeneralRepository interface {
	InitDatabase(ctx context.Context) error
	CreateIndex(ctx context.Context, name, columns string, model any) error
	IsRecordNotFound(err error) bool

	// Save - performs insert or update items.
	Save(ctx context.Context, items []Model) error
	BulkDelete(ctx context.Context, models []Model) error

	TablesExist(ctx context.Context) bool
	// Drop - drops full database
	Drop(ctx context.Context) error
}
