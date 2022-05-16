package bigmapaction

// Repository -
type Repository interface {
	Get(ptr, limit, offset int64) ([]BigMapAction, error)
}
