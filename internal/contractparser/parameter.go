package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
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

	Metadata meta.Metadata

	Tags helpers.Set
}

func newParameter(v gjson.Result) (Parameter, error) {
	if !v.IsArray() {
		return Parameter{}, fmt.Errorf("Parameter is not array")
	}
	p := Parameter{
		parser: &parser{},
		Tags:   make(helpers.Set),
	}
	p.primHandler = p.handlePrimitive
	if err := p.parse(v); err != nil {
		return p, err
	}

	m, err := meta.ParseMetadata(v)
	if err != nil {
		return p, err
	}
	p.Metadata = m

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

func (p *Parameter) handlePrimitive(n node.Node) error {
	if n.Is(consts.CONTRACT) {
		p.Tags.Append(consts.ViewMethodTag)
	}
	return nil
}

// Entrypoints -
func (p *Parameter) Entrypoints() []string {
	root, ok := p.Metadata["0"]
	if !ok {
		return nil
	}
	if root.Prim != consts.OR {
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
	if root.Prim == consts.LAMBDA || root.Prim == consts.CONTRACT {
		return []interface{}{root.Parameter}
	}

	args := make([]interface{}, 0)
	for i := 0; i < 2; i++ {
		subTree := fmt.Sprintf("%s%d", path, i)
		m, ok := p.Metadata[subTree]
		if !ok {
			continue
		}

		if m.Prim == consts.PAIR || m.Prim == consts.OR || m.Prim == consts.OPTION {
			args = append(args, map[string]interface{}{
				"prim": m.Prim,
				"args": p.getEntrypointArgs(subTree),
			})
		} else if m.Prim == consts.LAMBDA || m.Prim == consts.CONTRACT {
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
	if root.Prim != consts.OR {
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
