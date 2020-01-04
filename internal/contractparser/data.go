package contractparser

import "strings"

// Fail -
type Fail struct {
	With string
}

func parseFail(args []interface{}) *Fail {
	if len(args) != 2 {
		return nil
	}

	if m, ok := args[1].(map[string]interface{}); ok {
		nodeFail := newNode(m)
		if !nodeFail.Is("FAILWITH") {
			return nil
		}
		s := ""
		if w, ok := args[0].(map[string]interface{}); ok {
			nodeWith := newNode(w)
			s = nodeWith.GetString()
			if s == "" && nodeWith.Is("PUSH") {
				if len(nodeWith.Args) == 2 {
					nodeValue := newNode(nodeWith.Args[1].(map[string]interface{}))
					s = nodeValue.GetString()
				}
			} else {
				s = nodeWith.Prim
			}
			return &Fail{
				With: strings.ToLower(s),
			}
		}
	}
	return nil
}
