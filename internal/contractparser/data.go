package contractparser

import "strings"

import "github.com/tidwall/gjson"

// Fail -
type Fail struct {
	With string
}

func parseFail(args []gjson.Result) *Fail {
	if len(args) != 2 {
		return nil
	}

	if args[1].IsObject() {
		nodeFail := newNodeJSON(args[1])
		if !nodeFail.Is("FAILWITH") {
			return nil
		}
		s := ""
		if args[0].IsObject() {
			nodeWith := newNodeJSON(args[0])
			s = nodeWith.GetString()
			if s == "" && nodeWith.Is("PUSH") {
				arr := nodeWith.Args.Array()
				if len(arr) == 2 {
					nodeValue := newNodeJSON(arr[1])
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
