package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
)

const (
	defaultEntrypoint = "entrypoint_%d"
)

type internalNode struct {
	*Node
	InternalArgs []internalNode `json:"-,omitempty"`
	Nested       bool           `json:"-"`
}

// Entrypoint -
type Entrypoint struct {
	Name string        `json:"name"`
	Prim string        `json:"prim"`
	Args []interface{} `json:"args,omitempty"`
}

// Metadata -
type Metadata struct {
	TypeName  string                 `json:"type,omitempty"`
	FieldName string                 `json:"field,omitempty"`
	Prim      string                 `json:"prim,omitempty"`
	Entry     string                 `json:"entry,omitempty"`
	Parameter map[string]interface{} `json:"parameter,omitempty"`
	Args      []string               `json:"args,omitempty"`
}

// Parameter -
type Parameter struct {
	Metadata map[string]*Metadata
	Hash     string

	fuzzyReader *HashReader

	Tags Set
}

func newParameter(v []interface{}) (Parameter, error) {
	p := Parameter{
		Metadata:    make(map[string]*Metadata),
		fuzzyReader: NewHashReader(),
		Tags:        make(Set),
	}
	err := p.parse(v)
	return p, err
}

func (p *Parameter) parse(v []interface{}) error {
	if len(v) == 1 {
		parent := Node{
			Prim: PARAMETER,
			Path: "0",
		}
		p.parseNode(v[0].(map[string]interface{}), parent, "0", "")

		h, err := tlsh.HashReader(p.fuzzyReader)
		if err != nil {
			return err
		}
		p.Hash = h.String()

		e, err := p.EntrypointStructs()
		if err != nil {
			return err
		}

		tags, err := endpointsTags(e)
		if err != nil {
			return err
		}
		p.Tags.Append(tags...)
		return nil
	}
	return fmt.Errorf("Invalid parameter length: %d", len(v))
}

func (p *Parameter) getAnnotation(x []string, prefix byte) string {
	for i := range x {
		if x[i][0] == prefix {
			return x[i][1:]
		}
	}
	return ""
}

func (p *Parameter) getFlatNested(data internalNode) []internalNode {
	nodes := make([]internalNode, 0)
	for _, arg := range data.InternalArgs {
		if data.Node.Is(arg.Node.Prim) && len(arg.InternalArgs) > 0 && arg.Nested {
			nodes = append(nodes, p.getFlatNested(arg)...)
		} else {
			nodes = append(nodes, arg)
		}
	}
	return nodes
}

func (p *Parameter) parseNode(v map[string]interface{}, parent Node, path, entry string) internalNode {
	n := newNode(v)
	n.Path = path

	if n.Is(CONTRACT) {
		p.Tags.Append(ViewMethodTag)
	}

	fieldName := p.getAnnotation(n.Annotations, '%')
	typeName := p.getAnnotation(n.Annotations, ':')

	if _, ok := p.Metadata[path]; !ok {
		p.Metadata[path] = &Metadata{
			Prim:      n.Prim,
			TypeName:  typeName,
			FieldName: fieldName,
			Entry:     entry,
			Args:      make([]string, 0),
		}
	}

	p.fuzzyReader.WriteString(n.Prim)

	if n.Is(LAMBDA) || n.Is(CONTRACT) {
		if len(n.Args) > 0 {
			arg := n.Args[0].(map[string]interface{})
			m := p.Metadata[path]
			m.Parameter = arg
		}
		return internalNode{
			Node: &n,
		}
	} else if n.Is(OPTION) {
		arg := n.Args[0].(map[string]interface{})
		return p.parseNode(arg, parent, path+"0", fieldName)
	}

	args := make([]internalNode, 0)
	for i := range n.Args {
		argPath := fmt.Sprintf("%s%d", path, i)
		a := n.Args[i].(map[string]interface{})
		args = append(args, p.parseNode(a, n, argPath, entry))
	}

	if n.Is(PAIR) || n.Is(OR) {
		res := internalNode{
			Node:         &n,
			InternalArgs: args,
			Nested:       true,
		}
		isStruct := n.Is(PAIR) && (typeName != "" || fieldName != "")
		if isStruct || n.Prim != parent.Prim {
			args = p.getFlatNested(res)
		} else {
			return res
		}
	}

	m := p.Metadata[path]
	for _, a := range args {
		m.Args = append(m.Args, a.Node.Path)
	}

	return internalNode{
		Node:         &n,
		InternalArgs: args,
	}
}

// Entrypoints -
func (p *Parameter) Entrypoints() []string {
	root, ok := p.Metadata["0"]
	if !ok {
		return nil
	}
	if root.Prim != OR {
		s := root.FieldName
		if s == "" {
			s = fmt.Sprintf(defaultEntrypoint, 0)
		}
		return []string{s}
	}

	res := make([]string, len(root.Args))
	for i := range root.Args {
		m := p.Metadata[root.Args[i]]
		if m.FieldName != "" {
			res[i] = m.FieldName
		} else {
			res[i] = fmt.Sprintf(defaultEntrypoint, i)
		}
	}
	return res
}

func (p *Parameter) buildEntrypoint(path string, idx int) (e Entrypoint, err error) {
	m, ok := p.Metadata[path]
	if !ok {
		return e, fmt.Errorf("Unknown path: %s", path)
	}

	name := m.FieldName
	if name == "" {
		name = fmt.Sprintf(defaultEntrypoint, idx)
	}

	e.Name = name
	e.Prim = m.Prim
	e.Args = p.getEntrypointArgs(path)
	return
}

func (p *Parameter) getEntrypointArgs(path string) []interface{} {
	root := p.Metadata[path]
	if root.Prim == LAMBDA || root.Prim == CONTRACT {
		return []interface{}{root.Parameter}
	}

	args := make([]interface{}, 0)
	for i := 0; i < 2; i++ {
		subTree := fmt.Sprintf("%s%d", path, i)
		m, ok := p.Metadata[subTree]
		if !ok {
			continue
		}

		if m.Prim == PAIR || m.Prim == OR || m.Prim == OPTION {
			args = append(args, map[string]interface{}{
				"prim": m.Prim,
				"args": p.getEntrypointArgs(subTree),
			})
		} else if m.Prim == LAMBDA || m.Prim == CONTRACT {
			args = append(args, map[string]interface{}{
				"prim": m.Prim,
				"args": []interface{}{
					m.Parameter,
				},
			})
		} else {
			args = append(args, map[string]interface{}{
				"prim": m.Prim,
			})
		}
	}
	return args
}

// EntrypointStructs -
func (p *Parameter) EntrypointStructs() ([]Entrypoint, error) {
	root, ok := p.Metadata["0"]
	if !ok {
		return nil, fmt.Errorf("Unknown root metadata")
	}
	if root.Prim != OR {
		e, err := p.buildEntrypoint("0", 0)
		if err != nil {
			return nil, err
		}
		return []Entrypoint{e}, nil

	}

	res := make([]Entrypoint, len(root.Args))
	for i := range root.Args {
		e, err := p.buildEntrypoint(root.Args[i], i)
		if err != nil {
			return nil, err
		}
		res[i] = e
	}
	return res, nil
}
