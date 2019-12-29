package contractparser

var fa12 = []Entrypoint{
	Entrypoint{
		Name: "transfer",
		Args: []string{
			"address",
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "approve",
		Args: []string{
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "getAllowance",
		Args: []string{
			"address",
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "getBalance",
		Args: []string{
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "getTotalSupply",
		Args: []string{
			"unit",
			"nat",
		},
	},
}
