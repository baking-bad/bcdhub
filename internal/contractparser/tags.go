package contractparser

import (
	"fmt"
	"io/ioutil"
	"strings"

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
	files, err := ioutil.ReadDir("./interfaces/")
	if err != nil {
		return err
	}

	for _, f := range files {
		path := fmt.Sprintf("./interfaces/%s", f.Name())
		var e []Entrypoint
		if err := jsonload.StructFromFile(path, &e); err != nil {
			return err
		}
		name := strings.Split(f.Name(), ".")[0]
		interfaces[name] = e
	}
	return nil
}

func endpointsTags(endpoints []Entrypoint) ([]string, error) {
	if len(interfaces) == 0 {
		if err := loadInterfaces(); err != nil {
			return nil, err
		}
	}
	res := make([]string, 0)
	for tag, i := range interfaces {
		if findInterface(endpoints, i) {
			res = append(res, tag)
		}
	}
	return res, nil
}

func findInterface(entrypoints []Entrypoint, i []Entrypoint) bool {
	for _, ie := range i {
		found := false
		for _, e := range entrypoints {
			if compareEntrypoints(ie, e) {
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

func deepEqual(a, b map[string]interface{}) bool {
	for ak, av := range a {
		bv, ok := b[ak]
		if !ok {
			return false
		}

		switch ak {
		case consts.KeyArgs:
			ava := av.([]interface{})
			bva := bv.([]interface{})

			if len(ava) != len(bva) {
				return false
			}

			for j := range ava {
				avam := ava[j].(map[string]interface{})
				bvam := bva[j].(map[string]interface{})
				if !deepEqual(avam, bvam) {
					return false
				}
			}
		case consts.KeyPrim:
			if av != bv {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func compareEntrypoints(a, b Entrypoint) bool {
	if a.Name != b.Name || a.Prim != b.Prim || len(a.Args) != len(b.Args) {
		return false
	}

	for i := range a.Args {
		ai := a.Args[i].(map[string]interface{})
		bi := b.Args[i].(map[string]interface{})
		if !deepEqual(ai, bi) {
			return false
		}
	}
	return true
}
