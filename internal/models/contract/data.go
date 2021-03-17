package contract

// SameResponse -
type SameResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}

// Similar -
type Similar struct {
	*Contract
	Count int64 `json:"count"`
}

// DiffTask -
type DiffTask struct {
	Network1 string
	Address1 string
	Network2 string
	Address2 string
}

// Address -
type Address struct {
	Address string
	Network string
}

// Stats -
type Stats struct {
	SameCount    int64
	SimilarCount int64
}
