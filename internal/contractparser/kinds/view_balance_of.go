package kinds

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
