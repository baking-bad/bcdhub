package kinds

// IContractKind -
type IContractKind interface {
	GetJSON() string
	GetName() string
	IsRoot() bool
}
