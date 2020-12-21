package bigmapdiff

// GetContext -
type GetContext struct {
	Network string
	Ptr     *int64
	Query   string
	Size    int64
	Offset  int64
	Level   *int64

	To int64
}
