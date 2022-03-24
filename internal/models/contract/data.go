package contract

// SameResponse -
type SameResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}
