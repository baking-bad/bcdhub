package bigmapaction

// Repository -
type Repository interface {
	Get(ptr int64, network string) ([]BigMapAction, error)
}
