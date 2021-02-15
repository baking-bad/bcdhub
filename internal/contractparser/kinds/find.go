package kinds

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

// Find - finds `interfaces` in metadata and return array of found tags
func Find(metadata meta.Metadata, interfaces map[string]ContractKind) ([]string, error) {
	if len(interfaces) == 0 {
		return nil, nil
	}

	res := make([]string, 0)

	for tag, i := range interfaces {
		if findInterface(metadata, i) {
			res = append(res, tag)
		}
	}

	return res, nil
}

func findInterface(metadata meta.Metadata, kind ContractKind) bool {
	root, ok := metadata["0"]
	if !ok {
		return false
	}

	if len(kind.Entrypoints) == 1 && kind.IsRoot {
		return compareEntrypoints(metadata, kind.Entrypoints[0], *root, "0")
	}

	for _, ie := range kind.Entrypoints {
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

func compareEntrypoints(metadata meta.Metadata, in Entrypoint, en meta.NodeMetadata, path string) bool {
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
		if in.Parameter == nil {
			return false
		}
		inObj := in.Parameter.(map[string]interface{})
		return typesCompare(inObj, en.Parameter)
	case consts.LAMBDA:
		if in.ReturnValue == nil || in.Parameter == nil {
			return false
		}
		inParamObj := in.Parameter.(map[string]interface{})
		inReturnObj := in.ReturnValue.(map[string]interface{})
		return typesCompare(inParamObj, en.Parameter) &&
			typesCompare(inReturnObj, en.ReturnValue)
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

func typesCompare(inObj map[string]interface{}, en string) bool {
	var enObj map[string]interface{}
	if err := json.Unmarshal([]byte(en), &enObj); err != nil {
		return false
	}

	return partCompare(inObj, enObj)
}

func partCompare(in, en map[string]interface{}) bool {
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
		if !partCompare(iargs[idx].(map[string]interface{}), eargs[idx].(map[string]interface{})) {
			return false
		}
	}

	return true
}
