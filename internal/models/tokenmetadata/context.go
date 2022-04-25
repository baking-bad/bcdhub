package tokenmetadata

// GetContext -
type GetContext struct {
	Contract string
	TokenID  *uint64
	MaxLevel int64
	MinLevel int64
	Creator  string
	Name     string
}
