package contractparser

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
)

var interfaces = map[string][]Entrypoint{}

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

func loadInterfaces() error {
	files, err := ioutil.ReadDir("interfaces/")
	if err != nil {
		return err
	}

	for _, f := range files {
		path := fmt.Sprintf("interfaces/%s", f.Name())
		var e []Entrypoint
		if err := jsonload.StructFromFile(path, &e); err != nil {
			return err
		}
		name := strings.Split(f.Name(), ".")[0]
		interfaces[name] = e
	}
	return nil
}

func endpointsTags(metadata meta.Metadata) ([]string, error) {
	if len(interfaces) == 0 {
		if err := loadInterfaces(); err != nil {
			return nil, err
		}
	}
	res := make([]string, 0)
	for tag, i := range interfaces {
		if findInterface(metadata, i) {
			res = append(res, tag)
		}
	}
	return res, nil
}

func findInterface(metadata meta.Metadata, i []Entrypoint) bool {
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

func compareEntrypoints(metadata meta.Metadata, in Entrypoint, en meta.NodeMetadata, path string) bool {
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
