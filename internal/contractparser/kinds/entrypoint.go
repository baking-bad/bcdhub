package kinds

// Entrypoint -
type Entrypoint struct {
	Name        string       `json:"name"`
	Prim        string       `json:"prim"`
	Args        []Entrypoint `json:"args,omitempty"`
	Parameter   interface{}  `json:"parameter"`
	ReturnValue interface{}  `json:"return_value,omitempty"`
}
