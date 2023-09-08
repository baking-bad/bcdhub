package bigmapaction

//go:generate mockgen -source=$GOFILE -destination=../mock/bigmapaction/mock.go -package=bigmapaction -typed
type Repository interface {
	Get(ptr, limit, offset int64) ([]BigMapAction, error)
}
