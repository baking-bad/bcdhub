package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
	"github.com/tidwall/gjson"
)

const (
	defaultEntrypoint = "entrypoint_%d"
)

// Entrypoint -
type Entrypoint struct {
	Name string        `json:"name"`
	Prim string        `json:"prim"`
	Args []interface{} `json:"args,omitempty"`
}

// Parameter -
type Parameter struct {
	*parser

	Metadata Metadata

	Hash string
	hash []byte

	Tags Set
}

func newParameter(v gjson.Result) (Parameter, error) {
	if !v.IsArray() {
		return Parameter{}, fmt.Errorf("Parameter is not array")
	}
	p := Parameter{
		parser: &parser{},
		hash:   make([]byte, 0),
		Tags:   make(Set),
	}
	p.primHandler = p.handlePrimitive
	if err := p.parse(v); err != nil {
		return p, err
	}

	m, err := ParseMetadata(v)
	if err != nil {
		return p, err
	}
	p.Metadata = m

	if len(p.hash) == 0 {
		p.hash = append(p.hash, 0)
	}
	h, err := tlsh.HashBytes(p.hash)
	if err != nil {
		return p, err
	}
	p.Hash = h.String()

	e, err := p.EntrypointStructs()
	if err != nil {
		return p, err
	}

	tags, err := endpointsTags(e)
	if err != nil {
		return p, err
	}
	p.Tags.Append(tags...)

	return p, err
}

func (p *Parameter) handlePrimitive(n Node) error {
	p.hash = append(p.hash, []byte(n.Prim)...)

	if n.Is(CONTRACT) {
		p.Tags.Append(ViewMethodTag)
	}
	return nil
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
