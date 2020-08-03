package kinds

// ViewAddress -
type ViewAddress struct{}

// GetJSON -
func (v ViewAddress) GetJSON() string {
	return `[
		{
			"prim": "address"
		}
	]`
}

// GetName -
func (v ViewAddress) GetName() string {
	return "view_address"
}
