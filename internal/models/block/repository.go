package block

//go:generate mockgen -source=$GOFILE -destination=../mock/block/mock.go -package=block -typed
type Repository interface {
	Get(level int64) (Block, error)
	Last() (Block, error)
}
