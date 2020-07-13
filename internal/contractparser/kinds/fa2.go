package kinds

// Fa2 -
type Fa2 struct{}

// GetName -
func (fa2 Fa2) GetName() string {
	return "fa2"
}

// GetJSON -
func (fa2 Fa2) GetJSON() string {
	return `
	[
	{
		"name": "transfer",
		"prim": "list",
		"args": [
		{
			"prim": "pair",
			"args": [
			{
				"prim": "address"
			},
			{
				"prim": "list",
				"args": [
				{
					"prim": "pair",
					"args": [
					{
						"prim": "address"
					},
					{
						"prim": "pair",
						"args": [
						{
							"prim": "nat"
						},
						{
							"prim": "nat"
						}
						]
					}
					]
				}
				]
			}
			]
		}
		]
	},
	{
		"name": "token_metadata",
		"prim": "pair",
		"args": [
		{
			"prim": "list",
			"args": [
			{
				"prim": "nat"
			}
			]
		},
		{
			"prim": "lambda",
			"args": [
			{
				"prim": "list",
				"args": [
				{
					"prim": "pair",
					"args": [
					{
						"prim": "nat"
					},
					{
						"prim": "pair",
						"args": [
						{
							"prim": "string"
						},
						{
							"prim": "pair",
							"args": [
							{
								"prim": "string"
							},
							{
								"prim": "pair",
								"args": [
								{
									"prim": "nat"
								},
								{
									"prim": "map",
									"args": [
									{
										"prim": "string"
									},
									{
										"prim": "string"
									}
									]
								}
								]
							}
							]
						}
						]
					}
					]
				}
				]
			},
			{
				"prim": "unit"
			}
			]
		}
		]
	},
	{
		"name": "update_operators",
		"prim": "list",
		"args": [
		{
			"prim": "or",
			"args": [
			{
				"prim": "pair",
				"args": [
				{
					"prim": "address"
				},
				{
					"prim": "address"
				}
				]
			},
			{
				"prim": "pair",
				"args": [
				{
					"prim": "address"
				},
				{
					"prim": "address"
				}
				]
			}
			]
		}
		]
	},
	{
		"name": "balance_of",
		"prim": "pair",
		"args": [
		{
			"prim": "list",
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
			}
			]
		},
		{
			"prim": "contract",
			"args": [
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
			]
		}
		]
	}
	]`
}
