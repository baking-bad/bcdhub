package kinds

// FA1Name - name of tag
const FA1Name = "fa1"

// Fa1 -
type Fa1 struct{}

// GetJSON -
func (fa1 Fa1) GetJSON() string {
	return `
	[
		{
			"name": "getBalance",
			"prim": "pair",
			"args": [
				{
					"prim": "address"
				},
				{
					"parameter": {
						"prim": "nat"
					},
					"prim": "contract"
				}
			]
		},
		{
			"name": "getTotalSupply",
			"prim": "pair",
			"args": [
				{
					"prim": "unit"
				},
				{
					"parameter": {
						"prim": "nat"
					},
					"prim": "contract"
				}
			]
		},
		{
			"name": "transfer",
			"prim": "pair",
			"args": [
				{
					"prim": "address"
				},
				{
					"args": [
						{
							"prim": "address"
						},
						{
							"prim": "nat"
						}
					],
					"prim": "pair"
				}
			]
		}
	]`
}

// GetName -
func (fa1 Fa1) GetName() string {
	return FA1Name
}

// IsRoot -
func (fa1 Fa1) IsRoot() bool {
	return false
}
