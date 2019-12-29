package contractparser

import "reflect"

func primTags(obj map[string]interface{}) string {
	if prim, ok := obj["prim"]; ok {
		switch prim.(string) {
		case "contract":
			return ViewMethodTag
		case "CREATE_CONTRACT":
			return ContractFactoryTag
		case "SET_DELEGATE":
			return DelegatableTag
		case "CHECK_SIGNATURE":
			return CheckSigTag
		case "CHAIN_ID", "chain_id":
			return ChainAwareTag
		}
	}
	return ""
}

func endpointsTags(endpoints []Entrypoint) []string {
	res := make([]string, 0)
	if findFA12(endpoints) {
		res = append(res, FA12Tag)
	}
	return res
}

func findFA12(endpoints []Entrypoint) bool {
	for _, v := range fa12 {
		found := false
		for _, e := range endpoints {
			if reflect.DeepEqual(e, v) {
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
