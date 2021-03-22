package bigmapdiff

// Bucket -
type Bucket struct {
	BigMapDiff

	KeysCount int64
}

// Stats -
type Stats struct {
	Total    int64
	Active   int64
	Contract string
}
