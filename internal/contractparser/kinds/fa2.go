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
					"parameter": {
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
			"name": "token_metadata_registry",
			"prim": "contract",
			"parameter": {
				"prim": "address"
			}
		}
	]`
}
