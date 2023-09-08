package migration

//go:generate mockgen -source=$GOFILE -destination=../mock/migration/mock.go -package=migration -typed
type Repository interface {
	Get(contractID int64) ([]Migration, error)
}
