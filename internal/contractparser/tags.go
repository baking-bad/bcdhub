package contractparser

import (
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

	return true
}
