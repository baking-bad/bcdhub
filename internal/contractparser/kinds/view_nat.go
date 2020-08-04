package kinds

// ViewNatName - name of tag
const ViewNatName = "view_nat"

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

// IsRoot -
func (v ViewNat) IsRoot() bool {
	return true
}
