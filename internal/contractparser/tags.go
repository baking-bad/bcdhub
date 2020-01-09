package contractparser

import "strings"

func primTags(node Node) string {
	switch strings.ToUpper(node.Prim) {
	case "CREATE_CONTRACT":
		return ContractFactoryTag
	case "SET_DELEGATE":
		return DelegatableTag
	case "CHECK_SIGNATURE":
		return CheckSigTag
	case "CHAIN_ID", "chain_id":
		return ChainAwareTag
	}
	return ""
}

var handlers = map[string]func(entrypoint []Entrypoint) bool{
	FA12Tag: findFA12,
}

func endpointsTags(endpoints []Entrypoint) []string {
	res := make([]string, 0)
	for tag, handler := range handlers {
		if handler(endpoints) {
			res = append(res, tag)
		}
	}
	return res
}

func compareStringArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findInterface(entrypoints []Entrypoint, i []Entrypoint) bool {
	for _, ie := range i {
		found := false
		for _, e := range entrypoints {
			if e.Name == ie.Name && e.Type == ie.Type {
				if compareStringArrays(e.Args, ie.Args) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func findFA12(entrypoints []Entrypoint) bool {
	return findInterface(entrypoints, fa12)
}
