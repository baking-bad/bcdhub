package kinds

// ViewBalanceOfName - name of tag
const ViewBalanceOfName = "view_balance_of"

// ViewBalanceOf -
type ViewBalanceOf struct{}

// GetJSON -
func (v ViewBalanceOf) GetJSON() string {
	return `[
		{
			"prim": "list",
			"args": [
				{
					"prim": "pair",
					"args": [
						{
							"prim": "pair",
							"args": [
								{
									"prim": "address"
								},
								{
									"prim": "nat"
								}
							]
						},
						{
							"prim": "nat"
						}
					]
				}
			]
		}
	]`
}

// GetName -
func (v ViewBalanceOf) GetName() string {
	return "view_balance_of"
}

// IsRoot -
func (v ViewBalanceOf) IsRoot() bool {
	return true
}
