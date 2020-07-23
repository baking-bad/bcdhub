package contractparser

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
)

func primTags(node node.Node) string {
	switch node.Prim {
	case consts.CREATECONTRACT:
		return consts.ContractFactoryTag
	case consts.SETDELEGATE:
		return consts.DelegatableTag
	case consts.CHECKSIGNATURE:
		return consts.CheckSigTag
	case consts.CHAINID:
		return consts.ChainAwareTag
	}
	return ""
}

func endpointsTags(metadata meta.Metadata, interfaces map[string][]kinds.Entrypoint) ([]string, error) {
	res := make([]string, 0)

	for tag, i := range interfaces {
		if findInterface(metadata, i) {
			res = append(res, tag)
		}
	}

	return res, nil
}

func findInterface(metadata meta.Metadata, i []kinds.Entrypoint) bool {
	root := metadata["0"]

	for _, ie := range i {
		found := false
		for _, e := range root.Args {
			entrypointMeta := metadata[e]
			if compareEntrypoints(metadata, ie, *entrypointMeta, e) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func compareEntrypoints(metadata meta.Metadata, in kinds.Entrypoint, en meta.NodeMetadata, path string) bool {
	if in.Name != "" && en.Name != in.Name {
		return false
	}
	// fmt.Printf("[in] %+v\n[en] %+v\n\n", in, en)
	if in.Prim != en.Prim {
		return false
	}

	switch en.Prim {
	case consts.MAP:
		if len(in.Args) != 2 {
			return false
		}

		for idx, suffix := range []string{"k", "v"} {
			enPath := fmt.Sprintf("%s/%s", path, suffix)

			enMeta, ok := metadata[enPath]
			if !ok {
				return false
			}
			if !compareEntrypoints(metadata, in.Args[idx], *enMeta, enPath) {
				return false
			}
		}
	case consts.LIST:
		if len(in.Args) != 1 {
			return false
		}
		enPath := fmt.Sprintf("%s/l", path)
		enMeta, ok := metadata[enPath]
		if !ok {
			return false
		}
		if !compareEntrypoints(metadata, in.Args[0], *enMeta, enPath) {
			return false
		}
	case consts.SET:
		if len(in.Args) != 1 {
			return false
		}
		enPath := fmt.Sprintf("%s/s", path)
		enMeta, ok := metadata[enPath]
		if !ok {
			return false
		}
		if !compareEntrypoints(metadata, in.Args[0], *enMeta, enPath) {
			return false
		}
	case consts.OPTION:
		if len(in.Args) != 1 {
			return false
		}
		enPath := fmt.Sprintf("%s/o", path)
		enMeta, ok := metadata[enPath]
		if !ok {
			return false
		}
		if !compareEntrypoints(metadata, in.Args[0], *enMeta, enPath) {
			return false
		}
	case consts.CONTRACT:
		inObj := in.Parameter.(map[string]interface{})
		var enObj map[string]interface{}
		err := json.Unmarshal([]byte(en.Parameter), &enObj)
		if err != nil {
			return false
		}

		return contractParameterCompare(inObj, enObj)
	default:
		for i, inArg := range in.Args {
			enPath := fmt.Sprintf("%s/%d", path, i)
			enMeta, ok := metadata[enPath]
			if !ok {
				return false
			}
			if !compareEntrypoints(metadata, inArg, *enMeta, enPath) {
				return false
			}
		}
	}

	return true
}

func contractParameterCompare(in, en map[string]interface{}) bool {
	if in["prim"] != en["prim"] {
		return false
	}

	inArgs, iok := in["args"]
	enArgs, eok := en["args"]

	if iok != eok {
		return false
	}

	if !iok {
		return true
	}

	iargs := inArgs.([]interface{})
	eargs := enArgs.([]interface{})

	if len(iargs) != len(eargs) {
		return false
	}

	for idx := range iargs {
		if !contractParameterCompare(iargs[idx].(map[string]interface{}), eargs[idx].(map[string]interface{})) {
			return false
		}
	}

	return true
}
