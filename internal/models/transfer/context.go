package transfer

// GetContext -
type GetContext struct {
	Contracts []string
	AccountID int64

	Start       uint
	End         uint
	SortOrder   string
	LastID      string
	Size        int64
	Offset      int64
	TokenID     *uint64
	OperationID *int64
}
