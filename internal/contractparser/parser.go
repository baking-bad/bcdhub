package contractparser

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

type onArray func(arr gjson.Result) error
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
			if err := p.arrayHandler(v); err != nil {
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

// MakeStorageParser -
func MakeStorageParser(rpc noderpc.INode, es elastic.IElastic, protocol string, isSimulating bool) (storage.Parser, error) {
	if isSimulating {
		return storage.NewSimulate(rpc, es), nil
	}

	protoSymLink, err := meta.GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	switch protoSymLink {
	case consts.MetadataBabylon:
		return storage.NewBabylon(rpc, es), nil
	case consts.MetadataAlpha:
		return storage.NewAlpha(), nil
	default:
		return nil, fmt.Errorf("Unknown protocol %s", protocol)
	}
}
