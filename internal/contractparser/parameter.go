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
	Name      string       `json:"name"`
	Prim      string       `json:"prim"`
	Args      []Entrypoint `json:"args,omitempty"`
	Parameter interface{}  `json:"parameter"`
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

	tags, err := endpointsTags(m)
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
