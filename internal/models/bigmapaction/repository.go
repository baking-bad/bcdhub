package bigmapaction

// Repository -
type Repository interface {
	Get(ptr int64) ([]BigMapAction, error)
}
