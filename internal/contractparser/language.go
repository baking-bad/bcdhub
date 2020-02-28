package contractparser

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/node"
)

func detectLiquidity(n node.Node, entries []string) bool {
	for _, a := range n.Annotations {
		if strings.Contains(a, "_slash_") {
			return true
		}
	}

	for _, e := range entries {
		if strings.Contains(e, "_Liq_entry") {
			return true
		}
	}
	return false
}

func detectPython(n node.Node) bool {
	str := n.GetString()
	return str != "" && (strings.Contains(str, "SmartPy") || strings.Contains(str, "self.") || strings.Contains(str, "sp.")) || strings.Contains(str, "WrongCondition")
}

func detectLIGO(n node.Node) bool {
	str := n.GetString()
	return str == "GET_FORCE" || str == "get_force"
}
