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

var handlers = map[string]func(entrypoint Entrypoints) bool{
	FA12Tag: findFA12,
}

func endpointsTags(endpoints Entrypoints) []string {
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

func findInterface(entrypoints Entrypoints, i Entrypoints) bool {
	for k, v := range i {
		found := false
		for e, a := range entrypoints {
			if e == k {
				if compareStringArrays(a, v) {
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

func findFA12(entrypoints Entrypoints) bool {
	return findInterface(entrypoints, fa12)
}
