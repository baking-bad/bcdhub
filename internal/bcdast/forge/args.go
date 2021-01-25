package forge

type args struct {
	Args []*Node

	count int
}

func newArgs(count int) *args {
	return &args{
		Args: make([]*Node, 0),

		count: count,
	}
}

func newArgsFromNodes(nodes []*Node) *args {
	return &args{
		Args:  nodes,
		count: len(nodes),
	}
}

// Unforge -
func (a *args) Unforge(data []byte) (int, error) {
	var length int
	for i := 0; i < a.count; i++ {
		unforger := NewMichelson()

		n, err := unforger.Unforge(data)
		if err != nil {
			return n, err
		}
		length += n

		data = data[n:]
		a.Args = append(a.Args, unforger.Nodes...)
	}
	return length, nil
}

// Forge -
func (a *args) Forge() ([]byte, error) {
	forger := NewMichelson()
	forger.Nodes = a.Args
	return forger.Forge()
}
