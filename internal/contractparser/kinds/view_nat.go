package kinds

// ViewNat -
type ViewNat struct{}

// GetJSON -
func (v ViewNat) GetJSON() string {
	return `[
		{
			"prim": "nat"
		}
	]`
}

// GetName -
func (v ViewNat) GetName() string {
	return "view_nat"
}
