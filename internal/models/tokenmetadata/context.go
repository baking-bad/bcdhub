package tokenmetadata

// GetContext -
type GetContext struct {
	Contract string
	Network  string
	TokenID  int64
	MaxLevel int64
	MinLevel int64
}
