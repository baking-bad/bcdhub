package transfer

// GetContext -
type GetContext struct {
	Contracts []string
	Network   string
	Address   string
	Hash      string
	Start     uint
	End       uint
	SortOrder string
	LastID    string
	Size      int64
	Offset    int64
	TokenID   int64
}
