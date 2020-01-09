package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
)

const (
	defaultEntrypoint = "entrypoint_%d"
)

type internalNode struct {
	Node   Node
	Args   []internalNode
	Nested bool
}

// Entrypoint -
type Entrypoint struct {
	Name string   `json:"name"`
	Type string   `json:"type"`
	Args []string `json:"args"`
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
	Value    []map[string]interface{}
	Metadata map[string]*Metadata
	Hash     string

	fuzzyReader *HashReader
}

func newParameter(v []map[string]interface{}) Parameter {
	return Parameter{
		Value:       v,
		Metadata:    make(map[string]*Metadata),
		fuzzyReader: NewHashReader(),
	}
}

func (p *Parameter) parse() error {
	if len(p.Value) == 1 {
		parent := Node{
			Prim: "PARAMETER",
			Path: "0",
		}
		p.parseNode(p.Value[0], parent, "0", "")

		h, err := tlsh.HashReader(p.fuzzyReader)
		if err != nil {
			return err
		}
		p.Hash = h.String()
		return nil
	}
	return fmt.Errorf("Invalid parameter length: %d", len(p.Value))
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
	for _, arg := range data.Args {
		if data.Node.Is(arg.Node.Prim) && len(arg.Args) > 0 && arg.Nested {
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
			Node: n,
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
			Node:   n,
			Args:   args,
			Nested: true,
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
		Node: n,
		Args: args,
	}
}

// Entrypoints -
func (p *Parameter) Entrypoints() []string {
	root, ok := p.Metadata["0"]
	if !ok {
		return nil
	}
	if root.Prim != OR {
		if root.FieldName == "" {
			return []string{
				fmt.Sprintf(defaultEntrypoint, 0),
			}
		}
		return []string{
			root.FieldName,
		}
	}

	res := make([]string, len(root.Args))
	isEmpty := false
	for i := range root.Args {
		m := p.Metadata[root.Args[i]]
		if m.FieldName != "" {
			res[i] = m.FieldName
		} else {
			isEmpty = true
			break
		}
	}

	if isEmpty {
		for i := range root.Args {
			res[i] = fmt.Sprintf(defaultEntrypoint, i)
		}
	}
	return res
}
