package bigmapdiff

// Bucket -
type Bucket struct {
	BigMapDiff

	Count int64
}

// Stats -
type Stats struct {
	Total    int64
	Active   int64
	Address  string
	Protocol string
}
