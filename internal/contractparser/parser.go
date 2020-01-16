package contractparser

import "fmt"

type onArray func(arr []interface{}) error
type onPrim func(n Node) error

type parser struct {
	arrayHandler onArray
	primHandler  onPrim
}

func newParser(onArr onArray, onPrimitive onPrim) parser {
	return parser{
		arrayHandler: onArr,
		primHandler:  onPrimitive,
	}
}

func (p *parser) parse(v interface{}) error {
	switch t := v.(type) {
	case []interface{}:
		for _, a := range t {
			if err := p.parse(a); err != nil {
				return err
			}
		}
		if p.arrayHandler != nil {
			if err := p.arrayHandler(t); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		node := newNode(t)
		for i := range node.Args {
			p.parse(node.Args[i])
		}
		if p.primHandler != nil {
			if err := p.primHandler(node); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unknown value type: %T", t)
	}
	return nil
}
