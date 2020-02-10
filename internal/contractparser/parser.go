package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

type onArray func(arr []gjson.Result) error
type onPrim func(n node.Node) error

type parser struct {
	arrayHandler onArray
	primHandler  onPrim
}

func (p *parser) parse(v gjson.Result) error {
	if v.IsArray() {
		arr := v.Array()
		for _, a := range arr {
			if err := p.parse(a); err != nil {
				return err
			}
		}
		if p.arrayHandler != nil {
			if err := p.arrayHandler(arr); err != nil {
				return err
			}
		}
	} else if v.IsObject() {
		node := node.NewNodeJSON(v)
		for _, a := range node.Args.Array() {
			if err := p.parse(a); err != nil {
				return err
			}
		}
		if p.primHandler != nil {
			if err := p.primHandler(node); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Unknown value type: %T", v.Type)
	}
	return nil
}
