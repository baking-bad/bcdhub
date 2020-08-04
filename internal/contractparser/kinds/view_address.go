package kinds

// ViewAddressName - name of tag
const ViewAddressName = "view_address"

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
	return ViewAddressName
}

// IsRoot -
func (v ViewAddress) IsRoot() bool {
	return true
}
